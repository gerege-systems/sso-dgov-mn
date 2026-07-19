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
	"time"

	"template/internal/apperror"
	"template/internal/business/domain"
	"template/pkg/hydra"
)

// hydraClients нь Hydra admin-ийн OAuth2 client CRUD + жагсаалт (*hydra.Admin
// үүнийг хангана) — usecase-ыг тест хийхэд эвтэйхэн. Applications-ийн ЦОРЫН ГАНЦ
// эх сурвалж нь Hydra: апп бүр нэг OAuth2 client. Overlay (tags/enabled/
// app_type)-ыг Hydra client-ийн metadata-д хадгална.
type hydraClients interface {
	ListClients(ctx context.Context) ([]hydra.ClientListEntry, error)
	CreateClient(ctx context.Context, body hydra.ClientCreate) (*hydra.ClientCreate, error)
	GetClient(ctx context.Context, clientID string) (*hydra.ClientListEntry, error)
	UpdateClient(ctx context.Context, clientID string, body hydra.ClientUpdate) (*hydra.ClientCreate, error)
	DeleteClient(ctx context.Context, clientID string) error
}

// serviceScopeResolver нь gateway service id ↔ OAuth scope хооронд хөрвүүлнэ
// (gateway_services хүснэгт). Апп-ыг Hydra эзэмшдэг тул DB-ээс зөвхөн service
// scope-ийг л резолв хийнэ (application-ийн бусад өгөгдлийг DB-д хадгалахаа больсон).
type serviceScopeResolver interface {
	ServiceScopes(ctx context.Context, serviceIDs []string) ([]string, error)
	ServiceIDsForScopes(ctx context.Context, scopes []string) ([]string, error)
}

type usecase struct {
	svc   serviceScopeResolver
	hydra hydraClients
}

// NewUsecase нь applications usecase-ыг буцаана. hydra нь Hydra admin client
// (ProviderConfigured үед л энэ usecase залгагдана); svc нь gateway service
// scope resolver.
func NewUsecase(svc serviceScopeResolver, h hydraClients) Usecase {
	return &usecase{svc: svc, hydra: h}
}

func (u *usecase) List(ctx context.Context) ([]domain.Application, error) {
	clients, err := u.hydra.ListClients(ctx)
	if err != nil {
		return nil, apperror.InternalCause(fmt.Errorf("list oauth clients: %w", err))
	}
	out := make([]domain.Application, 0, len(clients))
	for i := range clients {
		app, err := u.clientToApp(ctx, &clients[i])
		if err != nil {
			return nil, err
		}
		out = append(out, app)
	}
	return out, nil
}

func (u *usecase) Get(ctx context.Context, id string) (domain.Application, error) {
	c, err := u.hydra.GetClient(ctx, id)
	if err != nil {
		if isNotFound(err) {
			return domain.Application{}, apperror.NotFound("application not found")
		}
		return domain.Application{}, apperror.InternalCause(err)
	}
	return u.clientToApp(ctx, c)
}

func (u *usecase) Create(ctx context.Context, in Input) (domain.Application, error) {
	app, err := validate(in)
	if err != nil {
		return domain.Application{}, err
	}
	app.ClientID = "app-" + randomHex(8)
	app.ID = app.ClientID // Hydra-д тусдаа UUID байхгүй — client_id нь танигч.

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
	// Public (spa/native)-д secret байхгүй; confidential (web/m2m)-д зөвхөн энэ
	// хариунд НЭГ удаа (дараа нь Hydra эзэмшинэ).
	if !domain.AppIsPublic(app.AppType) {
		app.Secret = firstNonEmpty(created.ClientSecret, secret)
	}
	app.CreatedAt = time.Now()
	return app, nil
}

