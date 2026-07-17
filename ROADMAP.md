# ROADMAP — DAN-Government SSO

> **DAN-Government SSO** ([dan.dgov.mn](https://dan.dgov.mn)) — eID-д суурилсан
> улсын нэгдсэн нэвтрэлт (Single Sign-On), **Government Template Platform V3.0**
> стек дээр бүтээгдсэн. Энэ файл нь хийгдсэн ажил ба урагшлах төлөвлөгөөг
> харуулна. Дэлгэрэнгүй баримт: [README.md](README.md#documentation).

**Одоогийн байдал:** eID нэвтрэлт, Google холболт, dgov SSO consumer, DAN-ий
өөрийн OIDC provider (Hydra), байгууллага/гишүүнчлэл, төрийн үйлчилгээ, API
gateway, PAdES гарын үсэг, интеграци, audit, RBAC/superadmin, сайтын харагдац —
бүгд production-д ([dan.dgov.mn](https://dan.dgov.mn)) ажиллаж байгаа (CI ногоон).

---

## ✅ Хийгдсэн

### Суурь платформ (Government Template Platform V3.0)
- Clean Architecture Go backend: chi (net/http) + pgx (ORM-гүй) + PostgreSQL + Redis
- RBAC: динамик role/permission + каталог; Postgres RLS (ENABLE+FORCE, non-superuser app role)
- Observability: OTel tracing + Prometheus + Zap; security headers, CORS, rate limiting, server timeouts
- Next.js 15 BFF frontend: httpOnly cookie session, mn/en i18n, TanStack Query
- CI: gofmt + vet + race tests + swag drift + frontend lint/build + gitleaks; CI дараа Deploy

### AI pipeline (Gemini)
- `pkg/gemini` — SDK-гүй REST client (retry + backoff); function-calling чат (Монгол fallback)
- Voice: audio ойлголт (дуут мессеж), STT, TTS (PCM→WAV), live орчуулга
- 3 давхаргат system prompt: hardcoded guardrails + DB scope/instructions; `search_knowledge` tool (`ai_knowledge`)
- Admin UI + API (`/admin/ai/prompts`, `settings.manage`)

### Танилт — eID + Google + dgov SSO
- **eID Mongolia RP нэвтрэлт** цорын ганц нэвтрэх арга болов (нууц үг/OTP/бүртгэл хасагдсан):
  QR (`/eid/start`), иргэний РД push (`/eid/start-id`), long-poll session (`/eid/poll`)
- **Google OAuth холболт** — Google account-ийг eID хэрэглэгчид холбож, дараа нь түүгээр нэвтрэх; салгах
- **dgov SSO (OIDC) consumer** (`sso.dgov.mn`) — start / callback / native (мобайл PKCE) / logout
- Landing = eID нэвтрэх дэлгэц; hard-redirect засвар
- Session: JWT access + refresh (rotation, `kind` guard); logout = refresh revoke + access deny-list

### eID PKI профайл
- Нэвтэрсэн иргэний eID identity: холбоотой байгууллага ба эрх бүхий гарын үсэг зурагчид,
  гэрчилгээ, бүртгэлтэй төхөөрөмж, идэвх (`/me/*`, `/users/me/eid/*`)

### Байгууллага, төрийн үйлчилгээ
- **Байгууллага + гишүүнчлэл** — үүсгэх/хайх (Gerege Verify/XYP улсын бүртгэл лавлагаа),
  гишүүн/эрх удирдах; RLS-ээр хэрэглэгч тус бүрт хамгаалагдсан
- **Төрийн үйлчилгээний портал** — каталог, хүсэлт, лавлагаа, мэдэгдэл, төлбөр, цаг захиалга (`/gov/*`)
- **Gerege Core find** — user/org лавлагааны wrap (`/core/*`)

### DAN нь OIDC provider (SSO issuer)
- **Ory Hydra** урдтай login/consent/logout цөм (`/provider/*`) — зөвхөн Hydra тохируулагдсан үед идэвхжинэ
- RP OAuth2 client бүртгэл/удирдлагын `/admin` гадаргуу + admin API key
- First-party client-д consent алгасах; consent-ийг сануулах (эхний удаад л асуух)
- RP-д Google linkage claim (`google_sub/email/name/picture`) release; RP-нэрийг login дэлгэцэнд харуулах
- DAN-ий өөрийн дизайнтай `/oauth` дэлгэц (SigninShell + LoginForm)

### API gateway
- services / routes / consumers / API keys / policies CRUD + overview + logs телеметр (`/gateway/*`)

### Баримт бичгийн гарын үсэг (PAdES)
- eID Mongolia `/v3`-ээр server талын PDF гарын үсэг; байнгын Document-Signer гэрчилгээ (production-д fail-closed)
- **Sign relay** (`/rp/sign/*`) — 3 дагч RP-ууд DAN-ий eID RP креденшлээр ДАМЖИН гарын үсэг зурах

### Интеграци, хадгалалт
- Хэрэглэгчийн OAuth интеграци (Google Drive/Meet, Dropbox) — токен AES-256-GCM-ээр шифрлэгдсэн (`/integrations/*`)
- **Gerege Space** — апп-ын өөрийн SFTP хадгалалт, хэрэглэгч тус бүр квоттой (`/gspace/*`)
- Гарын үсэг/тамгын asset (`/assets/*`)

### Аюулгүй байдал, audit, удирдлага
- **Audit log** — hash-chain холбоост, зөвхөн-нэмэх; админ унших + бүрэн бүтэн байдал шалгах (`/audit`, `/audit/verify`)
- **Security events** ingest (`/security/events`)
- **RBAC + super admin** — 4-үүрэгт загвар (superadmin → admin → manager → user, migration `23`);
  super admin нь админ хэрэглэгчийг удирдах цорын ганц эрх (`/superadmin/*`)
- Security hardening: HTTP server timeouts + MaxHeaderBytes, RLS boot guard, BFF давхар CSRF + route validation,
  production-д `/metrics` + `/swagger` bearer token-оор хаагдана

### Сайтын харагдац
- Админ тохируулдаг сайт-даяар харагдац (accent / font / density / theme) нийтийн хуудсанд (`/site/appearance`)
- Admin (нийтийн хуудас) ба per-user (апп) scope-ыг тусгаарласан

### Deploy
- [dan.dgov.mn](https://dan.dgov.mn) дээр production deploy (docker compose: db + redis + migrate + api + web + Hydra)
- Бүх док EN/MN хосоор шинэчлэгдсэн; DEPLOYMENT(_MN).md, AI_PIPELINE(_MN).md, CLAUDE.md

---

## 🔜 Дараагийн (ач холбогдлоор)

### SSO / provider төлөвшил
- [ ] RP self-service portal (`/admin`-ийг бүрэн UI болгох: client CRUD, redirect/scope удирдлага)
- [ ] Мобайл native урсгалыг (PKCE public client) баримтжуулж, жишиг апп гаргах
- [ ] Session удирдлага: идэвхтэй холболтуудыг харах/цуцлах (back-channel logout)

### API gateway enforcement
- [ ] Тохируулсан route/policy-г бодит reverse-proxy болгож хэрэгжүүлэх (одоо удирдлага + телеметр)
- [ ] Rate-limit / quota-г consumer/API key түвшинд мөрдүүлэх; ашиглалтын тайлан

### AI сайжруулалт
- [ ] Knowledge base хайлтыг tsvector (full-text); том санд pgvector (semantic)
- [ ] Чатын streaming хариу (SSE); чат түүхийг server талд хадгалах сонголт
- [ ] Нэмэлт tools: хэрэглэгчийн профайл (RLS-тэй), системийн статистик (admin); prompt хувилбарын audit

### Security (ASVS L2 үлдэгдэл)
- [ ] CSP-г nonce-based болгох (одоо 'unsafe-inline')
- [ ] `govulncheck` + container scan CI-д; golangci-lint-ийг CI-д буцаах
- [ ] Secrets manager/KMS интеграц (production-д .env-ийн оронд); INTEGRATION_ENC_KEY эргэлт

### Ops
- [ ] DB автомат backup + restore тест (cron + offsite)
- [ ] Staging орчин + deploy автоматжуулалтыг гүнзгийрүүлэх
- [ ] Pool/error alert (Prometheus alertmanager); Interactive Swagger UI

### Backlog (хэрэгцээ гарвал)
- [ ] Олон tenant-ийн RLS (`tenant_id`), field-level PII шифрлэлт
- [ ] Нэмэлт интеграци (OneDrive г.м.); Gerege Space квот/хуваалцах бодлого
- [ ] Frontend: error boundaries, bundle analyzer, nonce CSP-тэй hydration аудит

---

**Government Template Platform V3.0** — Co-developed by the **Gerege Systems
Development Team** and **Claude AI**, 2026.
