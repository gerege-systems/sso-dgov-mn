// Government Template Platform V3.0
// Gerege Systems Development Team болон Claude AI хамтран бүтээв, 2026.

// Package gateway нь API Gateway-ийн admin удирдлагын business logic —
// services/routes/consumers/api keys/policies-ийн CRUD, validation, API key
// үүсгэлт (hash) болон dashboard-ийн нэгтгэлийг хариуцна.
package gateway

import (
	"context"
	"encoding/json"

	"template/internal/business/domain"
)

type Usecase interface {
	// Services
	ListServices(ctx context.Context) ([]domain.GatewayService, error)
	CreateService(ctx context.Context, in ServiceInput) (domain.GatewayService, error)
	UpdateService(ctx context.Context, id string, in ServiceInput) (domain.GatewayService, error)
	DeleteService(ctx context.Context, id string) error

	// Routes
	ListRoutes(ctx context.Context) ([]domain.GatewayRoute, error)
	CreateRoute(ctx context.Context, in RouteInput) (domain.GatewayRoute, error)
	UpdateRoute(ctx context.Context, id string, in RouteInput) (domain.GatewayRoute, error)
	DeleteRoute(ctx context.Context, id string) error

	// Policies
	ListPolicies(ctx context.Context) ([]domain.GatewayPolicy, error)
	CreatePolicy(ctx context.Context, in PolicyInput) (domain.GatewayPolicy, error)
	UpdatePolicy(ctx context.Context, id string, in PolicyInput) (domain.GatewayPolicy, error)
	DeletePolicy(ctx context.Context, id string) error

	// Telemetry
	ListRequestLogs(ctx context.Context, limit int) ([]domain.GatewayRequestLog, error)
	Overview(ctx context.Context) (domain.GatewayOverview, error)
}

type (
	ServiceInput struct {
		Name           string
		Protocol       string
		Host           string
		Port           int
		Path           string
		Retries        int
		ConnectTimeout int
		Tags           []string
		Enabled        bool
	}
	RouteInput struct {
		ServiceID    string
		Name         string
		Methods      []string
		Paths        []string
		StripPath    bool
		PreserveHost bool
		Enabled      bool
	}
	PolicyInput struct {
		RouteID *string
		Type    string
		Config  json.RawMessage
		Enabled bool
	}
)
