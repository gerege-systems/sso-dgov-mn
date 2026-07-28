# Quick Start

## Prerequisites

- Go 1.26+, Node 20+, PostgreSQL 15+, Redis 7+
- Docker (recommended for the full stack)

## Option A — the full stack with Docker Compose

Brings up `db`, `redis`, a one-off `migrate` job, `api`, `web`, and (when
configured) Ory `hydra` for OIDC-provider mode.

```bash
docker compose up -d --build
```

Open **http://localhost:3007** (the `web` service publishes on a loopback port)
and choose **Login with eID** — scan the QR / open the eID mobile app, or enter a
national ID to receive a push. Google account-linking and the OIDC provider
appear only when their credentials are configured.

!!! warning "Compose runs `ENVIRONMENT=development` on purpose"
    The internal database has no TLS, and the production guard requires
    `sslmode=verify-full`. As a result the compose stack does **not** exercise
    the HSTS / sslmode / observability-gate production paths — those engage only
    with `ENVIRONMENT=production` behind nginx. See [Deployment](../operations/deployment.md).

## Option B — run the services directly

```bash
# 1) Backend  →  http://localhost:8080
cd backend
cp internal/config/.env.example internal/config/.env
#   set JWT_SECRET (≥32 chars), DB, Redis, EID_* RP credentials
go run ./cmd/api

# 2) Frontend →  http://localhost:3000
cd ../frontend
cp .env.example .env.local          # BACKEND_URL=http://localhost:8080
npm install
npm run dev
```

## Backend developer commands

Run from `backend/`:

```bash
go build ./...          # build
go test ./...           # unit tests (mocks only, fast)
make test-integration   # testcontainers (needs Docker)
make swag               # regenerate swagger after touching handler annotations
make pre-push           # mirror CI: lint + test + swag drift + build
```

## Frontend commands

Run from `frontend/`:

```bash
npm run dev             # local dev
npm run build           # build + lint + typecheck (what CI runs)
```

## First login requires eID RP credentials

Because eID is the only login method, a working login needs a registered eID
Mongolia **Relying Party** (RP UUID + secret) in the backend env
(`EID_RP_UUID`, `EID_RP_NAME`, `EID_RP_SECRET`, `EID_BASE_URL`). Without them the
API still boots, but the eID start endpoints return an error. See
[Configuration](../operations/configuration.md) for the full env schema.
