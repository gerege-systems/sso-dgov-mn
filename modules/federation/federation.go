/*
 * Gerege SSO
 * Copyright (c) 2026 Gerege Systems Development Team.
 * Distributed under the Apache 2.0 License.
 */

// Package federation is where an organisation says which outside identity
// providers it trusts.
//
// The core already federates: platform/ssoclient will hand a sign-in off to an
// upstream issuer, and every deployment that does so today names that issuer in
// SSO_CLIENT_ISSUER. That is one issuer, for the whole installation, changed by
// a redeploy. An SSO product needs the other shape — several, per organisation,
// changed by an administrator — and this module is where those live.
//
// # What this module does not do yet
//
// ponytail: the core reads SSO_CLIENT_ISSUER from the environment at start-up
// and knows nothing about these rows, so a provider registered here is not yet
// consulted during an actual sign-in. Closing that needs a hook in the core —
// a FederationSource on platform.Options that ssoclient consults instead of
// the environment — which is an upstream pull request, not something a module
// can reach in from outside. Until it lands, this is the register: what an
// organisation has declared it trusts, and who has arrived through each one.
//
// Registering it here first is deliberate rather than premature. The hook needs
// something to read, and a schema that has been used is a better argument for
// an interface than a proposal.
package federation

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/url"
	"strings"

	"github.com/gerege-systems/open-gerege-nexus/backend/pkg/nexus"
	"github.com/gerege-systems/sso-gerege-nexus/internal/secret"
	"github.com/go-chi/chi/v5"
)

// ID is the catalogue identifier.
const ID = "io.gerege.nexus.sso_federation"

type Module struct {
	store  *Store
	secret *secret.Box
}

// New builds the module and registers it in the compile-time app registry.
//
// A missing SSO_SECRET_KEY is logged and carried, not fatal: an installation
// that federates with nobody has no credential to protect, and refusing to
// start would make an unused feature everybody's problem. The refusal happens
// at the write that would otherwise store a secret in the clear.
func New(p nexus.Platform) *Module {
	box, err := secret.Open()
	if err != nil && !errors.Is(err, secret.ErrNoKey) {
		slog.Error("federation: the credential key was rejected; providers cannot be registered", "error", err)
	} else if err != nil {
		slog.Info("federation: SSO_SECRET_KEY is unset; providers cannot be registered on this installation")
	}
	m := &Module{store: NewStore(p.DB()), secret: box}
	nexus.Register(m)
	return m
}

func (m *Module) ID() string      { return ID }
func (m *Module) Name() string    { return "Federation" }
func (m *Module) Version() string { return "1.0.0" }

func (m *Module) Dependencies() []nexus.Dependency { return nil }

// Permissions splits looking from changing.
//
// Both are administrative — AdminOnly on the read as well, which is not the
// installer's default for a `.read`. What this reads is the list of every
// organisation this one has agreed to accept people from, and that is a map of
// its trust relationships rather than its own data.
func (m *Module) Permissions() []nexus.PermissionDefinition {
	return []nexus.PermissionDefinition{
		{Code: "sso_federation.read", Name: "See federated providers", AdminOnly: true,
			Description: "View the identity providers this organisation trusts"},
		{Code: "sso_federation.manage", Name: "Manage federated providers", AdminOnly: true,
			Description: "Register, edit, disable and remove trusted identity providers"},
	}
}

func (m *Module) Menus() []nexus.MenuDefinition {
	return []nexus.MenuDefinition{{
		ID: "sso_federation", Label: "Federation",
		Path: "/module/sso-federation", Icon: "network", Order: 10,
		Labels: map[string]string{
			"mn": "Холбоос нэвтрэлт", "ar": "الاتحاد", "zh": "身份联合",
			"fr": "Fédération", "ru": "Федерация", "es": "Federación",
		},
	}}
}

// MenuPermission and RoutePermissionPrefix are this module's half of
// nexus.AccessPolicy: the platform gates every route below on
// sso_federation.read for a GET and sso_federation.manage for anything else.
func (m *Module) MenuPermission() string        { return "sso_federation.read" }
func (m *Module) RoutePermissionPrefix() string { return "sso_federation" }

