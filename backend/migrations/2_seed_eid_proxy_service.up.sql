-- Government Template Platform V3.0
-- eID proxy service-ийг API gateway catalog-д бүртгэнэ. Ингэснээр admin gateway
-- UI-д харагдаж, enable/disable хийж болно (route нь энэ enabled флагийг runtime-д
-- шалгадаг). Бүртгэгдсэн апп-ууд /rp/eid/* -ээр SSO-ий eID service-үүдийг proxy-оор
-- дуудна.
INSERT INTO gateway_services (name, protocol, host, port, path, retries, connect_timeout_ms, tags, enabled, scope)
VALUES ('eid-proxy', 'https', 'sso.dgov.mn', 443, '/rp/eid', 3, 60000, '{eid,proxy}', true, 'svc:eid-proxy')
ON CONFLICT (name) DO NOTHING;
