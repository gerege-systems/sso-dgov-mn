-- Government Template Platform V3.0
-- API Gateway consumer + SSO RP (developer_apps) бүртгэлийг НЭГ 'applications'
-- загвар болгон нэгтгэнэ. Application бүр = Hydra OAuth2 client (RP =
-- authorization_code, m2m = client_credentials). Чимэглэлийн (runtime-д хэзээ ч
-- шалгагддаггүй) gateway_api_keys-ийг хасаж, credential-ийг Hydra client_secret
-- болгоно. Аппад ямар gateway service ашиглаж болохыг application_services-ээр
-- (OAuth scope) оноодог. RLS-гүй нийтийн config (gateway-тэй ижил); app CRUD хийнэ.

-- Нэгдсэн бүртгэл.
CREATE TABLE IF NOT EXISTS applications (
    id            uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
    client_id     text UNIQUE NOT NULL,           -- Hydra OAuth2 client_id
    name          text NOT NULL,
    app_type      text NOT NULL DEFAULT 'm2m',    -- web | spa | native | m2m
    tags          text[] NOT NULL DEFAULT '{}',
    redirect_uris text[] NOT NULL DEFAULT '{}',   -- Hydra-гийн толин тусгал (display)
    enabled       boolean NOT NULL DEFAULT true,
    created_by    text NOT NULL DEFAULT '',
    created_at    timestamptz NOT NULL DEFAULT now(),
    updated_at    timestamptz
);

-- gateway service бүрд OAuth scope нэр (name-ээс backfill).
ALTER TABLE gateway_services ADD COLUMN IF NOT EXISTS scope text NOT NULL DEFAULT '';
UPDATE gateway_services SET scope = 'svc:' || name WHERE scope = '';

-- Аппад зөвшөөрсөн service-үүд (байгаа мөр = зөвшөөрөгдсөн). App-ийн Hydra
-- client-ийн scope нь эдгээр service-ийн scope-уудаас бүрдэнэ.
CREATE TABLE IF NOT EXISTS application_services (
    application_id uuid NOT NULL REFERENCES applications(id) ON DELETE CASCADE,
    service_id     uuid NOT NULL REFERENCES gateway_services(id) ON DELETE CASCADE,
    PRIMARY KEY (application_id, service_id)
);
CREATE INDEX IF NOT EXISTS idx_application_services_service ON application_services (service_id);

-- Seed хийсэн demo gateway_consumers-ыг applications руу шилжүүлнэ (id-г хадгалж
-- gateway_request_logs.consumer_id-г хүчинтэй үлдээнэ). Эдгээр demo мөр нь
-- ЗӨВХӨН display — энэ SQL Hydra client үүсгэхгүй; admin UI-аар үүсгэсэн апп
-- жинхэнэ Hydra client-тай болно. client_id = 'demo-'||username.
INSERT INTO applications (id, client_id, name, app_type, tags, enabled, created_by, created_at)
SELECT c.id, 'demo-' || c.username, c.username, 'm2m', c.tags, c.enabled, 'seed', c.created_at
FROM gateway_consumers c
ON CONFLICT (id) DO NOTHING;

-- Demo апп бүрд жишээ service хандалт оноож харуулна.
INSERT INTO application_services (application_id, service_id)
SELECT a.id, s.id
FROM applications a
JOIN gateway_services s ON s.name IN ('eid-core', 'payments')
WHERE a.created_by = 'seed'
ON CONFLICT DO NOTHING;

-- Чимэглэлийн gateway API key-г хасна (runtime-д хэзээ ч шалгагддаггүй байсан).
DROP TABLE IF EXISTS gateway_api_keys;

-- gateway_consumers-ыг applications орлосон тул хасна. request_logs.consumer_id
-- одоо applications(id)-г заана (id-г дээр хадгалсан). FK-г дахин холбоно.
ALTER TABLE gateway_request_logs DROP CONSTRAINT IF EXISTS gateway_request_logs_consumer_id_fkey;
DROP TABLE IF EXISTS gateway_consumers;
ALTER TABLE gateway_request_logs
    ADD CONSTRAINT gateway_request_logs_consumer_id_fkey
    FOREIGN KEY (consumer_id) REFERENCES applications(id) ON DELETE SET NULL;
