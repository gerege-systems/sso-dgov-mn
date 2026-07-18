-- Government Template Platform V3.0
INSERT INTO gateway_services (name, protocol, host, port, path, retries, connect_timeout_ms, tags, enabled, scope)
VALUES ('dan-sso', 'https', 'sso.dgov.mn', 443, '/oauth2', 3, 60000, '{sso,oidc}', true, 'svc:dan-sso')
ON CONFLICT (name) DO NOTHING;
