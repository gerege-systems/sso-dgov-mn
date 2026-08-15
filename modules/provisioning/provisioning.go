/*
 * Gerege SSO
 * Copyright (c) 2026 Gerege Systems Development Team.
 * Distributed under the Apache 2.0 License.
 */

// Package provisioning pushes this platform's people into the systems that
// need to know about them, over SCIM 2.0.
//
// Single sign-on answers "can this person get in". It does not answer "does
// this person have an account here at all", and a surprising number of systems
// need the second before the first is any use — a mailbox, a shared drive, a
// line-of-business application that will authenticate anybody it already has a
// row for and nobody it does not. Provisioning is that row.
//
// Deliberately one-directional. Reading users back out of a target and merging
// them is a different product with a conflict-resolution problem in the middle;
// this one has a source of truth and sends it outward.
package provisioning

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gerege-systems/open-gerege-nexus/backend/pkg/nexus"
	"github.com/gerege-systems/sso-gerege-nexus/internal/secret"
	"github.com/go-chi/chi/v5"
)

// ID is the catalogue identifier.
const ID = "io.gerege.nexus.sso_provisioning"

type Module struct {
	store  *Store
	secret *secret.Box
}

// New builds the module, registers it, and starts the queue worker.
//
// The worker runs for the life of the process. platform.Options hands a module
// no shutdown hook, so it ends when the process does — which is correct for
// what it is: a loop whose every iteration is a transaction the database
// releases on disconnect.
func New(p nexus.Platform) *Module {
	box, err := secret.Open()
	if err != nil && !errors.Is(err, secret.ErrNoKey) {
		slog.Error("provisioning: the credential key was rejected; targets cannot be registered", "error", err)
	} else if err != nil {
		slog.Info("provisioning: SSO_SECRET_KEY is unset; targets cannot be registered on this installation")
	}

	m := &Module{store: NewStore(p.DB()), secret: box}
	// No database, no worker. That is the shape a test builds — nexus.NewPlatform(nil, nil)
	// to ask a module what it declares — and a loop querying a nil pool would
	// turn every such test into a panic thirty seconds after it passed.
	if p.DB() != nil {
		w := &worker{
			store:  m.store,
			reveal: box.Reveal,
			// A timeout, because the other end is somebody else's server and a
			// request with none is a goroutine that never comes back.
			client: &http.Client{Timeout: 20 * time.Second},
		}
		go w.run(context.Background())
	}

	nexus.Register(m)
	return m
}

func (m *Module) ID() string      { return ID }
func (m *Module) Name() string    { return "Provisioning" }
func (m *Module) Version() string { return "1.0.0" }

func (m *Module) Dependencies() []nexus.Dependency { return nil }

func (m *Module) Permissions() []nexus.PermissionDefinition {
	return []nexus.PermissionDefinition{
		{Code: "sso_provisioning.read", Name: "See provisioning targets", AdminOnly: true,
			Description: "View the systems this organisation pushes users into, and what was sent"},
		{Code: "sso_provisioning.manage", Name: "Manage provisioning", AdminOnly: true,
			Description: "Register targets, test them, and start a resynchronisation"},
	}
}

func (m *Module) Menus() []nexus.MenuDefinition {
	return []nexus.MenuDefinition{{
		ID: "sso_provisioning", Label: "Provisioning",
		Path: "/module/sso-provisioning", Icon: "refresh-cw", Order: 40,
		Labels: map[string]string{
			"mn": "Хэрэглэгч нийлүүлэлт", "ar": "التزويد", "zh": "用户预配",
			"fr": "Provisionnement", "ru": "Подготовка учётных записей",
			"es": "Aprovisionamiento",
		},
	}}
}

func (m *Module) MenuPermission() string        { return "sso_provisioning.read" }
func (m *Module) RoutePermissionPrefix() string { return "sso_provisioning" }

