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
	"errors"
	"net/http"
	"net/url"
	"strings"

	oidcuc "template/internal/business/usecases/oidc"
	"template/pkg/logger"
)

type Handler struct {
	keys   *oidcuc.KeyManager
	svc    *oidcuc.Service
	issuer string
}

func NewHandler(keys *oidcuc.KeyManager, svc *oidcuc.Service, issuer string) Handler {
	return Handler{keys: keys, svc: svc, issuer: strings.TrimRight(issuer, "/")}
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

// Authorize godoc
// @Summary      OAuth2 authorization endpoint
// @Tags         oidc
// @Param        client_id      query  string  true   "Client ID"
// @Param        redirect_uri   query  string  true   "Registered redirect URI"
// @Param        response_type  query  string  true   "code"
// @Param        scope          query  string  false  "Space-separated scopes"
// @Param        state          query  string  false  "Opaque RP state"
// @Success      302
// @Router       /oauth2/auth [get]
func (h Handler) Authorize(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	req := oidcuc.AuthorizeRequest{
		ClientID:            q.Get("client_id"),
		RedirectURI:         q.Get("redirect_uri"),
		ResponseType:        q.Get("response_type"),
		Scope:               q.Get("scope"),
		State:               q.Get("state"),
		Nonce:               q.Get("nonce"),
		CodeChallenge:       q.Get("code_challenge"),
		CodeChallengeMethod: q.Get("code_challenge_method"),
		Prompt:              q.Get("prompt"),
	}

	challenge, _, err := h.svc.Authorize(r.Context(), req)
	if err != nil {
		var authErr *oidcuc.AuthorizeError
		if errors.As(err, &authErr) {
			// Зөвхөн service-ийн БАТАЛГААЖУУЛСАН хаяг руу чиглүүлнэ. Хүсэлтээс
			// ирсэн түүхий redirect_uri-г энд огт ашиглахгүй — client эсвэл
			// redirect_uri буруу бол алдааг шууд харуулна.
			if !authErr.CanRedirect() {
				writeError(w, r, http.StatusBadRequest, authErr.Code, authErr.Description)
				return
			}
			http.Redirect(w, r, authErr.RedirectURL(), http.StatusFound)
			return
		}
		logger.ErrorWithContext(r.Context(), "OIDC: authorize амжилтгүй", logger.Fields{"error": err.Error()})
		writeError(w, r, http.StatusInternalServerError, "server_error", "could not start the authorization request")
		return
	}

	// Нэвтрэх хуудас руу. Session байвал тэр хуудас шууд accept руу шилжинэ.
	http.Redirect(w, r, h.issuer+"/oauth/login?login_challenge="+url.QueryEscape(challenge), http.StatusFound)
}

// Token godoc
// @Summary      OAuth2 token endpoint
// @Tags         oidc
// @Accept       x-www-form-urlencoded
// @Produce      json
// @Param        grant_type  formData  string  true  "authorization_code | refresh_token | client_credentials"
// @Success      200  {object}  map[string]any
// @Router       /oauth2/token [post]
func (h Handler) Token(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		writeError(w, r, http.StatusBadRequest, "invalid_request", "malformed request body")
		return
	}

	req := oidcuc.TokenRequest{
		GrantType:    r.PostFormValue("grant_type"),
		Code:         r.PostFormValue("code"),
		RedirectURI:  r.PostFormValue("redirect_uri"),
		CodeVerifier: r.PostFormValue("code_verifier"),
		RefreshToken: r.PostFormValue("refresh_token"),
		Scope:        r.PostFormValue("scope"),
		ClientID:     r.PostFormValue("client_id"),
		ClientSecret: r.PostFormValue("client_secret"),
	}
	// HTTP Basic нь биетээс давуу — хоёулаа ирвэл Basic-ыг авна (RFC 6749 §2.3.1).
	if id, secret, ok := basicClientAuth(r); ok {
		req.ClientID, req.ClientSecret, req.SecretFromBasic = id, secret, true
	}

	// Token нь ХЭЗЭЭ Ч кэшлэгдэх ёсгүй.
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Pragma", "no-cache")

	resp, err := h.svc.Token(r.Context(), req)
	if err != nil {
		var tokErr *oidcuc.TokenError
		if errors.As(err, &tokErr) {
			if tokErr.Code == "invalid_client" {
				// RFC 6749 §5.2 — Basic ашигласан бол WWW-Authenticate буцаана.
				w.Header().Set("WWW-Authenticate", `Basic realm="oauth2"`)
			}
			writeError(w, r, tokErr.Status, tokErr.Code, tokErr.Description)
			return
		}
		logger.ErrorWithContext(r.Context(), "OIDC: token гаргаж чадсангүй", logger.Fields{
			"error": err.Error(), "grant_type": req.GrantType,
		})
		writeError(w, r, http.StatusInternalServerError, "server_error", "could not issue a token")
		return
	}
	writeJSON(w, r, http.StatusOK, resp)
}

// basicClientAuth нь Authorization: Basic-аас client итгэмжлэлийг задална.
//
// RFC 6749 §2.3.1 нь client_id/secret-ыг base64-ийн ӨМНӨ form-urlencode хийхийг
// шаарддаг — тусгай тэмдэгттэй secret зөв ажиллахын тулд буцааж decode хийнэ.
func basicClientAuth(r *http.Request) (clientID, clientSecret string, ok bool) {
	id, secret, ok := r.BasicAuth()
	if !ok {
		return "", "", false
	}
	decodedID, err := url.QueryUnescape(id)
	if err != nil {
		decodedID = id
	}
	decodedSecret, err := url.QueryUnescape(secret)
	if err != nil {
		decodedSecret = secret
	}
	return decodedID, decodedSecret, true
}