func (m *Module) RegisterRoutes(r chi.Router, gate func(http.Handler) http.Handler) {
	r.Route("/api/v1/sso/federation", func(fed chi.Router) {
		fed.Use(gate)
		fed.Get("/providers", m.handleList)
		fed.Post("/providers", m.handleCreate)
		fed.Put("/providers/{id}", m.handleUpdate)
		fed.Put("/providers/{id}/enabled", m.handleSetEnabled)
		fed.Delete("/providers/{id}", m.handleDelete)
		fed.Get("/providers/{id}/links", m.handleLinks)
	})
}

// Input is what a caller may set on a provider.
type Input struct {
	DisplayName  string            `json:"display_name"`
	Issuer       string            `json:"issuer"`
	ClientID     string            `json:"client_id"`
	ClientSecret string            `json:"client_secret"`
	Scopes       string            `json:"scopes"`
	AttributeMap map[string]string `json:"attribute_map"`
}

// validate rejects what cannot work before it reaches the database.
//
// The issuer check is the one that matters. An OIDC issuer is an https origin —
// discovery appends /.well-known/openid-configuration to it — so a value with a
// path, a query or a trailing slash produces a discovery URL that quietly 404s
// at the first sign-in, which is a long way from here.
func (in *Input) validate(requireSecret bool) error {
	in.DisplayName = strings.TrimSpace(in.DisplayName)
	in.Issuer = strings.TrimRight(strings.TrimSpace(in.Issuer), "/")
	in.ClientID = strings.TrimSpace(in.ClientID)
	in.Scopes = strings.TrimSpace(in.Scopes)
	if in.Scopes == "" {
		in.Scopes = "openid profile email"
	}
	if in.AttributeMap == nil {
		in.AttributeMap = map[string]string{}
	}

	if in.DisplayName == "" {
		return errors.New("display_name is required")
	}
	if in.ClientID == "" {
		return errors.New("client_id is required")
	}
	if requireSecret && in.ClientSecret == "" {
		return errors.New("client_secret is required")
	}
	issuer, err := url.Parse(in.Issuer)
	if err != nil || issuer.Host == "" {
		return errors.New("issuer must be a URL")
	}
	if issuer.Scheme != "https" {
		// http is allowed nowhere, including in development: a credential
		// travels over this connection, and an installation that learns the
		// habit locally keeps it in production.
		return errors.New("issuer must be https")
	}
	if issuer.Path != "" || issuer.RawQuery != "" || issuer.Fragment != "" {
		return errors.New("issuer must be an origin, with no path or query")
	}
	return nil
}

func decode(w http.ResponseWriter, r *http.Request, into any) bool {
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<16)).Decode(into); err != nil {
		nexus.Error(w, http.StatusBadRequest, "malformed request body")
		return false
	}
	return true
}

func (m *Module) handleList(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := nexus.RequireTenant(w, r)
	if !ok {
		return
	}
	providers, err := m.store.List(r.Context(), tenantID)
	if err != nil {
		slog.Error("federation: could not list providers", "error", err)
		nexus.Error(w, http.StatusInternalServerError, "could not load the providers")
		return
	}
	nexus.JSON(w, http.StatusOK, providers)
}

func (m *Module) handleCreate(w http.ResponseWriter, r *http.Request) {
	tenantID, claims, ok := actor(w, r)
	if !ok {
		return
	}
	var in Input
	if !decode(w, r, &in) {
		return
	}
	if err := in.validate(true); err != nil {
		nexus.Error(w, http.StatusBadRequest, err.Error())
		return
	}
	sealed, err := m.secret.Seal(in.ClientSecret)
	if err != nil {
		nexus.Error(w, http.StatusServiceUnavailable, err.Error())
		return
	}

	provider, err := m.store.Create(r.Context(), tenantID, claims.UserID, in, sealed)
	if err != nil {
		if isDuplicate(err) {
			nexus.Error(w, http.StatusConflict, "this organisation already trusts that issuer")
			return
		}
		slog.Error("federation: could not register a provider", "error", err)
		nexus.Error(w, http.StatusInternalServerError, "could not register the provider")
		return
	}
	nexus.Audit(r.Context(), tenantID, claims.UserID, "sso_federation.registered", "federation_provider",
		map[string]any{"provider_id": provider.ID, "issuer": provider.Issuer})
	nexus.JSON(w, http.StatusCreated, provider)
}

