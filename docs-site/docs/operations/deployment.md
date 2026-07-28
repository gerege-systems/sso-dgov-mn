# Deployment

The full runbook lives in `docs/DEPLOYMENT.md`; this is the operational summary.

## The docker-compose stack

`docker-compose.yml` defines seven services; only loopback host ports are
published, and `db`/`redis` never leave the internal network.

| Service | Image / build | Role | Ports |
|---------|---------------|------|-------|
| `db` | `postgres:16-alpine` | Postgres — hosts both the app DB and the `hydra` DB | internal |
| `redis` | `redis:7-alpine` (`--requirepass`) | cache / state | internal |
| `migrate` | build `./backend` | one-off `migrate -up`, then exits | `restart: "no"` |
| `api` | build `./backend` | Go API (`:8080`) | `127.0.0.1:8091:8080` (relay) |
| `web` | build `./frontend` | Next.js BFF | `127.0.0.1:3007:3000` |
| `hydra-migrate` | `oryd/hydra:v2.2.0` | applies Hydra schema, exits | `restart: "no"` |
| `hydra` | `oryd/hydra:v2.2.0` | OIDC issuer | `:4444` public (proxied), `:4445` admin loopback-only |

Dependency ordering: `api` waits on `db` healthy + `redis` healthy + `migrate`
completed; `web` waits on `api` healthy; `hydra` waits on `db` + `hydra-migrate`.
The `db` healthcheck uses `pg_isready -h 127.0.0.1` (TCP) so it stays unhealthy
until initdb finishes — preventing `migrate` from racing in during the
socket-only first-boot window.

## Two DB roles

- `migrate` / `hydra-migrate` connect as the **superuser** (`POSTGRES_USER`) —
  they need `CREATE EXTENSION`, RLS DDL, create-database.
- `api` connects as `APP_DB_USER` (`NOSUPERUSER NOBYPASSRLS`) so Postgres RLS is
  actually enforced. The API verifies its role at boot and **fails to start in
  production if it is superuser/BYPASSRLS**.

The app role and the hydra DB are auto-created on first init by
`backend/deploy/initdb/10-create-app-user.sh` and `20-create-hydra-db.sh`.

## nginx reverse proxy + TLS

nginx runs on the host, terminates TLS (Let's Encrypt), and fans out three
loopback upstreams:

- `/oauth2/*`, `/userinfo`, `/.well-known/*` → Hydra public (`:4444`)
- `/rp/sign/*` → api relay (`:8091`)
- everything else (app, BFF `/api/*`, the DAN-rendered OIDC login/consent UI) →
  web (`:3007`)

HTTPS is mandatory: compose sets `COOKIE_SECURE=true` and Hydra runs
`SameSite=None` (which requires `Secure`), so browsers drop the auth/OIDC cookies
over plain HTTP. The Hydra admin port `:4445` is deliberately **not** proxied.
Set `TRUSTED_PROXIES=172.16.0.0/12,127.0.0.1` or per-IP rate limits collapse into
one bucket.

## Migrations

`migrate` runs `/app/migrate -up` on every `docker compose up`; already-applied
migrations are skipped. Files live in `backend/migrations/` as `N_name.up.sql` /
`.down.sql` — the current tree is a consolidated `1_init_schema` pair plus an
`old/` folder of the pre-consolidation files. `hydra-migrate` applies Hydra's own
schema separately.

## Update & rollback

**Manual update:** `git pull --ff-only` → `docker compose build` → `up -d`.
`db`/`redis` keep running; both migrate jobs re-run and skip applied migrations.

**Automated:** `deploy.yml` SSHes as a non-root `deploy` user, `git reset --hard`
to the exact CI-passed SHA, then runs `deploy/deploy.sh` — which idempotently
writes `INTEGRATION_ENC_KEY` only if absent (never overwrites — rotating it would
break all AES-GCM-encrypted data), builds, brings up `--remove-orphans`, waits up
to ~150 s for `api`/`web` to report healthy (dumping logs on failure), then prunes
images. Concurrency group `deploy-production` serializes deploys.

**Rollback:** `git reset --hard <commit>` then rebuild. Migrations are
forward-only in this flow — to revert a schema change, apply the matching
`down.sql` by hand before rolling code back past it.
