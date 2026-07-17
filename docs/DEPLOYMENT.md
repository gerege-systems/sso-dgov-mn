# Deployment Guide

> 🌐 **English** · [Монгол](DEPLOYMENT_MN.md)

How to deploy **DAN-Government SSO** (dan.dgov.mn) to a single VPS with Docker
Compose behind nginx. The stack is Postgres + Redis + Go API + Next.js BFF web
+ **Ory Hydra** (the OIDC issuer that turns dan into an SSO provider). This is
the runbook used for the reference deployment.

## Topology

Three host loopback ports are published; nginx terminates TLS and reverse-proxies
each to the right container. `db` and `redis` never leave the internal compose
network, and the Hydra **admin** API is loopback-only (never proxied).

```
Internet ──► nginx (80/443, TLS via Let's Encrypt)
   │
   ├─ /oauth2/*, /.well-known/openid-configuration, /userinfo, /health/ready
   │      ─────────────────────────► hydra  127.0.0.1:${HYDRA_PUBLIC_PORT}   (Ory Hydra — OIDC issuer, PUBLIC API)
   │
   ├─ /rp/sign/*  (eID sign relay for 3rd-party Relying Parties)
   │      ─────────────────────────► api    127.0.0.1:${API_RELAY_PORT}      (backend :8080, loopback relay)
   │
   └─ everything else — app, BFF /api/*, and the OIDC login/consent UI
      (/oauth/login, /oauth/consent, /oauth/logout, /oauth/error)
          ─────────────────────────► web    127.0.0.1:${WEB_PORT}            (Next.js BFF)
                                       │ BACKEND_URL=http://api:8080
                                       ▼
   internal compose network (no public host ports):
        api ──► db (Postgres 16 — gerege_template + hydra databases) + redis (7)
        hydra ──► db (hydra database)   admin :4445 = LOOPBACK ONLY, never proxied
        hydra-migrate (one-off), migrate (one-off) — apply schema, then exit
```

So `web` is **not** the only exposed container: nginx must also front the Hydra
public API (`:4444`) and the api sign relay (`:8091`). The browser reaches `web`
for the app and its BFF; the OIDC protocol endpoints are served by Hydra; and
the OAuth *login/consent* pages (which dan renders itself, after authenticating
the citizen with eID) live on `web` under `/oauth/*`. A one-off `migrate`
container applies SQL migrations on every `up`; `hydra-migrate` applies Hydra's
own schema into the separate `hydra` database and exits.

## Prerequisites

- A VPS with Docker + the compose plugin (`docker compose version`)
- nginx + certbot on the host (or any reverse proxy that terminates TLS)
- A DNS record for `dan.dgov.mn` pointing at the server

## 1. Get the code

```bash
git clone https://github.com/gerege-systems/dan-dgov-mn.git /srv/dan
cd /srv/dan
```

## 2. Create the two env files (both gitignored)

### `./.env` — compose interpolation

Everything compose interpolates lives here. The Hydra secrets marked **REQUIRED**
use `${VAR:?}` in `docker-compose.yml`, so **compose refuses to start** if they
are unset or empty.

```env
# --- Postgres / Redis ---
POSTGRES_USER=postgres            # superuser — used by migrate + hydra-migrate only
POSTGRES_PASSWORD=<random>
POSTGRES_DB=gerege_template
APP_DB_USER=app_user              # least-privilege role the api connects as
APP_DB_PASSWORD=<random>
APP_DB_DSN=host=db port=5432 user=app_user password=<same> dbname=gerege_template sslmode=disable
REDIS_PASS=<random>

# --- App / origin ---
APP_ORIGIN=https://dan.dgov.mn    # exact public origin (CSRF origin check)
WEB_PORT=3007                     # loopback port nginx proxies the app to
API_RELAY_PORT=8091               # loopback port nginx proxies /rp/sign to (api :8080)

# --- Ory Hydra (OIDC issuer) ---
HYDRA_PUBLIC_PORT=4444            # loopback port nginx proxies the OIDC public API to
HYDRA_ADMIN_PORT=4445             # Hydra admin API — bound to loopback, NEVER proxied
HYDRA_PUBLIC_URL=https://dan.dgov.mn          # REQUIRED — OIDC issuer / self URL
HYDRA_POST_LOGOUT_REDIRECT=https://dan.dgov.mn/   # optional; defaults to HYDRA_PUBLIC_URL/
HYDRA_SYSTEM_SECRET=<≥32 random chars>        # REQUIRED — Hydra system secret
HYDRA_COOKIE_SECRET=<≥32 random chars>        # REQUIRED — Hydra cookie secret
HYDRA_PAIRWISE_SALT=<random>                  # REQUIRED — pairwise subject salt

# --- OAuth client IDs/secrets used by the web BFF (empty = that button/card inert) ---
GOOGLE_CLIENT_ID=<…>              # Google account-linking (also set in backend.env)
GOOGLE_DRIVE_CLIENT_ID=<…>        # third-party integrations; BFF does the token
GOOGLE_DRIVE_CLIENT_SECRET=<…>    # exchange, so the secrets belong here too.
DROPBOX_CLIENT_ID=<…>             # redirect_uri = ${APP_ORIGIN}/api/integrations/<provider>/callback
DROPBOX_CLIENT_SECRET=<…>
GOOGLE_MEET_CLIENT_ID=<…>
GOOGLE_MEET_CLIENT_SECRET=<…>
```

