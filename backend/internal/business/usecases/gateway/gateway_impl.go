// Government Template Platform V3.0
// Gerege Systems Development Team болон Claude AI хамтран бүтээв, 2026.

package gateway

import (
	"context"
	"encoding/json"
	"strings"

	"template/internal/apperror"
	"template/internal/business/domain"
	repointerface "template/internal/datasources/repositories/interface"
)

type usecase struct {
	repo repointerface.GatewayRepository
}

func NewUsecase(repo repointerface.GatewayRepository) Usecase {
	return &usecase{repo: repo}
}

// ── Services ────────────────────────────────────────────────────────────────

func (u *usecase) ListServices(ctx context.Context) ([]domain.GatewayService, error) {
	return u.repo.ListServices(ctx)
}

func (u *usecase) CreateService(ctx context.Context, in ServiceInput) (domain.GatewayService, error) {
	svc, err := in.toDomain()
	if err != nil {
		return domain.GatewayService{}, err
	}
	return u.repo.CreateService(ctx, &svc)
}

func (u *usecase) UpdateService(ctx context.Context, id string, in ServiceInput) (domain.GatewayService, error) {
	svc, err := in.toDomain()
	if err != nil {
		return domain.GatewayService{}, err
	}
	svc.ID = id
	return u.repo.UpdateService(ctx, &svc)
}

func (u *usecase) DeleteService(ctx context.Context, id string) error {
	return u.repo.DeleteService(ctx, id)
}

func (in ServiceInput) toDomain() (domain.GatewayService, error) {
	name := strings.TrimSpace(in.Name)
	host := strings.TrimSpace(in.Host)
	if name == "" {
		return domain.GatewayService{}, apperror.BadRequest("service name is required")
	}
	if host == "" {
		return domain.GatewayService{}, apperror.BadRequest("service host is required")
	}
	protocol := strings.ToLower(strings.TrimSpace(in.Protocol))
	if protocol != "http" && protocol != "https" {
		protocol = "https"
	}
	port := in.Port
	if port <= 0 || port > 65535 {
		if protocol == "http" {
			port = 80
		} else {
			port = 443
		}
	}
	path := strings.TrimSpace(in.Path)
	if path == "" {
		path = "/"
	}
	retries := in.Retries
	if retries < 0 {
		retries = 0
	}
	timeout := in.ConnectTimeout
	if timeout <= 0 {
		timeout = 60000
	}
	return domain.GatewayService{
		Name: name, Protocol: protocol, Host: host, Port: port, Path: path,
		Retries: retries, ConnectTimeout: timeout, Tags: cleanTags(in.Tags), Enabled: in.Enabled,
	}, nil
}

// ── Routes ──────────────────────────────────────────────────────────────────

func (u *usecase) ListRoutes(ctx context.Context) ([]domain.GatewayRoute, error) {
	return u.repo.ListRoutes(ctx)
}

func (u *usecase) CreateRoute(ctx context.Context, in RouteInput) (domain.GatewayRoute, error) {
	rt, err := in.toDomain()
	if err != nil {
		return domain.GatewayRoute{}, err
	}
	return u.repo.CreateRoute(ctx, &rt)
}

func (u *usecase) UpdateRoute(ctx context.Context, id string, in RouteInput) (domain.GatewayRoute, error) {
	rt, err := in.toDomain()
	if err != nil {
		return domain.GatewayRoute{}, err
	}
	rt.ID = id
	return u.repo.UpdateRoute(ctx, &rt)
}

func (u *usecase) DeleteRoute(ctx context.Context, id string) error {
	return u.repo.DeleteRoute(ctx, id)
}

var allowedMethods = map[string]bool{
	"GET": true, "POST": true, "PUT": true, "PATCH": true, "DELETE": true, "HEAD": true, "OPTIONS": true,
}

