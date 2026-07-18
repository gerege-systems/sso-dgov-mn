# Deploy хийх заавар

> 🌐 [English](DEPLOYMENT.md) · **Монгол**

**Government SSO** (sso.dgov.mn)-ийг нэг VPS дээр Docker Compose-оор,
nginx-ийн ард deploy хийх заавар. Стек нь Postgres + Redis + Go API + Next.js
BFF web + **Ory Hydra** (dan-ийг SSO provider болгодог OIDC issuer). Жишиг
deployment-д ашигласан бодит runbook.

## Топологи

Хост дээр гурван loopback port гаргана; nginx нь TLS-ыг төгсгөж, тус бүрийг
зөв контейнер руу reverse-proxy хийнэ. `db`, `redis` нь дотоод compose
сүлжээнээс хэзээ ч гарахгүй, Hydra-гийн **admin** API нь зөвхөн loopback (хэзээ
ч проксилохгүй).

```
Internet ──► nginx (80/443, Let's Encrypt TLS)
   │
   ├─ /oauth2/*, /.well-known/openid-configuration, /userinfo, /health/ready
   │      ─────────────────────────► hydra  127.0.0.1:${HYDRA_PUBLIC_PORT}   (Ory Hydra — OIDC issuer, PUBLIC API)
   │
   ├─ /rp/sign/*  (3 дагч Relying Party-ийн eID sign relay)
   │      ─────────────────────────► api    127.0.0.1:${API_RELAY_PORT}      (backend :8080, loopback relay)
   │
   └─ бусад бүх зүйл — app, BFF /api/*, ба OIDC login/consent UI
      (/oauth/login, /oauth/consent, /oauth/logout, /oauth/error)
          ─────────────────────────► web    127.0.0.1:${WEB_PORT}            (Next.js BFF)
                                       │ BACKEND_URL=http://api:8080
                                       ▼
   дотоод compose сүлжээ (нийтийн host port байхгүй):
        api ──► db (Postgres 16 — gerege_template + hydra database) + redis (7)
        hydra ──► db (hydra database)   admin :4445 = ЗӨВХӨН LOOPBACK, хэзээ ч проксилохгүй
        hydra-migrate (нэг удаа), migrate (нэг удаа) — schema түрхээд гардаг
```

Тэгэхээр `web` нь гадагш нээгддэг ЦОРЫН ГАНЦ контейнер **биш**: nginx нь Hydra-гийн
public API (`:4444`) болон api sign relay (`:8091`)-ыг мөн урдаас барих ёстой.
Browser нь app болон BFF-д `web`-ээр хүрнэ; OIDC протоколын endpoint-уудыг Hydra
үйлчилнэ; OAuth *login/consent* хуудсууд (dan өөрөө иргэнийг eID-ээр
баталгаажуулаад render хийдэг) нь `web` дээр `/oauth/*` дор байрлана. Нэг удаагийн
`migrate` контейнер `up` бүр дээр SQL migration-уудыг түрхэнэ; `hydra-migrate` нь
Hydra-гийн өөрийн schema-г тусдаа `hydra` database руу түрхээд гардаг.

## Шаардлага

- Docker + compose plugin-тэй VPS (`docker compose version`)
- Хост дээр nginx + certbot (эсвэл TLS terminate хийдэг дурын reverse proxy)
- Сервер рүү заасан `sso.dgov.mn` DNS бичлэг

## 1. Кодоо татах

```bash
git clone https://github.com/gerege-systems/dan-dgov-mn.git /srv/dan
cd /srv/dan
```

## 2. Хоёр env файл үүсгэх (хоёулаа gitignored)

### `./.env` — compose interpolation

Compose-ийн interpolate хийдэг бүхэн энд байна. **REQUIRED** гэж тэмдэглэсэн
Hydra нууцууд `docker-compose.yml`-д `${VAR:?}` хэлбэртэй тул тэдгээрийг
тохируулаагүй/хоосон бол **compose асахаас татгалзана**.

