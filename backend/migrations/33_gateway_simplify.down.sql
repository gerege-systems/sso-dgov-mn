-- Government Template Platform V3.0
-- 33-ийн буцаалт — gateway_routes + gateway_policies-ийг (22-ийн схемээр) сэргээж,
-- gateway_request_logs-ийн route_id/consumer_id баганыг буцаана. (Seed сэргээхгүй.)
CREATE TABLE IF NOT EXISTS gateway_routes (
    id            uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
    service_id    uuid NOT NULL REFERENCES gateway_services(id) ON DELETE CASCADE,
    name          TEXT NOT NULL,
    methods       TEXT[] NOT NULL DEFAULT '{GET}',
    paths         TEXT[] NOT NULL DEFAULT '{}',
    strip_path    BOOLEAN NOT NULL DEFAULT true,
    preserve_host BOOLEAN NOT NULL DEFAULT false,
    enabled       BOOLEAN NOT NULL DEFAULT true,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at    TIMESTAMPTZ
);
CREATE INDEX IF NOT EXISTS idx_gateway_routes_service ON gateway_routes (service_id);

CREATE TABLE IF NOT EXISTS gateway_policies (
    id         uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
    route_id   uuid REFERENCES gateway_routes(id) ON DELETE CASCADE,
    type       TEXT NOT NULL,
    config     JSONB NOT NULL DEFAULT '{}',
    enabled    BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ
);
CREATE INDEX IF NOT EXISTS idx_gateway_policies_route ON gateway_policies (route_id);

ALTER TABLE gateway_request_logs ADD COLUMN IF NOT EXISTS route_id uuid REFERENCES gateway_routes(id) ON DELETE SET NULL;
ALTER TABLE gateway_request_logs ADD COLUMN IF NOT EXISTS consumer_id uuid REFERENCES applications(id) ON DELETE SET NULL;
