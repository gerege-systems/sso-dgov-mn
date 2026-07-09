# Deploy хийх заавар

> 🌐 [English](DEPLOYMENT.md) · **Монгол**

Бүтэн стекийг (Postgres + Redis + Go API + Next.js web) нэг VPS дээр Docker
Compose-оор, nginx-ийн ард deploy хийх заавар. Жишиг deployment-д ашигласан
бодит runbook.

## Топологи

```
Internet ──► nginx (80/443, Let's Encrypt TLS)
                │ proxy_pass
                ▼
        web  127.0.0.1:${WEB_PORT}   (Next.js BFF — гадагш нээгддэг ЦОРЫН ГАНЦ контейнер)
                │ BACKEND_URL=http://api:8080   (дотоод compose сүлжээ)
                ▼
        api ──► db (Postgres 16) + redis (7)    (нийтийн port байхгүй)
```

Browser зөвхөн `web`-д хүрнэ; `api`, `db`, `redis` дотоод compose сүлжээнд
үлдэнэ. Нэг удаагийн `migrate` контейнер `up` бүр дээр SQL migration-уудыг
түрхээд unтардаг.

## Шаардлага

- Docker + compose plugin-тэй VPS (`docker compose version`)
- Хост дээр nginx + certbot (эсвэл TLS terminate хийдэг дурын reverse proxy)
- Сервер рүү заасан DNS бичлэг

## 1. Кодоо татах

```bash
git clone https://github.com/gerege-systems/template-gerege-mn.git /srv/template-gerege-mn
cd /srv/template-gerege-mn
```

## 2. Хоёр env файл үүсгэх (хоёулаа gitignored)

### `./.env` — compose interpolation

```env
POSTGRES_USER=postgres            # superuser — зөвхөн migrate хэрэглэнэ
POSTGRES_PASSWORD=<санамсаргүй>
POSTGRES_DB=gerege_template
APP_DB_USER=app_user              # api-ийн холбогддог хамгийн бага эрхт role
APP_DB_PASSWORD=<санамсаргүй>
APP_DB_DSN=host=db port=5432 user=app_user password=<мөн адил> dbname=gerege_template sslmode=disable
REDIS_PASS=<санамсаргүй>
APP_ORIGIN=https://your.domain.mn # яг нийтийн origin (CSRF origin шалгалт)
WEB_PORT=3007                     # nginx-ийн proxy хийдэг loopback port
```

### `./backend.env` — `api` + `migrate`-д `/app/.env` болж mount хийгдэнэ

```env
PORT=8080
ENVIRONMENT=development           # compose стек dev горимоор ажиллана: дотоод
                                  # DB TLS-гүй (prod guard нь sslmode=verify-full
                                  # шаарддаг); TLS нь nginx дээр төгсдөг
DEBUG=false
DB_POSTGRE_DRIVER=postgres
DB_POSTGRE_DSN=postgres://postgres:<POSTGRES_PASSWORD>@db:5432/gerege_template?sslmode=disable
                                  # ^ superuser DSN — MIGRATE (DDL) хэрэглэнэ.
                                  # api-д APP_DB_DSN-ээр дарж бичигдэнэ.
JWT_SECRET=<≥32 санамсаргүй тэмдэгт>
JWT_EXPIRED=24                    # цаг
JWT_ISSUER=your.domain.mn
JWT_REFRESH_EXPIRED=7             # хоног
OTP_MAX_ATTEMPTS=5
BCRYPT_COST=12
REDIS_HOST=redis:6379
REDIS_PASS=<.env-тэй ижил>
REDIS_EXPIRED=5                   # минут (OTP request_id-ийн TTL)
ALLOWED_ORIGINS=https://your.domain.mn
TRUSTED_PROXIES=172.16.0.0/12,127.0.0.1   # XFF-д зөвхөн docker сүлжээ + nginx-ээс итгэнэ.
                                  # Proxy-гийн ард ЗААВАЛ: api нийтийн порт-гүй
                                  # тул бүх хүсэлт web/nginx peer-ээс ирнэ. BFF нь
                                  # жинхэнэ клиент IP-г X-Forwarded-For-оор
                                  # дамжуулдаг; итгэмжит proxy жагсаалтгүй бол api
                                  # түүнийг үл тоож, per-IP rate-limit бүгд нэг
                                  # bucket-д уначихна.
VERIFY_API_BASE=https://verify.gecloud.mn/v1
VERIFY_API_KEY=<gck_live_…>       # имэйл/SMS OTP — үгүй бол бүртгэл ажиллахгүй
VERIFY_CHANNEL=email
GEMINI_API_KEY=<AIza…>            # AI боломжууд; хоосон бол AI endpoint 500
```

Нууцуудыг `openssl rand -hex 24`-өөр үүсгэ.

## 3. Яагаад хоёр DB role вэ (анхны boot-оос ӨМНӨ унш)

