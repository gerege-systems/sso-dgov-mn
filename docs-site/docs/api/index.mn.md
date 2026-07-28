# REST API Reference

Үнэний эх сурвалж: Go route файлууд (`internal/http/routes/route_*.go`), тэдгээрийг
`swagger.json` болон `API_CONTRACT.md`-тай харьцуулан шалгасан. Интерактив
харагдацын хувьд [API Explorer](explorer.md)-г үзнэ үү.

## Глобал конвенцууд

### Base path ба хувилбаржуулалт

Бизнесийн endpoint бүр **`/api`** дор байрлах ба домэйн бүлэг тус бүр
**`/v1/...`** дэд мод холбодог тул үр дүнгийн зам нь `/api/v1/<group>/<path>` —
жишээ нь `POST /api/v1/auth/eid/start`. Хоёр бүлэг нь холболтын угтвараараа
ялгаатай (гэхдээ мөн `/api/v1` дор): хувийн хөрөнгө `/api/v1/me/*` дор, eID
профайл `/api/v1/users/me/eid/*` дор.

### Нэвтрэлт

`Authorization: Bearer <token>` дэх Bearer JWT-г auth middleware
баталгаажуулна (мөн Redis дэх хүчингүй болголтын жагсаалтыг шалгана). Эрхийн
хаалтууд дээр нь давхарлана: `RequirePermission(<perm>)`, `RequireAdmin()`,
`RequireSuperAdmin()`.

### Хариуны бүрхүүл

```json
{ "status": true, "message": "…", "data": {…}, "request_id": "…" }
```

Алдаанууд нь `apperror` төрлүүдийг статус руу буулгана (404/401/403/409/400, бусад
нь 500); валидацийн алдаа нь `data.errors` талбарын map бүхий **422** буцаана. 5xx
шалтгаануудыг логлож, `"internal server error"`-ээр орлуулна.

### Глобал middleware гинж

`Tracing → RequestID → Recoverer → Metrics → SecurityHeaders → CORS →
BodySizeLimit(26 MiB) → AccessLog → Timeout(30 s)`. `/api` дэд мод нь async
gateway хүсэлтийн логийг нэмнэ.

### Биеийн хэмжээний хязгаар

| Хамрах хүрээ | Хязгаар |
|-------|-----|
| Глобал | 26 MiB (гарын үсгийн upload-д шаардлагатай) |
| Auth бүлгүүд (`/auth`, `/auth/superadmin`, `/provider`) | 4 KiB |
| Энгийн JSON (`DecodeBody`) | 1 MiB |

### Хурдны хязгаар (клиент IP тус бүр)

| Limiter | Хурд / burst | Хамаарах |
|---------|--------------|------------|
| auth | ~5/минут, burst 5 | `/auth/*`, `/auth/superadmin/*` |
| ai | ~20/минут, burst 10 | `/ai/*` |
| poll | ~60/минут, burst 30 | eID long-poll endpoint-ууд |
| gov-write | ~30/минут, burst 15 | `/gov`, `/me`, `/gspace`, eID дэх өөрчлөлтүүд |

### Дэд бүтцийн endpoint-ууд (`/api`-гаас гадуур)

| Method | Path | Auth | Зорилго |
|--------|------|------|---------|
| GET | `/health` | none | Амьд эсэх шалгалт (DB + Redis) |
| GET | `/ready` | none | Бэлэн эсэх |
| GET | `/metrics` | ObservabilityGate | Prometheus (token-гүй production дээр 404) |
| GET | `/swagger/doc.json` | ObservabilityGate | OpenAPI баримт |

---

## Auth

Бүлэг `/api/v1/auth`; биеийн хязгаар 4 KiB + `ServiceRLSContext`. eID бол цорын
ганц нэвтрэх арга. [Authentication](../authentication/index.md)-г үзнэ үү.

