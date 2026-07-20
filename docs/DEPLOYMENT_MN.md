# Deploy хийх заавар

> 🌐 [English](DEPLOYMENT.md) · **Монгол**

**Government SSO** (sso.dgov.mn)-ийг нэг VPS дээр Docker Compose-оор,
nginx-ийн ард deploy хийх заавар. Стек нь Postgres + Redis + Go API + Next.js
BFF web; **OAuth2/OIDC provider нь Go backend дотроо хэрэгжсэн** (тусдаа issuer
процесс байхгүй). Жишиг deployment-д ашигласан бодит runbook.

## Топологи

Хост дээр хоёр loopback port гаргана; nginx нь TLS-ыг төгсгөж, тус бүрийг
зөв контейнер руу reverse-proxy хийнэ. `db`, `redis` нь дотоод compose
сүлжээнээс хэзээ ч гарахгүй. Compose-д яг таван service ажиллана: `db`,
`migrate` (нэг удаа), `redis`, `api`, `web`.

```
Internet ──► nginx (80/443, Let's Encrypt TLS)
   │
   ├─ /oauth2/*, /.well-known/openid-configuration, /.well-known/jwks.json, /userinfo
   │  болон /rp/sign/*  (3 дагч Relying Party-ийн eID sign relay)
   │      ─────────────────────────► api    127.0.0.1:${API_RELAY_PORT}      (backend :8080 — OIDC issuer + relay)
   │
   └─ бусад бүх зүйл — app, BFF /api/*, ба OIDC login/consent UI
      (/oauth/login, /oauth/consent, /oauth/logout, /oauth/error)
          ─────────────────────────► web    127.0.0.1:${WEB_PORT}            (Next.js BFF)
                                       │ BACKEND_URL=http://api:8080
                                       ▼
   дотоод compose сүлжээ (нийтийн host port байхгүй):
        api ──► db (Postgres 16 — ганц gerege_template database) + redis (7)
        migrate (нэг удаа) — SQL migration түрхээд гардаг
```

Тэгэхээр `web` нь гадагш нээгддэг ЦОРЫН ГАНЦ контейнер **биш**: nginx нь api-ийн
loopback портыг (`API_RELAY_PORT`, энэ deployment-д `8081`) мөн урдаас барих
ёстой — тэр порт одоо OIDC
протоколын endpoint болон sign relay хоёуланг үйлчилнэ. Browser нь app болон
BFF-д `web`-ээр хүрнэ; OAuth *login/consent* хуудсууд (dan өөрөө иргэнийг
eID-ээр баталгаажуулаад render хийдэг) нь `web` дээр `/oauth/*` дор байрлана.
Нэг удаагийн `migrate` контейнер `up` бүр дээр SQL migration-уудыг түрхэнэ.

api өөрөө үйлчилдэг endpoint-ууд (канон замууд нь
`backend/internal/business/usecases/oidc/discovery.go`-д):

| Endpoint | Зориулалт |
|----------|-----------|
| `/oauth2/auth` | authorization endpoint |
| `/oauth2/token` | token endpoint |
| `/oauth2/revoke` | токен цуцлах |
| `/oauth2/introspect` | токен шалгах (introspection) |
| `/oauth2/sessions/logout` | RP-ээс эхэлсэн logout (end session) |
| `/userinfo` | UserInfo |
| `/.well-known/openid-configuration` | discovery баримт |
| `/.well-known/jwks.json` | id_token шалгах нийтийн түлхүүрүүд |

Access token нь **opaque** хэвээр (RP-үүд `/oauth2/introspect`-ээр шалгана);
id_token нь **RS256 JWT** бөгөөд `/.well-known/jwks.json`-оор офлайн шалгагдана.

Хамгаалах шаардлагатай **тусдаа admin порт байхгүй боллоо**: client CRUD болон
login/consent/logout цөм нь api дотоод дуудлага болж, ердийн route
middleware-ээр хамгаалагдана.

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

Compose-ийн interpolate хийдэг бүхэн энд байна.