Row-Level Security-г superuser **чимээгүй алгасдаг**. Тиймээс стек хоёр
role ашиглана:

- `migrate` нь `POSTGRES_USER`-ээр (superuser — `CREATE EXTENSION`, RLS
  DDL-д хэрэгтэй) холбогдоно.
- `api` нь `APP_DB_USER`-ээр (`NOSUPERUSER NOBYPASSRLS`) холбогдоно —
  **хоосон data volume-ийн анхны init дээр**
  `backend/deploy/initdb/10-create-app-user.sh` автоматаар үүсгэдэг.

api үүнийг **boot үед шалгадаг**: role нь superuser/BYPASSRLS бол
production горимд асахаас татгалзаж, development горимд warning логдоно.
*Одоо байгаа* DB рүү deploy хийж байгаа бол role + grant-уудыг гараар
үүсгээд (initdb script-ийг үз) `APP_DB_DSN`-ийг түүн рүү заа.

## 4. Анхны deploy

```bash
docker compose up -d --build      # api+web-ийг бүтээж, migration түрхэж, бүгдийг асаана
docker compose ps                 # db/redis/api/web healthy, migrate Exited (0) байх ёстой
```

### nginx vhost (хост дээр)

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

Дараа нь TLS: `certbot --nginx -d your.domain.mn`. Compose файл
`COOKIE_SECURE=true` тавьдаг тул сайт **заавал HTTPS-ээр** үйлчлэх ёстой —
эс бөгөөс browser auth cookie-г хадгалахгүй.

## 5. Ажиллаж буй deployment-ийг шинэчлэх

```bash
cd /srv/template-gerege-mn
git pull --ff-only origin main
docker compose build              # api + web + migrate
docker compose up -d              # өөрчлөгдсөн контейнеруудыг сэргээнэ; migrate
                                  # дахин ажиллана (түрхэгдсэн файлуудыг алгасна)
```

`db`, `redis` хэвээр ажиллана — өгөгдөл хөндөгдөхгүй. Зөвхөн тохиргоо
өөрчилсөн бол: `backend.env` / `.env`-ээ засаад `docker compose up -d api web`.

### Автомат deploy (CI/CD)

Дээрх алхмуудыг GitHub Actions-д холбосон тул `main`-д push хийхэд автоматаар
deploy хийгдэнэ. [`.github/workflows/ci.yml`](../.github/workflows/ci.yml)-ийн
`deploy` job нь `backend`, `frontend`, `secrets-scan` job-ууд **амжилттай
давсны дараа л** ажиллаж, энэ VPS руу SSH-ээр орж
[`deploy/deploy.sh`](../deploy/deploy.sh)-ийг ажиллуулна (rebuild → `up -d` →
эрүүл болтол хүлээх → prune). `db`/`redis` тасрахгүй; migration дахин ажиллаж
түрхэгдсэн файлуудыг алгасна.

Нэг удаагийн тохиргоо — **Settings → Secrets and variables → Actions** дор
гурван repo secret нэмнэ:

| Secret | Утга |
|--------|------|
| `DEPLOY_HOST` | VPS-ийн IP / hostname |
| `DEPLOY_USER` | repo + docker эрхтэй SSH хэрэглэгч (жишиг deploy-д `root`) |
| `DEPLOY_SSH_KEY` | тусгай deploy keypair-ийн **private** түлхүүр; public түлхүүрийг серверийн `~/.ssh/authorized_keys`-д нэмсэн байна |
| `DEPLOY_PORT` | *(заавал биш)* SSH порт, default нь `22` |

Keypair-ийг `ssh-keygen -t ed25519 -f deploy_key -N ''`-ээр үүсгэж,
`deploy_key.pub`-ийг серверийн `authorized_keys`-д нэмээд, private `deploy_key`-г
`DEPLOY_SSH_KEY`-д хийнэ. Код өөрчлөхгүйгээр Actions таб-аас гараар deploy
дуудаж болно (**Run workflow** — `workflow_dispatch`), эсвэл сервер дээр
`bash deploy/deploy.sh`-ийг гараар ажиллуулж болно.

## 6. Баталгаажуулах

```bash
docker compose ps                                      # бүгд healthy
docker logs temp-gerege-mn-migrate-1 | tail -3         # "migration [up] success"
docker logs temp-gerege-mn-api-1 2>&1 | grep -i error  # хоосон байх ёстой
curl -s -o /dev/null -w '%{http_code}\n' https://your.domain.mn/   # 200
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
- `GEMINI_API_KEY` / `VERIFY_API_KEY`-г консолоос нь rotate хийгээд
  `backend.env`-д сольж `docker compose up -d api` хийнэ.

---

**Government Template Platform V3.0** — **Gerege Systems Development Team** болон **Claude AI** хамтран бүтээв, 2026.