| Method | Path | Auth | Зорилго |
|--------|------|------|---------|
| POST | `/auth/eid/start` | none · 5/min | eID нэвтрэлт эхлүүлэх (QR / deep-link) |
| POST | `/auth/eid/start-id` | none · 5/min | Иргэний үнэмлэхийн дугаараар eID нэвтрэлт эхлүүлэх (push) |
| POST | `/auth/eid/poll` | none · ~60/min | eID нэвтрэлтийн төлөвийг long-poll хийх |
| POST | `/auth/google` | none · 5/min | Google OAuth callback → холбох/нэвтрэх |
| POST | `/auth/refresh` | none · 5/min | Session эргүүлэх (нэг удаагийн refresh) |
| POST | `/auth/logout` | none · 5/min | Session хүчингүй болгох |
| DELETE | `/auth/google/link` | Bearer | Холбогдсон Google акаунтыг салгах |

## Хэрэглэгчид ба eID профайл

| Method | Path | Auth | Зорилго |
|--------|------|------|---------|
| GET | `/users/me` | Bearer | Одоогийн хэрэглэгчийн профайл |
| GET | `/users/me/eid/summary` | Bearer | eID акаунтын хураангуй |
| GET | `/users/me/eid/certificates` | Bearer | eID сертификатууд |
| GET | `/users/me/eid/devices` | Bearer | Бүртгэлтэй төхөөрөмжүүд |
| GET | `/users/me/eid/activity` | Bearer | eID үйл ажиллагааны лог |
| GET/POST | `/users/me/eid/organizations` | Bearer | Байгууллага жагсаах / холбох |
| DELETE | `/users/me/eid/organizations/{regNo}` | Bearer | Байгууллага салгах |
| GET/POST/DELETE | `/users/me/eid/organizations/{regNo}/signers` | Bearer | Байгууллагын гарын үсэг зурагчдыг удирдах |

## Хувийн хөрөнгө — гарын үсэг ба тамга

Бүлэг `/api/v1/me`; бичих үйлдэл хурдны хязгаартай.

| Method | Path | Auth | Зорилго |
|--------|------|------|---------|
| GET/PUT/DELETE | `/me/signature` | Bearer | Хувийн гарын үсгийн зураг |
| PUT | `/me/latin-name` | Bearer | Өөрийн нэрний латин галиглалыг засах |
| PUT | `/me/org-name-latin/{regNo}` | Bearer | Байгууллагын латин нэрийг засах |
| GET/PUT/DELETE | `/me/orgstamp/{regNo}` | Bearer (PUT: org ADMIN) | Байгууллагын тамганы зураг |

## RBAC / Үүргүүд

Бүлэг `/api/v1/rbac`. Удирдлагад `roles.manage` шаардлагатай (админ автоматаар
давна).

| Method | Path | Auth | Зорилго |
|--------|------|------|---------|
| GET | `/rbac/me` | Bearer | Одоогийн хэрэглэгчийн эрхүүд (цэс шүүх) |
| GET | `/rbac/roles` | `roles.manage` | Үүргүүдийг жагсаах |
| GET | `/rbac/permissions` | `roles.manage` | Эрхүүдийг жагсаах |
| POST | `/rbac/roles` | `roles.manage` | Үүрэг үүсгэх |
| PUT | `/rbac/roles/{id}` | `roles.manage` | Үүрэг шинэчлэх |
| PUT | `/rbac/roles/{id}/permissions` | `roles.manage` | Үүргийн эрхүүдийг тохируулах |
| DELETE | `/rbac/roles/{id}` | `roles.manage` | Үүрэг устгах |

## Админ — хэрэглэгчид ба AI prompt-ууд

Бүлэг `/api/v1/admin`.

| Method | Path | Auth | Зорилго |
|--------|------|------|---------|
| GET | `/admin/users` | `users.manage` | Хэрэглэгчдийг жагсаах |
| PUT | `/admin/users/{id}/role` | `users.manage` | Хэрэглэгчийн үүргийг өөрчлөх |
| PUT | `/admin/users/{id}/active` | `users.manage` | Хэрэглэгчийг идэвхжүүлэх/идэвхгүй болгох |
| DELETE | `/admin/users/{id}` | `users.manage` | Хэрэглэгч устгах |
| GET | `/admin/ai/prompts` | `settings.manage` | AI prompt давхаргын тохиргоог жагсаах |
| PUT | `/admin/ai/prompts/{key}` | `settings.manage` | AI prompt-ийн scope/зааврыг тохируулах |

