# Data Model & Row-Level Security

Ажиллаж буй схем нь нэгтгэсэн ганц migration болох
`backend/migrations/1_init_schema.up.sql` юм (өмнөх migration 1–37-ийн
`pg_dump`; эх хувилбарууд нь `migrations/old/` дор архивлагдсан). `uuid-ossp`
extension нь `uuid_generate_v4()` default-уудыг хангадаг.

## Row-Level Security (RLS)

RLS нь хэрэглэгч тус бүрийн тусгаарлалтыг үүрдэг гол хамгаалалтын хил юм. Энэ нь
репозиторийн аль хэдийн бичсэн `WHERE user_id = …` нөхцлийн **доор** байрлах тул
query-ийн алдаа гарлаа ч өөр хэрэглэгчийн мөрийг задруулж чадахгүй.

### Context дээрх identity

`internal/datasources/rls/rls.go` нь `Identity{ UserID string, Role Role }`-г
зөөж явдаг leaf package юм (зөвхөн stdlib `context`). `Role` нь **SQL policy-ийн
литералуудтай яг таарах ёстой** гурван string тогтмолын нэг:

| Role | Тогтмол | Утга |
|------|----------|-------|
| `service` | `RoleService` | итгэмжлэгдсэн pre-auth / систем урсгал (eID upsert, refresh lookup) — бүрэн хандалт |
| `admin` | `RoleAdmin` | бүх мөрөнд бүрэн хандалт |
| `user` | `RoleUser` | зөвхөн дуудагчийн өөрийн мөрүүд |

`WithService` / `WithUser` / `WithAdmin` туслах функцууд нь context-ыг тамгална.
Auth middleware нь admin JWT-д `WithAdmin`, бусад тохиолдолд `WithUser`-ыг
тохируулна; `ServiceRLSContext` нь нэрлэгдээгүй (anonymous) `/auth` бүлэгт
`service` роль суулгана. `FromContext` нь юу ч тохируулагдаагүй үед тэг `Identity`
буцаана → хоосон GUC-ууд → fail-closed.

### `withRLS` transaction загвар

Хэрэглэгч тус бүрийн query бүр эхлээд дараахыг ажиллуулдаг transaction дотор
боож өнхрүүлэгддэг:

```sql
SELECT set_config('app.user_id', $1, true),
       set_config('app.user_role', $2, true);
```

Гуравдахь аргумент `true` = `is_local`, өөрөөр хэлбэл `SET LOCAL` семантик —
GUC-ууд нь transaction-д хамрагдах ба pool хийсэн холболтуудаар дамжин задарч
чадахгүй. Дутуу identity нь хоосон GUC-ыг үүсгэх ба энэ нь ямар ч policy-тэй
таарахгүй → бүх мөр нуугдаж, бүх бичилт татгалзагдана.

### Хүснэгт тус бүрийн policy-ууд

RLS хүснэгт бүр `ENABLE ROW LEVEL SECURITY` **дээр нэмээд** `FORCE ROW LEVEL
SECURITY`-г ашигладаг (FORCE нь хүснэгтийн эзэнд ч RLS-ыг хэрэглэдэг). Self
policy загвар:

```sql
CREATE POLICY users_self ON public.users
  USING     ((current_setting('app.user_role', true) = 'user')
             AND (id = (NULLIF(current_setting('app.user_id', true), ''))::uuid))
  WITH CHECK (…same…);
```

`NULLIF(…, '')` нь хоосон GUC-ыг `NULL` болгодог тул `::uuid` cast хэзээ ч алдаа
гаргахгүй бөгөөд мөр нь зүгээр л хасагдана.

| Хүснэгт(үүд) | Policy-ууд |
|----------|----------|
| `users` | service / admin / self |
| `organizations`, `organization_memberships` | admin / service / member (рекурсийг таслахын тулд `SECURITY DEFINER` `app_is_org_member()`-ээр) |
| `gov_applications`, `gov_appointments`, `gov_notifications`, `gov_payments`, `gov_references` | admin / service / self |
| `user_integrations` | admin / service / self |
| `user_recovery_codes` | admin / service / self |
| `security_events` | admin / service + `user`-role зөвхөн INSERT self |
| `superadmin_accounts` | зөвхөн service / admin (self policy байхгүй) |
| `audit_log` | зөвхөн admin / service |

Глобал/админ-тохиргооны хүснэгтүүд (`gov_services`, `gateway_*`, `applications`,
`themes`, `site_appearance`, `roles`/`permissions`, `developer_apps`,
`admin_api_keys`, `login_events`) нь зориудаар RLS-ээр хамгаалагдаагүй **биш** —
тэдгээр нь usecase/handler давхаргад хамгаалагдаж, `app_user` роль эсрэг
table-privilege `REVOKE`-ийн DB backstop-той.

