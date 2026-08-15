/*
 * Gerege SSO
 * Copyright (c) 2026 Gerege Systems Development Team.
 * Distributed under the Apache 2.0 License.
 */

// Package accessreview asks, on a schedule somebody sets, whether the access
// people hold is still the access they need.
//
// Access accumulates. Somebody covers a colleague's maternity leave, moves
// department, is added to a project that ended two years ago — and nothing in
// an ordinary permissions system ever asks whether any of it should still be
// true. The answer everywhere is the same: periodically freeze who holds what,
// put it in front of the person who would know, and record what they said.
//
// # What a decision does, and does not do
//
// ponytail: a revoked item is recorded, not enforced. Removing the access is a
// change to RBAC, which is the core's, and a module reaching into another
// repository's tables to take permissions away is the kind of write that is
// discovered during an incident. What this produces is the attestation record —
// which is what an audit asks for — and a list an administrator acts on.
// Enforcing it needs a RoleStore on pkg/nexus, upstream.
package accessreview

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/gerege-systems/open-gerege-nexus/backend/pkg/nexus"
	"github.com/go-chi/chi/v5"
)

// ID is the catalogue identifier.
const ID = "io.gerege.nexus.sso_access_review"

type Module struct {
	store *Store
	// permissions is kept from the constructor because RegisterRoutes is not
	// handed one, and this module gates three route groups on three different
	// permissions rather than on the platform's read/manage rule.
	permissions nexus.PermissionStore
}

func New(p nexus.Platform) *Module {
	m := &Module{store: NewStore(p.DB()), permissions: p.Permissions()}
	nexus.Register(m)
	return m
}

func (m *Module) ID() string      { return ID }
func (m *Module) Name() string    { return "Access Review" }
func (m *Module) Version() string { return "1.0.0" }

func (m *Module) Dependencies() []nexus.Dependency { return nil }

// Permissions separates running a campaign from answering one.
//
// Deliberately not a read/manage pair, because the interesting split here is
// not by verb. A reviewer decides on items and must not be able to open or
// close the campaign their decisions are recorded in — that is the same reason
// the App Store keeps submitting and publishing apart.
func (m *Module) Permissions() []nexus.PermissionDefinition {
	return []nexus.PermissionDefinition{
		{Code: "sso_access_review.read", Name: "See access reviews", AdminOnly: true,
			Description: "View campaigns and the access they cover"},
		{Code: "sso_access_review.decide", Name: "Decide on access", AdminOnly: true,
			Description: "Keep or revoke an access a campaign asks about"},
		{Code: "sso_access_review.manage", Name: "Run access reviews", AdminOnly: true,
			Description: "Create, open and close review campaigns"},
	}
}

func (m *Module) Menus() []nexus.MenuDefinition {
	return []nexus.MenuDefinition{{
		ID: "sso_access_review", Label: "Access Review",
		Path: "/module/sso-access-review/campaigns", Icon: "clipboard-check", Order: 30,
		Labels: map[string]string{
			"mn": "Эрхийн хяналт", "ar": "مراجعة الصلاحيات", "zh": "权限复核",
			"fr": "Revue des accès", "ru": "Пересмотр доступа", "es": "Revisión de accesos",
		},
	}}
}

func (m *Module) MenuPermission() string { return "sso_access_review.read" }

// RoutePermissionPrefix is empty on purpose.
//
// The platform's rule — read for a GET, manage for anything else — cannot
// express this module's split, where deciding on an item and closing the
// campaign it belongs to are both POSTs held by different people. So the gate
// below is the installation check, and each writing route names its own
// permission.
func (m *Module) RoutePermissionPrefix() string { return "" }

