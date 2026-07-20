// Government Template Platform V3.0
// Gerege Systems Development Team болон Claude AI хамтран бүтээв, 2026.

// Package oidc нь өөрийн OAuth2/OIDC provider-ийн НИЙТИЙН endpoint-уудыг
// үйлчилнэ (`/oauth2/*`, `/userinfo`, `/.well-known/*`).
//
// АНХААР: эдгээр нь OAuth2/OIDC-ийн стандарт гэрээ тул платформын ердийн
// `v1.BaseResponse` бүрхүүлийг ХЭРЭГЛЭХГҮЙ — RP-ийн сангууд задлахгүй. Хариу нь
// RFC-ийн заасан JSON биетэй, алдаа нь RFC 6749 §5.2-ийн `{"error": ...}` хэлбэртэй.
package oidc

import (
	"encoding/json"
	"net/http"

	oidcuc "template/internal/business/usecases/oidc"
	"template/pkg/logger"
)

type Handler struct {
	keys   *oidcuc.KeyManager
	issuer string
}

func NewHandler(keys *oidcuc.KeyManager, issuer string) Handler {
	return Handler{keys: keys, issuer: issuer}
}

// Discovery godoc
// @Summary      OpenID Connect discovery баримт
// @Tags         oidc
// @Produce      json
// @Success      200  {object}  map[string]any
// @Router       /.well-known/openid-configuration [get]
func (h Handler) Discovery(w http.ResponseWriter, r *http.Request) {
	// Discovery нь ховор өөрчлөгддөг ба RP-үүд кэшилдэг.
	w.Header().Set("Cache-Control", "public, max-age=3600")
	writeJSON(w, r, http.StatusOK, oidcuc.BuildDiscovery(h.issuer))
}

// JWKS godoc
// @Summary      id_token шалгах нийтийн түлхүүрүүд (JWK Set)
// @Tags         oidc
// @Produce      json
// @Success      200  {object}  map[string]any
// @Router       /.well-known/jwks.json [get]
func (h Handler) JWKS(w http.ResponseWriter, r *http.Request) {
	set, err := h.keys.JWKS(r.Context())
	if err != nil {
		logger.ErrorWithContext(r.Context(), "OIDC: JWKS-ийг уншиж чадсангүй", logger.Fields{"error": err.Error()})
		writeError(w, r, http.StatusInternalServerError, "server_error", "could not load signing keys")
		return
	}
	// Түлхүүр эргэлт нь шинэ kid авчирдаг тул RP-үүд удаан кэшлэх ёсгүй.
	w.Header().Set("Cache-Control", "public, max-age=300")
	writeJSON(w, r, http.StatusOK, set)
}

func writeJSON(w http.ResponseWriter, r *http.Request, status int, body any) {
	w.Header().Set("Content-Type", "application/json;charset=UTF-8")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(body); err != nil {
		logger.ErrorWithContext(r.Context(), "OIDC: хариу бичихэд алдаа", logger.Fields{"error": err.Error()})
	}
}

// writeError нь RFC 6749 §5.2-ийн алдааны биетийг буцаана.
func writeError(w http.ResponseWriter, r *http.Request, status int, code, description string) {
	writeJSON(w, r, status, map[string]string{
		"error":             code,
		"error_description": description,
	})
}
