// Government Template Platform V3.0
// Gerege Systems Development Team болон Claude AI хамтран бүтээв, 2026.

package requests

import (
	"encoding/json"
)

// GatewayServiceRequest нь upstream service үүсгэх/шинэчлэх body.
type GatewayServiceRequest struct {
	Name           string   `json:"name" validate:"required,min=2,max=80"`
	Protocol       string   `json:"protocol" validate:"omitempty,oneof=http https"`
	Host           string   `json:"host" validate:"required,max=255"`
	Port           int      `json:"port" validate:"omitempty,min=1,max=65535"`
	Path           string   `json:"path" validate:"omitempty,max=255"`
	Retries        int      `json:"retries" validate:"omitempty,min=0,max=10"`
	ConnectTimeout int      `json:"connect_timeout_ms" validate:"omitempty,min=100,max=600000"`
	Tags           []string `json:"tags" validate:"omitempty,dive,max=40"`
	Enabled        bool     `json:"enabled"`
}

// GatewayRouteRequest нь route үүсгэх/шинэчлэх body.
type GatewayRouteRequest struct {
	ServiceID    string   `json:"service_id" validate:"required,uuid"`
	Name         string   `json:"name" validate:"required,min=2,max=80"`
	Methods      []string `json:"methods" validate:"omitempty,dive,max=10"`
	Paths        []string `json:"paths" validate:"required,min=1,dive,max=255"`
	StripPath    bool     `json:"strip_path"`
	PreserveHost bool     `json:"preserve_host"`
	Enabled      bool     `json:"enabled"`
}

// GatewayPolicyRequest нь policy (plugin) үүсгэх/шинэчлэх body. route_id хоосон
// бол global policy. config нь plugin-тус-бүрийн JSON.
type GatewayPolicyRequest struct {
	RouteID *string         `json:"route_id" validate:"omitempty,uuid"`
	Type    string          `json:"type" validate:"required,oneof=rate-limit key-auth cors ip-restriction request-transform"`
	Config  json.RawMessage `json:"config" validate:"omitempty"`
	Enabled bool            `json:"enabled"`
}
