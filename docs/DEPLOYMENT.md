# Deployment Guide

> 🌐 **English** · [Монгол](DEPLOYMENT_MN.md)

How to deploy **Government SSO** (sso.dgov.mn) to a single VPS with Docker
Compose behind nginx. The stack is Postgres + Redis + Go API + Next.js BFF web;
the **OAuth2/OIDC provider is implemented inside the Go API itself** (there is
no separate issuer process). This is the runbook used for the reference
deployment.

## Topology

Two host loopback ports are published; nginx terminates TLS and reverse-proxies
each to the right container. `db` and `redis` never leave the internal compose
network. Compose runs exactly five services: `db`, `migrate` (one-off), `redis`,
`api`, `web`.

```
Internet ──► nginx (80/443, TLS via Let's Encrypt)
   │
   ├─ /oauth2/*, /.well-known/openid-configuration, /.well-known/jwks.json, /userinfo
   │  and /rp/sign/*  (eID sign relay for 3rd-party Relying Parties)
   │      ─────────────────────────► api    127.0.0.1:${API_RELAY_PORT}      (backend :8080 — OIDC issuer + relay)
   │
   └─ everything else — app, BFF /api/*, and the OIDC login/consent UI
      (/oauth/login, /oauth/consent, /oauth/logout, /oauth/error)
          ─────────────────────────► web    127.0.0.1:${WEB_PORT}            (Next.js BFF)
                                       │ BACKEND_URL=http://api:8080
                                       ▼
   internal compose network (no public host ports):
        api ──► db (Postgres 16 — single gerege_template database) + redis (7)
        migrate (one-off) — apply SQL migrations, then exit
```

So `web` is **not** the only exposed container: nginx must also front the api
loopback port (`API_RELAY_PORT`, `8081` in this deployment), which now serves both
the OIDC protocol endpoints and
the sign relay. The browser reaches `web` for the app and its BFF; the OAuth
*login/consent* pages (which dan renders itself, after authenticating the
citizen with eID) live on `web` under `/oauth/*`. A one-off `migrate` container
applies SQL migrations on every `up`.

The endpoints the api serves directly (canonical paths in
`backend/internal/business/usecases/oidc/discovery.go`):

| Endpoint | Purpose |
|----------|---------|
| `/oauth2/auth` | authorization endpoint |
| `/oauth2/token` | token endpoint |
| `/oauth2/revoke` | token revocation |
| `/oauth2/introspect` | token introspection |
| `/oauth2/sessions/logout` | RP-initiated logout (end session) |
| `/userinfo` | UserInfo |
| `/.well-known/openid-configuration` | discovery document |
| `/.well-known/jwks.json` | id_token verification keys |

Access tokens stay **opaque** (relying parties validate them at
`/oauth2/introspect`); id_tokens are **RS256 JWTs** verifiable offline against
`/.well-known/jwks.json`.

There is **no separate admin port** to firewall any more: client CRUD and the
login/consent/logout core are internal calls inside the api, guarded by the
regular route middleware.

## Prerequisites

- A VPS with Docker + the compose plugin (`docker compose version`)
- nginx + certbot on the host (or any reverse proxy that terminates TLS)
- A DNS record for `sso.dgov.mn` pointing at the server

## 1. Get the code

```bash
git clone https://github.com/gerege-systems/dan-dgov-mn.git /srv/dan
cd /srv/dan
```

## 2. Create the two env files (both gitignored)

### `./.env` — compose interpolation

Everything compose interpolates lives here.

