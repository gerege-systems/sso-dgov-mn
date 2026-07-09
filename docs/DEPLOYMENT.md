# Deployment Guide

> 🌐 **English** · [Монгол](DEPLOYMENT_MN.md)

How to deploy the full stack (Postgres + Redis + Go API + Next.js web) to a
single VPS with Docker Compose behind nginx. This is the runbook used for the
reference deployment.

## Topology

```
Internet ──► nginx (80/443, TLS via Let's Encrypt)
                │ proxy_pass
                ▼
        web  127.0.0.1:${WEB_PORT}   (Next.js BFF — the ONLY exposed container)
                │ BACKEND_URL=http://api:8080   (internal compose network)
                ▼
        api ──► db (Postgres 16) + redis (7)    (no public ports)
```

The browser only ever reaches `web`; `api`, `db`, `redis` stay on the
internal compose network. A one-off `migrate` container applies SQL
migrations on every `up` and exits.

## Prerequisites

- A VPS with Docker + the compose plugin (`docker compose version`)
- nginx + certbot on the host (or any reverse proxy that terminates TLS)
- A DNS record pointing at the server

## 1. Get the code

```bash
git clone https://github.com/gerege-systems/template-gerege-mn.git /srv/template-gerege-mn
cd /srv/template-gerege-mn
```

## 2. Create the two env files (both gitignored)

### `./.env` — compose interpolation

```env
POSTGRES_USER=postgres            # superuser — used by migrate only
POSTGRES_PASSWORD=<random>
POSTGRES_DB=gerege_template
APP_DB_USER=app_user              # least-privilege role the api connects as
APP_DB_PASSWORD=<random>
APP_DB_DSN=host=db port=5432 user=app_user password=<same> dbname=gerege_template sslmode=disable
REDIS_PASS=<random>
APP_ORIGIN=https://your.domain.mn # exact public origin (CSRF origin check)
WEB_PORT=3007                     # loopback port nginx proxies to
```

### `./backend.env` — mounted into `api` + `migrate` at `/app/.env`

```env
PORT=8080
ENVIRONMENT=development           # the compose stack runs dev mode: internal
                                  # DB has no TLS (the prod guard requires
                                  # sslmode=verify-full); TLS terminates at nginx
DEBUG=false
DB_POSTGRE_DRIVER=postgres
DB_POSTGRE_DSN=postgres://postgres:<POSTGRES_PASSWORD>@db:5432/gerege_template?sslmode=disable
                                  # ^ superuser DSN — used by MIGRATE (DDL).
                                  # The api overrides this with APP_DB_DSN.
JWT_SECRET=<≥32 random chars>
JWT_EXPIRED=24                    # hours
JWT_ISSUER=your.domain.mn
JWT_REFRESH_EXPIRED=7             # days
OTP_MAX_ATTEMPTS=5
BCRYPT_COST=12
REDIS_HOST=redis:6379
REDIS_PASS=<same as .env>
REDIS_EXPIRED=5                   # minutes (OTP request_id TTL)
ALLOWED_ORIGINS=https://your.domain.mn
TRUSTED_PROXIES=172.16.0.0/12,127.0.0.1   # trust XFF only from docker net + nginx.
                                  # REQUIRED behind the proxy: the api has no
                                  # public port, so every request arrives from
                                  # the web/nginx peer. The BFF forwards the real
                                  # client IP as X-Forwarded-For; without a
                                  # trusted-proxy list the api ignores it and all
                                  # per-IP rate limits collapse into one bucket.
VERIFY_API_BASE=https://verify.gecloud.mn/v1
VERIFY_API_KEY=<gck_live_…>       # email/SMS OTP — registration won't work without it
VERIFY_CHANNEL=email
GEMINI_API_KEY=<AIza…>            # AI features; empty = AI endpoints return 500
```

Generate secrets with `openssl rand -hex 24`.

## 3. Why two DB roles (read before first boot)

Row-Level Security is **silently bypassed** by superusers. The stack
therefore uses two roles:

- `migrate` connects as `POSTGRES_USER` (superuser — needed for
  `CREATE EXTENSION`, RLS DDL).
