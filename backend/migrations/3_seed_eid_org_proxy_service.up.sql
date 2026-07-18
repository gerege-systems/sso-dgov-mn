-- Government Template Platform V3.0
-- Байгууллагатай холбоотой eID service-үүдийг (organizations/signers) тусад нь
-- proxy service болгож API gateway catalog-д бүртгэнэ. Admin gateway UI-аас
-- хувь хүний eID (eid-proxy)-ээс ТУСДАА enable/disable хийж болно. Route нь энэ
-- enabled флагийг runtime-д шалгадаг. Public зам: /rp/eid-org/*.
INSERT INTO gateway_services (name, protocol, host, port, path, retries, connect_timeout_ms, tags, enabled, scope)
VALUES ('eid-org-proxy', 'https', 'sso.dgov.mn', 443, '/rp/eid-org', 3, 60000, '{eid,org,proxy}', true, 'svc:eid-org-proxy')
ON CONFLICT (name) DO NOTHING;