func (m *Module) handleUpdate(w http.ResponseWriter, r *http.Request) {
	tenantID, claims, ok := actor(w, r)
	if !ok {
		return
	}
	var in Input
	if !decode(w, r, &in) {
		return
	}
	// The secret is optional on an edit: a console that cannot read one back
	// cannot resubmit it, so an absent field means "leave it as it was".
	if err := in.validate(false); err != nil {
		nexus.Error(w, http.StatusBadRequest, err.Error())
		return
	}
	var sealed []byte
	if in.ClientSecret != "" {
		var err error
		if sealed, err = m.secret.Seal(in.ClientSecret); err != nil {
			nexus.Error(w, http.StatusServiceUnavailable, err.Error())
			return
		}
	}

	id := chi.URLParam(r, "id")
	provider, err := m.store.Update(r.Context(), tenantID, id, in, sealed)
	switch {
	case errors.Is(err, ErrNotFound):
		nexus.Error(w, http.StatusNotFound, "no such provider")
		return
	case isDuplicate(err):
		nexus.Error(w, http.StatusConflict, "this organisation already trusts that issuer")
		return
	case err != nil:
		slog.Error("federation: could not update a provider", "error", err, "provider_id", id)
		nexus.Error(w, http.StatusInternalServerError, "could not update the provider")
		return
	}
	nexus.Audit(r.Context(), tenantID, claims.UserID, "sso_federation.updated", "federation_provider",
		map[string]any{"provider_id": id, "secret_replaced": sealed != nil})
	nexus.JSON(w, http.StatusOK, provider)
}

// handleSetEnabled is how trust is withdrawn without being forgotten.
func (m *Module) handleSetEnabled(w http.ResponseWriter, r *http.Request) {
	tenantID, claims, ok := actor(w, r)
	if !ok {
		return
	}
	var body struct {
		Enabled *bool `json:"enabled"`
	}
	if !decode(w, r, &body) {
		return
	}
	if body.Enabled == nil {
		nexus.Error(w, http.StatusBadRequest, "enabled must be true or false")
		return
	}

	id := chi.URLParam(r, "id")
	if err := m.store.SetEnabled(r.Context(), tenantID, id, *body.Enabled); err != nil {
		if errors.Is(err, ErrNotFound) {
			nexus.Error(w, http.StatusNotFound, "no such provider")
			return
		}
		nexus.Error(w, http.StatusInternalServerError, "could not change the provider")
		return
	}
	nexus.Audit(r.Context(), tenantID, claims.UserID, "sso_federation.enabled_changed", "federation_provider",
		map[string]any{"provider_id": id, "enabled": *body.Enabled})
	nexus.JSON(w, http.StatusOK, map[string]any{"enabled": *body.Enabled})
}

func (m *Module) handleDelete(w http.ResponseWriter, r *http.Request) {
	tenantID, claims, ok := actor(w, r)
	if !ok {
		return
	}
	id := chi.URLParam(r, "id")
	if err := m.store.Delete(r.Context(), tenantID, id); err != nil {
		if errors.Is(err, ErrNotFound) {
			nexus.Error(w, http.StatusNotFound, "no such provider")
			return
		}
		nexus.Error(w, http.StatusInternalServerError, "could not remove the provider")
		return
	}
	nexus.Audit(r.Context(), tenantID, claims.UserID, "sso_federation.removed", "federation_provider",
		map[string]any{"provider_id": id})
	w.WriteHeader(http.StatusNoContent)
}

func (m *Module) handleLinks(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := nexus.RequireTenant(w, r)
	if !ok {
		return
	}
	links, err := m.store.Links(r.Context(), tenantID, chi.URLParam(r, "id"))
	if err != nil {
		nexus.Error(w, http.StatusInternalServerError, "could not load the links")
		return
	}
	nexus.JSON(w, http.StatusOK, links)
}

// actor resolves the two things every writing handler needs.
func actor(w http.ResponseWriter, r *http.Request) (string, nexus.UserClaims, bool) {
	tenantID, ok := nexus.RequireTenant(w, r)
	if !ok {
		return "", nexus.UserClaims{}, false
	}
	claims, err := nexus.UserFromContext(r.Context())
	if err != nil {
		nexus.Error(w, http.StatusUnauthorized, "unauthorized")
		return "", nexus.UserClaims{}, false
	}
	return tenantID, claims, true
}

// isDuplicate reports whether err is the unique constraint on (tenant, issuer).
//
// Matched on the SQLSTATE rather than the message, which is localised by the
// server's lc_messages and would stop matching the day somebody set it.
func isDuplicate(err error) bool {
	if err == nil {
		return false
	}
	var pgErr interface{ SQLState() string }
	return errors.As(err, &pgErr) && pgErr.SQLState() == "23505"
}