!!! danger "Boot үеийн хэрэгжүүлэлтийн guard"
    RLS нь Postgres superuser болон `BYPASSRLS` роль-уудаар чимээгүйхэн
    тойрогдож (bypass) болдог. Эхлэх үед `guardRLSEnforceable`
    (`drivers/driver_pgx.go`) нь `current_user`-ыг `pg_roles`-оос асуудаг. Хэрэв
    роль нь `rolsuper` эсвэл `rolbypassrls`-тэй бол: **продакшн boot амжилтгүй
    болно** (pool хаагдаж, процесс тасарна); development нь анхааруулга лог хийж,
    үргэлжилнэ. Тиймээс API нь продакшнд хамгийн бага эрхтэй non-superuser роль
    болон холбогдох ёстой.

## Домэйнээр ангилсан хүснэгтүүд

### Identity & хэрэглэгчид

- **`users`** — бүртгэл. UUID PK; `username`, сонголтоор `email`/`password`
  (eID хэрэглэгчдэд nullable), `role_id` (smallint FK → `roles`), Монгол + латин
  нэрийн хос, eID identity багана (`national_id`, `civil_id`, `kyc_level`,
  `document_number`, `cert_*`), Google-холбоос багана (`google_sub`,
  `google_email`, …), soft-delete `deleted_at`, `password_changed_at`
  (token-revocation-ийн зааг). `civil_id`, `national_id`, `google_sub`, `email`,
  `username` дээрх partial-unique index-үүд.
- **`superadmin_accounts`** — `user_id`-ээр түлхүүрлэгдсэн satellite хүснэгт.
  `users` мөр зориудаар орхигдуулсан super-admin credential-уудыг агуулна:
  `civil_id`, `email_verified`, `mfa_enabled`, AES-GCM-ээр шифрлэгдсэн
  `totp_secret`, `invited_by`, `onboarded_at`. Энэ тусгаарлалт нь нэг хүн eID
  admin болон Google super-admin хоёулаа болох боломжийг `civil_id` index дээр
  мөргөлдөлгүйгээр олгодог. [Super Admin
  MFA](../authentication/super-admin-mfa.md)-ийг үз.
- **`superadmin_invites`** — онбординг wizard-ын имэйл зөвшөөрлийн жагсаалт.
- **`user_recovery_codes`** — 2FA recovery код, зөвхөн SHA-256 hash,
  хэрэглэгч тус бүрийн RLS.

### RBAC

- **`roles`** — динамик роль; superadmin(1)/admin(2)/manager(3)/user(4)-ээр seed
  хийгдсэн.
- **`permissions`** — эрхийн каталог (жишээ нь `users.manage`, `gateway.manage`).
- **`role_permissions`** — роль↔эрхийн join.

### Байгууллагууд

- **`organizations`** — eID-тэй холбогдсон байгууллага; том-жижиг үсэг ялгадаггүй
  `reg_no`, гишүүнчлэлээр RLS. **`organization_memberships`** — (org, user)
  роль текст-тэй. **`org_stamps`** — байгууллагын тамганы зураг, usecase-ээр
  хамгаалагдсан.

### Төрийн үйлчилгээний портал

- **`gov_services`** — нийтийн үйлчилгээний каталог (RLS байхгүй).
  **`gov_applications`**, **`gov_references`**, **`gov_notifications`**,
  **`gov_payments`**, **`gov_appointments`** — хэрэглэгч тус бүрийн мөрүүд, тус
  бүр RLS-force-той.

### API gateway & OAuth client-ууд

- **`gateway_services`** — upstream backend-ууд, тус бүр OAuth `scope`-той.
  **`gateway_request_logs`** — хүсэлтийн telemetry. **`applications`** —
  нэгдсэн OAuth2 client registry (`client_id`, `app_type`, `redirect_uris`).
  **`application_services`** — application↔service grant-ууд.

### OIDC provider операторын гадаргуу

- **`developer_apps`** — legacy RP overlay store. **`admin_api_keys`** —
  bootstrap admin key (SHA-256 hash). **`login_events`** — provider нэвтрэлтийн
  audit.

### AI pipeline

- **`ai_prompts`** — `key`-ээр түлхүүрлэгдсэн тохируулж болох prompt давхаргууд
  (`scope`, `instructions`); апп зөвхөн UPDATE хийнэ. **`ai_knowledge`** —
  `search_knowledge` tool-д зориулсан мэдлэгийн сан; апп зөвхөн уншина.

### Audit & гадаад төрх

- **`audit_log`** — зөвхөн нэмдэг (append-only), hash-chain хэлбэрийн
  (`prev_hash`/`chain_hash`). **`security_events`** — RASP-маягийн ingest.
  **`site_appearance`** — singleton загварын тохиргоо. **`themes`** — нэрлэсэн
  landing theme-үүд (`config` jsonb).