```env
# --- Postgres / Redis ---
POSTGRES_USER=postgres            # superuser — зөвхөн migrate хэрэглэнэ
POSTGRES_PASSWORD=<санамсаргүй>
POSTGRES_DB=gerege_template
APP_DB_USER=app_user              # api-ийн холбогддог хамгийн бага эрхт role
APP_DB_PASSWORD=<санамсаргүй>
APP_DB_DSN=host=db port=5432 user=app_user password=<мөн адил> dbname=gerege_template sslmode=disable
REDIS_PASS=<санамсаргүй>

# --- App / origin ---
APP_ORIGIN=https://sso.dgov.mn    # яг нийтийн origin (CSRF origin шалгалт)
WEB_PORT=3007                     # nginx app руу проксилдог loopback port
API_RELAY_PORT=8081               # nginx-ийн барьдаг ГАНЦ api порт — OIDC,
                                  # loopback port (api :8080)

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
SSO_NATIVE_CLIENT_ID=dan-dgov-mn-ios   # mobile PKCE урсгалын provider client_id

# --- OIDC PROVIDER тал (dan нь ӨӨРӨӨ OAuth2/OIDC issuer — api үйлчилнэ) ---
OAUTH_ISSUER=https://sso.dgov.mn       # issuer: id_token-ий `iss`, discovery баримт
                                       # болон зарлагдах бүх endpoint URL үүнээс
                                       # гарна. RP-үүдийн тохируулсантай ЯГ таарна.
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
INTEGRATION_ENC_KEY=<≥32 санамсаргүй тэмдэгт> # хадгалсан OAuth токен БОЛОН id_token-ий
                                       # RSA гарын үсгийн түлхүүрийг (`oauth_signing_keys`)
                                       # шифрлэх AES-256-GCM түлхүүр — "Нууцын эрүүл ахуй"-г үз
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

- `migrate` нь `POSTGRES_USER`-ээр (superuser — `CREATE EXTENSION`, RLS DDL-д
  хэрэгтэй) холбогдоно.
- `api` нь `APP_DB_USER`-ээр (`NOSUPERUSER NOBYPASSRLS`) холбогдоно —
  **хоосон data volume-ийн анхны init дээр**
  `backend/deploy/initdb/10-create-app-user.sh` автоматаар үүсгэдэг.

api үүнийг **boot үед шалгадаг**: role нь superuser/BYPASSRLS бол production горимд
асахаас татгалзаж, development горимд warning логдоно. *Одоо байгаа* DB рүү deploy
хийж байгаа бол app role + grant-уудыг гараар үүсгээд (initdb script-ийг үз),
`APP_DB_DSN`-ийг app role руу заа.

OAuth2/OIDC протоколын төлөв — authorization code, access/refresh token,
login/consent challenge болон хадгалсан consent — мөн энэ л `gerege_template`
database дотор RLS-тэйгээр (service/admin/self бодлого) байрлана, тиймээс хоёр
дахь database бэлдэх шаардлагагүй. Code болон token нь **зөвхөн sha256 hash**
хэлбэрээр хадгалагдана; ил утга нь client руу буцаах хормыг л оршино.

## 4. Анхны deploy

```bash
docker compose up -d --build      # api+web бүтээж, migrate job-ыг ажиллуулж, бүгдийг асаана
docker compose ps                 # db/redis/api/web healthy эсвэл running,
                                  # migrate Exited (0) байх ёстой
```

### nginx vhost (хост дээр)

OIDC issuer замууд болон `/rp/sign` хоёулаа api-ийн loopback порт руу, бусад
бүхэн `web` руу очно.

```nginx
upstream dan_web { server 127.0.0.1:3007; }   # = WEB_PORT
upstream dan_api { server 127.0.0.1:8081; }   # = API_RELAY_PORT (api :8080).
                                              # ГАНЦ api порт: OIDC, /rp/eid/ болон
                                              # /rp/sign/ бүгд үүнийг хуваалцана —
                                              # OIDC шинэ порт НЭМЭЭГҮЙ.

