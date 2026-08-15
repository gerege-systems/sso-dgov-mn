package accessreview

import (
	"context"
	"errors"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
)

// The tests that need rows.
//
// Open() is the reason this file exists. It is a six-table join into the core's
// RBAC with two scope filters written as `$3 <> 'app' OR …`, and nothing about
// it can be checked by reading: a filter that silently matches everything
// produces a campaign that looks busy and asks the wrong question.
func testStore(t *testing.T) (*Store, *pgxpool.Pool) {
	t.Helper()
	dsn := os.Getenv("SSO_TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("SSO_TEST_DATABASE_URL is not set")
	}
	pool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	t.Cleanup(pool.Close)
	return NewStore(pool), pool
}

// org builds an organisation with one member holding one role, and that role
// holding the permissions named.
func org(t *testing.T, pool *pgxpool.Pool, roleCode string, codes ...string) (string, string) {
	t.Helper()
	ctx := context.Background()
	var tenantID, userID, membershipID, roleID string

	if err := pool.QueryRow(ctx, `
		INSERT INTO tenants (slug, name) VALUES ('t-' || gen_random_uuid(), 'Test') RETURNING id`).
		Scan(&tenantID); err != nil {
		t.Fatalf("tenant: %v", err)
	}
	t.Cleanup(func() { _, _ = pool.Exec(context.Background(), `DELETE FROM tenants WHERE id = $1`, tenantID) })

	if err := pool.QueryRow(ctx, `
		INSERT INTO users (email, password_hash, name)
		VALUES (gen_random_uuid() || '@example.mn', 'x', 'Test Person') RETURNING id`).
		Scan(&userID); err != nil {
		t.Fatalf("user: %v", err)
	}
	t.Cleanup(func() { _, _ = pool.Exec(context.Background(), `DELETE FROM users WHERE id = $1`, userID) })

	if err := pool.QueryRow(ctx,
		`INSERT INTO memberships (tenant_id, user_id) VALUES ($1, $2) RETURNING id`,
		tenantID, userID).Scan(&membershipID); err != nil {
		t.Fatalf("membership: %v", err)
	}
	if err := pool.QueryRow(ctx,
		`INSERT INTO roles (tenant_id, code, name) VALUES ($1, $2, $2) RETURNING id`,
		tenantID, roleCode).Scan(&roleID); err != nil {
		t.Fatalf("role: %v", err)
	}
	if _, err := pool.Exec(ctx,
		`INSERT INTO membership_roles (membership_id, role_id) VALUES ($1, $2)`,
		membershipID, roleID); err != nil {
		t.Fatalf("membership role: %v", err)
	}

	for _, code := range codes {
		var permissionID string
		// The permissions table is global and seeded by the core, so a code
		// this test wants may already be there.
		if err := pool.QueryRow(ctx, `
			INSERT INTO permissions (code, name) VALUES ($1, $1)
			ON CONFLICT (code) DO UPDATE SET code = EXCLUDED.code
			RETURNING id`, code).Scan(&permissionID); err != nil {
			t.Fatalf("permission %s: %v", code, err)
		}
		if _, err := pool.Exec(ctx,
			`INSERT INTO role_permissions (role_id, permission_id) VALUES ($1, $2)`,
			roleID, permissionID); err != nil {
			t.Fatalf("role permission: %v", err)
		}
	}
	return tenantID, userID
}

func openCampaign(t *testing.T, store *Store, tenantID string, in Input) Campaign {
	t.Helper()
	ctx := context.Background()
	if err := in.validate(); err != nil {
		t.Fatalf("validate: %v", err)
	}
	campaign, err := store.CreateCampaign(ctx, tenantID, "", in)
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if campaign.Status != "draft" || campaign.Total != 0 {
		t.Fatalf("a new campaign should be an empty draft: %+v", campaign)
	}
	opened, err := store.Open(ctx, tenantID, campaign.ID)
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	return opened
}

// An "all" campaign asks about every permission a member holds by any route.
//
// That includes the ones nobody assigned. The core gives every new membership a
// default `user` role — a trigger, in migration 00008 — so the member here holds
// its five permissions as well as the three the fixture granted through `clerk`.
// Counting them would be asserting the core's seed data; what this checks is
// that the three granted here arrived, that the union is what a union should be,
// and that nothing belonging to anybody else did.
func TestOpeningACampaignCopiesEveryAccessInScope(t *testing.T) {
	store, pool := testStore(t)
	tenantID, userID := org(t, pool, "clerk", "documents.read", "documents.manage", "sso_sessions.read")

	campaign := openCampaign(t, store, tenantID, Input{Name: "Everything", Scope: "all"})
	if campaign.Pending != campaign.Total {
		t.Fatalf("a freshly opened campaign has %d of %d pending", campaign.Pending, campaign.Total)
	}

	items, err := store.Items(context.Background(), tenantID, campaign.ID, "")
	if err != nil {
		t.Fatalf("items: %v", err)
	}
	held := map[string]bool{}
	for _, item := range items {
		held[item.PermissionCode] = true
		if item.UserID != userID {
			t.Errorf("an item names somebody who is not a member: %s", item.UserID)
		}
		if item.RoleName == "" || item.UserEmail == "" {
			t.Errorf("an item does not carry who and how: %+v", item)
		}
	}
	for _, code := range []string{"documents.read", "documents.manage", "sso_sessions.read"} {
		if !held[code] {
			t.Errorf("%s was granted and did not reach the campaign", code)
		}
	}
}

