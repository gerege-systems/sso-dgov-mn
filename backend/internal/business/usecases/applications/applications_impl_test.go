// Government Template Platform V3.0
// Gerege Systems Development Team болон Claude AI хамтран бүтээв, 2026.

package applications

import (
	"context"
	"testing"

	"template/pkg/hydra"
)

// fakeRepo нь serviceScopeResolver-ийн санах-ой хувилбар (gateway service scope
// ↔ id). Апп-ыг одоо Hydra эзэмшдэг тул тест зөвхөн scope резолвыг хуурамчилна.
type fakeRepo struct {
	scopes     []string // ServiceScopes буцаах утга
	serviceIDs []string // ServiceIDsForScopes буцаах утга
}

func (f *fakeRepo) ServiceScopes(context.Context, []string) ([]string, error) {
	return f.scopes, nil
}
func (f *fakeRepo) ServiceIDsForScopes(context.Context, []string) ([]string, error) {
	return f.serviceIDs, nil
}

// fakeHydra нь hydraClients-ийн тест хувилбар — сүүлд илгээсэн body-г хадгална.
type fakeHydra struct {
	lastCreate hydra.ClientCreate
	list       []hydra.ClientListEntry
	getErr     error
}

func (h *fakeHydra) ListClients(context.Context) ([]hydra.ClientListEntry, error) {
	return h.list, nil
}
func (h *fakeHydra) CreateClient(_ context.Context, b hydra.ClientCreate) (*hydra.ClientCreate, error) {
	h.lastCreate = b
	out := b
	out.ClientSecret = "s3cr3t-echoed"
	return &out, nil
}
func (h *fakeHydra) GetClient(context.Context, string) (*hydra.ClientListEntry, error) {
	return &hydra.ClientListEntry{}, h.getErr
}
func (h *fakeHydra) UpdateClient(_ context.Context, _ string, b hydra.ClientCreate) (*hydra.ClientCreate, error) {
	h.lastCreate = b
	return &b, nil
}
func (h *fakeHydra) DeleteClient(context.Context, string) error { return nil }

func TestCreateM2MProvisionsClientAndReturnsSecret(t *testing.T) {
	repo := &fakeRepo{scopes: []string{"svc:eid-core"}}
	fh := &fakeHydra{}
	uc := NewUsecase(repo, fh)

	app, err := uc.Create(context.Background(), Input{
		Name: "analytics-job", AppType: "m2m", ServiceIDs: []string{"id-1"}, Enabled: true,
	})
	if err != nil {
		t.Fatalf("Create m2m: %v", err)
	}
	if app.Secret == "" {
		t.Fatal("expected a one-time secret for a confidential (m2m) app")
	}
	if app.ID == "" || app.ID != app.ClientID {
		t.Fatalf("app id should equal client_id, got id=%q client_id=%q", app.ID, app.ClientID)
	}
	if got := fh.lastCreate.GrantTypes; len(got) != 1 || got[0] != "client_credentials" {
		t.Fatalf("m2m should use client_credentials grant, got %v", got)
	}
	if fh.lastCreate.Scope != "svc:eid-core" {
		t.Fatalf("scope should come from allowed services, got %q", fh.lastCreate.Scope)
	}
	if len(fh.lastCreate.RedirectURIs) != 0 {
		t.Fatalf("m2m must not carry redirect_uris, got %v", fh.lastCreate.RedirectURIs)
	}
	// Overlay (tags/enabled/app_type) нь Hydra metadata-д хадгалагдана.
	if fh.lastCreate.Metadata["app_type"] != "m2m" || fh.lastCreate.Metadata["enabled"] != true {
		t.Fatalf("overlay must be stored in metadata, got %v", fh.lastCreate.Metadata)
	}
}

func TestCreateWebRequiresRedirectAndBaseScopes(t *testing.T) {
	repo := &fakeRepo{}
	uc := NewUsecase(repo, &fakeHydra{})

	// redirect дутуу → алдаа
	if _, err := uc.Create(context.Background(), Input{Name: "portal", AppType: "web"}); err == nil {
		t.Fatal("web app without redirect_uri should fail validation")
	}

	fh := &fakeHydra{}
	uc = NewUsecase(&fakeRepo{}, fh)
	app, err := uc.Create(context.Background(), Input{
		Name: "portal", AppType: "web", RedirectURIs: []string{"https://rp.example.mn/cb"},
	})
	if err != nil {
		t.Fatalf("Create web: %v", err)
	}
	if app.Secret == "" {
		t.Fatal("web (confidential) app should get a secret")
	}
	if got := fh.lastCreate.Scope; got == "" || got[:6] != "openid" {
		t.Fatalf("RP app should include base OIDC scopes, got %q", got)
	}
	if len(fh.lastCreate.RedirectURIs) != 1 {
		t.Fatalf("web app should carry its redirect_uri, got %v", fh.lastCreate.RedirectURIs)
	}
}

