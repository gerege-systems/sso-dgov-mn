// Government Template Platform V3.0
// Gerege Systems Development Team болон Claude AI хамтран бүтээв, 2026.

package applications

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"net/url"
	"strings"

	"template/internal/apperror"
	"template/internal/business/domain"
	repointerface "template/internal/datasources/repositories/interface"
	"template/pkg/hydra"
)

// hydraClients нь Hydra admin-ийн ашиглах OAuth2 client CRUD дэд олонлог
// (*hydra.Admin үүнийг хангана) — usecase-ыг тест хийхэд эвтэйхэн.
type hydraClients interface {
	CreateClient(ctx context.Context, body hydra.ClientCreate) (*hydra.ClientCreate, error)
	GetClient(ctx context.Context, clientID string) (*hydra.ClientListEntry, error)
	UpdateClient(ctx context.Context, clientID string, body hydra.ClientUpdate) (*hydra.ClientCreate, error)
	DeleteClient(ctx context.Context, clientID string) error
}

type usecase struct {
	repo  repointerface.ApplicationRepository
	hydra hydraClients
}

// NewUsecase нь applications usecase-ыг буцаана. hydra нь Hydra admin client
// (ProviderConfigured үед л энэ usecase залгагдана).
func NewUsecase(repo repointerface.ApplicationRepository, h hydraClients) Usecase {
	return &usecase{repo: repo, hydra: h}
}

func (u *usecase) List(ctx context.Context) ([]domain.Application, error) {
	return u.repo.List(ctx)
}

func (u *usecase) Get(ctx context.Context, id string) (domain.Application, error) {
	return u.repo.Get(ctx, id)
}

func (u *usecase) Create(ctx context.Context, in Input) (domain.Application, error) {
	app, err := validate(in)
	if err != nil {
		return domain.Application{}, err
	}
	app.ClientID = "app-" + randomHex(8)

	scopes, err := u.scopesFor(ctx, app.AppType, app.ServiceIDs)
	if err != nil {
		return domain.Application{}, err
	}
	secret := ""
	if !domain.AppIsPublic(app.AppType) {
		secret = randomToken(40)
	}
	created, err := u.hydra.CreateClient(ctx, buildClient(app, scopes, secret))
	if err != nil {
		return domain.Application{}, apperror.InternalCause(fmt.Errorf("create oauth client: %w", err))
	}

	stored, err := u.repo.Create(ctx, &app)
	if err != nil {
		// Overlay бичих амжилтгүй бол Hydra client-ыг цэвэрлэж orphan үлдээхгүй.
		_ = u.hydra.DeleteClient(ctx, app.ClientID)
		return domain.Application{}, err
	}
	// Public (spa/native) client-д secret байхгүй — Hydra юу буцаахаас үл хамааран
	// хэзээ ч гаргахгүй. Confidential (web/m2m)-д зөвхөн энэ хариунд НЭГ удаа.
	if !domain.AppIsPublic(app.AppType) {
		stored.Secret = firstNonEmpty(created.ClientSecret, secret)
	}
	return stored, nil
}

func (u *usecase) Update(ctx context.Context, id string, in Input) (domain.Application, error) {
	app, err := validate(in)
	if err != nil {
		return domain.Application{}, err
	}
	existing, err := u.repo.Get(ctx, id)
	if err != nil {
		return domain.Application{}, err
	}
	app.ID = id
	app.ClientID = existing.ClientID

	scopes, err := u.scopesFor(ctx, app.AppType, app.ServiceIDs)
	if err != nil {
		return domain.Application{}, err
	}
	if err := u.syncClient(ctx, app, scopes, ""); err != nil {
		return domain.Application{}, err
	}
	return u.repo.Update(ctx, &app)
}

func (u *usecase) Delete(ctx context.Context, id string) error {
	app, err := u.repo.Get(ctx, id)
	if err != nil {
		return err
	}
	if err := u.hydra.DeleteClient(ctx, app.ClientID); err != nil && !isNotFound(err) {
		return apperror.InternalCause(fmt.Errorf("delete oauth client: %w", err))
	}
	return u.repo.Delete(ctx, id)
}