```env
# --- Postgres / Redis ---
POSTGRES_USER=postgres            # superuser — зөвхөн migrate + hydra-migrate хэрэглэнэ
POSTGRES_PASSWORD=<санамсаргүй>
POSTGRES_DB=gerege_template
APP_DB_USER=app_user              # api-ийн холбогддог хамгийн бага эрхт role
APP_DB_PASSWORD=<санамсаргүй>
APP_DB_DSN=host=db port=5432 user=app_user password=<мөн адил> dbname=gerege_template sslmode=disable
REDIS_PASS=<санамсаргүй>

# --- App / origin ---
APP_ORIGIN=https://sso.dgov.mn    # яг нийтийн origin (CSRF origin шалгалт)
WEB_PORT=3007                     # nginx app руу проксилдог loopback port
API_RELAY_PORT=8091               # nginx /rp/sign-ыг проксилдог loopback port (api :8080)

# --- Ory Hydra (OIDC issuer) ---
HYDRA_PUBLIC_PORT=4444            # nginx OIDC public API руу проксилдог loopback port
HYDRA_ADMIN_PORT=4445             # Hydra admin API — loopback дээр, ХЭЗЭЭ Ч проксилохгүй
HYDRA_PUBLIC_URL=https://sso.dgov.mn          # REQUIRED — OIDC issuer / self URL
HYDRA_POST_LOGOUT_REDIRECT=https://sso.dgov.mn/   # заавал биш; default нь HYDRA_PUBLIC_URL/
HYDRA_SYSTEM_SECRET=<≥32 санамсаргүй тэмдэгт>  # REQUIRED — Hydra system secret
HYDRA_COOKIE_SECRET=<≥32 санамсаргүй тэмдэгт>  # REQUIRED — Hydra cookie secret
HYDRA_PAIRWISE_SALT=<санамсаргүй>              # REQUIRED — pairwise subject salt

# --- web BFF-ийн хэрэглэдэг OAuth client ID/secret (хоосон = тэр товч/карт идэвхгүй) ---
GOOGLE_CLIENT_ID=<…>              # Google account холболт (backend.env-д мөн тавина)
GOOGLE_DRIVE_CLIENT_ID=<…>        # гуравдагч интеграци; token exchange-ыг BFF хийдэг тул
GOOGLE_DRIVE_CLIENT_SECRET=<…>    # secret ч энд орно.
DROPBOX_CLIENT_ID=<…>             # redirect_uri = ${APP_ORIGIN}/api/integrations/<provider>/callback
DROPBOX_CLIENT_SECRET=<…>
GOOGLE_MEET_CLIENT_ID=<…>
GOOGLE_MEET_CLIENT_SECRET=<…>
```

### `./backend.env` — `api` + `migrate`-д `/app/.env` болж mount хийгдэнэ

Энэ нь backend-ийн config файл (viper уншина). eID Relying-Party креденшл, SSO/OIDC
provider тохиргоо болон бүх интеграцийн нууцыг агуулна. Бүрэн schema нь
`backend/internal/config/config.go`; eID SSO deployment-ийн гол түлхүүрүүд:

