/*
 * Gerege SSO
 * Copyright (c) 2026 Gerege Systems Development Team.
 * Distributed under the Apache 2.0 License.
 */

package sessions

import (
	"context"
	"errors"

	"github.com/gerege-systems/open-gerege-nexus/backend/pkg/nexus"
	"github.com/jackc/pgx/v5"
)

// ErrNotFound is what the store answers for a session this organisation cannot
// see, which is the same answer as for one that never existed.
var ErrNotFound = errors.New("no such session")

// Session is one live sign-in.
//
// No token, no hash, not even a prefix. What this module exists to do is end a
// session, and ending one needs its id; a console that also displayed the
// credential would be a console that could be screenshotted into an account
// takeover.
type Session struct {
	ID         string `json:"id"`
	UserID     string `json:"user_id"`
	Email      string `json:"email"`
	Name       string `json:"name"`
	AuthMethod string `json:"auth_method"`
	UserAgent  string `json:"user_agent"`
	IPAddress  string `json:"ip_address"`
	CreatedAt  string `json:"created_at"`
	LastSeenAt string `json:"last_seen_at"`
	ExpiresAt  string `json:"expires_at"`
}

// Event is something this module did to a session.
type Event struct {
	ID        string `json:"id"`
	SessionID string `json:"session_id"`
	UserID    string `json:"user_id"`
	Action    string `json:"action"`
	ActorID   string `json:"actor_id"`
	Reason    string `json:"reason"`
	CreatedAt string `json:"created_at"`
}

// Store is every line of SQL this module runs.
//
// It is one file on purpose. Two of the tables it reads — `sessions` and
// `users` — belong to the core, and a module in another repository reading them
// is a coupling that survives only as long as their columns do. Keeping the
// coupling in one place means the day a core release moves a column, this is
// the file that changes; scattered through handlers it would be a search.
//
// ponytail: reads the core's schema directly. The upgrade path is a SessionStore
// on pkg/nexus — list and revoke, with the columns behind it — which is an
// upstream pull request rather than something reachable from here.
type Store struct{ db nexus.DB }

func NewStore(db nexus.DB) *Store { return &Store{db: db} }

// live is the definition of an active session, used by every query below so
// the list, the count and the revoke cannot disagree about what they are about.
// Written without a table alias so the same string can be pasted into the
// SELECT and into the two UPDATEs. It carried an `s.` prefix once and the
// UPDATEs, which have no alias to prefix, failed at runtime with a message
// about a missing FROM clause — caught by the test below rather than by
// anything the compiler can see, which is what a string of SQL costs.
const live = `revoked_at IS NULL AND expires_at > NOW()`

const sessionColumns = `sessions.id, sessions.user_id, u.email, u.name, sessions.auth_method,
	sessions.user_agent, sessions.ip_address,
	to_char(sessions.created_at, 'YYYY-MM-DD"T"HH24:MI:SSOF'),
	to_char(sessions.last_seen_at, 'YYYY-MM-DD"T"HH24:MI:SSOF'),
	to_char(sessions.expires_at, 'YYYY-MM-DD"T"HH24:MI:SSOF')`

// List answers the active sessions in an organisation, most recently seen
// first — the order in which somebody looking for a session that should not be
// there will find it.
func (s *Store) List(ctx context.Context, tenantID, userID string) ([]Session, error) {
	query := `SELECT ` + sessionColumns + `
		FROM sessions JOIN users u ON u.id = sessions.user_id
		WHERE sessions.tenant_id = $1 AND ` + live
	args := []any{tenantID}
	if userID != "" {
		query += ` AND sessions.user_id = $2`
		args = append(args, userID)
	}
	query += ` ORDER BY sessions.last_seen_at DESC`

	rows, err := s.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	found := []Session{}
	for rows.Next() {
		var item Session
		if err := rows.Scan(&item.ID, &item.UserID, &item.Email, &item.Name, &item.AuthMethod,
			&item.UserAgent, &item.IPAddress, &item.CreatedAt, &item.LastSeenAt, &item.ExpiresAt); err != nil {
			return nil, err
		}
		found = append(found, item)
	}
	return found, rows.Err()
}

// Revoke ends one session and records that it was ended.
//
// Both halves in one transaction. A revocation that happened without a record
// of who ordered it is the shape this module exists to prevent, and a record of
// a revocation that did not happen is worse than no record at all.
func (s *Store) Revoke(ctx context.Context, tenantID, sessionID, actorID, reason string) error {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var userID string
	err = tx.QueryRow(ctx, `
		UPDATE sessions SET revoked_at = NOW()
		WHERE tenant_id = $1 AND id = $2 AND `+live+`
		RETURNING user_id`, tenantID, sessionID).Scan(&userID)
	if errors.Is(err, pgx.ErrNoRows) {
		// Either it is not there or it was already over. Both mean the caller
		// asked to end something that is not running, and neither is an error
		// worth a different answer.
		//
		// Only ErrNoRows, though. This swallowed every error once, and a
		// malformed statement came back as a polite "no such session" — the
		// bug and its own cover story in one return.
		return ErrNotFound
	}
	if err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, `
		INSERT INTO sso_session_events (tenant_id, session_id, user_id, action, actor_id, reason)
		VALUES ($1, $2, $3, 'revoked', $4, $5)`,
		tenantID, sessionID, userID, nullUUID(actorID), reason); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

// RevokeAllFor ends every session one person holds in this organisation and
// returns how many that was.
func (s *Store) RevokeAllFor(ctx context.Context, tenantID, userID, actorID, reason string) (int, error) {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return 0, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	rows, err := tx.Query(ctx, `
		UPDATE sessions SET revoked_at = NOW()
		WHERE tenant_id = $1 AND user_id = $2 AND `+live+`
		RETURNING id`, tenantID, userID)
	if err != nil {
		return 0, err
	}
	ids := []string{}
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			rows.Close()
			return 0, err
		}
		ids = append(ids, id)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return 0, err
	}

	for _, id := range ids {
		if _, err := tx.Exec(ctx, `
			INSERT INTO sso_session_events (tenant_id, session_id, user_id, action, actor_id, reason)
			VALUES ($1, $2, $3, 'revoked', $4, $5)`,
			tenantID, id, userID, nullUUID(actorID), reason); err != nil {
			return 0, err
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return 0, err
	}
	return len(ids), nil
}

// Events is the revocation history, newest first.
func (s *Store) Events(ctx context.Context, tenantID string, limit int) ([]Event, error) {
	rows, err := s.db.Query(ctx, `
		SELECT id, COALESCE(session_id::text, ''), user_id, action, COALESCE(actor_id::text, ''), reason,
			to_char(created_at, 'YYYY-MM-DD"T"HH24:MI:SSOF')
		FROM sso_session_events
		WHERE tenant_id = $1
		ORDER BY created_at DESC
		LIMIT $2`, tenantID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	events := []Event{}
	for rows.Next() {
		var event Event
		if err := rows.Scan(&event.ID, &event.SessionID, &event.UserID, &event.Action,
			&event.ActorID, &event.Reason, &event.CreatedAt); err != nil {
			return nil, err
		}
		events = append(events, event)
	}
	return events, rows.Err()
}

func nullUUID(id string) any {
	if id == "" {
		return nil
	}
	return id
}
