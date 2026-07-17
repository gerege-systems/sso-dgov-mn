-- Government Template Platform V3.0
-- 32-ийн буцаалт — бодит RP-үүдийг хасаж, mock service-үүдийг сэргээнэ.
-- (Demo application-уудыг буцааж сэргээхгүй.)
DELETE FROM applications WHERE created_by = 'seed-rp';

DELETE FROM gateway_services WHERE name IN ('dan-sso', 'eid-sign');

INSERT INTO gateway_services (name, protocol, host, port, path, tags, scope)
SELECT * FROM (VALUES
    ('eid-core',   'https', 'api.eiddgov.mn', 443, '/gerege/v1', ARRAY['eid', 'core']::text[], 'svc:eid-core'),
    ('payments',   'https', 'pay.dgov.mn',    443, '/v2',        ARRAY['billing']::text[],     'svc:payments'),
    ('ai-gateway', 'https', 'ai.dgov.mn',     443, '/v1',        ARRAY['ai']::text[],          'svc:ai-gateway')
) AS v(name, protocol, host, port, path, tags, scope)
ON CONFLICT (name) DO NOTHING;