```env
# --- Үндсэн runtime ---
PORT=8080
ENVIRONMENT=development           # compose стек dev горимоор ажиллана: дотоод DB
                                  # TLS-гүй (prod guard нь sslmode=verify-full
                                  # шаарддаг); TLS нь nginx дээр төгсдөг
DEBUG=false
DB_POSTGRE_DRIVER=postgres
DB_POSTGRE_DSN=postgres://postgres:<POSTGRES_PASSWORD>@db:5432/gerege_template?sslmode=disable
                                  # ^ superuser DSN — MIGRATE (DDL) хэрэглэнэ.
                                  # api-д APP_DB_DSN-ээр дарж бичигдэнэ (§3-ыг үз).
JWT_SECRET=<≥32 санамсаргүй тэмдэгт>
JWT_EXPIRED=24                    # цаг (1–24)
JWT_ISSUER=sso.dgov.mn
JWT_REFRESH_EXPIRED=7             # хоног
BCRYPT_COST=12
OTP_MAX_ATTEMPTS=5
REDIS_HOST=redis:6379
REDIS_PASS=<.env-тэй ижил>
REDIS_EXPIRED=5                   # минут
ALLOWED_ORIGINS=https://sso.dgov.mn
TRUSTED_PROXIES=172.16.0.0/12,127.0.0.1   # XFF-д зөвхөн docker сүлжээ + nginx-ээс итгэнэ.
                                  # Proxy-гийн ард ЗААВАЛ: api нийтийн app порт-гүй тул
                                  # хүсэлт web/nginx peer-ээс ирнэ. Итгэмжит proxy
                                  # жагсаалтгүй бол api нь X-Forwarded-For-ыг үл тоож,
                                  # per-IP rate-limit бүгд нэг bucket-д уначихна.

# --- eID Relying Party (ЦОРЫН ГАНЦ интерактив нэвтрэх арга) ---
EID_BASE_URL=https://eidmongolia.mn/v3   # eID IdP base (default)
EID_RP_UUID=<eID Mongolia-гийн олгосон RP UUID>
EID_RP_NAME=dan-dgov-mn
EID_RP_SECRET=<RP secret>
EID_CERT_LEVEL=ADVANCED           # нэвтрэлтэд ADVANCED (гарын үсэгт QUALIFIED/QSCD)
EID_CALLBACK_URL=https://sso.dgov.mn/login/verify   # IdP-ийн allowlist-д байх ёстой
EID_DISPLAY_TEXT=sso.dgov.mn

# --- Google OAuth (eID account холболт; server талд code exchange) ---
GOOGLE_CLIENT_ID=<…>
GOOGLE_CLIENT_SECRET=<…>

# --- dgov SSO consumer (sso.dgov.mn OIDC — eID-ийн зэрэгцээ 2 дахь нэвтрэлт) ---
SSO_ISSUER=https://sso.dgov.mn
SSO_CLIENT_ID=<…>
SSO_CLIENT_SECRET=<…>
SSO_REDIRECT_URI=https://sso.dgov.mn/sso/callback
SSO_SCOPE=openid profile email
SSO_NATIVE_CLIENT_ID=dan-dgov-mn-ios   # mobile PKCE урсгалын Hydra client_id

# --- OIDC PROVIDER тал (dan нь Ory Hydra-г SSO issuer болгож урдаа тавина) ---
HYDRA_ADMIN_URL=http://hydra:4445      # admin API (client CRUD + login/consent/logout)
HYDRA_PUBLIC_URL=https://sso.dgov.mn   # redirect байгуулахад ашиглах issuer
SSO_STATE_KEY=<≥32 санамсаргүй тэмдэгт> # login/consent state cookie HMAC түлхүүр
SSO_FIRSTPARTY_CLIENTS=<csv client_id>    # эдгээрт consent дэлгэц алгасна
SSO_ADMIN_API_KEYS=<csv bootstrap key>    # /admin гадаргуугийн bootstrap key
SSO_ADMIN_SUBS=<csv eid_sub>              # superadmin эрхтэй eid_sub-ууд

# --- Gerege платформын үйлчилгээ ---
XYP_API_BASE=https://xyp.dgov.mn       # байгууллагын лавлагаа (HTTP Basic; сонголттой)
XYP_CLIENT_ID=<…>
XYP_CLIENT_SECRET=<…>
CORE_API_BASE=https://core.dgov.mn     # user/org хайлт
CORE_API_TOKEN=<service bearer>
GSPACE_HOST=<sftp host>                # Gerege Space хэрэглэгч тус бүрийн SFTP хадгалалт (сонголттой)
GSPACE_PORT=22
GSPACE_USER=<…>
GSPACE_PASSWORD=<…>
GSPACE_BASE_PATH=gerege-space
GSPACE_QUOTA_BYTES=2097152             # хэрэглэгч тус бүр 2 MB

# --- Шифрлэлт / гарын үсэг / observability ---
INTEGRATION_ENC_KEY=<≥32 санамсаргүй тэмдэгт> # хадгалсан OAuth токенд AES-256-GCM түлхүүр
SIGN_RELAY_TOKEN=<shared token>        # 3 дагч RP-д /rp/sign relay-г идэвхжүүлнэ (хоосон = унтраалттай)
SIGN_SIGNER_CERT_FILE=/app/certs/signer.crt   # PAdES document-signer гэрчилгээ (prod: REQUIRED,
SIGN_SIGNER_KEY_FILE=/app/certs/signer.key    #  fail-closed; dev-д self-signed руу шилжинэ)
OBSERVABILITY_TOKEN=<санамсаргүй>      # prod-д /metrics + /swagger/doc.json-ий bearer
GEMINI_API_KEY=<AIza…>                 # AI боломжууд; хоосон бол AI endpoint 500
```

Нууцуудыг `openssl rand -hex 24` (эсвэл `≥32` түлхүүрт `-hex 32`)-өөр үүсгэ.
`SIGN_SIGNER_CERT_FILE` / `SIGN_SIGNER_KEY_FILE` нь контейнер **дотор**-х замууд —
хэрэв тохируулбал PEM файлуудыг mount хий (жишээ `api` service-д read-only volume
нэм); compose dev стект хоосон үлдэж болох ба signer нь dev self-signed түлхүүр
хэрэглэнэ.

## 3. Яагаад хоёр DB role вэ (анхны boot-оос ӨМНӨ унш)

Row-Level Security-г superuser **чимээгүй алгасдаг**. Тиймээс стек хоёр role
ашиглана:

- `migrate` (болон `hydra-migrate`) нь `POSTGRES_USER`-ээр (superuser —
  `CREATE EXTENSION`, RLS DDL, `hydra` database үүсгэхэд хэрэгтэй) холбогдоно.