func (m *Module) RegisterRoutes(r chi.Router, gate func(http.Handler) http.Handler) {
	r.Route("/api/v1/sso/reviews", func(rev chi.Router) {
		rev.Use(gate)

		rev.Group(func(read chi.Router) {
			read.Use(nexus.RequirePermission(m.permissions, "sso_access_review.read"))
			read.Get("/campaigns", m.handleListCampaigns)
			read.Get("/campaigns/{id}", m.handleGetCampaign)
			read.Get("/campaigns/{id}/items", m.handleItems)
		})
		rev.Group(func(manage chi.Router) {
			manage.Use(nexus.RequirePermission(m.permissions, "sso_access_review.manage"))
			manage.Post("/campaigns", m.handleCreate)
			manage.Post("/campaigns/{id}/open", m.handleOpen)
			manage.Post("/campaigns/{id}/close", m.handleClose)
		})
		rev.Group(func(decide chi.Router) {
			decide.Use(nexus.RequirePermission(m.permissions, "sso_access_review.decide"))
			decide.Put("/items/{id}", m.handleDecide)
		})
	})
}

// Input is what a caller may set on a campaign.
type Input struct {
	Name     string `json:"name"`
	Scope    string `json:"scope"`
	ScopeRef string `json:"scope_ref"`
	DueDate  string `json:"due_date"`
}

func (in *Input) validate() error {
	in.Name = strings.TrimSpace(in.Name)
	in.Scope = strings.TrimSpace(in.Scope)
	in.ScopeRef = strings.TrimSpace(in.ScopeRef)
	in.DueDate = strings.TrimSpace(in.DueDate)

	if in.Name == "" {
		return errors.New("name is required")
	}
	if in.Scope == "" {
		in.Scope = "all"
	}
	switch in.Scope {
	case "all":
		in.ScopeRef = ""
	case "app", "role":
		if in.ScopeRef == "" {
			return errors.New("scope_ref is required for an app or role scope")
		}
	default:
		return errors.New(`scope must be "all", "app" or "role"`)
	}
	if in.DueDate != "" {
		if _, err := time.Parse("2006-01-02", in.DueDate); err != nil {
			return errors.New("due_date must be YYYY-MM-DD")
		}
	}
	return nil
}

func (m *Module) handleListCampaigns(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := nexus.RequireTenant(w, r)
	if !ok {
		return
	}
	campaigns, err := m.store.ListCampaigns(r.Context(), tenantID)
	if err != nil {
		slog.Error("access review: could not list campaigns", "error", err)
		nexus.Error(w, http.StatusInternalServerError, "could not load the campaigns")
		return
	}
	nexus.JSON(w, http.StatusOK, campaigns)
}

func (m *Module) handleGetCampaign(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := nexus.RequireTenant(w, r)
	if !ok {
		return
	}
	campaign, err := m.store.GetCampaign(r.Context(), tenantID, chi.URLParam(r, "id"))
	if errors.Is(err, ErrNotFound) {
		nexus.Error(w, http.StatusNotFound, "no such campaign")
		return
	}
	if err != nil {
		nexus.Error(w, http.StatusInternalServerError, "could not load the campaign")
		return
	}
	nexus.JSON(w, http.StatusOK, campaign)
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
	if err := in.validate(); err != nil {
		nexus.Error(w, http.StatusBadRequest, err.Error())
		return
	}
	campaign, err := m.store.CreateCampaign(r.Context(), tenantID, claims.UserID, in)
	if err != nil {
		slog.Error("access review: could not create a campaign", "error", err)
		nexus.Error(w, http.StatusInternalServerError, "could not create the campaign")
		return
	}
	nexus.Audit(r.Context(), tenantID, claims.UserID, "sso_access_review.created", "review_campaign",
		map[string]any{"campaign_id": campaign.ID, "scope": campaign.Scope})
	nexus.JSON(w, http.StatusCreated, campaign)
}

