/*
 * Gerege SSO
 * Copyright (c) 2026 Gerege Systems Development Team.
 * Distributed under the Apache 2.0 License.
 */

package accessreview

import (
	"context"
	"errors"

	"github.com/gerege-systems/open-gerege-nexus/backend/pkg/nexus"
	"github.com/jackc/pgx/v5"
)

var (
	// ErrNotFound covers a campaign or item this organisation cannot see.
	ErrNotFound = errors.New("no such campaign")
	// ErrNotDraft is the refusal to open something that is already open, or
	// closed. Both are states a campaign only leaves in one direction.
	ErrNotDraft = errors.New("only a draft campaign can be opened")
	// ErrNotOpen is the refusal to decide on an item in a campaign that is not
	// taking decisions.
	ErrNotOpen = errors.New("that campaign is not open")
)

// Campaign is one round of attestation.
type Campaign struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Scope     string `json:"scope"`
	ScopeRef  string `json:"scope_ref"`
	DueDate   string `json:"due_date"`
	Status    string `json:"status"`
	CreatedAt string `json:"created_at"`
	OpenedAt  string `json:"opened_at"`
	ClosedAt  string `json:"closed_at"`
	// The counts a console needs to draw a progress bar, and the only numbers
	// anybody asks a campaign for.
	Total   int `json:"total"`
	Pending int `json:"pending"`
	Revoked int `json:"revoked"`
}

// Item is one person's one permission, frozen at the moment the campaign opened.
type Item struct {
	ID             string `json:"id"`
	UserID         string `json:"user_id"`
	UserEmail      string `json:"user_email"`
	RoleID         string `json:"role_id"`
	RoleName       string `json:"role_name"`
	PermissionCode string `json:"permission_code"`
	Status         string `json:"status"`
	ReviewerID     string `json:"reviewer_id"`
	DecidedAt      string `json:"decided_at"`
	// Filled by Queue, which spans campaigns and so has to say which one each
	// row came from. Empty on a listing that was already given a campaign.
	CampaignID   string `json:"campaign_id,omitempty"`
	CampaignName string `json:"campaign_name,omitempty"`
	DueDate      string `json:"due_date,omitempty"`
}

// Store is every line of SQL this module runs.
//
// ponytail: Snapshot reads the core's RBAC tables directly — memberships,
// membership_roles, roles, role_permissions, permissions. The upgrade path is a
// RoleStore on pkg/nexus that answers "who holds what, right now", which is an
// upstream pull request. Until then the coupling lives in one query, in this
// file, so a core release that moves a column changes one place.
type Store struct{ db nexus.DB }

func NewStore(db nexus.DB) *Store { return &Store{db: db} }

const campaignColumns = `c.id, c.name, c.scope, c.scope_ref,
	COALESCE(to_char(c.due_date, 'YYYY-MM-DD'), ''),
	c.status,
	to_char(c.created_at, 'YYYY-MM-DD"T"HH24:MI:SSOF'),
	COALESCE(to_char(c.opened_at, 'YYYY-MM-DD"T"HH24:MI:SSOF'), ''),
	COALESCE(to_char(c.closed_at, 'YYYY-MM-DD"T"HH24:MI:SSOF'), ''),
	(SELECT COUNT(*) FROM sso_review_items i WHERE i.campaign_id = c.id),
	(SELECT COUNT(*) FROM sso_review_items i WHERE i.campaign_id = c.id AND i.status = 'pending'),
	(SELECT COUNT(*) FROM sso_review_items i WHERE i.campaign_id = c.id AND i.status = 'revoked')`

