// Government Template Platform V3.0
// Gerege Systems Development Team болон Claude AI хамтран бүтээв, 2026.

package middlewares

import (
	"context"
	"net/http"
	"strings"

	"template/internal/constants"
	"template/internal/datasources/rls"
	V1Handler "template/internal/http/handlers/v1"
	"template/pkg/hydra"
	"template/pkg/jwt"
	"template/pkg/logger"
)

// OAuthBearerMiddleware нь бүртгэгдсэн апп (relying party)-ийн Hydra access
// token-оор хамгаалдаг middleware — тухайн апп хэрэглэгчийнхээ өмнөөс энэ SSO-ий
// service-үүдийг ДАМЖУУЛАН (proxy) дуудна. Token-ыг Hydra introspection-оор
// (RFC 7662) шалгаж, active + subject-тэй эсэхийг баталгаажуулна. Token-ий `sub`
// нь энэ SSO-ий тогтвортой user ID (provider login-д тавьсан) тул downstream
// handler-ууд session-ий адил CurrentUserFromContext-оор хэрэглэгчийг олно.
//
// ЗӨВШӨӨРӨЛ: requiredServiceScope тавьсан бол апп нь тухайн gateway service-т
// ЗӨВШӨӨРӨГДСӨН байх ёстой. Зөвшөөрөл нь client-ийн Hydra allowed scope дахь
// service scope (жишээ "svc:eid-proxy")-ээр илэрхийлэгдэнэ — admin gateway UI-аас
// апп-д service олгоход энэ scope client-д нэмэгддэг (SetServices). Тиймээс token
// scope биш, харин client_id-ийн ОДООГИЙН олголтыг шалгана (олгох/цуцлах шууд
// хүчинтэй). Олгогдоогүй бол 403.
type OAuthBearerMiddleware struct {
	hydra                *hydra.Admin
	requiredServiceScope string // жишээ "svc:eid-proxy"; хоосон бол зөвхөн active token
}

// NewOAuthBearerMiddleware нь Hydra introspection + client олголт шалгах chi
// middleware буцаана. requiredServiceScope хоосон бол зөвхөн active token шаардана.
func NewOAuthBearerMiddleware(hydraAdmin *hydra.Admin, requiredServiceScope string) func(http.Handler) http.Handler {
	m := &OAuthBearerMiddleware{hydra: hydraAdmin, requiredServiceScope: strings.TrimSpace(requiredServiceScope)}
	return m.Handle
}

func (m *OAuthBearerMiddleware) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		token := bearerToken(r)
		if token == "" {
			_ = V1Handler.NewAbortResponse(w, r, "missing bearer token")
			return
		}

		intro, err := m.hydra.Introspect(ctx, token)
		if err != nil {
			logger.ErrorWithContext(ctx, "OAuth: token introspection unavailable", logger.Fields{
				"middleware": "OAuthBearerMiddleware", "error": err.Error(), "path": r.URL.Path,
			})
			_ = V1Handler.NewErrorResponse(w, r, http.StatusServiceUnavailable, "authorization temporarily unavailable")
			return
		}
		if !intro.Active {
			_ = V1Handler.NewAbortResponse(w, r, "token is inactive or expired")
			return
		}
		if intro.Sub == "" {
			// client_credentials зэрэг subject-гүй token — иргэний data-д хандаж болохгүй.
			_ = V1Handler.NewAbortResponse(w, r, "token has no subject")
			return
		}

		// ЗӨВШӨӨРӨЛ: апп нь энэ service-т олгогдсон эсэхийг client-ийн ОДООГИЙН
		// allowed scope-оор шалгана (admin gateway UI-аас олгосон/цуцалсан нь шууд
		// хүчинтэй). Token scope-д найдахгүй — token хуучин олголтоор олгогдсон
		// байж болзошгүй.
		if m.requiredServiceScope != "" {
			if intro.ClientID == "" {
				_ = V1Handler.NewAbortResponse(w, r, "token has no client")
				return
			}
			client, cerr := m.hydra.GetClient(ctx, intro.ClientID)
			if cerr != nil {
				logger.ErrorWithContext(ctx, "OAuth: client lookup unavailable", logger.Fields{
					"middleware": "OAuthBearerMiddleware", "client_id": intro.ClientID,
					"error": cerr.Error(), "path": r.URL.Path,
				})
				_ = V1Handler.NewErrorResponse(w, r, http.StatusServiceUnavailable, "authorization temporarily unavailable")
				return
			}
			if !scopeContains(client.Scope, m.requiredServiceScope) {
				logger.WarnWithContext(ctx, "OAuth: application not authorized for service", logger.Fields{
					"middleware": "OAuthBearerMiddleware", "client_id": intro.ClientID,
					"required_service_scope": m.requiredServiceScope, "path": r.URL.Path,
				})
				_ = V1Handler.NewErrorResponse(w, r, http.StatusForbidden, "application is not authorized for this service")
				return
			}
		}

		// Downstream handler-ууд session-ий адил ажиллахын тулд ижил context утга.
		claims := jwt.JwtCustomClaim{UserID: intro.Sub, Kind: "access"}
		ctx = context.WithValue(ctx, constants.CtxAuthenticatedUserKey, claims)
		ctx = rls.WithUser(ctx, intro.Sub)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// bearerToken нь Authorization header-ээс Bearer token-ыг гаргаж авна.
func bearerToken(r *http.Request) string {
	h := r.Header.Get("Authorization")
	if h == "" {
		return ""
	}
	const p = "bearer "
	if len(h) > len(p) && strings.EqualFold(h[:len(p)], p) {
		return strings.TrimSpace(h[len(p):])
	}
	return ""
}

// scopeContains нь space-separated scope мөрөнд target scope байгаа эсэхийг шалгана.
func scopeContains(scopes, target string) bool {
	for _, s := range strings.Fields(scopes) {
		if s == target {
			return true
		}
	}
	return false
}
