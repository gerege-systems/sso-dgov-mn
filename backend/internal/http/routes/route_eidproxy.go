// Government Template Platform V3.0
// Gerege Systems Development Team болон Claude AI хамтран бүтээв, 2026.

package routes

import (
	"net/http"

	authuc "template/internal/business/usecases/auth"
	v1 "template/internal/http/handlers/v1"
	eidprofilehandler "template/internal/http/handlers/v1/eidprofile"

	"github.com/go-chi/chi/v5"
)

// eidProxyRoute нь бүртгэгдсэн апп (relying party)-уудад ЗОРИУЛСАН eID service
// proxy-г холбоно (/v1/eid/*). Апп нь хэрэглэгчийнхээ Hydra access token-оор
// (Authorization: Bearer, "eid" scope) дуудна; SSO нь token-ий subject-ээр
// хэрэглэгчийг тогтоож, ӨӨРИЙН eidmongolia.mn RP creds-ээр өгөгдлийг татаж өгнө.
// Апп-д eID RP credential эзэмших шаардлагагүй. ЗӨВХӨН унших (mutation байхгүй).
type eidProxyRoute struct {
	handler         eidprofilehandler.Handler
	router          chi.Router
	oauthMiddleware func(http.Handler) http.Handler
}

func NewEIDProxyRoute(router chi.Router, authUC authuc.Usecase, oauthMiddleware func(http.Handler) http.Handler) *eidProxyRoute {
	return &eidProxyRoute{
		handler:         eidprofilehandler.NewHandler(authUC),
		router:          router,
		oauthMiddleware: oauthMiddleware,
	}
}

func (rt *eidProxyRoute) Routes() {
	rt.router.Route("/v1/eid", func(r chi.Router) {
		r.Use(rt.oauthMiddleware)
		r.Get("/summary", v1.Wrap(rt.handler.Summary))
		r.Get("/certificates", v1.Wrap(rt.handler.Certificates))
		r.Get("/devices", v1.Wrap(rt.handler.Devices))
		r.Get("/activity", v1.Wrap(rt.handler.Activity))
		r.Get("/organizations", v1.Wrap(rt.handler.Organizations))
		r.Get("/organizations/{regNo}/signers", v1.Wrap(rt.handler.OrgSigners))
	})
}