func TestValidateRejectsBadAppType(t *testing.T) {
	uc := NewUsecase(&fakeRepo{}, &fakeHydra{})
	if _, err := uc.Create(context.Background(), Input{Name: "x", AppType: "nonsense"}); err == nil {
		t.Fatal("invalid app_type should be rejected")
	}
}

func TestSpaIsPublicNoSecret(t *testing.T) {
	fh := &fakeHydra{}
	uc := NewUsecase(&fakeRepo{}, fh)
	app, err := uc.Create(context.Background(), Input{
		Name: "spa", AppType: "spa", RedirectURIs: []string{"https://app.example.mn/cb"},
	})
	if err != nil {
		t.Fatalf("Create spa: %v", err)
	}
	if app.Secret != "" {
		t.Fatal("public (spa) client must not return a secret")
	}
	if fh.lastCreate.TokenEndpointAuthMethod != "none" {
		t.Fatalf("spa should use token_endpoint_auth_method=none, got %q", fh.lastCreate.TokenEndpointAuthMethod)
	}
}

// TestListMapsHydraClients — List нь Hydra client-уудыг домэйн Application болгож
// буулгана: metadata-аас tags/enabled/app_type, svc:* scope-оос service id-ууд,
// metadata байхгүй бол grant-аас төрөл + идэвхтэй default.
func TestListMapsHydraClients(t *testing.T) {
	repo := &fakeRepo{serviceIDs: []string{"id-1"}}
	fh := &fakeHydra{list: []hydra.ClientListEntry{
		{
			ClientID: "template-dgov-mn", ClientName: "template.dgov.mn",
			GrantTypes: []string{"authorization_code", "refresh_token"}, TokenEndpointAuthMethod: "client_secret_basic",
			RedirectURIs: []string{"https://template.dgov.mn/cb"},
			Scope:        "openid profile email svc:eid-sign",
			Metadata:     map[string]any{"tags": []any{"rp"}, "enabled": true, "app_type": "web"},
		},
		{
			ClientID: "m2m-1", ClientName: "job",
			GrantTypes: []string{"client_credentials"}, TokenEndpointAuthMethod: "client_secret_basic",
			Scope: "svc:eid-core",
			// metadata байхгүй → app_type grant-аас m2m, enabled default true
		},
	}}
	apps, err := NewUsecase(repo, fh).List(context.Background())
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(apps) != 2 {
		t.Fatalf("want 2 apps, got %d", len(apps))
	}
	a := apps[0]
	if a.ID != "template-dgov-mn" || a.ClientID != "template-dgov-mn" {
		t.Fatalf("id/client_id should be the Hydra client_id, got %+v", a)
	}
	if a.Name != "template.dgov.mn" || a.AppType != "web" {
		t.Fatalf("name/type mapping wrong: %+v", a)
	}
	if len(a.Tags) != 1 || a.Tags[0] != "rp" || !a.Enabled {
		t.Fatalf("tags/enabled should come from metadata: tags=%v enabled=%v", a.Tags, a.Enabled)
	}
	if len(a.ServiceIDs) != 1 || a.ServiceIDs[0] != "id-1" {
		t.Fatalf("service ids should be resolved from svc:* scopes, got %v", a.ServiceIDs)
	}
	if apps[1].AppType != "m2m" || !apps[1].Enabled {
		t.Fatalf("no-metadata client should derive m2m + default enabled, got %+v", apps[1])
	}
}

func TestSetSecretWritesTheGivenSecretToHydra(t *testing.T) {
	fh := &fakeHydra{}
	uc := NewUsecase(&fakeRepo{}, fh)

	const want = "my-preconfigured-rp-secret"
	app, err := uc.SetSecret(context.Background(), "ring-dgov-mn", want)
	if err != nil {
		t.Fatalf("SetSecret: %v", err)
	}
	// Rotate-ээс ялгаатай нь — санамсаргүй биш, ЯГ өгсөн утга очно.
	if fh.lastCreate.ClientSecret != want {
		t.Fatalf("hydra should receive the given secret, got %q", fh.lastCreate.ClientSecret)
	}
	if app.Secret != want {
		t.Fatalf("response should echo the secret once, got %q", app.Secret)
	}
}

func TestSetSecretRejectsShortSecret(t *testing.T) {
	fh := &fakeHydra{}
	uc := NewUsecase(&fakeRepo{}, fh)

	if _, err := uc.SetSecret(context.Background(), "ring-dgov-mn", "  short  "); err == nil {
		t.Fatal("expected a bad-request error for a secret under the minimum length")
	}
	if fh.lastCreate.ClientSecret != "" {
		t.Fatal("hydra must not be called when validation fails")
	}
}