## Байгууллагууд

Бүлэг `/api/v1/org`.

| Method | Path | Auth | Зорилго |
|--------|------|------|---------|
| POST | `/org/` | Bearer | Байгууллага үүсгэх |
| GET | `/org/` | Bearer | Миний байгууллагуудыг жагсаах |
| GET | `/org/lookup/{regNo}` | Bearer | Регистрийн дугаараар хайх |
| GET | `/org/{id}` | Bearer | Байгууллага авах |
| GET/POST | `/org/{id}/members` | Bearer | Гишүүдийг жагсаах / нэмэх |
| PUT/DELETE | `/org/{id}/members/{userID}` | Bearer | Гишүүнийг шинэчлэх / хасах |

## Gateway — үйлчилгээ ба телеметр

Бүлэг `/api/v1/gateway`; бүхэл бүлэг `gateway.manage`.

| Method | Path | Зорилго |
|--------|------|---------|
| GET | `/gateway/overview` | Телеметрийн тойм |
| GET | `/gateway/logs` | Хүсэлтийн лог |
| GET/POST | `/gateway/services` | Үйлчилгээ жагсаах / үүсгэх |
| PUT/DELETE | `/gateway/services/{id}` | Үйлчилгээ шинэчлэх / устгах |

## Applications — OAuth2 client-ууд { #applications-oauth2-clients }

Бүлэг `/api/v1/applications`; бүхэл бүлэг `gateway.manage` (Hydra-хаалттай).

| Method | Path | Зорилго |
|--------|------|---------|
| GET/POST | `/applications/` | Application жагсаах / үүсгэх |
| GET/PUT/DELETE | `/applications/{id}` | Авах / шинэчлэх / устгах |
| POST | `/applications/{id}/rotate-secret` | Client secret эргүүлэх |
| PUT | `/applications/{id}/services` | Зөвшөөрөгдсөн үйлчилгээг тохируулах |

## Provider — OIDC login/consent/logout

Бүлэг `/api/v1/provider` (Hydra-хаалттай). [Sign in with DAN](../sso/index.md)-г
үзнэ үү.

## Төрийн үйлчилгээний портал

Бүлэг `/api/v1/gov`; өөрчлөлтүүд хурдны хязгаартай.

| Method | Path | Зорилго |
|--------|------|---------|
| GET | `/gov/services` · `/gov/overview` | Каталог · хяналтын самбар |
| GET/POST | `/gov/applications` | Жагсаах / хүсэлт гаргах |
| POST | `/gov/applications/{id}/cancel` | Хүсэлт цуцлах |
| GET/POST | `/gov/references` | Жагсаах / лавлагаа хүсэх |
| GET | `/gov/notifications` | Мэдэгдлүүдийг жагсаах |
| POST | `/gov/notifications/read-all` · `/{id}/read` | Уншсан гэж тэмдэглэх |
| GET | `/gov/payments` | Төлбөрүүдийг жагсаах |
| POST | `/gov/payments/{id}/pay` | Төлөх |
| GET/POST | `/gov/appointments` | Жагсаах / цаг захиалах |
| POST | `/gov/appointments/{id}/cancel` | Цаг цуцлах |

## AI pipeline

Бүлэг `/api/v1/ai`; Bearer + ~20/min. [AI Pipeline](../ai/index.md)-г үзнэ үү.

| Method | Path | Зорилго |
|--------|------|---------|
| POST | `/ai/chat` | Gemini чат (Монгол fallback руу буурна) |
| POST | `/ai/stt` | Яриаг текст рүү хөрвүүлэх |
| POST | `/ai/tts` | Текстийг яриа руу хөрвүүлэх |
| POST | `/ai/translate` | Шууд орчуулга |