- `api` connects as `APP_DB_USER` (`NOSUPERUSER NOBYPASSRLS`), created
  automatically by `backend/deploy/initdb/10-create-app-user.sh` **on first
  init of an empty data volume**.

The api **verifies this at boot**: if its role is superuser/BYPASSRLS it
fails to start in production mode and logs a warning in development mode.
If you deploy onto an *existing* database, create the role + grants by hand
(see the initdb script) and point `APP_DB_DSN` at it.

## 4. First deploy

```bash
docker compose up -d --build      # builds api+web, runs migrations, starts all
docker compose ps                 # expect: db/redis/api/web healthy, migrate Exited (0)
```

### nginx vhost (host)

```nginx
upstream gerege_web { server 127.0.0.1:3007; }   # = WEB_PORT
server {
    server_name your.domain.mn;
    location / {
        proxy_pass http://gerege_web;
        proxy_set_header Host $host;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

Then `certbot --nginx -d your.domain.mn` for TLS. The compose file sets
`COOKIE_SECURE=true`, so the site **must** be served over HTTPS or browsers
will drop the auth cookies.

## 5. Updating a running deployment

```bash
cd /srv/template-gerege-mn
git pull --ff-only origin main
docker compose build              # api + web + migrate
docker compose up -d              # recreates changed containers; migrate
                                  # re-runs (already-applied files are skipped)
```

`db` and `redis` keep running — data is untouched. Config-only change?
Edit `backend.env` / `.env` and `docker compose up -d api web`.

### Automated deploys (CI/CD)

The steps above are wired into GitHub Actions so a push to `main`
auto-deploys. The `deploy` job in
[`.github/workflows/ci.yml`](../.github/workflows/ci.yml) runs **only after**
the `backend`, `frontend` and `secrets-scan` jobs pass, then SSHes into this
VPS and runs [`deploy/deploy.sh`](../deploy/deploy.sh) (rebuild → `up -d` →
wait-for-healthy → prune). `db`/`redis` stay up; migrations re-run and skip
already-applied files.

One-time setup — add three repo secrets under **Settings → Secrets and
variables → Actions**:

| Secret | Value |
|--------|-------|
| `DEPLOY_HOST` | the VPS IP / hostname |
| `DEPLOY_USER` | SSH user with the repo + docker (`root` in the reference deploy) |
| `DEPLOY_SSH_KEY` | **private** key of a dedicated deploy keypair; its public key is appended to the server's `~/.ssh/authorized_keys` |
| `DEPLOY_PORT` | *(optional)* SSH port, defaults to `22` |

Generate the keypair with `ssh-keygen -t ed25519 -f deploy_key -N ''`, append
`deploy_key.pub` to the server's `authorized_keys`, and paste the private
`deploy_key` into `DEPLOY_SSH_KEY`. You can trigger a deploy without a code
change from the Actions tab (**Run workflow** — `workflow_dispatch`), or run
`bash deploy/deploy.sh` on the server by hand.

## 6. Verify

```bash
docker compose ps                                      # all healthy
docker logs temp-gerege-mn-migrate-1 | tail -3         # "migration [up] success"
docker logs temp-gerege-mn-api-1 2>&1 | grep -i error  # should be empty
curl -s -o /dev/null -w '%{http_code}\n' https://your.domain.mn/   # 200
```

## 7. Rollback

```bash
git log --oneline                 # find the last good commit
git checkout <commit> -- .        # or: git reset --hard <commit>
docker compose build && docker compose up -d
```

SQL migrations are forward-only in this flow; if a migration must be
reverted, apply the matching `N_*.down.sql` by hand before rolling the code
back past it.

## Secrets hygiene

- `.env` and `backend.env` are gitignored — never commit them.
- Rotate `JWT_SECRET` to force-logout everyone (all tokens invalidate).
- Rotate `GEMINI_API_KEY` / `VERIFY_API_KEY` from their consoles, update
  `backend.env`, then `docker compose up -d api`.

---

**Government Template Platform V3.0** — Co-developed by the **Gerege Systems Development Team** and **Claude AI**, 2026.
