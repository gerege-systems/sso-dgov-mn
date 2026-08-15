package sessions

import (
	"context"
	"errors"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
)

// The tests that need rows.
//
// Skipped without SSO_TEST_DATABASE_URL, which is the shape every DB-backed
// test in this ecosystem has — and the reason CI runs them in a job of their
// own with `-count=1`: a skipped assertion looks exactly like a passing one.
//
// What they are about is the half of this module that cannot be checked any
// other way: `live`, which decides what a session is, and the transaction that
// must not end one without recording it.
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

// tenant creates an organisation with one member, and takes them away again.
func tenant(t *testing.T, pool *pgxpool.Pool) (string, string) {
	t.Helper()
	ctx := context.Background()
	var tenantID, userID string
	if err := pool.QueryRow(ctx, `
		INSERT INTO tenants (slug, name) VALUES ('t-' || gen_random_uuid(), 'Test') RETURNING id`).
		Scan(&tenantID); err != nil {
		t.Fatalf("tenant: %v", err)
	}
	if err := pool.QueryRow(ctx, `
		INSERT INTO users (email, password_hash, name)
		VALUES (gen_random_uuid() || '@example.mn', 'x', 'Test Person') RETURNING id`).
		Scan(&userID); err != nil {
		t.Fatalf("user: %v", err)
	}
	if _, err := pool.Exec(ctx,
		`INSERT INTO memberships (tenant_id, user_id) VALUES ($1, $2)`, tenantID, userID); err != nil {
		t.Fatalf("membership: %v", err)
	}
	t.Cleanup(func() {
		_, _ = pool.Exec(context.Background(), `DELETE FROM tenants WHERE id = $1`, tenantID)
		_, _ = pool.Exec(context.Background(), `DELETE FROM users WHERE id = $1`, userID)
	})
	return tenantID, userID
}

func session(t *testing.T, pool *pgxpool.Pool, tenantID, userID, expires string) string {
	t.Helper()
	var id string
	if err := pool.QueryRow(context.Background(), `
		INSERT INTO sessions (token_hash, user_id, tenant_id, expires_at)
		VALUES (md5(gen_random_uuid()::text) || md5(gen_random_uuid()::text), $1, $2, NOW() + $3::interval)
		RETURNING id`, userID, tenantID, expires).Scan(&id); err != nil {
		t.Fatalf("session: %v", err)
	}
	return id
}

// An expired session is not an active one, and neither is a revoked one. Both
// would otherwise appear in a list an administrator acts on, and the button
// beside them would do nothing.
func TestListShowsOnlyLiveSessions(t *testing.T) {
	store, pool := testStore(t)
	ctx := context.Background()
	tenantID, userID := tenant(t, pool)

	live := session(t, pool, tenantID, userID, "1 hour")
	session(t, pool, tenantID, userID, "-1 hour")
	revoked := session(t, pool, tenantID, userID, "1 hour")
	if _, err := pool.Exec(ctx, `UPDATE sessions SET revoked_at = NOW() WHERE id = $1`, revoked); err != nil {
		t.Fatal(err)
	}

	found, err := store.List(ctx, tenantID, "")
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(found) != 1 || found[0].ID != live {
		t.Fatalf("expected only the live session, got %d", len(found))
	}
	if found[0].Email == "" {
		t.Error("the list should carry who the session belongs to")
	}
}

func TestRevokingEndsTheSessionAndSaysWhoDidIt(t *testing.T) {
	store, pool := testStore(t)
	ctx := context.Background()
	tenantID, userID := tenant(t, pool)
	id := session(t, pool, tenantID, userID, "1 hour")

	if err := store.Revoke(ctx, tenantID, id, userID, "laptop left on a train"); err != nil {
		t.Fatalf("revoke: %v", err)
	}
	found, err := store.List(ctx, tenantID, "")
	if err != nil {
		t.Fatal(err)
	}
	if len(found) != 0 {
		t.Fatalf("the session is still listed as active")
	}

	events, err := store.Events(ctx, tenantID, 10)
	if err != nil {
		t.Fatalf("events: %v", err)
	}
	if len(events) != 1 {
		t.Fatalf("expected one event, got %d", len(events))
	}
	if events[0].ActorID != userID || events[0].Reason != "laptop left on a train" {
		t.Fatalf("the event does not say who or why: %+v", events[0])
	}
}

// Revoking twice is not an error the second time round in any interesting
// sense, but it must not write a second event: a trail that records an act that
// did not happen is worse than no trail.
func TestRevokingSomethingAlreadyOverIsNotFound(t *testing.T) {
	store, pool := testStore(t)
	ctx := context.Background()
	tenantID, userID := tenant(t, pool)
	id := session(t, pool, tenantID, userID, "1 hour")

	if err := store.Revoke(ctx, tenantID, id, userID, "first"); err != nil {
		t.Fatalf("revoke: %v", err)
	}
	if err := store.Revoke(ctx, tenantID, id, userID, "second"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("second revoke: %v", err)
	}
	events, _ := store.Events(ctx, tenantID, 10)
	if len(events) != 1 {
		t.Fatalf("a second event was written for a session that was already over: %d", len(events))
	}
}

func TestRevokingEverythingSomebodyHoldsCountsThem(t *testing.T) {
	store, pool := testStore(t)
	ctx := context.Background()
	tenantID, userID := tenant(t, pool)
	session(t, pool, tenantID, userID, "1 hour")
	session(t, pool, tenantID, userID, "2 hours")
	session(t, pool, tenantID, userID, "-1 hour")

	count, err := store.RevokeAllFor(ctx, tenantID, userID, userID, "they have left")
	if err != nil {
		t.Fatalf("revoke all: %v", err)
	}
	if count != 2 {
		t.Fatalf("revoked %d, expected the two that were live", count)
	}
	events, _ := store.Events(ctx, tenantID, 10)
	if len(events) != 2 {
		t.Fatalf("expected two events, got %d", len(events))
	}
}

// Another organisation's sessions are not this one's, whatever id is asked for.
func TestOneOrganisationCannotEndAnothersSession(t *testing.T) {
	store, pool := testStore(t)
	ctx := context.Background()
	mine, _ := tenant(t, pool)
	theirs, theirUser := tenant(t, pool)
	id := session(t, pool, theirs, theirUser, "1 hour")

	if err := store.Revoke(ctx, mine, id, theirUser, "not mine to end"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("a session in another organisation was reachable: %v", err)
	}
}
