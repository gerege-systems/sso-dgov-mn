# Government Template Platform V3.0

> **eID-д суурилсан · AI-аар хүчирхэгжсэн** засгийн газрын үйлчилгээний платформ

> 🌐 [English](../README.md) · **Монгол**

[![Go](https://img.shields.io/badge/Go-1.26-blue.svg)](https://golang.org/)
[![chi](https://img.shields.io/badge/chi-v5-00ADD8.svg)](https://github.com/go-chi/chi)
[![pgx](https://img.shields.io/badge/pgx-v5-336791.svg)](https://github.com/jackc/pgx)
[![Next.js](https://img.shields.io/badge/Next.js-14-black.svg)](https://nextjs.org/)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)

Clean Architecture зарчмаар бүтээгдсэн, аюулгүй байдлыг хатууруулсан,
production-д бэлэн **full-stack template**. Go (**chi · net/http + pgx (pgxpool) +
PostgreSQL + Redis**) backend болон Next.js (**BFF**) frontend-ийг хослуулсан —
хооронд нь холбож, ямар ч систем рүү өргөтгөхөд бэлэн. Backend нь стандарт сангийн
`net/http`-ийг [go-chi/chi](https://github.com/go-chi/chi) router болон гар бичмэл
SQL-тэй [jackc/pgx](https://github.com/jackc/pgx) драйвертэй хослуулдаг — ORM
ашиглахгүй.

## 📌 Эх сурвалж ба нээлттэй эх

**Backend** нь нээлттэй эх
[snykk/go-rest-boilerplate](https://github.com/snykk/go-rest-boilerplate)
(MIT, Najib Fikri)-аас гаралтай; HTTP давхаргыг **Gin → chi (net/http)**, өгөгдлийн
давхаргыг **sqlx → pgx (pgxpool, гар бичмэл SQL)** болгож хөрвүүлсэн, бүх фичерийг
хадгалсан. Эх төслийн attribution-г [AUTHORS](../AUTHORS)-д хадгалсан. Энэ төсөл **MIT
лицензтэй** — [LICENSE](../LICENSE).

## Monorepo бүтэц

```
gerege-template/
├── backend/           # Go · chi (net/http) · pgx (pgxpool) · PostgreSQL · Redis · JWT/OTP танилт
│   └── docs/          # ARCHITECTURE · DEVELOPMENT · API_CONTRACT · SECURITY (EN/MN)
└── frontend/          # Next.js BFF (backend руу server талаас прокси; cookie session)
```

- **[backend/README_MN.md](../backend/README_MN.md)** — Clean Architecture Go API.
- **[frontend/README.md](../frontend/README.md)** — Next.js Backend-for-Frontend.

## Онцлог

- **Clean Architecture** — `handler → usecase → repository → domain`, back-import байхгүй; business core нь web framework-ийг import хийдэггүй.
- **Танилт** — JWT access + refresh (rotation), OTP-баталгаажуулсан бүртгэл, bcrypt, login lockout; logout хоёр токеныг хоёуланг хүчингүй болгоно (refresh + access deny-list).
- **AI pipeline (Gemini)** — SDK-гүй REST client + function calling: текст/дуут чат, яриа→текст (STT), текст→яриа (TTS), шууд орчуулга. Давхаргат system prompt (кодод хатуу suurь дүрэм + админ DB-ээс тохируулдаг хамрах хүрээ/заавар) туслахыг зөвхөн заасан хүрээнд барина; `search_knowledge` tool нь хариултыг `ai_knowledge` хүснэгтийн өгөгдөлд тулгуурлуулна.
- **Аюулгүй хатууруулсан** — security headers (CSP, HSTS, COOP/COEP/CORP), CORS allow-list, rate limiting, серверийн бүрэн timeout-ууд, parameterized query, Postgres Row-Level Security + boot-үеийн мөрдөлтийн guard. [SECURITY.md](../SECURITY.md)-г үз.
- **Observability** — OpenTelemetry trace + Prometheus metrics + Zap structured log.
- **Frontend BFF** — браузер зөвхөн ижил-origin Next.js route рүү залгаж, тэр нь server талаас backend руу проксиолдог (токен client JS-д хүрэхгүй); давхар CSRF хамгаалалт (custom header + origin), TanStack Query өгөгдлийн давхарга.
- **Тесттэй** — unit + testcontainers integration тест.

## Түргэн эхлүүлэх

**Шаардлага:** Go 1.26+, Node 20+, PostgreSQL 15+, Redis 7+.

```bash
# 1) Backend  →  http://localhost:8080
cd backend
cp internal/config/.env.example internal/config/.env   # JWT_SECRET (≥32), DB, Redis тохируул

# 2) Frontend →  http://localhost:3000
cd ../frontend
cp .env.example .env.local                              # BACKEND_URL=http://localhost:8080
npm install
npm run dev
```

**http://localhost:3000** нээж бүртгүүлэх / нэвтрэх.

## Баримтжуулалт

| Doc | Юу |
|-----|------|
| [backend/docs/ARCHITECTURE_MN.md](../backend/docs/ARCHITECTURE_MN.md) | Давхаргууд, dependency flow |
| [backend/docs/DEVELOPMENT_MN.md](../backend/docs/DEVELOPMENT_MN.md) | Фичер нэмэх заавар, тест, code style |
| [backend/docs/API_CONTRACT_MN.md](../backend/docs/API_CONTRACT_MN.md) | REST endpoint, request/response |
| [backend/docs/AI_PIPELINE_MN.md](../backend/docs/AI_PIPELINE_MN.md) | AI туслахын дотоод бүтэц: урсгал, prompt давхарга, tools, voice, өргөтгөх заавар |
| [backend/docs/SECURITY.md](../backend/docs/SECURITY.md) | Хэрэгжсэн хяналт + ASVS roadmap |
| [docs/DEPLOYMENT_MN.md](DEPLOYMENT_MN.md) | VPS deploy runbook (compose, env файлууд, nginx, шинэчлэх, rollback) |
| [SECURITY.md](../SECURITY.md) | Эмзэг байдлыг хэрхэн мэдээлэх |
| [CONTRIBUTING.md](../CONTRIBUTING.md) | Хэрхэн хувь нэмэр оруулах |

## Хувь нэмэр

Хувь нэмэр оруулахыг урьж байна — [CONTRIBUTING.md](../CONTRIBUTING.md) болон
[Code of Conduct](CODE_OF_CONDUCT.md)-ийг уншина уу.

## Лиценз

[MIT](../LICENSE) — snykk/go-rest-boilerplate (MIT)-ийн derivative; эх төслийн
attribution-г [AUTHORS](../AUTHORS)-д хадгалсан.

---

**Government Template Platform V3.0** — **Gerege Systems Development Team** болон
**Claude AI** хамтран бүтээв, 2026.
