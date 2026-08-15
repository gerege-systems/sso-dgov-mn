package federation

import (
	"context"
	"errors"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
)

// The tests that need rows: the COALESCE that lets an edit leave a secret it
// cannot read, and the isolation that makes another organisation's provider
// simply not exist.
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

func tenant(t *testing.T, pool *pgxpool.Pool) string {
	t.Helper()
	var id string
	if err := pool.QueryRow(context.Background(), `
		INSERT INTO tenants (slug, name) VALUES ('t-' || gen_random_uuid(), 'Test') RETURNING id`).
		Scan(&id); err != nil {
		t.Fatalf("tenant: %v", err)
	}
	t.Cleanup(func() { _, _ = pool.Exec(context.Background(), `DELETE FROM tenants WHERE id = $1`, id) })
	return id
}

func input() Input {
	return Input{
		DisplayName: "Ministry", Issuer: "https://id.example.mn",
		ClientID: "abc", ClientSecret: "s3cr3t",
		Scopes: "openid profile", AttributeMap: map[string]string{"email": "mail"},
	}
}

func TestARegisteredProviderComesBackWithoutItsSecret(t *testing.T) {
	store, pool := testStore(t)
	ctx := context.Background()
	tenantID := tenant(t, pool)

	in := input()
	if err := in.validate(true); err != nil {
		t.Fatalf("validate: %v", err)
	}
	provider, err := store.Create(ctx, tenantID, "", in, []byte("sealed"))
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if !provider.HasSecret {
		t.Error("the provider should report that a credential is stored")
	}
	if provider.AttributeMap["email"] != "mail" {
		t.Errorf("the attribute map did not survive the round trip: %+v", provider.AttributeMap)
	}

	listed, err := store.List(ctx, tenantID)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(listed) != 1 || listed[0].ID != provider.ID {
		t.Fatalf("expected the one provider, got %d", len(listed))
	}
}

// An edit with no secret must keep the stored one. Anything else would mean a
// console that cannot read a credential also cannot change a display name
// without destroying it.
func TestAnEditWithoutASecretKeepsTheStoredOne(t *testing.T) {
	store, pool := testStore(t)
	ctx := context.Background()
	tenantID := tenant(t, pool)

	in := input()
	_ = in.validate(true)
	provider, err := store.Create(ctx, tenantID, "", in, []byte("sealed"))
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	edit := input()
	edit.DisplayName = "Ministry of Everything"
	edit.ClientSecret = ""
	_ = edit.validate(false)
	if _, err := store.Update(ctx, tenantID, provider.ID, edit, nil); err != nil {
		t.Fatalf("update: %v", err)
	}

	var stored []byte
	if err := pool.QueryRow(ctx,
		`SELECT client_secret_encrypted FROM sso_federation_providers WHERE id = $1`,
		provider.ID).Scan(&stored); err != nil {
		t.Fatal(err)
	}
	if string(stored) != "sealed" {
		t.Fatalf("the stored credential was replaced with %q", stored)
	}
}

func TestOneOrganisationCannotReachAnothersProvider(t *testing.T) {
	store, pool := testStore(t)
	ctx := context.Background()
	mine, theirs := tenant(t, pool), tenant(t, pool)

	in := input()
	_ = in.validate(true)
	provider, err := store.Create(ctx, theirs, "", in, []byte("sealed"))
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	if _, err := store.Get(ctx, mine, provider.ID); !errors.Is(err, ErrNotFound) {
		t.Fatalf("get: %v", err)
	}
	if err := store.Delete(ctx, mine, provider.ID); !errors.Is(err, ErrNotFound) {
		t.Fatalf("delete: %v", err)
	}
	listed, _ := store.List(ctx, mine)
	if len(listed) != 0 {
		t.Fatalf("another organisation's provider is listed: %d", len(listed))
	}
}

// The same issuer twice in one organisation is refused, and the handler turns
// that into a 409 rather than a 500 — which is what isDuplicate is for.
func TestTheSameIssuerTwiceIsRefused(t *testing.T) {
	store, pool := testStore(t)
	ctx := context.Background()
	tenantID := tenant(t, pool)

	in := input()
	_ = in.validate(true)
	if _, err := store.Create(ctx, tenantID, "", in, []byte("sealed")); err != nil {
		t.Fatalf("create: %v", err)
	}
	_, err := store.Create(ctx, tenantID, "", in, []byte("sealed"))
	if err == nil {
		t.Fatal("the same issuer was registered twice")
	}
	if !isDuplicate(err) {
		t.Fatalf("the conflict was not recognised as one: %v", err)
	}
}