### `./backend.env` — mounted into `api` + `migrate` at `/app/.env`

This is the backend config file (viper reads it). It carries the eID Relying-Party
credentials, the SSO/OIDC provider settings and every integration secret. The full
schema is `backend/internal/config/config.go`; the load-bearing keys for an eID
SSO deployment:

```env
# --- Core runtime ---
PORT=8080
ENVIRONMENT=development           # the compose stack runs dev mode: the internal
                                  # DB has no TLS (the prod guard requires
                                  # sslmode=verify-full); TLS terminates at nginx
DEBUG=false
DB_POSTGRE_DRIVER=postgres
DB_POSTGRE_DSN=postgres://postgres:<POSTGRES_PASSWORD>@db:5432/gerege_template?sslmode=disable
                                  # ^ superuser DSN — used by MIGRATE (DDL).
                                  # The api overrides this with APP_DB_DSN (see §3).
JWT_SECRET=<≥32 random chars>
JWT_EXPIRED=24                    # hours (1–24)
JWT_ISSUER=dan.dgov.mn
JWT_REFRESH_EXPIRED=7             # days
BCRYPT_COST=12
OTP_MAX_ATTEMPTS=5
REDIS_HOST=redis:6379
REDIS_PASS=<same as .env>
REDIS_EXPIRED=5                   # minutes
ALLOWED_ORIGINS=https://dan.dgov.mn
TRUSTED_PROXIES=172.16.0.0/12,127.0.0.1   # trust XFF only from the docker net + nginx.
                                  # REQUIRED behind the proxy: the api has no public
                                  # app port, so requests arrive from the web/nginx
                                  # peer. Without a trusted-proxy list the api ignores
                                  # X-Forwarded-For and all per-IP rate limits collapse
                                  # into one bucket.

# --- eID Relying Party (the ONLY interactive login method) ---
EID_BASE_URL=https://eidmongolia.mn/v3   # eID IdP base (default)
EID_RP_UUID=<RP UUID issued by eID Mongolia>
EID_RP_NAME=dan-dgov-mn
EID_RP_SECRET=<RP secret>
EID_CERT_LEVEL=ADVANCED           # ADVANCED for login (QUALIFIED/QSCD for signing)
EID_CALLBACK_URL=https://dan.dgov.mn/login/verify   # must be allowlisted at the IdP
EID_DISPLAY_TEXT=dan.dgov.mn

# --- Google OAuth (eID account-linking; server-side code exchange) ---
GOOGLE_CLIENT_ID=<…>
GOOGLE_CLIENT_SECRET=<…>

# --- dgov SSO consumer (sso.dgov.mn OIDC — 2nd login alongside eID) ---
SSO_ISSUER=https://sso.dgov.mn
SSO_CLIENT_ID=<…>
SSO_CLIENT_SECRET=<…>
SSO_REDIRECT_URI=https://dan.dgov.mn/sso/callback
SSO_SCOPE=openid profile email
SSO_NATIVE_CLIENT_ID=dan-dgov-mn-ios   # Hydra client_id for the mobile PKCE flow

# --- OIDC PROVIDER side (dan fronts Ory Hydra as an SSO issuer) ---
HYDRA_ADMIN_URL=http://hydra:4445      # admin API (client CRUD + login/consent/logout)
HYDRA_PUBLIC_URL=https://dan.dgov.mn   # issuer used to build redirects
SSO_STATE_KEY=<≥32 random chars>       # login/consent state cookie HMAC key
SSO_FIRSTPARTY_CLIENTS=<csv client_ids>   # skip the consent screen for these
SSO_ADMIN_API_KEYS=<csv bootstrap keys>   # bootstrap keys for the /admin surface
SSO_ADMIN_SUBS=<csv eid_subs>             # eid_subs granted superadmin

# --- Gerege platform services ---
XYP_API_BASE=https://xyp.dgov.mn       # org lookup (HTTP Basic; optional)
XYP_CLIENT_ID=<…>
XYP_CLIENT_SECRET=<…>
CORE_API_BASE=https://core.dgov.mn     # user/org find
CORE_API_TOKEN=<service bearer>
GSPACE_HOST=<sftp host>                # Gerege Space per-user SFTP storage (optional)
GSPACE_PORT=22
GSPACE_USER=<…>
GSPACE_PASSWORD=<…>
GSPACE_BASE_PATH=gerege-space
GSPACE_QUOTA_BYTES=2097152             # 2 MB per user

# --- Encryption / signing / observability ---
INTEGRATION_ENC_KEY=<≥32 random chars> # AES-256-GCM key for stored OAuth tokens
SIGN_RELAY_TOKEN=<shared token>        # enables /rp/sign relay for 3rd-party RPs (empty = off)
SIGN_SIGNER_CERT_FILE=/app/certs/signer.crt   # PAdES document-signer cert (prod: REQUIRED,
SIGN_SIGNER_KEY_FILE=/app/certs/signer.key    #  fail-closed; dev falls back to self-signed)
OBSERVABILITY_TOKEN=<random>           # bearer for /metrics + /swagger/doc.json in prod
GEMINI_API_KEY=<AIza…>                 # AI features; empty = AI endpoints return 500
```