func (u *usecase) RotateSecret(ctx context.Context, id string) (domain.Application, error) {
	app, err := u.repo.Get(ctx, id)
	if err != nil {
		return domain.Application{}, err
	}
	if domain.AppIsPublic(app.AppType) {
		return domain.Application{}, apperror.BadRequest("public client (spa/native) has no secret to rotate")
	}
	if _, err := u.hydra.GetClient(ctx, app.ClientID); err != nil {
		if isNotFound(err) {
			return domain.Application{}, apperror.BadRequest("application has no OAuth client to rotate")
		}
		return domain.Application{}, apperror.InternalCause(err)
	}
	scopes, err := u.scopesFor(ctx, app.AppType, app.ServiceIDs)
	if err != nil {
		return domain.Application{}, err
	}
	secret := randomToken(40)
	if _, err := u.hydra.UpdateClient(ctx, app.ClientID, buildClient(app, scopes, secret)); err != nil {
		return domain.Application{}, apperror.InternalCause(fmt.Errorf("rotate oauth secret: %w", err))
	}
	app.Secret = secret
	return app, nil
}

func (u *usecase) SetServices(ctx context.Context, id string, serviceIDs []string) (domain.Application, error) {
	app, err := u.repo.Get(ctx, id)
	if err != nil {
		return domain.Application{}, err
	}
	app.ServiceIDs = cleanList(serviceIDs)
	scopes, err := u.scopesFor(ctx, app.AppType, app.ServiceIDs)
	if err != nil {
		return domain.Application{}, err
	}
	if err := u.syncClient(ctx, app, scopes, ""); err != nil {
		return domain.Application{}, err
	}
	if err := u.repo.SetServices(ctx, id, app.ServiceIDs); err != nil {
		return domain.Application{}, err
	}
	return u.repo.Get(ctx, id)
}

// seedRPMarker нь migration-аар seed хийсэн RP overlay мөрийг тэмдэглэнэ
// (32_real_gateway_services). Bootstrap ЗӨВХӨН эдгээрт Hydra client үүсгэнэ.
const seedRPMarker = "seed-rp"

// ReconcileClients нь seed хийсэн RP overlay мөрүүдийн Hydra client дутуу байвал
// үүсгэнэ (startup дээр нэг удаа). Байгаа client-ыг алгасна (idempotent); UI-аас
// устгасан RP-д мөр байхгүй тул дахин үүсэхгүй. Гарын үсэг зурах secret нь Hydra-д
// үүснэ; админ UI-аас "rotate secret"-ээр авна.
func (u *usecase) ReconcileClients(ctx context.Context) (int, error) {
	apps, err := u.repo.List(ctx)
	if err != nil {
		return 0, err
	}
	created := 0
	for _, app := range apps {
		if app.CreatedBy != seedRPMarker {
			continue
		}
		if _, err := u.hydra.GetClient(ctx, app.ClientID); err == nil {
			continue // Hydra client аль хэдийн бий
		} else if !isNotFound(err) {
			return created, apperror.InternalCause(err)
		}
		scopes, err := u.scopesFor(ctx, app.AppType, app.ServiceIDs)
		if err != nil {
			return created, err
		}
		secret := ""
		if !domain.AppIsPublic(app.AppType) {
			secret = randomToken(40)
		}
		if _, err := u.hydra.CreateClient(ctx, buildClient(app, scopes, secret)); err != nil {
			return created, apperror.InternalCause(fmt.Errorf("reconcile client %s: %w", app.ClientID, err))
		}
		created++
	}
	return created, nil
}

// syncClient нь апп-ын desired state-ыг Hydra client руу бичнэ. Hydra client
// байхгүй (demo/seed апп) бол чимээгүй алгасна — overlay-only.
func (u *usecase) syncClient(ctx context.Context, app domain.Application, scopes []string, secret string) error {
	if _, err := u.hydra.GetClient(ctx, app.ClientID); err != nil {
		if isNotFound(err) {
			return nil
		}
		return apperror.InternalCause(err)
	}
	if _, err := u.hydra.UpdateClient(ctx, app.ClientID, buildClient(app, scopes, secret)); err != nil {
		return apperror.InternalCause(fmt.Errorf("update oauth client: %w", err))
	}
	return nil
}

// scopesFor нь base OIDC scope (RP төрөлд) + зөвшөөрсөн service-үүдийн scope-г нэгтгэнэ.
func (u *usecase) scopesFor(ctx context.Context, appType string, serviceIDs []string) ([]string, error) {
	var out []string
	if domain.AppUsesRedirect(appType) {
		out = append(out, "openid", "profile", "email")
	}
	svc, err := u.repo.ServiceScopes(ctx, serviceIDs)
	if err != nil {
		return nil, apperror.InternalCause(err)
	}
	return dedup(append(out, svc...)), nil
}