func (s *Store) ListCampaigns(ctx context.Context, tenantID string) ([]Campaign, error) {
	rows, err := s.db.Query(ctx, `SELECT `+campaignColumns+`
		FROM sso_review_campaigns c WHERE c.tenant_id = $1 ORDER BY c.created_at DESC`, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	campaigns := []Campaign{}
	for rows.Next() {
		campaign, err := scanCampaign(rows)
		if err != nil {
			return nil, err
		}
		campaigns = append(campaigns, campaign)
	}
	return campaigns, rows.Err()
}

func (s *Store) GetCampaign(ctx context.Context, tenantID, id string) (Campaign, error) {
	row := s.db.QueryRow(ctx, `SELECT `+campaignColumns+`
		FROM sso_review_campaigns c WHERE c.tenant_id = $1 AND c.id = $2`, tenantID, id)
	campaign, err := scanCampaign(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return Campaign{}, ErrNotFound
	}
	return campaign, err
}

type scanRow interface{ Scan(dest ...any) error }

func scanCampaign(row scanRow) (Campaign, error) {
	var c Campaign
	err := row.Scan(&c.ID, &c.Name, &c.Scope, &c.ScopeRef, &c.DueDate, &c.Status,
		&c.CreatedAt, &c.OpenedAt, &c.ClosedAt, &c.Total, &c.Pending, &c.Revoked)
	return c, err
}

// CreateCampaign records a draft. Nothing is snapshotted yet: a draft is a plan,
// and the set of people it will ask about is whatever it is on the day it opens.
func (s *Store) CreateCampaign(ctx context.Context, tenantID, actorID string, in Input) (Campaign, error) {
	var id string
	err := s.db.QueryRow(ctx, `
		INSERT INTO sso_review_campaigns (tenant_id, name, scope, scope_ref, due_date, created_by)
		VALUES ($1, $2, $3, $4, NULLIF($5, '')::date, $6)
		RETURNING id`,
		tenantID, in.Name, in.Scope, in.ScopeRef, in.DueDate, nullUUID(actorID)).Scan(&id)
	if err != nil {
		return Campaign{}, err
	}
	return s.GetCampaign(ctx, tenantID, id)
}

// Open freezes the current state of RBAC into the campaign's items.
//
// The copy is the whole design. A campaign that queried RBAC live would change
// under the reviewer: a role edited on Tuesday would silently rewrite what
// somebody attested to on Monday, and the report at the end would describe a
// set of decisions nobody made.
//
// One person can reach one permission through two roles. The unique constraint
// takes the first and drops the second, so an item names *a* route to the
// permission rather than every route — which is the question a reviewer is
// being asked ("should they have this?") rather than the one they are not
// ("through which role?").
func (s *Store) Open(ctx context.Context, tenantID, id string) (Campaign, error) {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return Campaign{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var scope, scopeRef string
	err = tx.QueryRow(ctx, `
		UPDATE sso_review_campaigns SET status = 'open', opened_at = NOW()
		WHERE tenant_id = $1 AND id = $2 AND status = 'draft'
		RETURNING scope, scope_ref`, tenantID, id).Scan(&scope, &scopeRef)
	if errors.Is(err, pgx.ErrNoRows) {
		// Either it is not there or it is not a draft. Which of the two is a
		// question the caller can answer by reading it back.
		return Campaign{}, ErrNotDraft
	}
	if err != nil {
		return Campaign{}, err
	}

	if _, err := tx.Exec(ctx, `
		INSERT INTO sso_review_items
			(tenant_id, campaign_id, user_id, user_email, role_id, role_name, permission_code)
		SELECT $1, $2, u.id, u.email, r.id, r.name, p.code
		FROM memberships m
		JOIN users u ON u.id = m.user_id
		JOIN membership_roles mr ON mr.membership_id = m.id
		JOIN roles r ON r.id = mr.role_id
		JOIN role_permissions rp ON rp.role_id = r.id
		JOIN permissions p ON p.id = rp.permission_id
		WHERE m.tenant_id = $1
		  AND ($3 <> 'app'  OR p.code LIKE $4 || '.%')
		  AND ($3 <> 'role' OR r.code = $4)
		ON CONFLICT (campaign_id, user_id, permission_code) DO NOTHING`,
		tenantID, id, scope, scopeRef); err != nil {
		return Campaign{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return Campaign{}, err
	}
	return s.GetCampaign(ctx, tenantID, id)
}

// Close ends a campaign. There is no way back: reopening would let a decision
// change after it was reported as final.
func (s *Store) Close(ctx context.Context, tenantID, id string) (Campaign, error) {
	tag, err := s.db.Exec(ctx, `
		UPDATE sso_review_campaigns SET status = 'closed', closed_at = NOW()
		WHERE tenant_id = $1 AND id = $2 AND status = 'open'`, tenantID, id)
	if err != nil {
		return Campaign{}, err
	}
	if tag.RowsAffected() == 0 {
		return Campaign{}, ErrNotOpen
	}
	return s.GetCampaign(ctx, tenantID, id)
}

// Items lists a campaign's rows, optionally filtered by status. Pending first,
// because they are the work.
func (s *Store) Items(ctx context.Context, tenantID, campaignID, status string) ([]Item, error) {
	rows, err := s.db.Query(ctx, `
		SELECT i.id, i.user_id, i.user_email, COALESCE(i.role_id::text, ''), i.role_name,
			i.permission_code, i.status, COALESCE(i.reviewer_id::text, ''),
			COALESCE(to_char(i.decided_at, 'YYYY-MM-DD"T"HH24:MI:SSOF'), '')
		FROM sso_review_items i
		WHERE i.tenant_id = $1 AND i.campaign_id = $2 AND ($3 = '' OR i.status = $3)
		ORDER BY (i.status = 'pending') DESC, i.user_email, i.permission_code`,
		tenantID, campaignID, status)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := []Item{}
	for rows.Next() {
		var item Item
		if err := rows.Scan(&item.ID, &item.UserID, &item.UserEmail, &item.RoleID, &item.RoleName,
			&item.PermissionCode, &item.Status, &item.ReviewerID, &item.DecidedAt); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

// Queue is everything still waiting on a reviewer, across every open campaign.
//
// The per-campaign listing answers "what is in this campaign". This answers the
// question a reviewer actually has — "what is waiting for me" — which is not the
// same query with the filter dropped: it spans campaigns, it names the campaign
// each row came from, and it excludes the closed ones, where a pending item is
// a decision nobody will ever be asked for.
func (s *Store) Queue(ctx context.Context, tenantID string, limit int) ([]Item, error) {
	rows, err := s.db.Query(ctx, `
		SELECT i.id, i.user_id, i.user_email, COALESCE(i.role_id::text, ''), i.role_name,
			i.permission_code, i.status, COALESCE(i.reviewer_id::text, ''),
			COALESCE(to_char(i.decided_at, 'YYYY-MM-DD"T"HH24:MI:SSOF'), ''),
			c.id, c.name, COALESCE(to_char(c.due_date, 'YYYY-MM-DD'), '')
		FROM sso_review_items i
		JOIN sso_review_campaigns c ON c.id = i.campaign_id
		WHERE i.tenant_id = $1 AND i.status = 'pending' AND c.status = 'open'
		ORDER BY c.due_date NULLS LAST, i.user_email, i.permission_code
		LIMIT $2`, tenantID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := []Item{}
	for rows.Next() {
		var item Item
		if err := rows.Scan(&item.ID, &item.UserID, &item.UserEmail, &item.RoleID, &item.RoleName,
			&item.PermissionCode, &item.Status, &item.ReviewerID, &item.DecidedAt,
			&item.CampaignID, &item.CampaignName, &item.DueDate); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

// Decide records an answer, and keeps the previous ones.
//
// The item carries the current answer; sso_review_decisions carries every
// answer. A reviewer who keeps an access on Monday and revokes it on Thursday
// has made two decisions, and a trail that remembers only the second cannot say
// why the first was made.
func (s *Store) Decide(ctx context.Context, tenantID, itemID, decision, reviewerID, note string) (Item, error) {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return Item{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var campaignID string
	err = tx.QueryRow(ctx, `
		UPDATE sso_review_items i SET status = $3, reviewer_id = $4, decided_at = NOW()
		FROM sso_review_campaigns c
		WHERE i.tenant_id = $1 AND i.id = $2 AND c.id = i.campaign_id AND c.status = 'open'
		RETURNING i.campaign_id`, tenantID, itemID, decision, nullUUID(reviewerID)).Scan(&campaignID)
	if errors.Is(err, pgx.ErrNoRows) {
		return Item{}, ErrNotOpen
	}
	if err != nil {
		return Item{}, err
	}

	if _, err := tx.Exec(ctx, `
		INSERT INTO sso_review_decisions (tenant_id, item_id, decision, reviewer_id, note)
		VALUES ($1, $2, $3, $4, $5)`,
		tenantID, itemID, decision, nullUUID(reviewerID), note); err != nil {
		return Item{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return Item{}, err
	}

	items, err := s.Items(ctx, tenantID, campaignID, "")
	if err != nil {
		return Item{}, err
	}
	for _, item := range items {
		if item.ID == itemID {
			return item, nil
		}
	}
	return Item{}, ErrNotFound
}

func nullUUID(id string) any {
	if id == "" {
		return nil
	}
	return id
}