```env
# --- Postgres / Redis ---
POSTGRES_USER=postgres            # superuser — used by migrate only
POSTGRES_PASSWORD=<random>
POSTGRES_DB=gerege_template
APP_DB_USER=app_user              # least-privilege role the api connects as
APP_DB_PASSWORD=<random>
APP_DB_DSN=host=db port=5432 user=app_user password=<same> dbname=gerege_template sslmode=disable
REDIS_PASS=<random>

# --- App / origin ---
APP_ORIGIN=https://sso.dgov.mn    # exact public origin (CSRF origin check)
WEB_PORT=3007                     # loopback port nginx proxies the app to
API_RELAY_PORT=8081               # the ONE loopback api port nginx fronts — OIDC,
                                  # and /rp/sign to (api :8080)

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
JWT_ISSUER=sso.dgov.mn
JWT_REFRESH_EXPIRED=7             # days
BCRYPT_COST=12
OTP_MAX_ATTEMPTS=5
REDIS_HOST=redis:6379
REDIS_PASS=<same as .env>
REDIS_EXPIRED=5                   # minutes
ALLOWED_ORIGINS=https://sso.dgov.mn
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
EID_CALLBACK_URL=https://sso.dgov.mn/login/verify   # must be allowlisted at the IdP
EID_DISPLAY_TEXT=sso.dgov.mn

# --- Google OAuth (eID account-linking; server-side code exchange) ---
GOOGLE_CLIENT_ID=<…>
GOOGLE_CLIENT_SECRET=<…>

# --- dgov SSO consumer (sso.dgov.mn OIDC — 2nd login alongside eID) ---
SSO_ISSUER=https://sso.dgov.mn
SSO_CLIENT_ID=<…>
SSO_CLIENT_SECRET=<…>
SSO_REDIRECT_URI=https://sso.dgov.mn/sso/callback
SSO_SCOPE=openid profile email
SSO_NATIVE_CLIENT_ID=dan-dgov-mn-ios   # provider client_id for the mobile PKCE flow

# --- OIDC PROVIDER side (dan IS the OAuth2/OIDC issuer — served by the api) ---
OAUTH_ISSUER=https://sso.dgov.mn       # issuer: the `iss` claim, the discovery
                                       # document and every advertised endpoint URL
                                       # are built from it. Must match exactly what
                                       # relying parties have configured.
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
INTEGRATION_ENC_KEY=<≥32 random chars> # AES-256-GCM key for stored OAuth tokens AND
                                       # for the id_token RSA signing key at rest
                                       # (`oauth_signing_keys`) — see "Secrets hygiene"
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

- `migrate` connects as `POSTGRES_USER` (superuser — needed for
  `CREATE EXTENSION` and RLS DDL).
- `api` connects as `APP_DB_USER` (`NOSUPERUSER NOBYPASSRLS`), created
  automatically by `backend/deploy/initdb/10-create-app-user.sh` **on first init
  of an empty data volume**.

The api **verifies this at boot**: if its role is superuser/BYPASSRLS it fails to
start in production mode and logs a warning in development mode. If you deploy
onto an *existing* database, create the app role + grants by hand (see the initdb
script) and point `APP_DB_DSN` at the app role.

The OAuth2/OIDC protocol state — authorization codes, access and refresh tokens,
login/consent challenges and stored consents — lives in this same
`gerege_template` database under RLS (service/admin/self policies), so there is
no second database to provision. Codes and tokens are persisted **only as
sha256 hashes**; the plaintext value exists just long enough to hand it to the
client.

## 4. First deploy

```bash
docker compose up -d --build      # builds api+web, runs the migrate job, starts all
docker compose ps                 # expect: db/redis/api/web healthy or running,
                                  #         migrate Exited (0)
```

### nginx vhost (host)

The OIDC issuer paths and `/rp/sign` both reach the api loopback port; everything
else goes to `web`.

```nginx
upstream dan_web { server 127.0.0.1:3007; }   # = WEB_PORT
upstream dan_api { server 127.0.0.1:8081; }   # = API_RELAY_PORT (api :8080).
                                              # ONE api port: OIDC, /rp/eid/ and
                                              # /rp/sign/ all share it — the OIDC
                                              # endpoints did not add a new port.