func (u *usecase) Update(ctx context.Context, id string, in Input) (domain.Application, error) {
	app, err := validate(in)
	if err != nil {
		return domain.Application{}, err
	}
	if _, err := u.hydra.GetClient(ctx, id); err != nil {
		if isNotFound(err) {
			return domain.Application{}, apperror.NotFound("application not found")
		}
		return domain.Application{}, apperror.InternalCause(err)
	}
	app.ID = id
	app.ClientID = id

	scopes, err := u.scopesFor(ctx, app.AppType, app.ServiceIDs)
	if err != nil {
		return domain.Application{}, err
	}
	// secret хоосон → Hydra одоогийн secret-ыг хадгална (rotate биш).
	if _, err := u.hydra.UpdateClient(ctx, id, buildClient(app, scopes, "")); err != nil {
		return domain.Application{}, apperror.InternalCause(fmt.Errorf("update oauth client: %w", err))
	}
	return u.Get(ctx, id)
}

func (u *usecase) Delete(ctx context.Context, id string) error {
	if err := u.hydra.DeleteClient(ctx, id); err != nil && !isNotFound(err) {
		return apperror.InternalCause(fmt.Errorf("delete oauth client: %w", err))
	}
	return nil
}

func (u *usecase) RotateSecret(ctx context.Context, id string) (domain.Application, error) {
	c, err := u.hydra.GetClient(ctx, id)
	if err != nil {
		if isNotFound(err) {
			return domain.Application{}, apperror.BadRequest("application has no OAuth client to rotate")
		}
		return domain.Application{}, apperror.InternalCause(err)
	}
	app, err := u.clientToApp(ctx, c)
	if err != nil {
		return domain.Application{}, err
	}
	if domain.AppIsPublic(app.AppType) {
		return domain.Application{}, apperror.BadRequest("public client (spa/native) has no secret to rotate")
	}
	scopes, err := u.scopesFor(ctx, app.AppType, app.ServiceIDs)
	if err != nil {
		return domain.Application{}, err
	}
	secret := randomToken(40)
	if _, err := u.hydra.UpdateClient(ctx, id, buildClient(app, scopes, secret)); err != nil {
		return domain.Application{}, apperror.InternalCause(fmt.Errorf("rotate oauth secret: %w", err))
	}
	app.Secret = secret
	return app, nil
}

func (u *usecase) SetServices(ctx context.Context, id string, serviceIDs []string) (domain.Application, error) {
	c, err := u.hydra.GetClient(ctx, id)
	if err != nil {
		if isNotFound(err) {
			return domain.Application{}, apperror.NotFound("application not found")
		}
		return domain.Application{}, apperror.InternalCause(err)
	}
	app, err := u.clientToApp(ctx, c)
	if err != nil {
		return domain.Application{}, err
	}
	app.ServiceIDs = cleanList(serviceIDs)
	scopes, err := u.scopesFor(ctx, app.AppType, app.ServiceIDs)
	if err != nil {
		return domain.Application{}, err
	}
	if _, err := u.hydra.UpdateClient(ctx, id, buildClient(app, scopes, "")); err != nil {
		return domain.Application{}, apperror.InternalCause(fmt.Errorf("update oauth client: %w", err))
	}
	return u.Get(ctx, id)
}

// ReconcileClients нь хуучин DB overlay → Hydra тулгалт байсан. Апп-ыг одоо
// Hydra эзэмшдэг тул тулгах зүйлгүй — no-op (интерфейсийн нийцлийн төлөө үлдээв).
func (u *usecase) ReconcileClients(_ context.Context) (int, error) { return 0, nil }

// scopesFor нь base OIDC scope (RP төрөлд) + зөвшөөрсөн service-үүдийн scope-г нэгтгэнэ.
func (u *usecase) scopesFor(ctx context.Context, appType string, serviceIDs []string) ([]string, error) {
	var out []string
	if domain.AppUsesRedirect(appType) {
		out = append(out, "openid", "profile", "email")
	}
	svc, err := u.svc.ServiceScopes(ctx, serviceIDs)
	if err != nil {
		return nil, apperror.InternalCause(err)
	}
	return dedup(append(out, svc...)), nil
}