- `api` нь `APP_DB_USER`-ээр (`NOSUPERUSER NOBYPASSRLS`) холбогдоно —
  **хоосон data volume-ийн анхны init дээр**
  `backend/deploy/initdb/10-create-app-user.sh` автоматаар үүсгэдэг. Хоёр дахь
  initdb script `20-create-hydra-db.sh` нь Ory Hydra-д зориулж тусдаа `hydra`
  database үүсгэдэг.

api үүнийг **boot үед шалгадаг**: role нь superuser/BYPASSRLS бол production горимд
асахаас татгалзаж, development горимд warning логдоно. *Одоо байгаа* DB рүү deploy
хийж байгаа бол app role + grant-уудыг гараар үүсгээд (initdb script-ийг үз),
`hydra` database-ыг үүсгээд
(`docker compose exec db psql -U "$POSTGRES_USER" -c 'CREATE DATABASE hydra;'`),
`APP_DB_DSN`-ийг app role руу заа.

## 4. Анхны deploy

```bash
docker compose up -d --build      # api+web бүтээж, хоёр migrate job-ыг ажиллуулж, бүгдийг асаана
docker compose ps                 # db/redis/api/web/hydra healthy эсвэл running,
                                  # migrate + hydra-migrate Exited (0) байх ёстой
```

### nginx vhost (хост дээр)

OIDC issuer замууд Hydra руу, `/rp/sign` нь api relay руу, бусад бүхэн `web` руу
очно. Hydra admin port (`:4445`)-ыг энд **хэзээ ч** бичихгүй.

```nginx
upstream dan_web   { server 127.0.0.1:3007; }   # = WEB_PORT
upstream dan_hydra { server 127.0.0.1:4444; }   # = HYDRA_PUBLIC_PORT
upstream dan_relay { server 127.0.0.1:8091; }   # = API_RELAY_PORT (api :8080)

server {
    server_name sso.dgov.mn;

    # OIDC протоколын endpoint → Ory Hydra public API
    location /oauth2/                         { proxy_pass http://dan_hydra; include /etc/nginx/proxy_params; }
    location = /userinfo                      { proxy_pass http://dan_hydra; include /etc/nginx/proxy_params; }
    location /.well-known/openid-configuration { proxy_pass http://dan_hydra; include /etc/nginx/proxy_params; }
    location = /.well-known/jwks.json         { proxy_pass http://dan_hydra; include /etc/nginx/proxy_params; }

    # 3 дагч Relying Party-ийн eID sign relay → api loopback relay
    location /rp/sign/                        { proxy_pass http://dan_relay; include /etc/nginx/proxy_params; }

    # App, BFF /api/*, ба /oauth/login|consent|logout UI → web BFF
    location / {
        proxy_pass http://dan_web;
        proxy_set_header Host $host;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

(Хуваалцсан `proxy_set_header` мөрүүдийг `/etc/nginx/proxy_params`-д хийж
`include` хий, эсвэл block бүрт давт.) Дараа нь TLS:
`certbot --nginx -d sso.dgov.mn`. Compose файл `COOKIE_SECURE=true` тавьдаг ба
Hydra нь `SERVE_COOKIES_SAME_SITE_MODE=None` (Secure шаарддаг)-оор ажилладаг тул
сайт **заавал HTTPS-ээр** үйлчлэх ёстой — эс бөгөөс browser auth болон OIDC
cookie-г хадгалахгүй.

## 5. Ажиллаж буй deployment-ийг шинэчлэх

```bash
cd /srv/dan
git pull --ff-only origin main
docker compose build              # api + web + migrate
docker compose up -d              # өөрчлөгдсөн контейнеруудыг сэргээнэ; migrate + hydra-migrate
                                  # дахин ажиллана (түрхэгдсэн migration-уудыг алгасна)
