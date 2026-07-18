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
// eID service-үүдийг ДАМЖУУЛАН (proxy) дуудна. Token-ыг Hydra introspection-оор
// (RFC 7662) шалгаж, active + шаардлагатай scope-той эсэхийг баталгаажуулна.
// Token-ий `sub` нь энэ SSO-ий тогтвортой user ID (provider login-д тавьсан) тул
// downstream handler-ууд (eidprofile) session-ий адил CurrentUserFromContext-оор
// хэрэглэгчийг олж, sso-ий eID RP creds-ээр eidmongolia.mn-аас өгөгдлийг татна.
type OAuthBearerMiddleware struct {
	hydra         *hydra.Admin
	requiredScope string // жишээ "eid" — token-д энэ scope байх ёстой (хоосон бол scope шалгахгүй)
}

// NewOAuthBearerMiddleware нь Hydra admin introspection-д тулгуурласан chi
// middleware буцаана. requiredScope хоосон бол зөвхөн active token шаардана.
func NewOAuthBearerMiddleware(hydraAdmin *hydra.Admin, requiredScope string) func(http.Handler) http.Handler {
	m := &OAuthBearerMiddleware{hydra: hydraAdmin, requiredScope: strings.TrimSpace(requiredScope)}
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
			// Introspection endpoint-д хүрч чадсангүй — fail-closed, түр 503.
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
			// client_credentials зэрэг subject-гүй token — иргэний eID data-д хандаж болохгүй.
			_ = V1Handler.NewAbortResponse(w, r, "token has no subject")
			return
		}
		if m.requiredScope != "" && !scopeContains(intro.Scope, m.requiredScope) {
			logger.WarnWithContext(ctx, "OAuth: token missing required scope", logger.Fields{
				"middleware": "OAuthBearerMiddleware", "client_id": intro.ClientID,
				"required_scope": m.requiredScope, "token_scope": intro.Scope, "path": r.URL.Path,
			})
			_ = V1Handler.NewAbortResponse(w, r, "token missing required scope: "+m.requiredScope)
			return
		}

		// Downstream handler-ууд session-ий адил ажиллахын тулд ижил context
		// утгыг тавина: sub нь SSO-ий тогтвортой user ID (provider subject).
		claims := jwt.JwtCustomClaim{UserID: intro.Sub, Kind: "access"}
		ctx = context.WithValue(ctx, constants.CtxAuthenticatedUserKey, claims)
		// RLS: зөвхөн энэ хэрэглэгчийн мөр (админ эрх олгохгүй — гуравдагч апп).
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
