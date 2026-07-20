# Government SSO

> **eID-д суурилсан · AI-аар хүчирхэгжсэн** — Засгийн газрын үйлчилгээний нэгдсэн нэвтрэлт (Single Sign-On)

**Government Template Platform V3.0** стек (Clean-Architecture Go backend + Next.js
BFF frontend + Gemini AI pipeline) дээр бүтээгдэж, **Government SSO** нэрээр
[sso.dgov.mn](https://sso.dgov.mn)-д байршуулагдсан.

> 🌐 [English](../README.md) · **Монгол**

[![Go](https://img.shields.io/badge/Go-1.26-blue.svg)](https://golang.org/)
[![chi](https://img.shields.io/badge/chi-v5-00ADD8.svg)](https://github.com/go-chi/chi)
[![pgx](https://img.shields.io/badge/pgx-v5-336791.svg)](https://github.com/jackc/pgx)
[![Next.js](https://img.shields.io/badge/Next.js-15-black.svg)](https://nextjs.org/)
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
dan-dgov-mn/
├── backend/           # Go · chi (net/http) · pgx (pgxpool) · PostgreSQL · Redis · eID/Google/SSO танилт
│   └── docs/          # ARCHITECTURE · DEVELOPMENT · API_CONTRACT · SECURITY (EN/MN)
└── frontend/          # Next.js BFF (backend руу server талаас прокси; cookie session)
```

- **[backend/README_MN.md](../backend/README_MN.md)** — Clean Architecture Go API.
- **[frontend/README.md](../frontend/README.md)** — Next.js Backend-for-Frontend.

## Онцлог

- **Clean Architecture** — `handler → usecase → repository → domain`, back-import байхгүй; business core нь web framework-ийг import хийдэггүй.
- **Танилт — eID + Google + dgov SSO** — цорын ганц нэвтрэх арга бол **eID-ээр нэвтрэх** (eID Mongolia Relying Party: QR код / мобайл deep-link / иргэний РД push + long-poll session). Түүний зэрэгцээ **Google OAuth** account холболт болон **dgov SSO (OIDC) consumer** (`sso.dgov.mn`). Session нь JWT access + refresh (rotation); logout хоёуланг хүчингүй болгоно (refresh + access deny-list). Нууц үг / и-мэйл-OTP нэвтрэлт байхгүй.
- **eID PKI профайл** — нэвтэрсэн иргэний eID identity-г IdP-ээс уншина: холбоотой байгууллага ба эрх бүхий гарын үсэг зурагчид, гэрчилгээ, бүртгэлтэй төхөөрөмж, идэвх.
- **Байгууллага ба гишүүнчлэл** — байгууллага үүсгэх/хайх (улсын бүртгэлээс Gerege Verify/XYP-ээр лавлах) + гишүүд/эрх удирдах, хэрэглэгч тус бүрт RLS-ээр хамгаалагдсан.
- **Төрийн үйлчилгээний портал** — иргэн рүү харсан `Төрийн үйлчилгээ` гадаргуу: үйлчилгээний каталог, хүсэлт, лавлагаа, мэдэгдэл, төлбөр, цаг захиалга.
- **API gateway** — админ удирддаг services / routes / consumers / API key / policy + хүсэлтийн телеметр (overview + logs).
- **OIDC provider (SSO)** — Government SSO өөрөө identity provider болно: Go backend дотор хэрэгжсэн өөрийн OAuth2/OIDC provider (`usecases/oidc`) нь `/oauth2/*`, `/userinfo` болон `.well-known` баримтуудыг үйлчилж, login/consent/logout урсгалыг жолоодох тул relying party-ууд `Sign in with Government SSO` хийж чадна. Authorization code + PKCE (S256), эргэлддэг refresh token (дахин ашиглалт илрүүлэлттэй), `client_credentials`; access token нь opaque, id_token нь RS256. `OAUTH_ISSUER` тохируулагдсан үед идэвхжинэ.
- **Баримт бичгийн гарын үсэг (PAdES)** — eID Mongolia `/v3`-ээр PDF-д server талаас гарын үсэг зурна, байнгын Document-Signer гэрчилгээтэй; sign-relay нь 3 дагч RP-уудыг Government SSO-ийн eID креденшлээр дамжуулан гарын үсэг зурах боломж олгоно.
- **Гуравдагч этгээдийн интеграци** — хэрэглэгч тус бүрийн OAuth холболт (Google Drive/Meet, Dropbox), токеныг шифрлэн (AES-256-GCM) хадгална; мөн **Gerege Space** апп-ын өөрийн SFTP хадгалалт.
- **AI pipeline (Gemini)** — SDK-гүй REST client + function calling: текст/дуут чат, яриа→текст (STT), текст→яриа (TTS), шууд орчуулга. Давхаргат system prompt (кодод хатуу суурь дүрэм + админ DB-ээс тохируулдаг хамрах хүрээ/заавар) туслахыг зөвхөн заасан хүрээнд барина; `search_knowledge` tool нь хариултыг `ai_knowledge` хүснэгтийн өгөгдөлд тулгуурлуулна.
- **Audit log** — hash-chain холбоост, зөвхөн-нэмэх audit бүртгэл (админ-л унших + бүрэн бүтэн байдлыг шалгах).
- **RBAC ба super admin** — динамик role + permission каталог; 4-үүрэгт загвар (**superadmin → admin → manager → user**), super admin нь админ хэрэглэгчдийг удирдах цорын ганц үүрэг. Super admin нь **MFA-тай, тусдаа account**: onboarding шидтэнгээр (урилгын allow-list → Google → eID → и-мэйл OTP → TOTP + нөөц код) үүсч, Google identity-аар түлхүүрлэсэн тусдаа `superadmin_accounts` satellite хүснэгтэд хадгалагдана — тиймээс нэг хүн eID admin *ба* Google super admin зэрэг байж чадах бөгөөд super admin-ы **аль ч** нэвтрэлт MFA-гаар хаагдана. Админыг регистрийн дугаараар (local eID-ээр бүртгэлтэй хэрэглэгчээс) олгоно.
- **Сайтын харагдац** — админ тохируулдаг сайт-даяар харагдац (accent / font / density / theme) нийтийн хуудсанд, мөн хэрэглэгч тус бүрийн override.
- **Аюулгүй хатууруулсан** — security headers (CSP, HSTS, COOP/COEP/CORP), CORS allow-list, rate limiting, серверийн бүрэн timeout-ууд, parameterized query, Postgres Row-Level Security + boot-үеийн мөрдөлтийн guard. [SECURITY.md](../SECURITY.md)-г үз.
- **Observability** — OpenTelemetry trace + Prometheus metrics + Zap structured log; production-д `/metrics` ба `/swagger` bearer token-оор хаагдана.
- **Frontend BFF** — браузер зөвхөн ижил-origin Next.js route рүү залгаж, тэр нь server талаас backend руу проксиолдог (токен client JS-д хүрэхгүй); давхар CSRF хамгаалалт (custom header + origin), TanStack Query өгөгдлийн давхарга.
- **Тесттэй** — unit + testcontainers integration тест.

## Түргэн эхлүүлэх

**Шаардлага:** Go 1.26+, Node 20+, PostgreSQL 15+, Redis 7+ (бүтэн стекийг Docker-оор ажиллуулахыг зөвлөнө).

```bash
# 1) Backend  →  http://localhost:8080
cd backend
cp internal/config/.env.example internal/config/.env   # JWT_SECRET (≥32), DB, Redis, EID_* RP креденшл тохируул

# 2) Frontend →  http://localhost:3000
cd ../frontend
cp .env.example .env.local                              # BACKEND_URL=http://localhost:8080
npm install
npm run dev
```

Эсвэл бүтэн стекийг өргө (db + migrate + redis + api + web — OIDC provider нь `api` дотор ажилладаг тул тусдаа issuer контейнер байхгүй):

```bash
docker compose up -d --build
```

**http://localhost:3000** нээж **eID-ээр нэвтрэх**-ийг сонго (QR уншуулах / eID мобайл апп нээх, эсвэл иргэний РД оруулж push хүлээж авах). Google холболт ба dgov SSO нь тэдгээрийн креденшл тохируулагдсан үед харагдана.

## Баримтжуулалт

| Doc | Юу |
|-----|------|
| [backend/docs/ARCHITECTURE_MN.md](../backend/docs/ARCHITECTURE_MN.md) | Давхаргууд, dependency flow |
| [backend/docs/DEVELOPMENT_MN.md](../backend/docs/DEVELOPMENT_MN.md) | Фичер нэмэх заавар, тест, code style |
| [backend/docs/API_CONTRACT_MN.md](../backend/docs/API_CONTRACT_MN.md) | REST endpoint, request/response |
| [backend/docs/AI_PIPELINE_MN.md](../backend/docs/AI_PIPELINE_MN.md) | AI туслахын дотоод бүтэц: урсгал, prompt давхарга, tools, voice, өргөтгөх заавар |
| [backend/docs/SECURITY.md](../backend/docs/SECURITY.md) | Хэрэгжсэн хяналт + ASVS roadmap |
| [docs/DEPLOYMENT_MN.md](DEPLOYMENT_MN.md) | VPS deploy runbook (compose, env файлууд, nginx, шинэчлэх, rollback) |
| [ROADMAP.md](../ROADMAP.md) | Юу хийгдсэн, юу дараагийнх |
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
