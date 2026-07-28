# Аюулгүй байдал

Энэ хэсэг нь `backend/docs/SECURITY.md`-ээс товчилсон, **хэрэгжүүлсэн** аюулгүй
байдлын хяналтуудыг баримтжуулна. Платформ нь өөрийн хяналтуудыг OWASP ASVS / API
Top 10, NIST SP 800-63B / 800-218, болон CIS Controls-той харгалзуулсан.

## HTTP хилийн хяналтууд

### Аюулгүй байдлын хариултын header-ууд

`SecurityHeadersMiddleware`-ээр хариулт **бүрд** тохируулагдана:

| Header | Утга |
|--------|-------|
| `X-Content-Type-Options` | `nosniff` |
| `X-Frame-Options` | `DENY` |
| `Referrer-Policy` | `strict-origin-when-cross-origin` |
| `Content-Security-Policy` | `default-src 'none'; frame-ancestors 'none'` (API нь зөвхөн JSON) |
| `Permissions-Policy` | camera/geolocation/microphone/payment/usb/… -д бүгдийг хориглоно |
| `Cross-Origin-Opener-Policy` | `same-origin` |
| `Cross-Origin-Resource-Policy` | `same-site` |
| `Cross-Origin-Embedder-Policy` | `require-corp` |
| `Strict-Transport-Security` | `max-age=31536000; includeSubDomains` — **зөвхөн production** |

HSTS-ийг development-д саатуулдаг — ингэснээр localhost dev сервер нь энгийн HTTP-ийн
татгалзалыг нэг жилээр pin хийхгүй.

### CORS зөвшөөрлийн жагсаалт

`CORSMiddleware` нь Origin-ийг зөвхөн яг зөвшөөрлийн жагсаалтад байвал тусгана.
**Credential болон wildcard нь харилцан үл нийцнэ**:
`Access-Control-Allow-Credentials: true` нь зөвхөн зөвшөөрлийн жагсаалтын горимд
илгээгддэг, хэзээ ч `*`-тэй биш. Production нь тохиргооны баталгаажуулалт дээр хоосон
зөвшөөрлийн жагсаалтыг татгалздаг.

### Rate limiting

IP тус бүрийн token bucket-ууд:

| Хязгаарлагч | Хурд | Burst |
|---------|------|-------|
| `/auth` | ~5/мин | 5 |
| `/ai` | ~20/мин | 10 |
| long-poll | ~1/с | 30 |
| gov-write | ~30/мин | 15 |

Дуусахад → `429` бөгөөд `Retry-After` болон `X-RateLimit-*` header-тэй. Клиентийн IP
нь peer нь `TRUSTED_PROXIES`-д байгаа үед **л** `X-Forwarded-For`-т итгэн тодорхойлогдоно,
ингэснээр хуурамч XFF нь rate-limit эсвэл audit-ийн хамаарлыг хордуулж чадахгүй.

### Body-хэмжээний хязгаар ба серверийн timeout

- Дэлхийн дээд хязгаар 26 MiB (multipart PDF-sign зам-д зориулсан); энгийн JSON 1 MiB;
  auth зам-ууд **4 KiB**. Chi эцэг middleware нь зөвхөн хүүхэд хязгаарыг *чангалж*
  чадах тул 4 KiB auth хязгаар нь дэлхийн сүлжээн дээр давхарлагдана.
- `http.Server`: `ReadHeaderTimeout` 10 с, `ReadTimeout` 30 с, `WriteTimeout`
  60 с, `IdleTimeout` 120 с, `MaxHeaderBytes` 16 KiB. Хүсэлт тус бүрийн 30 с
  context deadline нь pgx query-нд тархах тул гацсан query-үүд цуцлагдана.

## Postgres Row-Level Security (RLS)

- **Superuser бус DB role заавал шаардлагатай** — RLS (`FORCE` хийсэн ч)
  superuser / `BYPASSRLS` role-уудаар чимээгүй тойрогддог. App нь `NOSUPERUSER
  NOBYPASSRLS` role-оор холбогддог; `migrate` контейнер нь DDL-ийн тулд superuser-ийг
  хадгалдаг.
- **Boot-time хэрэгжих чадварын хамгаалалт** — `guardRLSEnforceable` нь эхлэлд
  `pg_roles`-ийг асуудаг; хэрэв app-ийн role нь superuser/`BYPASSRLS` бол production-д
  **boot амжилтгүй болно** (fail-closed) ба development-д анхааруулна.
- Хэрэглэгч тус бүрийн хүснэгт бүрд RLS-ийг `ENABLE` + `FORCE` хийнэ; `withRLS`
  транзакцууд нь `SET LOCAL app.user_id / app.user_role`-аар identity-г нийтэлдэг.
  **Identity байхгүй ⇒ тэг мөр.**