// handleOpen is the moment the campaign stops being a plan.
func (m *Module) handleOpen(w http.ResponseWriter, r *http.Request) {
	tenantID, claims, ok := actor(w, r)
	if !ok {
		return
	}
	id := chi.URLParam(r, "id")
	campaign, err := m.store.Open(r.Context(), tenantID, id)
	if errors.Is(err, ErrNotDraft) {
		nexus.Error(w, http.StatusConflict, "only a draft campaign can be opened")
		return
	}
	if err != nil {
		slog.Error("access review: could not open a campaign", "error", err, "campaign_id", id)
		nexus.Error(w, http.StatusInternalServerError, "could not open the campaign")
		return
	}
	nexus.Audit(r.Context(), tenantID, claims.UserID, "sso_access_review.opened", "review_campaign",
		map[string]any{"campaign_id": id, "items": campaign.Total})
	nexus.JSON(w, http.StatusOK, campaign)
}

func (m *Module) handleClose(w http.ResponseWriter, r *http.Request) {
	tenantID, claims, ok := actor(w, r)
	if !ok {
		return
	}
	id := chi.URLParam(r, "id")
	campaign, err := m.store.Close(r.Context(), tenantID, id)
	if errors.Is(err, ErrNotOpen) {
		nexus.Error(w, http.StatusConflict, "only an open campaign can be closed")
		return
	}
	if err != nil {
		nexus.Error(w, http.StatusInternalServerError, "could not close the campaign")
		return
	}
	nexus.Audit(r.Context(), tenantID, claims.UserID, "sso_access_review.closed", "review_campaign",
		map[string]any{"campaign_id": id, "pending": campaign.Pending, "revoked": campaign.Revoked})
	nexus.JSON(w, http.StatusOK, campaign)
}

func (m *Module) handleItems(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := nexus.RequireTenant(w, r)
	if !ok {
		return
	}
	status := r.URL.Query().Get("status")
	switch status {
	case "", "pending", "kept", "revoked":
	default:
		nexus.Error(w, http.StatusBadRequest, `status must be "pending", "kept" or "revoked"`)
		return
	}
	items, err := m.store.Items(r.Context(), tenantID, chi.URLParam(r, "id"), status)
	if err != nil {
		nexus.Error(w, http.StatusInternalServerError, "could not load the items")
		return
	}
	nexus.JSON(w, http.StatusOK, items)
}

// handleDecide records one answer.
//
// Revoking needs a note; keeping does not. "Why did you take this away" is a
// question the person it was taken from will ask, and an answer written at the
// moment is the only one that will still be accurate in six months.
func (m *Module) handleDecide(w http.ResponseWriter, r *http.Request) {
	tenantID, claims, ok := actor(w, r)
	if !ok {
		return
	}
	var body struct {
		Decision string `json:"decision"`
		Note     string `json:"note"`
	}
	if !decode(w, r, &body) {
		return
	}
	body.Note = strings.TrimSpace(body.Note)
	switch body.Decision {
	case "kept":
	case "revoked":
		if body.Note == "" {
			nexus.Error(w, http.StatusBadRequest,
				"say why: a revoked access needs a reason the person can be told")
			return
		}
	default:
		nexus.Error(w, http.StatusBadRequest, `decision must be "kept" or "revoked"`)
		return
	}

	itemID := chi.URLParam(r, "id")
	item, err := m.store.Decide(r.Context(), tenantID, itemID, body.Decision, claims.UserID, body.Note)
	if errors.Is(err, ErrNotOpen) {
		nexus.Error(w, http.StatusConflict, "that item is not in an open campaign")
		return
	}
	if err != nil {
		slog.Error("access review: could not record a decision", "error", err, "item_id", itemID)
		nexus.Error(w, http.StatusInternalServerError, "could not record the decision")
		return
	}
	nexus.Audit(r.Context(), tenantID, claims.UserID, "sso_access_review."+body.Decision, "review_item",
		map[string]any{"item_id": itemID, "permission": item.PermissionCode,
			"subject_user_id": item.UserID, "note": body.Note})
	nexus.JSON(w, http.StatusOK, item)
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