Generate secrets with `openssl rand -hex 24` (or `-hex 32` for the `≥32` keys).
`SIGN_SIGNER_CERT_FILE` / `SIGN_SIGNER_KEY_FILE` are paths **inside** the container —
mount the PEM files (e.g. add a read-only volume to the `api` service) if you set
them; in the compose dev stack they may stay empty and the signer uses a dev
self-signed key.

## 3. Why two DB roles (read before first boot)

Row-Level Security is **silently bypassed** by superusers. The stack therefore
uses two roles:

- `migrate` (and `hydra-migrate`) connect as `POSTGRES_USER` (superuser — needed
  for `CREATE EXTENSION`, RLS DDL, and creating the `hydra` database).
- `api` connects as `APP_DB_USER` (`NOSUPERUSER NOBYPASSRLS`), created
  automatically by `backend/deploy/initdb/10-create-app-user.sh` **on first init
  of an empty data volume**. A second initdb script,
  `20-create-hydra-db.sh`, creates the separate `hydra` database for Ory Hydra.

The api **verifies this at boot**: if its role is superuser/BYPASSRLS it fails to
start in production mode and logs a warning in development mode. If you deploy
onto an *existing* database, create the app role + grants by hand (see the initdb
script), create the `hydra` database
(`docker compose exec db psql -U "$POSTGRES_USER" -c 'CREATE DATABASE hydra;'`),
and point `APP_DB_DSN` at the app role.

## 4. First deploy

```bash
docker compose up -d --build      # builds api+web, runs both migrate jobs, starts all
docker compose ps                 # expect: db/redis/api/web/hydra healthy or running,
                                  #         migrate + hydra-migrate Exited (0)
```

### nginx vhost (host)

The OIDC issuer paths must reach Hydra, `/rp/sign` must reach the api relay, and
everything else goes to `web`. The Hydra admin port (`:4445`) is **never** listed
here.

```nginx
upstream dan_web   { server 127.0.0.1:3007; }   # = WEB_PORT
upstream dan_hydra { server 127.0.0.1:4444; }   # = HYDRA_PUBLIC_PORT
upstream dan_relay { server 127.0.0.1:8091; }   # = API_RELAY_PORT (api :8080)

server {
    server_name dan.dgov.mn;

    # OIDC protocol endpoints → Ory Hydra public API
    location /oauth2/                         { proxy_pass http://dan_hydra; include /etc/nginx/proxy_params; }
    location = /userinfo                      { proxy_pass http://dan_hydra; include /etc/nginx/proxy_params; }
    location /.well-known/openid-configuration { proxy_pass http://dan_hydra; include /etc/nginx/proxy_params; }
    location = /.well-known/jwks.json         { proxy_pass http://dan_hydra; include /etc/nginx/proxy_params; }

    # eID sign relay for 3rd-party Relying Parties → api loopback relay
    location /rp/sign/                        { proxy_pass http://dan_relay; include /etc/nginx/proxy_params; }

    # App, BFF /api/*, and the /oauth/login|consent|logout UI → web BFF
    location / {
        proxy_pass http://dan_web;
        proxy_set_header Host $host;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

(Put the shared `proxy_set_header` lines in `/etc/nginx/proxy_params` and
`include` them, or repeat them per block.) Then `certbot --nginx -d dan.dgov.mn`
for TLS. The compose file sets `COOKIE_SECURE=true` and Hydra runs
`SERVE_COOKIES_SAME_SITE_MODE=None` (needs `Secure`), so the site **must** be
served over HTTPS or browsers will drop the auth and OIDC cookies.

## 5. Updating a running deployment

```bash
cd /srv/dan
git pull --ff-only origin main
docker compose build              # api + web + migrate
docker compose up -d              # recreates changed containers; migrate + hydra-migrate
                                  # re-run (already-applied migrations are skipped)