func (m *Module) RegisterRoutes(r chi.Router, gate func(http.Handler) http.Handler) {
	r.Route("/api/v1/sso/scim", func(scim chi.Router) {
		scim.Use(gate)
		scim.Get("/targets", m.handleList)
		scim.Post("/targets", m.handleCreate)
		scim.Put("/targets/{id}", m.handleUpdate)
		scim.Delete("/targets/{id}", m.handleDelete)
		scim.Post("/targets/{id}/test", m.handleTest)
		scim.Post("/targets/{id}/resync", m.handleResync)
		scim.Get("/runs", m.handleRuns)
	})
}

// Input is what a caller may set on a target.
type Input struct {
	Name    string `json:"name"`
	BaseURL string `json:"base_url"`
	Token   string `json:"token"`
}

func (in *Input) validate(requireToken bool) error {
	in.Name = strings.TrimSpace(in.Name)
	if in.Name == "" {
		return errors.New("name is required")
	}
	if requireToken && in.Token == "" {
		return errors.New("token is required")
	}
	cleaned, err := validBaseURL(in.BaseURL)
	if err != nil {
		return err
	}
	in.BaseURL = cleaned
	return nil
}

func (m *Module) handleList(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := nexus.RequireTenant(w, r)
	if !ok {
		return
	}
	targets, err := m.store.List(r.Context(), tenantID)
	if err != nil {
		slog.Error("provisioning: could not list targets", "error", err)
		nexus.Error(w, http.StatusInternalServerError, "could not load the targets")
		return
	}
	nexus.JSON(w, http.StatusOK, targets)
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
	sealed, err := m.secret.Seal(in.Token)
	if err != nil {
		nexus.Error(w, http.StatusServiceUnavailable, err.Error())
		return
	}

	target, err := m.store.Create(r.Context(), tenantID, claims.UserID, in, sealed)
	if err != nil {
		if isDuplicate(err) {
			nexus.Error(w, http.StatusConflict, "this organisation already has a target by that name")
			return
		}
		slog.Error("provisioning: could not register a target", "error", err)
		nexus.Error(w, http.StatusInternalServerError, "could not register the target")
		return
	}
	nexus.Audit(r.Context(), tenantID, claims.UserID, "sso_provisioning.target_registered", "scim_target",
		map[string]any{"target_id": target.ID, "base_url": target.BaseURL})
	nexus.JSON(w, http.StatusCreated, target)
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
	// The token is optional on an edit, for the same reason a federation
	// secret is: a console that cannot read one back cannot resubmit it.
	if err := in.validate(false); err != nil {
		nexus.Error(w, http.StatusBadRequest, err.Error())
		return
	}
	var sealed []byte
	if in.Token != "" {
		var err error
		if sealed, err = m.secret.Seal(in.Token); err != nil {
			nexus.Error(w, http.StatusServiceUnavailable, err.Error())
			return
		}
	}

	id := chi.URLParam(r, "id")
	target, err := m.store.Update(r.Context(), tenantID, id, in, sealed)
	switch {
	case errors.Is(err, ErrNotFound):
		nexus.Error(w, http.StatusNotFound, "no such target")
		return
	case isDuplicate(err):
		nexus.Error(w, http.StatusConflict, "this organisation already has a target by that name")
		return
	case err != nil:
		nexus.Error(w, http.StatusInternalServerError, "could not update the target")
		return
	}
	nexus.Audit(r.Context(), tenantID, claims.UserID, "sso_provisioning.target_updated", "scim_target",
		map[string]any{"target_id": id, "token_replaced": sealed != nil})
	nexus.JSON(w, http.StatusOK, target)
}

func (m *Module) handleDelete(w http.ResponseWriter, r *http.Request) {
	tenantID, claims, ok := actor(w, r)
	if !ok {
		return
	}
	id := chi.URLParam(r, "id")
	if err := m.store.Delete(r.Context(), tenantID, id); err != nil {
		if errors.Is(err, ErrNotFound) {
			nexus.Error(w, http.StatusNotFound, "no such target")
			return
		}
		nexus.Error(w, http.StatusInternalServerError, "could not remove the target")
		return
	}
	nexus.Audit(r.Context(), tenantID, claims.UserID, "sso_provisioning.target_removed", "scim_target",
		map[string]any{"target_id": id})
	w.WriteHeader(http.StatusNoContent)
}

