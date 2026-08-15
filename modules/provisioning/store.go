/*
 * Gerege SSO
 * Copyright (c) 2026 Gerege Systems Development Team.
 * Distributed under the Apache 2.0 License.
 */

package provisioning

import (
	"context"
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/gerege-systems/open-gerege-nexus/backend/pkg/nexus"
	"github.com/jackc/pgx/v5"
)

// ErrNotFound covers a target this organisation cannot see.
var ErrNotFound = errors.New("no such target")

// Target is a system this installation pushes users into.
//
// No token field. It is written through Create and Update and read only by the
// worker; a struct with a place to put it is a struct somebody will marshal.
type Target struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	BaseURL   string `json:"base_url"`
	Enabled   bool   `json:"enabled"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
	HasToken  bool   `json:"has_token"`
	// Pending is what a console needs to say whether this target is keeping up.
	Pending int `json:"pending"`
}

// Run is one attempt against a target, as the log recorded it.
type Run struct {
	ID         string `json:"id"`
	TargetID   string `json:"target_id"`
	Op         string `json:"op"`
	UserID     string `json:"user_id"`
	StatusCode int    `json:"status_code"`
	Response   string `json:"response"`
	CreatedAt  string `json:"created_at"`
}

// job is a claimed queue row, with everything the worker needs to act on it.
// Unexported: nothing outside this package has any use for a target's token.
type job struct {
	id       string
	tenantID string
	targetID string
	op       string
	userID   string
	payload  []byte
	attempts int
	baseURL  string
	token    []byte
}

type Store struct{ db nexus.DB }

func NewStore(db nexus.DB) *Store { return &Store{db: db} }

const targetColumns = `t.id, t.name, t.base_url, t.enabled,
	to_char(t.created_at, 'YYYY-MM-DD"T"HH24:MI:SSOF'),
	to_char(t.updated_at, 'YYYY-MM-DD"T"HH24:MI:SSOF'),
	octet_length(t.token_encrypted) > 0,
	(SELECT COUNT(*) FROM sso_scim_queue q WHERE q.target_id = t.id)`

func (s *Store) List(ctx context.Context, tenantID string) ([]Target, error) {
	rows, err := s.db.Query(ctx, `SELECT `+targetColumns+`
		FROM sso_scim_targets t WHERE t.tenant_id = $1 ORDER BY t.name`, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	targets := []Target{}
	for rows.Next() {
		var t Target
		if err := rows.Scan(&t.ID, &t.Name, &t.BaseURL, &t.Enabled,
			&t.CreatedAt, &t.UpdatedAt, &t.HasToken, &t.Pending); err != nil {
			return nil, err
		}
		targets = append(targets, t)
	}
	return targets, rows.Err()
}

func (s *Store) Get(ctx context.Context, tenantID, id string) (Target, error) {
	var t Target
	err := s.db.QueryRow(ctx, `SELECT `+targetColumns+`
		FROM sso_scim_targets t WHERE t.tenant_id = $1 AND t.id = $2`, tenantID, id).
		Scan(&t.ID, &t.Name, &t.BaseURL, &t.Enabled, &t.CreatedAt, &t.UpdatedAt, &t.HasToken, &t.Pending)
	if errors.Is(err, pgx.ErrNoRows) {
		return Target{}, ErrNotFound
	}
	return t, err
}

func (s *Store) Create(ctx context.Context, tenantID, actorID string, in Input, sealed []byte) (Target, error) {
	var id string
	err := s.db.QueryRow(ctx, `
		INSERT INTO sso_scim_targets (tenant_id, name, base_url, token_encrypted, created_by)
		VALUES ($1, $2, $3, $4, $5) RETURNING id`,
		tenantID, in.Name, in.BaseURL, sealed, nullUUID(actorID)).Scan(&id)
	if err != nil {
		return Target{}, err
	}
	return s.Get(ctx, tenantID, id)
}

// Update leaves the stored token alone when sealed is nil, which is what lets a
// console edit a target it can never read back.
func (s *Store) Update(ctx context.Context, tenantID, id string, in Input, sealed []byte) (Target, error) {
	tag, err := s.db.Exec(ctx, `
		UPDATE sso_scim_targets SET name = $3, base_url = $4,
			token_encrypted = COALESCE($5, token_encrypted), updated_at = NOW()
		WHERE tenant_id = $1 AND id = $2`, tenantID, id, in.Name, in.BaseURL, sealed)
	if err != nil {
		return Target{}, err
	}
	if tag.RowsAffected() == 0 {
		return Target{}, ErrNotFound
	}
	return s.Get(ctx, tenantID, id)
}

func (s *Store) Delete(ctx context.Context, tenantID, id string) error {
	tag, err := s.db.Exec(ctx, `DELETE FROM sso_scim_targets WHERE tenant_id = $1 AND id = $2`,
		tenantID, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// Credentials reads what the worker needs to talk to a target.
func (s *Store) Credentials(ctx context.Context, tenantID, id string) (string, []byte, error) {
	var baseURL string
	var token []byte
	err := s.db.QueryRow(ctx, `
		SELECT base_url, token_encrypted FROM sso_scim_targets
		WHERE tenant_id = $1 AND id = $2`, tenantID, id).Scan(&baseURL, &token)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", nil, ErrNotFound
	}
	return baseURL, token, err
}

// EnqueueAll puts every member of an organisation on a target's queue.
//
// ponytail: this is the only thing that enqueues. The core has no hook a module
// can subscribe to for "a user was created, changed, or removed", so a change
// made between two resyncs does not reach the target until the next one. The
// upgrade path is an event on pkg/nexus — upstream — after which this endpoint
// becomes what it should be: the repair tool, not the mechanism.
func (s *Store) EnqueueAll(ctx context.Context, tenantID, targetID string) (int, error) {
	tag, err := s.db.Exec(ctx, `
		INSERT INTO sso_scim_queue (tenant_id, target_id, op, user_id, payload)
		SELECT $1, $2, 'update', u.id,
			jsonb_build_object('userName', u.email, 'displayName', u.name, 'active', true)
		FROM memberships m JOIN users u ON u.id = m.user_id
		WHERE m.tenant_id = $1`, tenantID, targetID)
	if err != nil {
		return 0, err
	}
	return int(tag.RowsAffected()), nil
}

func (s *Store) Runs(ctx context.Context, tenantID string, limit int) ([]Run, error) {
	rows, err := s.db.Query(ctx, `
		SELECT id, target_id, op, COALESCE(user_id::text, ''), status_code, response_excerpt,
			to_char(created_at, 'YYYY-MM-DD"T"HH24:MI:SSOF')
		FROM sso_scim_log WHERE tenant_id = $1 ORDER BY created_at DESC LIMIT $2`, tenantID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	runs := []Run{}
	for rows.Next() {
		var run Run
		if err := rows.Scan(&run.ID, &run.TargetID, &run.Op, &run.UserID,
			&run.StatusCode, &run.Response, &run.CreatedAt); err != nil {
			return nil, err
		}
		runs = append(runs, run)
	}
	return runs, rows.Err()
}

// claim takes up to `limit` due jobs and hides them from any other worker.
//
// SKIP LOCKED rather than a lease column: two replicas of this product run the
// same worker, and a claim that is a row update is a claim that can be lost to
// a crash between the update and the send. A transaction that holds the rows
// for the duration of the batch is released by the database when the process
// dies, which is the only thing that is true whatever the process did.
//
// Runs with no tenant in the context — the platform path — so it sees the queue
// across every organisation, the way the core's own sweeps do.
func (s *Store) claim(ctx context.Context, limit int) (pgx.Tx, []job, error) {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return nil, nil, err
	}
	rows, err := tx.Query(ctx, `
		SELECT q.id, q.tenant_id, q.target_id, q.op, q.user_id, q.payload, q.attempts,
			t.base_url, t.token_encrypted
		FROM sso_scim_queue q
		JOIN sso_scim_targets t ON t.id = q.target_id AND t.enabled
		WHERE q.next_attempt_at <= NOW()
		ORDER BY q.created_at
		LIMIT $1
		FOR UPDATE OF q SKIP LOCKED`, limit)
	if err != nil {
		_ = tx.Rollback(ctx)
		return nil, nil, err
	}

	jobs := []job{}
	for rows.Next() {
		var j job
		if err := rows.Scan(&j.id, &j.tenantID, &j.targetID, &j.op, &j.userID,
			&j.payload, &j.attempts, &j.baseURL, &j.token); err != nil {
			rows.Close()
			_ = tx.Rollback(ctx)
			return nil, nil, err
		}
		jobs = append(jobs, j)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		_ = tx.Rollback(ctx)
		return nil, nil, err
	}
	return tx, jobs, nil
}

// finish removes a job that succeeded and records what happened.
func (s *Store) finish(ctx context.Context, tx pgx.Tx, j job, status int, body string) error {
	if _, err := tx.Exec(ctx, `DELETE FROM sso_scim_queue WHERE id = $1`, j.id); err != nil {
		return err
	}
	return logRun(ctx, tx, j, status, body)
}

// retry leaves a failed job on the queue, later. The delay doubles with each
// attempt and stops at six: a target that has refused six times is a target
// somebody has to look at, and a queue that retries for ever is a queue that
// hides that from them.
func (s *Store) retry(ctx context.Context, tx pgx.Tx, j job, status int, body string) error {
	if j.attempts+1 >= maxAttempts {
		if _, err := tx.Exec(ctx, `DELETE FROM sso_scim_queue WHERE id = $1`, j.id); err != nil {
			return err
		}
		return logRun(ctx, tx, j, status, "gave up after "+strconv.Itoa(maxAttempts)+" attempts: "+body)
	}
	delay := time.Duration(1<<uint(j.attempts)) * time.Minute
	if _, err := tx.Exec(ctx, `
		UPDATE sso_scim_queue SET attempts = attempts + 1, last_error = $2,
			next_attempt_at = NOW() + $3::interval
		WHERE id = $1`, j.id, truncate(body), delay.String()); err != nil {
		return err
	}
	return logRun(ctx, tx, j, status, body)
}

func logRun(ctx context.Context, tx pgx.Tx, j job, status int, body string) error {
	_, err := tx.Exec(ctx, `
		INSERT INTO sso_scim_log (tenant_id, target_id, op, user_id, status_code, response_excerpt)
		VALUES ($1, $2, $3, $4, $5, $6)`,
		j.tenantID, j.targetID, j.op, j.userID, status, truncate(body))
	return err
}

// truncate keeps a remote system's page of HTML from becoming a page in this
// database. Cut on a byte and repaired, because the cut lands mid-rune the
// first time somebody's error message is not in English.
func truncate(body string) string {
	const limit = 500
	if len(body) <= limit {
		return body
	}
	return strings.ToValidUTF8(body[:limit], "") + "…"
}

func nullUUID(id string) any {
	if id == "" {
		return nil
	}
	return id
}