```

`db` and `redis` keep running — data is untouched. Config-only change? Edit
`backend.env` / `.env` and `docker compose up -d api web` (restart `hydra` too if
you changed a `HYDRA_*` value).

### Automated deploys (CI/CD)

Deploy is **not** a job inside CI. Two workflows chain:

1. [`.github/workflows/ci.yml`](../.github/workflows/ci.yml) — the pre-flight gates
   (`backend`, `frontend`, `secrets-scan`) run on every push to `main` and every PR.
2. [`.github/workflows/deploy.yml`](../.github/workflows/deploy.yml) — a **separate**
   workflow triggered by `workflow_run` **after CI completes**, so CI and Deploy no
   longer run in parallel and a red build never ships. It only deploys when the
   chained CI run concluded `success` on `main` (or on manual `workflow_dispatch`).
   It SSHes into the VPS as a dedicated non-root `deploy` user, `git reset --hard`
   to the exact CI-passed commit, and runs
   [`deploy/deploy.sh`](../deploy/deploy.sh) (rebuild → `up -d` → wait-for-healthy
   → prune). `db`/`redis` stay up; migrations re-run and skip already-applied files.

One-time setup — add these repo secrets under **Settings → Secrets and variables →
Actions**:

| Secret | Value |
|--------|-------|
| `DEPLOY_HOST` | the VPS IP / hostname |
| `DEPLOY_USER` | dedicated **non-root** SSH user (`deploy`) that owns the repo checkout and can run docker |
| `DEPLOY_PATH` | repo path on the server; `deploy.yml` defaults to `/srv/dan` if unset |
| `DEPLOY_SSH_KEY` | **private** key of a dedicated deploy keypair; its public key is in the server's `~/.ssh/authorized_keys` |
| `DEPLOY_PORT` | *(optional)* SSH port, defaults to `22` |

Generate the keypair with `ssh-keygen -t ed25519 -f deploy_key -N ''`, append
`deploy_key.pub` to the `deploy` user's `authorized_keys`, and paste the private
`deploy_key` into `DEPLOY_SSH_KEY`. You can trigger a deploy without a code change
from the Actions tab (**Run workflow** — `workflow_dispatch` deploys `origin/main`
HEAD), or run `bash deploy/deploy.sh` on the server by hand.

## 6. Verify

```bash
docker compose ps                                       # all healthy / migrate jobs Exited(0)
docker logs dan-dgov-mn-migrate-1 | tail -3             # "migration [up] success"
docker logs dan-dgov-mn-hydra-migrate-1 | tail -3       # Hydra schema applied
docker logs dan-dgov-mn-api-1 2>&1 | grep -i error      # should be empty
curl -s -o /dev/null -w '%{http_code}\n' https://dan.dgov.mn/   # 200
curl -s https://dan.dgov.mn/.well-known/openid-configuration | head -c 80   # OIDC issuer JSON
```

## 7. Rollback

```bash
git log --oneline                 # find the last good commit
git checkout <commit> -- .        # or: git reset --hard <commit>
docker compose build && docker compose up -d
```

SQL migrations are forward-only in this flow; if a migration must be reverted,
apply the matching `N_*.down.sql` by hand before rolling the code back past it.

## Secrets hygiene

- `.env` and `backend.env` are gitignored — never commit them.
- Rotate `JWT_SECRET` to force-logout everyone (all tokens invalidate).
- Rotating `HYDRA_SYSTEM_SECRET` / `HYDRA_COOKIE_SECRET` invalidates existing
  OIDC sessions and consent — coordinate with downstream relying parties.
- Rotate `GEMINI_API_KEY` and the OAuth / `EID_RP_SECRET` / `CORE_API_TOKEN`
  credentials from their consoles, update `backend.env` / `.env`, then
  `docker compose up -d api web`.

---

**Government Template Platform V3.0** — Co-developed by the **Gerege Systems Development Team** and **Claude AI**, 2026.
