// Government Template Platform V3.0
// Gerege Systems Development Team болон Claude AI хамтран бүтээв, 2026.

package routes

import (
	"net/http"

	authuc "template/internal/business/usecases/auth"
	gatewayuc "template/internal/business/usecases/gateway"
	v1 "template/internal/http/handlers/v1"
	eidprofilehandler "template/internal/http/handlers/v1/eidprofile"

	"github.com/go-chi/chi/v5"
)

// EIDProxyServiceName нь API gateway catalog дахь энэ service-ийн нэр. Admin
// gateway UI-аас энэ нэрээр enable/disable хийхэд route бодитоор нөлөөлнө.
const EIDProxyServiceName = "eid-proxy"

// eidProxyRoute нь бүртгэгдсэн апп (relying party)-уудад ЗОРИУЛСАН eID service
// proxy-г холбоно (/v1/eid/*). Апп нь хэрэглэгчийнхээ Hydra access token-оор
// (Authorization: Bearer, "eid" scope) дуудна; SSO нь token-ий subject-ээр
// хэрэглэгчийг тогтоож, ӨӨРИЙН eidmongolia.mn RP creds-ээр өгөгдлийг татаж өгнө.
// Апп-д eID RP credential эзэмших шаардлагагүй. ЗӨВХӨН унших (mutation байхгүй).
// API gateway-ийн "eid-proxy" service-ээр admin enable/disable хийж болно.
type eidProxyRoute struct {
	handler         eidprofilehandler.Handler
	router          chi.Router
	gatewayUC       gatewayuc.Usecase
	oauthMiddleware func(http.Handler) http.Handler
}

func NewEIDProxyRoute(router chi.Router, authUC authuc.Usecase, gatewayUC gatewayuc.Usecase, oauthMiddleware func(http.Handler) http.Handler) *eidProxyRoute {
	return &eidProxyRoute{
		handler:         eidprofilehandler.NewHandler(authUC),
		router:          router,
		gatewayUC:       gatewayUC,
		oauthMiddleware: oauthMiddleware,
	}
}

// enabledGate нь API gateway catalog дахь "eid-proxy" service идэвхтэй эсэхийг
// шалгана — admin gateway UI-аас унтраасан бол 503. Catalog-д байхгүй бол
// fail-open (код-default идэвхтэй). OAuth шалгалтаас ӨМНӨ ажиллана.
func (rt *eidProxyRoute) enabledGate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if on, err := rt.gatewayUC.ServiceEnabled(r.Context(), EIDProxyServiceName); err == nil && !on {
			_ = v1.NewErrorResponse(w, r, http.StatusServiceUnavailable, "eID proxy service is disabled")
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (rt *eidProxyRoute) Routes() {
	rt.router.Route("/v1/eid", func(r chi.Router) {
		r.Use(rt.enabledGate)     // admin gateway toggle
		r.Use(rt.oauthMiddleware) // Bearer + "eid" scope
		r.Get("/summary", v1.Wrap(rt.handler.Summary))
		r.Get("/certificates", v1.Wrap(rt.handler.Certificates))
		r.Get("/devices", v1.Wrap(rt.handler.Devices))
		r.Get("/activity", v1.Wrap(rt.handler.Activity))
		r.Get("/organizations", v1.Wrap(rt.handler.Organizations))
		r.Get("/organizations/{regNo}/signers", v1.Wrap(rt.handler.OrgSigners))
	})
}