// handleTest asks the target for its own description of itself.
//
// /ServiceProviderConfig is the one endpoint every SCIM 2.0 server has and the
// only one that can be called without naming a user, so it answers the two
// questions an operator has — is the address right, is the token accepted —
// without creating anything.
func (m *Module) handleTest(w http.ResponseWriter, r *http.Request) {
	tenantID, claims, ok := actor(w, r)
	if !ok {
		return
	}
	id := chi.URLParam(r, "id")
	baseURL, sealed, err := m.store.Credentials(r.Context(), tenantID, id)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			nexus.Error(w, http.StatusNotFound, "no such target")
			return
		}
		nexus.Error(w, http.StatusInternalServerError, "could not read the target")
		return
	}
	token, err := m.secret.Reveal(sealed)
	if err != nil {
		nexus.Error(w, http.StatusServiceUnavailable, err.Error())
		return
	}

	probe := &worker{client: &http.Client{Timeout: 10 * time.Second}}
	status, body, err := probe.request(r.Context(), http.MethodGet,
		baseURL+"/ServiceProviderConfig", token, nil)
	if err != nil {
		nexus.JSON(w, http.StatusOK, map[string]any{"reachable": false, "error": err.Error()})
		return
	}
	nexus.Audit(r.Context(), tenantID, claims.UserID, "sso_provisioning.target_tested", "scim_target",
		map[string]any{"target_id": id, "status": status})
	nexus.JSON(w, http.StatusOK, map[string]any{
		"reachable": status >= 200 && status < 300,
		"status":    status,
		"response":  truncate(body),
	})
}

// handleResync puts every member on the queue.
func (m *Module) handleResync(w http.ResponseWriter, r *http.Request) {
	tenantID, claims, ok := actor(w, r)
	if !ok {
		return
	}
	id := chi.URLParam(r, "id")
	if _, err := m.store.Get(r.Context(), tenantID, id); err != nil {
		if errors.Is(err, ErrNotFound) {
			nexus.Error(w, http.StatusNotFound, "no such target")
			return
		}
		nexus.Error(w, http.StatusInternalServerError, "could not read the target")
		return
	}

	queued, err := m.store.EnqueueAll(r.Context(), tenantID, id)
	if err != nil {
		slog.Error("provisioning: could not enqueue a resync", "error", err, "target_id", id)
		nexus.Error(w, http.StatusInternalServerError, "could not start the resynchronisation")
		return
	}
	nexus.Audit(r.Context(), tenantID, claims.UserID, "sso_provisioning.resync", "scim_target",
		map[string]any{"target_id": id, "queued": queued})
	nexus.JSON(w, http.StatusAccepted, map[string]any{"queued": queued})
}

func (m *Module) handleRuns(w http.ResponseWriter, r *http.Request) {
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
	runs, err := m.store.Runs(r.Context(), tenantID, limit)
	if err != nil {
		nexus.Error(w, http.StatusInternalServerError, "could not load the history")
		return
	}
	nexus.JSON(w, http.StatusOK, runs)
}

func decode(w http.ResponseWriter, r *http.Request, into any) bool {
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<16)).Decode(into); err != nil {
		nexus.Error(w, http.StatusBadRequest, "malformed request body")
		return false
	}
	return true
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

// isDuplicate reports whether err is a unique-constraint violation, matched on
// the SQLSTATE rather than the message — which is localised by the server's
// lc_messages and would stop matching the day somebody set it.
func isDuplicate(err error) bool {
	if err == nil {
		return false
	}
	var pgErr interface{ SQLState() string }
	return errors.As(err, &pgErr) && pgErr.SQLState() == "23505"
}
