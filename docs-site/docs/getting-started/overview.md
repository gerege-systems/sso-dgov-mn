# Overview

DAN-Government SSO is a full-stack identity platform for government services,
deployed at [sso.dgov.mn](https://sso.dgov.mn). It is built on the reusable
**Government Template Platform V3.0** and branded/extended as DAN.

## Design principles

- **eID-first.** The only interactive login is *Login with eID*. There is no
  password, email, or OTP login path — the legacy usecases exist but are
  unreachable. Google is an account **link**, not a standalone login.
- **Clean Architecture.** Dependencies point inward only:
  `handler → usecase → repository → domain`. The business core never imports the
  web framework or the database driver, so either can be swapped.
- **No ORM.** All SQL is hand-written with pgx and parameterized; records are
  plain structs scanned by name. See [Backend Layers](../architecture/backend-layers.md).
- **Defense in depth.** Postgres Row-Level Security sits *beneath* the
  `WHERE user_id = …` clauses, so even a query bug cannot leak another user's
  rows. See [Security](../security/index.md).
- **BFF frontend.** The browser only talks to same-origin Next.js routes; JWTs
  live in httpOnly cookies and never reach client JavaScript.

## Monorepo structure

```
sso-dgov-mn/
├── backend/           # Go · chi (net/http) · pgx · PostgreSQL · Redis
│   ├── cmd/           # api (server + DI), migration, seed CLIs
│   ├── internal/      # business (domain + usecases), datasources, http
│   ├── pkg/           # framework-agnostic clients (eid, google, hydra, gemini…)
│   ├── migrations/    # numbered SQL (consolidated 1_init_schema)
│   └── docs/          # ARCHITECTURE · DEVELOPMENT · API_CONTRACT · SECURITY (EN/MN)
├── frontend/          # Next.js 15 BFF (server-side proxy; cookie sessions)
├── deploy/            # nginx, initdb, deploy scripts
└── docker-compose.yml # db · redis · migrate · api · web · hydra
```

## Feature map

| Area | Summary | Docs |
|------|---------|------|
| Authentication | eID login (QR / deep-link / national-ID push), Google linking, JWT sessions | [Authentication](../authentication/index.md) |
| SSO provider | DAN as an OIDC provider via Ory Hydra ("Sign in with DAN") | [Sign in with DAN](../sso/index.md) |
| eID PKI profile | Certificates, devices, activity, linked organizations & signers | [eID Login](../authentication/eid-login.md) |
| Organizations | Create/lookup orgs (state-registry lookup), members & roles, RLS-scoped | [API Reference](../api/index.md) |
| Gov services portal | Applications, references, notifications, payments, appointments | [API Reference](../api/index.md) |
| API gateway | Admin-managed services/routes/consumers with request telemetry | [API Reference](../api/index.md) |
| Document signing | Server-side PAdES via eID `/v3`; third-party sign-relay | [Document Signing](../signing/index.md) |
| AI assistant | Gemini chat/voice/translation with layered guardrails + knowledge base | [AI Pipeline](../ai/index.md) |
| RBAC & super admin | Dynamic roles + permission catalogue; MFA-secured super admins | [Super Admin MFA](../authentication/super-admin-mfa.md) |
| Audit | Hash-chained, append-only audit trail with integrity verification | [Security](../security/index.md) |

## Next steps

- Bring it up locally → [Quick Start](quick-start.md)
- Understand the runtime environments → [Environments](environments.md)
