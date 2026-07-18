# DAN-Government SSO

> **eID based · AI enabled** — Засгийн газрын үйлчилгээний нэгдсэн нэвтрэлт (Single Sign-On)

Built on the **Government Template Platform V3.0** stack (Clean-Architecture Go
backend + Next.js BFF frontend + Gemini AI pipeline), branded and deployed as
**DAN-Government SSO** at [sso.dgov.mn](https://sso.dgov.mn).

> 🌐 **English** · [Монгол](docs/README_MN.md)

[![Go](https://img.shields.io/badge/Go-1.26-blue.svg)](https://golang.org/)
[![chi](https://img.shields.io/badge/chi-v5-00ADD8.svg)](https://github.com/go-chi/chi)
[![pgx](https://img.shields.io/badge/pgx-v5-336791.svg)](https://github.com/jackc/pgx)
[![Next.js](https://img.shields.io/badge/Next.js-15-black.svg)](https://nextjs.org/)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)
[![PRs Welcome](https://img.shields.io/badge/PRs-welcome-brightgreen.svg)](CONTRIBUTING.md)

A production-ready, security-hardened **full-stack template** built on Clean
Architecture. It pairs a Go (**chi · net/http + pgx (pgxpool) + PostgreSQL + Redis**)
backend with a Next.js (**BFF**) frontend, wired together and ready to extend into
any system. The backend uses the standard library `net/http` with the
[go-chi/chi](https://github.com/go-chi/chi) router and the
[jackc/pgx](https://github.com/jackc/pgx) driver with hand-written SQL — no ORM.

## 📌 Origin & Open Source

The **backend** is derived from the open-source
[snykk/go-rest-boilerplate](https://github.com/snykk/go-rest-boilerplate)
(MIT, by Najib Fikri); we ported the HTTP layer **Gin → chi (net/http)** and the
data layer **sqlx → pgx (pgxpool, hand-written SQL)**, keeping the full feature
set. Upstream attribution is retained in [AUTHORS](AUTHORS). This project is
**MIT-licensed** — see [LICENSE](LICENSE).

## Monorepo structure

```
dan-dgov-mn/
├── backend/           # Go · chi (net/http) · pgx (pgxpool) · PostgreSQL · Redis · eID/Google/SSO auth
│   └── docs/          # ARCHITECTURE · DEVELOPMENT · API_CONTRACT · SECURITY (EN/MN)
└── frontend/          # Next.js BFF (server-side proxy to the backend; cookie sessions)
```

- **[backend/README.md](backend/README.md)** — Clean Architecture Go API.
- **[frontend/README.md](frontend/README.md)** — Next.js Backend-for-Frontend.

## Features

- **Clean Architecture** — `handler → usecase → repository → domain`, no back-imports; the business core never imports the web framework.
- **Auth — eID + Google + dgov SSO** — the only login method is **Login with eID** (eID Mongolia Relying Party: QR code / mobile deep-link / national-ID push with a long-poll session). Alongside it, **Google OAuth** account-linking and a **dgov SSO (OIDC) consumer** (`sso.dgov.mn`). Sessions are JWT access + refresh (rotation); logout revokes both (refresh + access deny-list). There is no password or email/OTP login.
- **eID PKI profile** — reads the signed-in citizen's eID identity from the IdP: linked organizations & authorized signers, certificates, registered devices, and activity.
- **Organizations & membership** — create/lookup organizations (state-registry lookup via Gerege Verify/XYP) and manage members/roles, RLS-scoped per user.
- **Government services portal** — the citizen-facing `Төрийн үйлчилгээ` surface: service catalogue, applications, references, notifications, payments, appointments.
- **API gateway** — admin-managed services / routes / consumers / API keys / policies with request telemetry (overview + logs).
- **OIDC provider (SSO)** — DAN itself acts as an identity provider: an [Ory Hydra](https://www.ory.sh/hydra/) front-end drives the login/consent/logout flows so relying parties can `Sign in with DAN`. Enabled only when Hydra is configured.
- **Document signing (PAdES)** — server-side PDF signing through eID Mongolia `/v3`, with a persistent Document-Signer certificate; a sign-relay lets third-party RPs sign through DAN's eID credentials.
- **Third-party integrations** — per-user OAuth links (Google Drive/Meet, Dropbox) with tokens encrypted at rest (AES-256-GCM), plus **Gerege Space** app-native SFTP storage.
- **AI pipeline (Gemini)** — SDK-free REST client with function calling: text/voice chat, speech-to-text, text-to-speech, live translation. Layered system prompt (hardcoded guardrails + admin-configurable scope/instructions in the DB) keeps the assistant inside its configured domain; a `search_knowledge` tool grounds answers in the `ai_knowledge` table.
- **Audit log** — hash-chained, append-only audit trail (admin-only read + integrity verify).
- **RBAC & super admin** — dynamic roles + permission catalogue; a 4-role model (**superadmin → admin → manager → user**) where super admin is the only role that can manage admin accounts. Super admins are **MFA-secured, separate accounts**: created through an onboarding wizard (invite allow-list → Google → eID → email OTP → TOTP + recovery codes) and stored in a dedicated `superadmin_accounts` satellite table keyed by Google identity — so one person can be an eID admin *and* a Google super admin at once, and **every** super admin login is MFA-gated. Admins are granted by register number against local eID-registered users.
- **Site appearance** — admin-configurable site-wide look (accent / font / density / theme) for public pages, plus per-user overrides.
- **Security-hardened** — strict security headers (CSP, HSTS, COOP/COEP/CORP), CORS allow-list, rate limiting, full HTTP server timeouts, parameterized queries, Postgres Row-Level Security with a boot-time enforceability guard. See [SECURITY.md](SECURITY.md).
- **Observability** — OpenTelemetry tracing + Prometheus metrics + structured Zap logs; `/metrics` and `/swagger` are gated behind a bearer token in production.
- **Frontend BFF** — the browser talks only to same-origin Next.js routes, which proxy to the backend server-side (tokens never reach client JS); double CSRF defense (custom header + origin check), TanStack Query data layer.
- **Tested** — unit tests + testcontainers integration tests.

## Quick start

**Prerequisites:** Go 1.26+, Node 20+, PostgreSQL 15+, Redis 7+ (Docker recommended for the full stack).

```bash
# 1) Backend  →  http://localhost:8080
cd backend
cp internal/config/.env.example internal/config/.env   # set JWT_SECRET (≥32 chars), DB, Redis, EID_* RP creds

# 2) Frontend →  http://localhost:3000
cd ../frontend
cp .env.example .env.local                              # BACKEND_URL=http://localhost:8080
npm install
npm run dev
```

Or bring up the whole stack (db + redis + migrate + api + web, and Hydra for OIDC-provider mode):

```bash
docker compose up -d --build
```

Open **http://localhost:3000** and choose **Login with eID** (scan the QR / open the eID mobile app, or enter a national ID to receive a push). Google account-linking and dgov SSO appear when their credentials are configured.

## Documentation

| Doc | What |
|-----|------|
| [backend/docs/ARCHITECTURE.md](backend/docs/ARCHITECTURE.md) | Layers, dependency flow, components |
| [backend/docs/DEVELOPMENT.md](backend/docs/DEVELOPMENT.md) | Add-a-feature guide, testing, code style |
| [backend/docs/API_CONTRACT.md](backend/docs/API_CONTRACT.md) | REST endpoints, request/response shapes |
| [backend/docs/AI_PIPELINE.md](backend/docs/AI_PIPELINE.md) | AI assistant internals: flows, prompt layers, tools, voice, how to extend |
| [backend/docs/SECURITY.md](backend/docs/SECURITY.md) | Implemented controls + ASVS roadmap |
| [docs/DEPLOYMENT.md](docs/DEPLOYMENT.md) | VPS deployment runbook (compose, env files, nginx, Hydra, updates, rollback) |
| [ROADMAP.md](ROADMAP.md) | What's shipped and what's next |
| [SECURITY.md](SECURITY.md) | How to report a vulnerability |
| [CONTRIBUTING.md](CONTRIBUTING.md) | How to contribute |

## Contributing

Contributions are welcome — please read [CONTRIBUTING.md](CONTRIBUTING.md) and
the [Code of Conduct](docs/CODE_OF_CONDUCT.md).

## License

[MIT](LICENSE) — derivative of snykk/go-rest-boilerplate (MIT); upstream
attribution is retained in [AUTHORS](AUTHORS).

---

**Government Template Platform V3.0** — Co-developed by the **Gerege Systems
Development Team** and **Claude AI**, 2026.

<!-- submodule sync test: dgov-mn-projects -->
