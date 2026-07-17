-- Government Template Platform V3.0
-- Жишээ (mock) gateway service болон demo application-уудыг DAN-ий БОДИТ,
-- гуравдагч талд өгч болох service/RP-ээр солино.
--   Services: ai-gateway/eid-core/payments (mock) → dan-sso (OIDC "Login with DAN")
--             + eid-sign (eID цахим гарын үсэг relay).
--   Applications: demo (migration 31-д consumer-ээс шилжсэн) → бодит RP
--             template.dgov.mn, developer.dgov.mn (тогтвортой client_id-тай).
-- Тэмдэглэл: SQL нь Hydra client үүсгэж чадахгүй тул эдгээр RP-ийн OAuth-ыг
-- бүрэн идэвхжүүлэхийн тулд админ UI-аас secret эргүүлэх/дахин үүсгэнэ.

-- 1) Mock service-үүдийг устгана (тэдний demo route/policy/grant cascade-аар).
DELETE FROM gateway_services WHERE name IN ('ai-gateway', 'eid-core', 'payments');

INSERT INTO gateway_services (name, protocol, host, port, path, tags, scope)
SELECT * FROM (VALUES
    ('dan-sso',  'https', 'dan.dgov.mn', 443, '/oauth2',  ARRAY['sso', 'oidc']::text[], 'svc:dan-sso'),
    ('eid-sign', 'https', 'dan.dgov.mn', 443, '/rp/sign', ARRAY['eid', 'sign']::text[], 'svc:eid-sign')
) AS v(name, protocol, host, port, path, tags, scope)
ON CONFLICT (name) DO NOTHING;

-- 2) Demo application-уудыг устгаж, бодит RP-үүдийг нэмнэ.
DELETE FROM applications WHERE created_by = 'seed';

INSERT INTO applications (client_id, name, app_type, tags, redirect_uris, enabled, created_by)
SELECT * FROM (VALUES
    ('template-dgov-mn',  'template.dgov.mn',  'web', ARRAY['rp']::text[],
        ARRAY['https://template.dgov.mn/auth/callback']::text[], true, 'seed-rp'),
    ('developer-dgov-mn', 'developer.dgov.mn', 'web', ARRAY['rp', 'developer']::text[],
        ARRAY['https://developer.dgov.mn/auth/callback']::text[], true, 'seed-rp')
) AS v(client_id, name, app_type, tags, redirect_uris, enabled, created_by)
ON CONFLICT (client_id) DO NOTHING;

-- 3) Бодит RP-үүдэд eID гарын үсэг (eid-sign) service-ийн хандалт олгож,
--    үйлчилгээ-оноох боломжийг харуулна.
INSERT INTO application_services (application_id, service_id)
SELECT a.id, s.id
FROM applications a
JOIN gateway_services s ON s.name = 'eid-sign'
WHERE a.created_by = 'seed-rp'
ON CONFLICT DO NOTHING;
