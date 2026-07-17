// Government Template Platform V3.0
// Gerege Systems Development Team болон Claude AI хамтран бүтээв, 2026.

package applications

import (
	"context"
	"testing"

	"template/internal/business/domain"
	"template/pkg/hydra"
)

// fakeRepo нь ApplicationRepository-ийн санах-ой хувилбар (тест).
type fakeRepo struct {
	created domain.Application
	scopes  []string
}

func (f *fakeRepo) List(context.Context) ([]domain.Application, error) { return nil, nil }
func (f *fakeRepo) Get(_ context.Context, id string) (domain.Application, error) {
	a := f.created
	a.ID = id
	return a, nil
}
func (f *fakeRepo) Create(_ context.Context, a *domain.Application) (domain.Application, error) {
	f.created = *a
	f.created.ID = "app-row-1"
	return f.created, nil
}
func (f *fakeRepo) Update(_ context.Context, a *domain.Application) (domain.Application, error) {
	f.created = *a
	return f.created, nil
}
func (f *fakeRepo) Delete(context.Context, string) error                { return nil }
func (f *fakeRepo) SetServices(context.Context, string, []string) error { return nil }
func (f *fakeRepo) ServiceScopes(context.Context, []string) ([]string, error) {
	return f.scopes, nil
}

// fakeHydra нь hydraClients-ийн тест хувилбар — сүүлд илгээсэн body-г хадгална.
type fakeHydra struct {
	lastCreate hydra.ClientCreate
	getErr     error
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
	if got := fh.lastCreate.GrantTypes; len(got) != 1 || got[0] != "client_credentials" {
		t.Fatalf("m2m should use client_credentials grant, got %v", got)
	}
	if fh.lastCreate.Scope != "svc:eid-core" {
		t.Fatalf("scope should come from allowed services, got %q", fh.lastCreate.Scope)
	}
	if len(fh.lastCreate.RedirectURIs) != 0 {
		t.Fatalf("m2m must not carry redirect_uris, got %v", fh.lastCreate.RedirectURIs)
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
