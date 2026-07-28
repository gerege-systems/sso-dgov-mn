# Configuration

Configuration is read by Viper (`backend/internal/config/config.go`). Secrets
live in two **gitignored** env files; only `*.env.example` is committed.

- **`./.env`** — compose interpolation only (`POSTGRES_*`, `APP_DB_*`,
  `REDIS_PASS`, `APP_ORIGIN`, `WEB_PORT`, `API_RELAY_PORT`, `HYDRA_*` ports/secrets,
  web-side OAuth client IDs). Hydra secrets use `${VAR:?}` so compose refuses to
  start if unset.
- **`./backend.env`** — mounted read-only into `api` / `migrate` at `/app/.env`.

!!! danger "Never commit secrets"
    `.env`, `backend.env`, and `backend/internal/config/.env*` are gitignored.
    Document new variables in the READMEs, not in the repo's committed config.
    In production, prefer a real secret store / KMS over flat `.env` files.

## Key backend variables

| Variable | Purpose |
|----------|---------|
| `ENVIRONMENT` | `development` \| `production` — gates HSTS, sslmode, RLS guard, observability gate |
| `JWT_SECRET` | HMAC signing key (≥ 32 chars) |
| `JWT_EXPIRED` / `JWT_REFRESH_EXPIRED` | access TTL (hours) / refresh TTL (days, default 7) |
| `DB_POSTGRE_DSN` / `DB_POSTGRE_URL` | dev DSN / prod URL (prod requires `sslmode=verify-full`) |
| `DB_MAX_OPEN_CONNS` / `DB_MAX_IDLE_CONNS` / `DB_CONN_MAX_LIFE_MINS` | pool sizing (25 / 5 / 15) |
| `REDIS_*` | cache / session-revocation store |
| `TRUSTED_PROXIES` | CIDRs whose `X-Forwarded-For` is trusted |
| `EID_RP_UUID` / `EID_RP_NAME` / `EID_RP_SECRET` / `EID_BASE_URL` | eID Relying Party identity |
| `SIGN_SIGNER_CERT_FILE` / `SIGN_SIGNER_KEY_FILE` | persistent Document-Signer PEM (required in prod) |
| `SIGN_RELAY_TOKEN` | enables the third-party sign-relay (with `EID_RP_SECRET`) |
| `INTEGRATION_ENC_KEY` | AES-256-GCM key for integration tokens + super-admin TOTP; enables super-admin onboarding |
| `HYDRA_ADMIN_URL` / `HYDRA_PUBLIC_URL` / `SSO_STATE_KEY` (≥32) | enable the OIDC provider |
| `SSO_FIRSTPARTY_CLIENTS` | client IDs that skip the consent UI |
| `GEMINI_API_KEY` / `GEMINI_MODEL` / `GEMINI_TTS_MODEL` / `GEMINI_VOICE` | AI pipeline |
| `AI_SCOPE_PROMPT` | fallback scope for the AI guardrail layer |
| `OBSERVABILITY_TOKEN` | bearer for `/metrics` + `/swagger/doc.json` in prod |
| `OTEL_EXPORTER_OTLP_ENDPOINT` | OTLP tracing endpoint (prod) |

## Key frontend variables

| Variable | Purpose |
|----------|---------|
| `BACKEND_URL` | server-side base for the BFF (`+ /api/v1`); `http://api:8080` in compose |
| `APP_ORIGIN` | expected Origin for the CSRF origin check |
| `COOKIE_SECURE` | `true` in production (HTTPS-only cookies) |
| `NEXT_PUBLIC_DOCS_URL` | overrides the "Docs" link target (defaults to this documentation site) |

!!! note "Feature flags are config presence"
    Several subsystems mount only when configured: the OIDC provider needs the
    `HYDRA_*` + `SSO_STATE_KEY` trio; super-admin onboarding needs
    `INTEGRATION_ENC_KEY`; the sign-relay needs `SIGN_RELAY_TOKEN` + `EID_RP_SECRET`.
