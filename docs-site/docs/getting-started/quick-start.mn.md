# Quick Start

## Урьдчилсан шаардлага

- Go 1.26+, Node 20+, PostgreSQL 15+, Redis 7+
- Docker (бүх стект зөвлөмж болгоно)

## A хувилбар — Docker Compose ашиглан бүх стек

`db`, `redis`, нэг удаагийн `migrate` job, `api`, `web`, мөн (тохируулагдсан үед)
OIDC-provider горимд зориулсан Ory `hydra`-г ажиллуулна.

```bash
docker compose up -d --build
```

**http://localhost:3007**-ыг нээж (`web` service нь loopback порт дээр
нийтэлдэг) **Login with eID**-ийг сонго — QR-ыг уншуулах / eID гар утасны апп-ыг
нээх, эсвэл иргэний бүртгэлийн дугаар оруулан push хүлээн ав. Google холболт
болон OIDC provider нь тэдгээрийн credential тохируулагдсан үед л харагдана.

!!! warning "Compose нь зориудаар `ENVIRONMENT=development`-ээр ажилладаг"
    Дотоод өгөгдлийн санд TLS байхгүй бөгөөд продакшны хамгаалалт нь
    `sslmode=verify-full`-ыг шаарддаг. Үүний улмаас compose стек нь HSTS /
    sslmode / observability-gate зэрэг продакшны замуудыг **ажиллуулдаггүй** —
    эдгээр нь зөвхөн nginx-ийн ард `ENVIRONMENT=production` үед идэвхжинэ.
    [Deployment](../operations/deployment.md)-ийг үз.

## B хувилбар — service-үүдийг шууд ажиллуулах

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

## Backend хөгжүүлэгчийн командууд

`backend/`-ээс ажиллуул:

```bash
go build ./...          # build
go test ./...           # unit tests (mocks only, fast)
make test-integration   # testcontainers (needs Docker)
make swag               # regenerate swagger after touching handler annotations
make pre-push           # mirror CI: lint + test + swag drift + build
```

## Frontend командууд

`frontend/`-ээс ажиллуул:

```bash
npm run dev             # local dev
npm run build           # build + lint + typecheck (what CI runs)
```

## Анхны нэвтрэлтэд eID RP credential шаардлагатай

eID нь цорын ганц нэвтрэх арга тул ажиллах нэвтрэлт нь backend env дотор
бүртгэгдсэн eID Mongolia **Relying Party**-г (RP UUID + secret) шаарддаг
(`EID_RP_UUID`, `EID_RP_NAME`, `EID_RP_SECRET`, `EID_BASE_URL`). Тэдгээргүйгээр
API нь дуудаж ажиллах ч eID start endpoint-ууд алдаа буцаана. Бүрэн env схемийн
хувьд [Configuration](../operations/configuration.md)-ийг үз.