```

`db`, `redis` хэвээр ажиллана — өгөгдөл хөндөгдөхгүй. Зөвхөн тохиргоо өөрчилсөн
бол: `backend.env` / `.env`-ээ засаад `docker compose up -d api web` (хэрэв
`HYDRA_*` утга өөрчилсөн бол `hydra`-г мөн дахин асаа).

### Автомат deploy (CI/CD)

Deploy нь CI дотор job **биш**. Хоёр workflow гинжлэгдэнэ:

1. [`.github/workflows/ci.yml`](../.github/workflows/ci.yml) — pre-flight gate-үүд
   (`backend`, `frontend`, `secrets-scan`) нь `main`-д push бүр болон PR бүр дээр
   ажиллана.
2. [`.github/workflows/deploy.yml`](../.github/workflows/deploy.yml) — **CI дууссаны
   дараа** `workflow_run`-аар өдөөгдөх **тусдаа** workflow, ингэснээр CI ба Deploy
   зэрэгцэн ажиллахаа больж, унасан build хэзээ ч ship хийгдэхгүй. Зөвхөн гинжлэгдсэн
   CI ажиллагаа `main` дээр `success` дүгнэгдсэн үед (эсвэл гараар
   `workflow_dispatch`) deploy хийнэ. Тусгай **non-root** `deploy` хэрэглэгчээр VPS
   руу SSH-ээр орж, CI давсан яг тэр commit руу `git reset --hard` хийгээд
   [`deploy/deploy.sh`](../deploy/deploy.sh)-ийг ажиллуулна (rebuild → `up -d` →
   эрүүл болтол хүлээх → prune). `db`/`redis` тасрахгүй; migration дахин ажиллаж
   түрхэгдсэн файлуудыг алгасна.

Нэг удаагийн тохиргоо — **Settings → Secrets and variables → Actions** дор дараах
repo secret-уудыг нэмнэ:

| Secret | Утга |
|--------|------|
| `DEPLOY_HOST` | VPS-ийн IP / hostname |
| `DEPLOY_USER` | repo checkout-ыг эзэмшдэг, docker ажиллуулах эрхтэй тусгай **non-root** SSH хэрэглэгч (`deploy`) |
| `DEPLOY_PATH` | серверийн repo зам; `deploy.yml`-д тохируулаагүй бол default нь `/srv/dan` |
| `DEPLOY_SSH_KEY` | тусгай deploy keypair-ийн **private** түлхүүр; public түлхүүр нь серверийн `~/.ssh/authorized_keys`-д байна |
| `DEPLOY_PORT` | *(заавал биш)* SSH порт, default нь `22` |

Keypair-ийг `ssh-keygen -t ed25519 -f deploy_key -N ''`-ээр үүсгэж,
`deploy_key.pub`-ийг `deploy` хэрэглэгчийн `authorized_keys`-д нэмээд, private
`deploy_key`-г `DEPLOY_SSH_KEY`-д хийнэ. Код өөрчлөхгүйгээр Actions таб-аас гараар
deploy дуудаж болно (**Run workflow** — `workflow_dispatch` нь `origin/main` HEAD-ыг
deploy хийнэ), эсвэл сервер дээр `bash deploy/deploy.sh`-ийг гараар ажиллуулж болно.

## 6. Баталгаажуулах

```bash
docker compose ps                                       # бүгд healthy / migrate job Exited(0)
docker logs dan-dgov-mn-migrate-1 | tail -3             # "migration [up] success"
docker logs dan-dgov-mn-hydra-migrate-1 | tail -3       # Hydra schema түрхэгдсэн
docker logs dan-dgov-mn-api-1 2>&1 | grep -i error      # хоосон байх ёстой
curl -s -o /dev/null -w '%{http_code}\n' https://sso.dgov.mn/   # 200
curl -s https://sso.dgov.mn/.well-known/openid-configuration | head -c 80   # OIDC issuer JSON
```

## 7. Буцаах (Rollback)

```bash
git log --oneline                 # сүүлийн зөв commit-оо ол
git checkout <commit> -- .        # эсвэл: git reset --hard <commit>
docker compose build && docker compose up -d
```

Энэ урсгалд SQL migration зөвхөн урагшаа; migration буцаах шаардлагатай бол
тохирох `N_*.down.sql`-ийг гараар түрхээд дараа нь кодоо буцаана.

## Нууцын эрүүл ахуй

- `.env`, `backend.env` gitignored — хэзээ ч commit хийхгүй.
- `JWT_SECRET` солих = бүх хэрэглэгчийг хүчээр logout хийнэ (бүх токен хүчингүй).
- `HYDRA_SYSTEM_SECRET` / `HYDRA_COOKIE_SECRET` солих нь одоо байгаа OIDC session
  болон consent-ыг хүчингүй болгоно — доод талын relying party-уудтай зохицуул.
- `GEMINI_API_KEY` болон OAuth / `EID_RP_SECRET` / `CORE_API_TOKEN` креденшлүүдийг
  консолоос нь rotate хийгээд `backend.env` / `.env`-д сольж
  `docker compose up -d api web` хийнэ.

---

**Government Template Platform V3.0** — **Gerege Systems Development Team** болон **Claude AI** хамтран бүтээв, 2026.