server {
    server_name sso.dgov.mn;

    # OIDC протоколын endpoint → api (энэ backend нь ӨӨРӨӨ issuer)
    location /oauth2/                         { proxy_pass http://dan_api; include /etc/nginx/proxy_params; }
    location = /userinfo                      { proxy_pass http://dan_api; include /etc/nginx/proxy_params; }
    location /.well-known/openid-configuration { proxy_pass http://dan_api; include /etc/nginx/proxy_params; }
    location = /.well-known/jwks.json         { proxy_pass http://dan_api; include /etc/nginx/proxy_params; }

    # 3 дагч Relying Party-ийн eID sign relay → мөн ижил api upstream
    location /rp/sign/                        { proxy_pass http://dan_api; include /etc/nginx/proxy_params; }

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
`certbot --nginx -d sso.dgov.mn`. Compose файл `COOKIE_SECURE=true` тавьдаг тул
сайт **заавал HTTPS-ээр** үйлчлэх ёстой — эс бөгөөс browser auth болон OIDC
cookie-г хадгалахгүй.

> **Одоо байгаа vhost-ыг шинэчлэх.** Location блокууд өөрчлөгдөөгүй — backend
> дотоод provider нь яг ижил замуудыг үйлчилнэ. Зөвхөн upstream солигдоно:
> `upstream dan_hydra { server 127.0.0.1:4444; }`-ыг устгаад дөрвөн OIDC блокыг
> api upstream руу заа. `4444` / `4445` портууд байхаа больсон бөгөөд "admin
> портыг хэзээ ч проксилохгүй" гэсэн хуучин дүрэм ч утгагүй болсон — admin порт
> гэж байхгүй.

## 5. Ажиллаж буй deployment-ийг шинэчлэх

```bash
cd /srv/dan
git pull --ff-only origin main
docker compose build              # api + web + migrate
docker compose up -d              # өөрчлөгдсөн контейнеруудыг сэргээнэ; migrate дахин
                                  # ажиллана (түрхэгдсэн migration-уудыг алгасна)
```

`db`, `redis` хэвээр ажиллана — өгөгдөл хөндөгдөхгүй. Зөвхөн тохиргоо өөрчилсөн
бол: `backend.env` / `.env`-ээ засаад `docker compose up -d api web`.

### Өмнөх Ory Hydra стекээс шилжих (cutover)

Хэрэв сервер өмнө нь Hydra-д суурилсан стекээр ажиллаж байсан бол бүртгэлтэй
OAuth2 client-уудыг cutover-ийн **ӨМНӨ** api-ийн өөрийн `oauth_clients` руу зөө:

```bash
./scripts/migrate-hydra-clients.sh    # DRY_RUN=1 … бол зөвхөн уншиж шалгана
```

Client secret нь хэвээрээ зөөгддөг тул relying party-ууд тохиргоогоо огт
өөрчлөхгүйгээр үргэлжлүүлэн ажиллана. Үүний дараа `hydra` / `hydra-migrate`
контейнер, тусдаа `hydra` database болон бүх `HYDRA_*` хувьсагчийг устгаж,
nginx vhost-ыг дахин чиглүүлж болно (§4-ийг үз).

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
docker logs dan-dgov-mn-api-1 2>&1 | grep -i error      # хоосон байх ёстой
curl -s -o /dev/null -w '%{http_code}\n' https://sso.dgov.mn/   # 200
curl -s https://sso.dgov.mn/.well-known/openid-configuration | head -c 80   # OIDC issuer JSON
curl -s https://sso.dgov.mn/.well-known/jwks.json | head -c 80              # RS256 нийтийн түлхүүр(үүд)
```

Discovery-гийн `issuer` нь `OAUTH_ISSUER`-тэй ЯГ таарах ёстой ба `jwks.json` нь
дор хаяж нэг RS256 түлхүүр буцаана. Түлхүүр бэлтгэх нь fail-closed: api эхний
ажиллагаанд гарын үсгийн түлхүүрээ үүсгэдэг бөгөөд чадахгүй бол **асахаас
татгалзана** (api лог дотроос `oidc: ensure signing key`-г хай — ихэвчлэн буруу
`INTEGRATION_ENC_KEY`).

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
- `SSO_STATE_KEY` солих нь дундаа явж буй login/consent state-ыг хүчингүй
  болгоно — тэр үед authorization хийж байсан browser урсгалаа дахин эхлүүлнэ.
- **`INTEGRATION_ENC_KEY` одоо хоёр дахь үүрэгтэй.** Гуравдагч OAuth токеныг
  битүүмжлэхээс гадна id_token-ий гарын үсгийн түлхүүрийг (RSA-2048, эхний
  ажиллагаанд үүсдэг) `oauth_signing_keys`-д AES-256-GCM-ээр шифрлэдэг. Үүнийг
  сольсон тохиолдолд хадгалсан гарын үсгийн түлхүүр **задрахаа болино** — api нь
  өөрөө засахгүй (идэвхтэй түлхүүрийн мөр хэвээр байгаа тул), улмаар түлхүүрийг
  солих хүртэл id_token гаргах нь бүтэлгүйтнэ. Сэргээхийн тулд `oauth_signing_keys`
  дэх идэвхтэй мөрийг retire/устгавал дараагийн boot дээр шинэ түлхүүр үүснэ;
  `kid` өөрчлөгдөх тул relying party-ууд `/.well-known/jwks.json`-оо дахин татах
  ба хадгалсан интеграцийн токенуудыг дахин холбох шаардлагатай болно.
  `INTEGRATION_ENC_KEY` солихыг ердийн нууц солилт биш, зарлаж хийх issuer
  түлхүүрийн rollover гэж төлөвлө.
- `GEMINI_API_KEY` болон OAuth / `EID_RP_SECRET` / `CORE_API_TOKEN` креденшлүүдийг
  консолоос нь rotate хийгээд `backend.env` / `.env`-д сольж
  `docker compose up -d api web` хийнэ.

---

**Government Template Platform V3.0** — **Gerege Systems Development Team** болон **Claude AI** хамтран бүтээв, 2026.