// clientToApp нь Hydra OAuth2 client-ыг домэйн Application болгоно. tags/enabled/
// app_type-ыг client metadata-аас (байхгүй бол grant-аас дүгнэж / default), service
// id-уудыг svc:* scope-оос сэргээнэ.
func (u *usecase) clientToApp(ctx context.Context, c *hydra.ClientListEntry) (domain.Application, error) {
	serviceIDs, err := u.svc.ServiceIDsForScopes(ctx, filterSvcScopes(c.Scope))
	if err != nil {
		return domain.Application{}, apperror.InternalCause(err)
	}
	app := domain.Application{
		ID:           c.ClientID,
		ClientID:     c.ClientID,
		Name:         c.ClientName,
		AppType:      appTypeOf(c),
		Tags:         metaTags(c.Metadata),
		RedirectURIs: c.RedirectURIs,
		Enabled:      metaEnabled(c.Metadata),
		ServiceIDs:   cleanList(serviceIDs),
		CreatedAt:    parseTime(c.CreatedAt),
	}
	if ut := parseTime(c.UpdatedAt); !ut.IsZero() {
		app.UpdatedAt = &ut
	}
	return app, nil
}

// buildClient нь домэйн апп-аас Hydra ClientCreate/Update body-г угсарна. secret
// хоосон бол Hydra одоогийн secret-ыг хадгална (update); шинэ secret бол сольно.
// tags/enabled/app_type-ыг metadata-д тавина.
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
		Metadata: map[string]any{
			"app_type": app.AppType,
			"tags":     arrStr(app.Tags),
			"enabled":  app.Enabled,
		},
	}
	if domain.AppUsesRedirect(app.AppType) {
		b.RedirectURIs = app.RedirectURIs
	}
	return b
}

// appTypeOf нь Hydra client-ээс апп төрлийг тодорхойлно: metadata.app_type байвал
// түүнийг (spa/native ялгааг хадгална), эс бол grant/auth-method-оос дүгнэнэ.
func appTypeOf(c *hydra.ClientListEntry) string {
	if t, ok := c.Metadata["app_type"].(string); ok && domain.AppTypes[t] {
		return t
	}
	if contains(c.GrantTypes, "client_credentials") {
		return "m2m"
	}
	if c.TokenEndpointAuthMethod == "none" {
		return "spa" // public authorization_code (native-г metadata-гүйгээр ялгах боломжгүй)
	}
	return "web"
}

// filterSvcScopes нь scope мөрөөс зөвхөн gateway service scope-уудыг (svc:*) авна.
func filterSvcScopes(scope string) []string {
	var out []string
	for _, s := range strings.Fields(scope) {
		if strings.HasPrefix(s, "svc:") {
			out = append(out, s)
		}
	}
	return out
}

func metaTags(m map[string]any) []string {
	raw, ok := m["tags"].([]any)
	if !ok {
		return nil
	}
	out := make([]string, 0, len(raw))
	for _, v := range raw {
		if s, ok := v.(string); ok && strings.TrimSpace(s) != "" {
			out = append(out, s)
		}
	}
	return out
}

// metaEnabled нь metadata.enabled-ийг уншина; байхгүй бол default идэвхтэй (хуучин
// metadata-гүй client-уудыг идэвхтэй гэж үзнэ).
func metaEnabled(m map[string]any) bool {
	if v, ok := m["enabled"].(bool); ok {
		return v
	}
	return true
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

func arrStr(in []string) []string {
	if in == nil {
		return []string{}
	}
	return in
}

func contains(list []string, v string) bool {
	for _, s := range list {
		if s == v {
			return true
		}
	}
	return false
}

func parseTime(s string) time.Time {
	if s == "" {
		return time.Time{}
	}
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return time.Time{}
	}
	return t
}

func firstNonEmpty(a, b string) string {
	if a != "" {
		return a
	}
	return b
}

// isNotFound нь Hydra 404 (client байхгүй)-г таних.
func isNotFound(err error) bool {
	return err != nil && (strings.Contains(err.Error(), "404") || strings.Contains(err.Error(), "Not Found"))
}