server {
    server_name sso.dgov.mn;

    # OIDC protocol endpoints → the api (this backend IS the issuer)
    location /oauth2/                         { proxy_pass http://dan_api; include /etc/nginx/proxy_params; }
    location = /userinfo                      { proxy_pass http://dan_api; include /etc/nginx/proxy_params; }
    location /.well-known/openid-configuration { proxy_pass http://dan_api; include /etc/nginx/proxy_params; }
    location = /.well-known/jwks.json         { proxy_pass http://dan_api; include /etc/nginx/proxy_params; }

    # eID sign relay for 3rd-party Relying Parties → same api upstream
    location /rp/sign/                        { proxy_pass http://dan_api; include /etc/nginx/proxy_params; }

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
`include` them, or repeat them per block.) Then `certbot --nginx -d sso.dgov.mn`
for TLS. The compose file sets `COOKIE_SECURE=true`, so the site **must** be
served over HTTPS or browsers will drop the auth and OIDC cookies.

> **Upgrading an existing vhost.** The location blocks are unchanged — the
> in-backend provider serves the exact same paths. Only the upstream moves:
> delete `upstream dan_hydra { server 127.0.0.1:4444; }` and repoint the four
> OIDC blocks at the api upstream. Ports `4444` / `4445` no longer exist, and the
> old rule "never proxy the admin port" is moot — there is no admin port.

## 5. Updating a running deployment

```bash
cd /srv/dan
git pull --ff-only origin main
docker compose build              # api + web + migrate
docker compose up -d              # recreates changed containers; migrate re-runs
                                  # (already-applied migrations are skipped)
```

`db` and `redis` keep running — data is untouched. Config-only change? Edit
`backend.env` / `.env` and `docker compose up -d api web`.

### Cutting over from the previous Ory Hydra stack

If the server previously ran the Hydra-based stack, move the registered OAuth2
clients into the api's own `oauth_clients` table **before** the cutover:

```bash
./scripts/migrate-hydra-clients.sh    # DRY_RUN=1 … to inspect without writing
```

Client secrets are copied as-is, so relying parties keep working without
reconfiguration. Afterwards the `hydra` / `hydra-migrate` containers, the
separate `hydra` database and every `HYDRA_*` variable can be dropped, and the
nginx vhost repointed (see §4).

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
docker compose ps                                       # all healthy / migrate job Exited(0)
docker logs dan-dgov-mn-migrate-1 | tail -3             # "migration [up] success"
docker logs dan-dgov-mn-api-1 2>&1 | grep -i error      # should be empty
curl -s -o /dev/null -w '%{http_code}\n' https://sso.dgov.mn/   # 200
curl -s https://sso.dgov.mn/.well-known/openid-configuration | head -c 80   # OIDC issuer JSON
curl -s https://sso.dgov.mn/.well-known/jwks.json | head -c 80              # RS256 public key(s)
```

The discovery `issuer` must equal `OAUTH_ISSUER` exactly, and `jwks.json` must
return at least one RS256 key. Key provisioning is fail-closed: the api creates
the signing key on first boot and **refuses to start** if it cannot (look for
`oidc: ensure signing key` in the api logs — usually a bad
`INTEGRATION_ENC_KEY`).

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
- Rotating `SSO_STATE_KEY` invalidates in-flight login/consent state — any
  browser mid-authorization has to restart the flow.
- **`INTEGRATION_ENC_KEY` now has a second job.** Besides sealing stored
  third-party OAuth tokens, it encrypts the id_token signing key (RSA-2048,
  generated on first boot) at rest in `oauth_signing_keys` (AES-256-GCM).
  Rotating it makes the stored signing key **undecryptable** — and the api does
  *not* self-heal, because an active key row still exists, so id_token issuance
  fails until the key is replaced. To recover, retire/delete the active
  `oauth_signing_keys` row so the next boot generates a fresh one; the `kid`
  changes, relying parties must re-fetch `/.well-known/jwks.json`, and stored
  integration tokens need re-linking. Plan an `INTEGRATION_ENC_KEY` rotation as
  an announced issuer key rollover, not a routine secret swap.
- Rotate `GEMINI_API_KEY` and the OAuth / `EID_RP_SECRET` / `CORE_API_TOKEN`
  credentials from their consoles, update `backend.env` / `.env`, then
  `docker compose up -d api web`.

---

**Government Template Platform V3.0** — Co-developed by the **Gerege Systems Development Team** and **Claude AI**, 2026.