// buildClient нь домэйн апп-аас Hydra ClientCreate/Update body-г угсарна. secret
// хоосон бол Hydra одоогийн secret-ыг хадгална (update); шинэ secret бол сольно.
func buildClient(app domain.Application, scopes []string, secret string) hydra.ClientCreate {
	grants, responseTypes, authMethod := grantsFor(app.AppType)
	b := hydra.ClientCreate{
		ClientID:                app.ClientID,
		ClientName:              app.Name,
		ClientSecret:            secret,
		GrantTypes:              grants,
		ResponseTypes:           responseTypes,
		Scope:                   strings.Join(scopes, " "),
		TokenEndpointAuthMethod: authMethod,
		SubjectType:             "public",
	}
	if domain.AppUsesRedirect(app.AppType) {
		b.RedirectURIs = app.RedirectURIs
	}
	return b
}

// grantsFor нь апп төрлөөр Hydra grant_types / response_types / auth method-ыг өгнө.
func grantsFor(appType string) (grants, responseTypes []string, authMethod string) {
	switch appType {
	case "m2m":
		return []string{"client_credentials"}, []string{}, "client_secret_basic"
	case "spa", "native":
		return []string{"authorization_code", "refresh_token"}, []string{"code"}, "none"
	default: // web
		return []string{"authorization_code", "refresh_token"}, []string{"code"}, "client_secret_basic"
	}
}

// validate нь Input-ыг шалгаж нормчилсон домэйн апп болгоно.
func validate(in Input) (domain.Application, error) {
	name := strings.TrimSpace(in.Name)
	if name == "" {
		return domain.Application{}, apperror.BadRequest("application name is required")
	}
	if len(name) > 128 {
		return domain.Application{}, apperror.BadRequest("application name too long (max 128)")
	}
	appType := strings.TrimSpace(in.AppType)
	if appType == "" {
		appType = "m2m"
	}
	if !domain.AppTypes[appType] {
		return domain.Application{}, apperror.BadRequest("app_type must be web, spa, native or m2m")
	}

	redirects := cleanList(in.RedirectURIs)
	if domain.AppUsesRedirect(appType) {
		if len(redirects) == 0 {
			return domain.Application{}, apperror.BadRequest("at least one redirect_uri is required for this app type")
		}
		for _, u := range redirects {
			if err := validateRedirectURI(u, appType == "native"); err != nil {
				return domain.Application{}, apperror.BadRequest(err.Error())
			}
		}
	} else {
		redirects = nil // m2m нь redirect ашиглахгүй
	}

	return domain.Application{
		Name:         name,
		AppType:      appType,
		Tags:         cleanList(in.Tags),
		RedirectURIs: redirects,
		Enabled:      in.Enabled,
		ServiceIDs:   cleanList(in.ServiceIDs),
	}, nil
}

func validateRedirectURI(raw string, allowPrivateScheme bool) error {
	u, err := url.Parse(raw)
	if err != nil {
		return errors.New("redirect_uri: invalid URL")
	}
	if !u.IsAbs() {
		return errors.New("redirect_uri: must be absolute")
	}
	if u.Fragment != "" {
		return errors.New("redirect_uri: fragments not allowed (RFC 6749 §3.1.2)")
	}
	switch u.Scheme {
	case "https":
		return nil
	case "http":
		host := u.Hostname()
		if host != "localhost" && host != "127.0.0.1" && host != "::1" {
			return errors.New("redirect_uri: http only allowed on loopback")
		}
		return nil
	default:
		if allowPrivateScheme && u.Scheme != "" {
			return nil // native (RFC 8252) private-use scheme, e.g. geregetemp://oauth2/callback
		}
		return errors.New("redirect_uri: scheme must be https (or http on loopback)")
	}
}

// ── helpers ──────────────────────────────────────────────────────────────—

func randomHex(n int) string {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		panic("applications: crypto/rand failed: " + err.Error())
	}
	return hex.EncodeToString(b)
}

func randomToken(n int) string {
	const alphabet = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		panic("applications: crypto/rand failed: " + err.Error())
	}
	for i := range b {
		b[i] = alphabet[int(b[i])%len(alphabet)]
	}
	return string(b)
}

// cleanList нь trim + хоосон/давхардсаныг хасна (дараалал хадгална).
func cleanList(in []string) []string {
	seen := map[string]bool{}
	out := make([]string, 0, len(in))
	for _, s := range in {
		s = strings.TrimSpace(s)
		if s == "" || seen[s] {
			continue
		}
		seen[s] = true
		out = append(out, s)
	}
	return out
}

func dedup(in []string) []string { return cleanList(in) }

func firstNonEmpty(a, b string) string {
	if a != "" {
		return a
	}
	return b
}

// isNotFound нь Hydra 404 (client байхгүй)-г таних — demo апп-д client байхгүй.
func isNotFound(err error) bool {
	return err != nil && (strings.Contains(err.Error(), "404") || strings.Contains(err.Error(), "Not Found"))
}
