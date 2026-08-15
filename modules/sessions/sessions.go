/*
 * Gerege SSO
 * Copyright (c) 2026 Gerege Systems Development Team.
 * Distributed under the Apache 2.0 License.
 */

// Package sessions is who is signed in right now, and the button that ends it.
//
// Single sign-on makes one session worth more than it used to be: it is not
// access to one application, it is access to every application the person was
// granted. The corollary is that ending it has to be possible from somewhere,
// by somebody, within seconds — a laptop left on a train is a support call, not
// a redeploy — and that is what this module is.
//
// Everything it does is recorded. Cutting a colleague off mid-sentence is an
// act somebody has to be able to be asked about afterwards, so a revocation
// carries who ordered it and, when it was more than one person's session, why.
package sessions

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/gerege-systems/open-gerege-nexus/backend/pkg/nexus"
	"github.com/go-chi/chi/v5"
)

// ID is the catalogue identifier.
const ID = "io.gerege.nexus.sso_sessions"

type Module struct{ store *Store }

func New(p nexus.Platform) *Module {
	m := &Module{store: NewStore(p.DB())}
	nexus.Register(m)
	return m
}

func (m *Module) ID() string      { return ID }
func (m *Module) Name() string    { return "Sessions" }
func (m *Module) Version() string { return "1.0.0" }

func (m *Module) Dependencies() []nexus.Dependency { return nil }

// Permissions keeps reading and revoking apart, and withholds both by default.
//
// Reading is administrative despite ending in `.read`: the list says who is
// working, from where, at what hour. That is a colleague's movements, and the
// installer's rule of thumb — a `.read` is safe for every member — is the wrong
// default for it.
func (m *Module) Permissions() []nexus.PermissionDefinition {
	return []nexus.PermissionDefinition{
		{Code: "sso_sessions.read", Name: "See active sessions", AdminOnly: true,
			Description: "View who is currently signed in and from where"},
		{Code: "sso_sessions.manage", Name: "End sessions", AdminOnly: true,
			Description: "Sign somebody out of this organisation remotely"},
	}
}

func (m *Module) Menus() []nexus.MenuDefinition {
	return []nexus.MenuDefinition{{
		ID: "sso_sessions", Label: "Sessions",
		Path: "/module/sso-sessions", Icon: "monitor-smartphone", Order: 20,
		Labels: map[string]string{
			"mn": "Сессүүд", "ar": "الجلسات", "zh": "会话",
			"fr": "Sessions", "ru": "Сеансы", "es": "Sesiones",
		},
	}}
}

func (m *Module) MenuPermission() string        { return "sso_sessions.read" }
func (m *Module) RoutePermissionPrefix() string { return "sso_sessions" }

func (m *Module) RegisterRoutes(r chi.Router, gate func(http.Handler) http.Handler) {
	r.Route("/api/v1/sso/sessions", func(ses chi.Router) {
		ses.Use(gate)
		ses.Get("/", m.handleList)
		ses.Get("/events", m.handleEvents)
		ses.Get("/users/{userID}", m.handleListForUser)
		ses.Delete("/users/{userID}", m.handleRevokeAllForUser)
		// Last, because chi matches in order and `/{id}` would otherwise
		// swallow `/events`.
		ses.Delete("/{id}", m.handleRevoke)
	})
}

func (m *Module) handleList(w http.ResponseWriter, r *http.Request) {
	m.list(w, r, "")
}

func (m *Module) handleListForUser(w http.ResponseWriter, r *http.Request) {
	m.list(w, r, chi.URLParam(r, "userID"))
}

func (m *Module) list(w http.ResponseWriter, r *http.Request, userID string) {
	tenantID, ok := nexus.RequireTenant(w, r)
	if !ok {
		return
	}
	found, err := m.store.List(r.Context(), tenantID, userID)
	if err != nil {
		slog.Error("sessions: could not list", "error", err)
		nexus.Error(w, http.StatusInternalServerError, "could not load the sessions")
		return
	}
	nexus.JSON(w, http.StatusOK, found)
}

func (m *Module) handleRevoke(w http.ResponseWriter, r *http.Request) {
	tenantID, claims, ok := actor(w, r)
	if !ok {
		return
	}
	sessionID := chi.URLParam(r, "id")
	reason := reasonFrom(w, r)

	if err := m.store.Revoke(r.Context(), tenantID, sessionID, claims.UserID, reason); err != nil {
		if errors.Is(err, ErrNotFound) {
			nexus.Error(w, http.StatusNotFound, "that session is not running")
			return
		}
		slog.Error("sessions: could not revoke", "error", err, "session_id", sessionID)
		nexus.Error(w, http.StatusInternalServerError, "could not end the session")
		return
	}
	nexus.Audit(r.Context(), tenantID, claims.UserID, "sso_sessions.revoked", "session",
		map[string]any{"session_id": sessionID, "reason": reason})
	w.WriteHeader(http.StatusNoContent)
}

// handleRevokeAllForUser signs one person out of everything.
//
// This one insists on a reason. Ending a single session is often housekeeping —
// a device that was replaced — but ending all of somebody's is an intervention,
// and the person it happened to is entitled to be told what it was about.
func (m *Module) handleRevokeAllForUser(w http.ResponseWriter, r *http.Request) {
	tenantID, claims, ok := actor(w, r)
	if !ok {
		return
	}
	reason := reasonFrom(w, r)
	if reason == "" {
		nexus.Error(w, http.StatusBadRequest,
			"say why: ending every session somebody holds needs a reason they can be told")
		return
	}

	userID := chi.URLParam(r, "userID")
	count, err := m.store.RevokeAllFor(r.Context(), tenantID, userID, claims.UserID, reason)
	if err != nil {
		slog.Error("sessions: could not revoke all", "error", err, "user_id", userID)
		nexus.Error(w, http.StatusInternalServerError, "could not end the sessions")
		return
	}
	nexus.Audit(r.Context(), tenantID, claims.UserID, "sso_sessions.revoked_all", "user",
		map[string]any{"user_id": userID, "count": count, "reason": reason})
	nexus.JSON(w, http.StatusOK, map[string]any{"revoked": count})
}

func (m *Module) handleEvents(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := nexus.RequireTenant(w, r)
	if !ok {
		return
	}
	limit := 100
	if raw := r.URL.Query().Get("limit"); raw != "" {
		if parsed, err := strconv.Atoi(raw); err == nil && parsed > 0 && parsed <= 500 {
			limit = parsed
		}
	}
	events, err := m.store.Events(r.Context(), tenantID, limit)
	if err != nil {
		nexus.Error(w, http.StatusInternalServerError, "could not load the history")
		return
	}
	nexus.JSON(w, http.StatusOK, events)
}

// reasonFrom reads the reason out of a body that is allowed to be absent.
//
// A DELETE with no body is a legitimate request, so a body that will not parse
// is read as no reason rather than as a bad request — the caller who sent one
// and got it wrong is told by the handler that requires it.
func reasonFrom(w http.ResponseWriter, r *http.Request) string {
	var body struct {
		Reason string `json:"reason"`
	}
	_ = json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<14)).Decode(&body)
	return strings.TrimSpace(body.Reason)
}

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
