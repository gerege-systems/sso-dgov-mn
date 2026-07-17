-- Government Template Platform V3.0
-- 31_applications-ийн буцаалт — gateway_consumers/gateway_api_keys-ийг сэргээж,
-- applications/application_services болон gateway_services.scope-ыг хасна.

-- gateway_consumers-ыг (22-ийн схемээр) сэргээнэ.
CREATE TABLE IF NOT EXISTS gateway_consumers (
    id         uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
    username   TEXT UNIQUE NOT NULL,
    custom_id  TEXT NOT NULL DEFAULT '',
    tags       TEXT[] NOT NULL DEFAULT '{}',
    enabled    BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ
);
CREATE TABLE IF NOT EXISTS gateway_api_keys (
    id           uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
    consumer_id  uuid NOT NULL REFERENCES gateway_consumers(id) ON DELETE CASCADE,
    label        TEXT NOT NULL DEFAULT '',
    key_prefix   TEXT NOT NULL,
    key_hash     TEXT NOT NULL,
    last_used_at TIMESTAMPTZ,
    expires_at   TIMESTAMPTZ,
    revoked      BOOLEAN NOT NULL DEFAULT false,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_gateway_api_keys_consumer ON gateway_api_keys (consumer_id);
CREATE UNIQUE INDEX IF NOT EXISTS idx_gateway_api_keys_hash ON gateway_api_keys (key_hash);

-- Demo апп-уудыг consumers руу буцаана (id хадгална).
INSERT INTO gateway_consumers (id, username, custom_id, tags, enabled, created_at)
SELECT id, name, '', tags, enabled, created_at FROM applications WHERE created_by = 'seed'
ON CONFLICT (id) DO NOTHING;

-- request_logs FK-г consumers руу буцаана.
ALTER TABLE gateway_request_logs DROP CONSTRAINT IF EXISTS gateway_request_logs_consumer_id_fkey;
ALTER TABLE gateway_request_logs
    ADD CONSTRAINT gateway_request_logs_consumer_id_fkey
    FOREIGN KEY (consumer_id) REFERENCES gateway_consumers(id) ON DELETE SET NULL;

DROP TABLE IF EXISTS application_services;
DROP TABLE IF EXISTS applications;
ALTER TABLE gateway_services DROP COLUMN IF EXISTS scope;
