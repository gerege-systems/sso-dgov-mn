# Байршуулалт

Бүрэн runbook нь `docs/DEPLOYMENT.md`-д байдаг; энэ бол үйл ажиллагааны товч тойм юм.

## docker-compose stack

`docker-compose.yml` нь долоон үйлчилгээ тодорхойлдог; зөвхөн loopback host port-ууд
нийтлэгдэх ба `db`/`redis` нь дотоод сүлжээнээс хэзээ ч гардаггүй.

| Үйлчилгээ | Image / build | Үүрэг | Port-ууд |
|---------|---------------|------|-------|
| `db` | `postgres:16-alpine` | Postgres — app DB болон `hydra` DB хоёуланг hosting хийдэг | дотоод |
| `redis` | `redis:7-alpine` (`--requirepass`) | cache / state | дотоод |
| `migrate` | build `./backend` | нэг удаагийн `migrate -up`, дараа нь гарна | `restart: "no"` |
| `api` | build `./backend` | Go API (`:8080`) | `127.0.0.1:8091:8080` (relay) |
| `web` | build `./frontend` | Next.js BFF | `127.0.0.1:3007:3000` |
| `hydra-migrate` | `oryd/hydra:v2.2.0` | Hydra schema хэрэглэж, гарна | `restart: "no"` |
| `hydra` | `oryd/hydra:v2.2.0` | OIDC issuer | `:4444` нийтийн (proxy хийсэн), `:4445` admin зөвхөн loopback |

Хамаарлын дараалал: `api` нь `db` healthy + `redis` healthy + `migrate` дуусахыг
хүлээнэ; `web` нь `api` healthy-г хүлээнэ; `hydra` нь `db` + `hydra-migrate`-ийг
хүлээнэ. `db` healthcheck нь `pg_isready -h 127.0.0.1` (TCP) ашигладаг тул initdb
дуустал unhealthy хэвээр байна — ингэснээр `migrate` нь socket-only анхны boot-ийн
цонхон дундуур уралдаж орохоос сэргийлнэ.

## Хоёр DB role

- `migrate` / `hydra-migrate` нь **superuser**-ээр (`POSTGRES_USER`) холбогддог —
  тэдэнд `CREATE EXTENSION`, RLS DDL, create-database хэрэгтэй.
- `api` нь `APP_DB_USER`-ээр (`NOSUPERUSER NOBYPASSRLS`) холбогддог тул Postgres RLS
  бодитоор хэрэгжинэ. API нь boot дээр өөрийн role-ийг шалгаж, **production-д хэрэв
  superuser/BYPASSRLS бол эхлэхгүй**.

App role болон hydra DB нь анхны init дээр
`backend/deploy/initdb/10-create-app-user.sh` болон `20-create-hydra-db.sh`-ээр
автоматаар үүсгэгддэг.

## nginx reverse proxy + TLS

nginx нь host дээр ажиллаж, TLS-ийг (Let's Encrypt) terminate хийж, гурван loopback
upstream руу тараана:

- `/oauth2/*`, `/userinfo`, `/.well-known/*` → Hydra public (`:4444`)
- `/rp/sign/*` → api relay (`:8091`)
- бусад бүх зүйл (app, BFF `/api/*`, DAN-render хийсэн OIDC login/consent UI) →
  web (`:3007`)

HTTPS заавал: compose нь `COOKIE_SECURE=true` тохируулдаг ба Hydra нь
`SameSite=None` (энэ нь `Secure` шаарддаг) ажилладаг тул хөтчүүд auth/OIDC cookie-г
энгийн HTTP дээр хаядаг. Hydra admin port `:4445` нь зориудаар proxy **хийгддэггүй**.
`TRUSTED_PROXIES=172.16.0.0/12,127.0.0.1` тохируул, эс бол IP тус бүрийн rate limit
нэг bucket болон нурна.

## Migration-ууд

`migrate` нь `docker compose up` бүрд `/app/migrate -up` ажиллуулна; аль хэдийн
хэрэглэсэн migration-ууд алгасагдана. Файлууд `backend/migrations/`-д `N_name.up.sql` /
`.down.sql` байдаг — одоогийн мод нь нэгтгэсэн `1_init_schema` хос дээр нэмэн
нэгтгэхээс өмнөх файлуудын `old/` фолдер юм. `hydra-migrate` нь Hydra-ийн өөрийн
schema-г тусад нь хэрэглэдэг.

## Шинэчлэх ба буцаах

**Гараар шинэчлэх:** `git pull --ff-only` → `docker compose build` → `up -d`.
`db`/`redis` ажиллаж хэвээр байна; хоёр migrate ажил дахин ажиллаж, хэрэглэсэн
migration-уудыг алгасна.

**Автоматаар:** `deploy.yml` нь root бус `deploy` хэрэглэгчээр SSH хийж, CI-д
дассан яг SHA руу `git reset --hard` хийж, дараа нь `deploy/deploy.sh` ажиллуулна —
энэ нь `INTEGRATION_ENC_KEY` байхгүй үед л idempotent байдлаар бичдэг (хэзээ ч
дарж бичдэггүй — түүнийг эргүүлэх нь бүх AES-GCM-encrypt хийсэн өгөгдлийг эвдэнэ),
build хийж, `--remove-orphans`-ээр өргөж, `api`/`web` healthy гэж мэдээлэхийг ~150 с
хүртэл хүлээж (алдаа гарвал log хаяж), дараа нь image-үүдийг цэвэрлэдэг. Concurrency
group `deploy-production` нь deploy-уудыг цувуулна.

**Буцаах:** `git reset --hard <commit>` дараа нь дахин build. Migration-ууд энэ
урсгалд зөвхөн урагшаа явдаг — schema өөрчлөлтийг буцаахын тулд код руу буцахаас өмнө
харгалзах `down.sql`-ийг гараар хэрэглэ.