// An app scope must copy that app's permissions and nothing else. This is the
// filter that would be invisible if it were wrong: a campaign named "Documents"
// listing every permission in the organisation still looks like a campaign.
func TestAnAppScopeCopiesOnlyThatApp(t *testing.T) {
	store, pool := testStore(t)
	tenantID, _ := org(t, pool, "clerk", "documents.read", "documents.manage", "sso_sessions.read")

	campaign := openCampaign(t, store, tenantID, Input{Name: "Documents", Scope: "app", ScopeRef: "documents"})
	items, _ := store.Items(context.Background(), tenantID, campaign.ID, "")
	if len(items) != 2 {
		t.Fatalf("expected the two documents permissions, got %d", len(items))
	}
	for _, item := range items {
		if item.PermissionCode != "documents.read" && item.PermissionCode != "documents.manage" {
			t.Errorf("%s is not a documents permission", item.PermissionCode)
		}
	}
}

func TestARoleScopeCopiesOnlyThatRolesHolders(t *testing.T) {
	store, pool := testStore(t)
	tenantID, _ := org(t, pool, "clerk", "documents.read")

	matching := openCampaign(t, store, tenantID, Input{Name: "Clerks", Scope: "role", ScopeRef: "clerk"})
	if matching.Total != 1 {
		t.Fatalf("expected the clerk's one permission, got %d", matching.Total)
	}
	other := openCampaign(t, store, tenantID, Input{Name: "Auditors", Scope: "role", ScopeRef: "auditor"})
	if other.Total != 0 {
		t.Fatalf("a role nobody holds produced %d items", other.Total)
	}
}

// A campaign that is open takes decisions; one that is closed does not, and
// every answer ever given is kept.
func TestDecidingIsRecordedAndCanBeChangedUntilTheCampaignCloses(t *testing.T) {
	store, pool := testStore(t)
	ctx := context.Background()
	tenantID, reviewer := org(t, pool, "clerk", "documents.manage")
	// Scoped to the role the fixture made, so the item under test is the one it
	// granted rather than one of the default role's.
	campaign := openCampaign(t, store, tenantID, Input{Name: "Q3", Scope: "role", ScopeRef: "clerk"})

	items, _ := store.Items(ctx, tenantID, campaign.ID, "pending")
	if len(items) != 1 {
		t.Fatalf("expected one item, got %d", len(items))
	}
	item := items[0]

	if _, err := store.Decide(ctx, tenantID, item.ID, "kept", reviewer, ""); err != nil {
		t.Fatalf("keep: %v", err)
	}
	decided, err := store.Decide(ctx, tenantID, item.ID, "revoked", reviewer, "left the team")
	if err != nil {
		t.Fatalf("revoke: %v", err)
	}
	if decided.Status != "revoked" || decided.ReviewerID != reviewer {
		t.Fatalf("the item does not carry the latest decision: %+v", decided)
	}

	var decisions int
	if err := pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM sso_review_decisions WHERE item_id = $1`, item.ID).Scan(&decisions); err != nil {
		t.Fatal(err)
	}
	if decisions != 2 {
		t.Fatalf("expected both decisions to be kept, found %d", decisions)
	}

	if _, err := store.Close(ctx, tenantID, campaign.ID); err != nil {
		t.Fatalf("close: %v", err)
	}
	if _, err := store.Decide(ctx, tenantID, item.ID, "kept", reviewer, ""); !errors.Is(err, ErrNotOpen) {
		t.Fatalf("a closed campaign took a decision: %v", err)
	}
}

// draft → open → closed, and no way back.
func TestACampaignOnlyMovesForward(t *testing.T) {
	store, pool := testStore(t)
	ctx := context.Background()
	tenantID, _ := org(t, pool, "clerk", "documents.read")
	campaign := openCampaign(t, store, tenantID, Input{Name: "Q3", Scope: "role", ScopeRef: "clerk"})

	if _, err := store.Open(ctx, tenantID, campaign.ID); !errors.Is(err, ErrNotDraft) {
		t.Fatalf("an open campaign was opened again: %v", err)
	}
	if _, err := store.Close(ctx, tenantID, campaign.ID); err != nil {
		t.Fatalf("close: %v", err)
	}
	if _, err := store.Close(ctx, tenantID, campaign.ID); !errors.Is(err, ErrNotOpen) {
		t.Fatalf("a closed campaign was closed again: %v", err)
	}
	if _, err := store.Open(ctx, tenantID, campaign.ID); !errors.Is(err, ErrNotDraft) {
		t.Fatalf("a closed campaign was reopened: %v", err)
	}
}

// Opening a campaign must not copy another organisation's access.
func TestACampaignSeesOnlyItsOwnOrganisation(t *testing.T) {
	store, pool := testStore(t)
	mine, myUser := org(t, pool, "clerk", "documents.read")
	org(t, pool, "clerk", "documents.read", "documents.manage")

	campaign := openCampaign(t, store, mine, Input{Name: "Mine", Scope: "all"})
	items, err := store.Items(context.Background(), mine, campaign.ID, "")
	if err != nil {
		t.Fatalf("items: %v", err)
	}
	if len(items) == 0 {
		t.Fatal("the campaign copied nothing at all")
	}
	for _, item := range items {
		if item.UserID != myUser {
			t.Fatalf("the campaign reached another organisation's member: %s", item.UserID)
		}
	}
}