func (in RouteInput) toDomain() (domain.GatewayRoute, error) {
	name := strings.TrimSpace(in.Name)
	if name == "" {
		return domain.GatewayRoute{}, apperror.BadRequest("route name is required")
	}
	if strings.TrimSpace(in.ServiceID) == "" {
		return domain.GatewayRoute{}, apperror.BadRequest("route service is required")
	}
	methods := make([]string, 0, len(in.Methods))
	for _, m := range in.Methods {
		m = strings.ToUpper(strings.TrimSpace(m))
		if m == "" {
			continue
		}
		if !allowedMethods[m] {
			return domain.GatewayRoute{}, apperror.BadRequest("invalid HTTP method: " + m)
		}
		methods = append(methods, m)
	}
	if len(methods) == 0 {
		methods = []string{"GET"}
	}
	paths := make([]string, 0, len(in.Paths))
	for _, p := range in.Paths {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		if !strings.HasPrefix(p, "/") {
			return domain.GatewayRoute{}, apperror.BadRequest("path must start with '/': " + p)
		}
		paths = append(paths, p)
	}
	if len(paths) == 0 {
		return domain.GatewayRoute{}, apperror.BadRequest("at least one path is required")
	}
	return domain.GatewayRoute{
		ServiceID: in.ServiceID, Name: name, Methods: methods, Paths: paths,
		StripPath: in.StripPath, PreserveHost: in.PreserveHost, Enabled: in.Enabled,
	}, nil
}

// ── Policies ─────────────────────────────────────────────────────────────—

func (u *usecase) ListPolicies(ctx context.Context) ([]domain.GatewayPolicy, error) {
	return u.repo.ListPolicies(ctx)
}

func (u *usecase) CreatePolicy(ctx context.Context, in PolicyInput) (domain.GatewayPolicy, error) {
	p, err := in.toDomain()
	if err != nil {
		return domain.GatewayPolicy{}, err
	}
	return u.repo.CreatePolicy(ctx, &p)
}

func (u *usecase) UpdatePolicy(ctx context.Context, id string, in PolicyInput) (domain.GatewayPolicy, error) {
	p, err := in.toDomain()
	if err != nil {
		return domain.GatewayPolicy{}, err
	}
	p.ID = id
	return u.repo.UpdatePolicy(ctx, &p)
}

func (u *usecase) DeletePolicy(ctx context.Context, id string) error {
	return u.repo.DeletePolicy(ctx, id)
}

var allowedPolicyTypes = map[string]bool{
	domain.PolicyRateLimit: true, domain.PolicyKeyAuth: true, domain.PolicyCORS: true,
	domain.PolicyIPRestrict: true, domain.PolicyTransform: true,
}

func (in PolicyInput) toDomain() (domain.GatewayPolicy, error) {
	t := strings.ToLower(strings.TrimSpace(in.Type))
	if !allowedPolicyTypes[t] {
		return domain.GatewayPolicy{}, apperror.BadRequest("unknown policy type: " + in.Type)
	}
	cfg := []byte(in.Config)
	if len(cfg) == 0 {
		cfg = []byte("{}")
	}
	if !json.Valid(cfg) {
		return domain.GatewayPolicy{}, apperror.BadRequest("policy config must be valid JSON")
	}
	var routeID *string
	if in.RouteID != nil {
		if id := strings.TrimSpace(*in.RouteID); id != "" {
			routeID = &id
		}
	}
	return domain.GatewayPolicy{RouteID: routeID, Type: t, Config: cfg, Enabled: in.Enabled}, nil
}

// ── Telemetry ────────────────────────────────────────────────────────────—

func (u *usecase) ListRequestLogs(ctx context.Context, limit int) ([]domain.GatewayRequestLog, error) {
	if limit <= 0 || limit > 200 {
		limit = 100
	}
	return u.repo.ListRequestLogs(ctx, limit)
}

func (u *usecase) Overview(ctx context.Context) (domain.GatewayOverview, error) {
	return u.repo.Overview(ctx)
}

// cleanTags нь хоосон/давхардсан tag-уудыг арилгаж, эрэмбэ хадгална.
func cleanTags(tags []string) []string {
	if len(tags) == 0 {
		return []string{}
	}
	seen := make(map[string]bool, len(tags))
	out := make([]string, 0, len(tags))
	for _, t := range tags {
		t = strings.TrimSpace(t)
		if t == "" || seen[t] {
			continue
		}
		seen[t] = true
		out = append(out, t)
	}
	return out
}
