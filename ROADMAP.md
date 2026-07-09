# ROADMAP — Government Template Platform V3.0

> Төслийн phase-уудын явц, төлөвлөгөө. Phase бүр дуусахад энэ файлыг шинэчилж
> commit хийнэ. Дэлгэрэнгүй баримтууд: [README.md](README.md#documentation).

**Одоогийн байдал:** v27 — бүх суурь систем + AI pipeline ажиллаж байгаа,
жишиг deployment: https://template.dgov.mn (CI ногоон).

---

## ✅ Дууссан phase-ууд

### Phase 1 — Core template (v27 суурь)
- Clean Architecture Go backend: chi (net/http) + pgx (ORM-гүй) + PostgreSQL + Redis
- Auth: JWT access+refresh (rotation), OTP бүртгэл (GeregeCloud Verify), bcrypt, lockout
- RBAC: динамик role/permission + каталог; Postgres RLS (ENABLE+FORCE, non-superuser app role)
- Observability: OTel tracing + Prometheus + Zap; security headers, CORS, rate limiting
- Next.js 15 BFF frontend: httpOnly cookie session, admin/manager/me системүүд, mn/en i18n
- CI: gofmt + vet + race tests + swag drift + frontend lint/build + gitleaks

### Phase 2 — Хэрэглэгчийн нэр (2026-06-10)
- Овог/нэр (Монгол + Латин) — register + хэл-мэдрэмжтэй харуулалт

### Phase 3 — AI pipeline: цөм (2026-06-11, `a4da698`)
- `pkg/gemini` — SDK-гүй REST client (3× retry + backoff)
- Function-calling чат (`/ai/chat`): AI шийднэ → backend гүйцэтгэнэ; Монгол fallback
- Frontend чат UI (/me/ai)

### Phase 4 — Security hardening (2026-06-11, `fb220da`…)
- HTTP server timeouts + MaxHeaderBytes (slowloris хаалт)
- Logout = refresh revoke + access token deny-list
- RLS boot guard (superuser/BYPASSRLS бол prod-д асахгүй)
- BFF: давхар CSRF (custom header + origin), route param/query validation,
  RSC refresh-token шатдаг bug засвар

### Phase 5 — AI voice (2026-06-11, `09d28f9`, `8f1f331`)
- Audio ойлголт (дуут чат мессеж), STT, TTS (PCM→WAV), live орчуулга
- Frontend: 🎤 дуут мессеж, 🔊 TTS playback, /me/translate live хуудас

### Phase 6 — Prompt давхарга + DB хайлт (2026-06-11, `426f851`)
- 3 давхаргат system prompt: hardcoded guardrails + DB scope/instructions
- `search_knowledge` tool (`ai_knowledge` хүснэгт) — DB-д тулгуурласан хариулт
- Admin UI + API (`/admin/ai/prompts`, settings.manage)

### Phase 7 — Frontend чанар (2026-06-11, `2ec5ef9`)
- TanStack Query (кэш/dedup/invalidation), admin pagination, CI Node 24

### Phase 8 — Deploy + баримтжуулалт (2026-06-11)
- template.dgov.mn дээр шинэчилсэн deploy (migration 11, Gemini key)
- Бүх док шинэчлэгдсэн + шинэ: AI_PIPELINE(_MN).md, DEPLOYMENT(_MN).md, CLAUDE.md
- Бүх relative .md холбоос скриптээр шалгагдаж засагдсан

---

## 🔜 Дараагийн phase-ууд (ач холбогдлоор)

### Phase 11 — AI сайжруулалтууд
- [ ] Knowledge base хайлтыг tsvector (full-text) болгох; том санд pgvector (semantic)
- [ ] Чатын streaming хариу (SSE) — урт хариултын UX
- [ ] Чат түүхийг server талд хадгалах сонголт (одоо stateless)
- [ ] Нэмэлт tools: хэрэглэгчийн өөрийн профайл асуух (RLS-тэй), системийн статистик (admin)
- [ ] AI prompt-ийн хувилбарын түүх (audit) — хэн хэзээ юу өөрчилсөн

### Phase 12 — Security (ASVS L2 үлдэгдэл)
- [ ] HIBP k-anonymity leaked-password шалгалт (config-gated, fail-open)
- [ ] CSP-г nonce-based болгох (одоо 'unsafe-inline' — Next.js-ийн хязгаарлалт)
- [ ] `govulncheck` + container scan CI-д
- [ ] golangci-lint-ийг Go 1.26 дэмжмэгц CI-д буцаах (одоо vet+gofmt)
- [ ] Secrets manager/KMS интеграц (production-д .env-ийн оронд)

### Phase 13 — Ops
- [ ] DB автомат backup + restore тест (cron + offsite)
- [ ] Interactive Swagger UI (одоо зөвхөн /swagger/doc.json)
- [ ] Staging орчин + deploy-г CI-ээс автоматжуулах (одоо гар runbook)
- [ ] Pool/error alert-ууд (Prometheus alertmanager)

### Backlog (хэрэгцээ гарвал)
- [ ] WebAuthn/passkeys, олон tenant-ийн RLS (`tenant_id`), field-level PII шифрлэлт
- [ ] Frontend: error boundaries, bundle analyzer, nonce CSP-тэй хамт hydration аудит
- [ ] AI: дуу хоолойн сонголтыг хэрэглэгчийн тохиргоонд

---

**Тэмдэглэл:** шинэ phase эхлэхдээ энд тэмдэглэж, дуусахад ✅ руу зөөнө.
Гүйцэтгэлийн дэлгэрэнгүй нь commit түүх болон холбогдох баримтуудад.