## eID гарын үсэг — PAdES

Бүлэг `/api/v1/sign`. [Document Signing](../signing/index.md)-г үзнэ үү.

| Method | Path | Зорилго |
|--------|------|---------|
| POST | `/sign/init` | PDF гарын үсэг эхлүүлэх (multipart, ≤25 MB) |
| GET | `/sign/{id}` | Гарын үсгийн төлөвийг poll хийх |
| GET | `/sign/{id}/download` | Гарын үсэг зурсан PDF татах |

## Integration-ууд ба storage

| Method | Path | Зорилго |
|--------|------|---------|
| GET/POST | `/integrations/` | Provider жагсаах / холбох |
| GET | `/integrations/{provider}/token` | Access token авах |
| DELETE | `/integrations/{provider}` | Салгах |
| GET | `/gspace/` · `/gspace/download` | Gerege Space тойм / татах |
| POST | `/gspace/upload` | Файл байршуулах |
| DELETE | `/gspace/` | Файл устгах |

## Аюулгүй байдал ба аудит

| Method | Path | Auth | Зорилго |
|--------|------|------|---------|
| POST | `/security/events` | Bearer | Аюулгүй байдлын үйл явдал бүртгэх (RLS-хамрагдсан) |
| GET | `/security/events` | Admin | Аюулгүй байдлын үйл явдлуудыг жагсаах |
| GET | `/audit/` | Admin | Аудитын лог бичлэгүүдийг жагсаах |
| GET | `/audit/verify` | Admin | Аудитын хэш гинжийг баталгаажуулах |

## Супер админ

Бүлэг `/api/v1/superadmin`; бүхэл бүлэг `RequireSuperAdmin`.

| Method | Path | Зорилго |
|--------|------|---------|
| GET/POST | `/superadmin/admins` | Админ жагсаах / үүсгэх |
| GET/POST | `/superadmin/admins/by-register` | Регистрийн дугаараар хайх / нэмэх |
| PUT | `/superadmin/admins/{id}/grant` | Админ эрх олгох |
| DELETE | `/superadmin/admins/{id}` | Админ эрх хураах |
| GET/POST | `/superadmin/invites` | Урилга жагсаах / үүсгэх (allow-list) |
| DELETE | `/superadmin/invites/{email}` | Урилга устгах |

Бүртгэл ба MFA endpoint-ууд (`/api/v1/auth/superadmin/*`)-ыг [Super Admin
MFA](../authentication/super-admin-mfa.md) дотор авч үзсэн.

## Сайт ба theme-ууд

| Method | Path | Auth | Зорилго |
|--------|------|------|---------|
| GET | `/site/appearance` | public | Нүүр/анонимын харагдац |
| PUT | `/site/appearance` | `settings.manage` | Харагдацыг шинэчлэх |
| GET | `/themes/active` | public | Идэвхтэй нүүр theme |
| GET/POST | `/themes` | `settings.manage` | Theme жагсаах / үүсгэх |
| GET/PUT/DELETE | `/themes/{id}` | `settings.manage` | Авах / шинэчлэх / устгах |
| PUT | `/themes/{id}/active` | `settings.manage` | Theme идэвхжүүлэх |

---

## Spec ба кодын зөрүү

Үүсгэсэн `swagger.json`-д мэдэгдэж буй дутагдал бий — route файлууд эрх бүхий:

- **Spec-д байхгүй:** RBAC болон Provider бүх бүлэг, мөн admin-users endpoint-ууд.
- **Spec-д хуучирсан:** password/OTP/бүртгэлийн auth route-ууд (`/auth/login`,
  `/auth/register`, `/auth/send-otp`, `/auth/password/*`) — эдгээрийг кодоос
  **устгасан**.
- Зарим spec зам нь давхардсан `/v1/` угтвар агуулдаг; жинхэнэ замууд нь `/api`
  дор ганц `/v1` байдаг. Spec-ийн `host` нь одоо ч `localhost:8080`.