Бүрэн policy хүснэгтийг [Data Model & RLS](../architecture/data-model.md)-аас үзнэ үү.

## Hash-гинжлэгдсэн зөвхөн-нэмэх audit log

- **Дизайн** (`pkg/audit/chain.go`): мөр бүр `chain_hash = SHA-256(
  prevHash ‖ canonical-json(entry) )`-ийг хадгална. `canonicalJSON` нь
  детерминист (`occurred_at`-ийг UTC unix-nanos-оор, эрэмбэлсэн `metadata`
  түлхүүрүүд, тогтмол талбарын дараалал) — **талбар нэмэх нь гинжийг таслах өөрчлөлт**.
- **Цуваа бичигчид** — нэмэлт бүрийн өмнө `pg_advisory_xact_lock`, ингэснээр зэрэгцээ
  нэмэлтүүд гинжийг салгаж чадахгүй.
- **Бүрэн бүтэн байдлын баталгаажуулалт** — `VerifyChain` нь admin GUC дор genesis-ээс
  дахин тооцоолдог, `GET /api/v1/audit/verify` (зөвхөн admin, `GET /api/v1/audit/`-ийн
  хажууд) байдлаар ил гардаг.

## Нууц мэдээллийн зохицуулалт

- Gitignore хийсэн env файлууд (`.env`, `backend.env`, `internal/config/.env*`);
  зөвхөн `*.env.example` хадгалагдана.
- CI нь working tree дээр **gitleaks** (`detect --no-git --redact`) ажиллуулна.
- Integration token болон super-admin TOTP нууцад **AES-256-GCM** амрах үед —
  түлхүүр нь `SHA-256(INTEGRATION_ENC_KEY)`-ээр гаргагдана, encrypt бүрд шинэ
  санамсаргүй nonce, `base64(nonce ‖ ciphertext)` болгон хадгалагдана.

## Production-ийн хатууруулах хамгаалалтууд

- **`sslmode=verify-full` шаардлагатай** — production-д `DB_POSTGRE_URL` нь sslmode
  нь `verify-full` (эсвэл дотоод `verify-ca`) байхаас бусад тохиолдолд татгалзагдана.
- **`/metrics` + `/swagger/doc.json` bearer-хамгаалалттай** — `ObservabilityGate` нь
  `Authorization: Bearer <OBSERVABILITY_TOKEN>`-ийг `crypto/subtle`-ээр харьцуулна;
  **аливаа таарахгүй нь 404 буцаана** (401 биш) — эндпойнтыг тагнуулаас нуухын тулд.
  `/health` + `/ready` нь нээлттэй хэвээр байна.
- **Compose нь `ENVIRONMENT=development` ажиллуулдаг** — зориудаар; дотоод DB нь
  TLS-гүй тул HSTS / sslmode / observability-gate production замууд нь production-д
  зөвхөн nginx-ийн ард идэвхжинэ.

## RBAC ба super-admin MFA (аюулгүй байдлын өнцөг)

**SuperAdmin / Admin / Manager / User** динамик role-ууд, зөвшөөрлийн каталогтой;
route хамгаалалтууд `RequirePermission` / `RequireAdmin` / `RequireSuperAdmin`
(энгийн admin нь super-admin хамгаалалтыг давж чадахгүй). Super-admin нэвтрэлт нь
бүтцийн хувьд MFA-хамгаалалттай — session нь зөвхөн TOTP/сэргээх баталгаажуулалтын
дараа олгогддог тул аливаа super-admin session нь угаасаа MFA-баталгаажсан байдаг.
[Super Admin MFA](../authentication/super-admin-mfa.md)-г үзнэ үү.

## ASVS замын зураглал

| Үе шат | Төлөв |
|-------|--------|
| **Phase 1 (ASVS L1)** | ✅ HTTPS + HSTS, зөвхөн eID нэвтрэлт, параметржүүлсэн query, аюулгүй байдлын header, чанд CORS, оролтын баталгаажуулалт, commit хийсэн нууцгүй · ⏳ CI дахь container scan / `govulncheck` |
| **Phase 2 (ASVS L2)** | ✅ rate limiting, refresh rotation, phishing-ээс тэсвэртэй eID device-link, хүсэлтийн timeout, encrypt хийсэн integration token, hash-гинжлэгдсэн audit · ⏳ WAF, SIEM, backup-restore тест, IR төлөвлөгөө |
| **Phase 3 (ASVS L3)** | ◻ талбар түвшний PII encryption (KMS), mTLS, SLSA L3 provenance, гадаад pentest |

## Эмзэг байдлыг мэдээлэх

Зохицуулсан ил тод байдлыг repository-root-ийн `SECURITY.md`-ийн дагуу зохицуулна.
Аюулгүй байдлын мэдээллийг олон нийтийн issue-д бүү нээгээрэй.
