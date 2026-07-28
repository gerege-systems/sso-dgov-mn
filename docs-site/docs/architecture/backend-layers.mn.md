# Backend Layers

## Директорын зураглал

```
backend/
├── cmd/
│   ├── api/main.go                     # entry point (config + logger init)
│   ├── api/server/server.go            # composition root — manual DI, all mounts
│   ├── migration/                      # SQL migration CLI (no AutoMigrate)
│   └── seed/                           # DB seeder CLI
├── internal/
│   ├── apperror/error.go               # typed DomainError → HTTP status
│   ├── config/config.go                # Viper config + production guard
│   ├── constants/                      # env/logger/error/endpoint constants
│   ├── business/
│   │   ├── domain/                     # entities (innermost ring)
│   │   └── usecases/                   # bounded contexts (interface + impl)
│   ├── datasources/
│   │   ├── drivers/driver_pgx.go       # pgxpool setup + RLS-enforceability guard
│   │   ├── caches/                     # Redis + Ristretto two-tier cache
│   │   ├── records/                    # pgx record structs + record↔domain mappers
│   │   ├── rls/rls.go                  # per-request RLS Identity on context
│   │   └── repositories/
│   │       ├── interface/              # gateway abstractions (package _interface)
│   │       └── postgres/<module>/      # pgx hand-written SQL adapters (withRLS)
│   ├── http/
│   │   ├── handlers/v1/                # HTTP handlers; handler_base_response.go
│   │   ├── routes/                     # route registration per module
│   │   ├── middlewares/                # global + per-group middleware
│   │   └── datatransfers/              # request/response DTOs (validate tags)
│   └── provider/                       # OIDC-provider operator surfaces
├── migrations/                         # 1_init_schema.up/down.sql (+ old/ history)
└── pkg/                                # eid, google, oidc, hydra, xyp, gemini, jwt…
```

`repositories/postgres/` доорх repository дэд package-ууд: `users`, `rbac`, `org`,
`gov`, `gateway`, `applications`, `audit`, `security`, `site`, `theme`, `ai`,
`orgstamp`, `superadminaccount`, `superadmininvite`, `recovery`,
`userintegrations`.

## Dependency injection — `server.go`

`NewApp()` нь гар аргаар хийсэн ганц DI composition root юм — DI container
байхгүй, глобал singleton байхгүй. Энэ нь тогтмол дараалал ажиллуулна:

1. **Infra init** — tracing, pgx pool (`drivers.SetupPgxPostgres`, RLS boot
   guard-ыг ажиллуулна), JWT service, Redis + Ristretto кэш.
2. **Гадаад client-ууд** — `verify`, `eid`, `google`, `xyp`, `gemini`, `hydra`,
   `gspace`.
3. **repo → usecase → route, гар аргаар**, bounded context тус бүрээр.

Каноник гурвал:

```go
userRepo := userspostgres.NewUserRepository(pool)     // repo (returns _interface.UserRepository)
usersUC  := users.NewUsecase(userRepo, ristretto, users.Config{
    BcryptCost: config.AppConfig.BcryptCost,
})                                                    // usecase (takes the interface)
// … inside r.Route("/api", …):
routes.NewUsersRoute(api, usersUC, authMiddleware).Routes()   // route (takes the usecase)
```

Модуль бүр энэ хэлбэрийг дагаж, нэг `r.Route("/api", …)` блок дор mount хийгддэг
бөгөөд route object бүр `/v1/<module>`-ыг бүртгэдэг.

### Нөхцөлт холболт (Conditional wiring)

Feature бүлгүүд нь зөвхөн тохиргоо нь байгаа үед бүтээгдэж, mount хийгддэг:
OIDC-provider usecase болон Applications registry нь Hydra-г
(`ProviderConfigured()`) шаарддаг; super-admin онбординг гадаргуу нь
`INTEGRATION_ENC_KEY`-г шаарддаг; `/rp/sign` relay нь relay token болон eID RP
secret хоёуланг шаарддаг.

### Gateway хүсэлтийн логжуулалт

Gateway request-log middleware нь дөрвөн worker goroutine-ээр асгардаг buffer-тэй
queue руу (cap 512) бичдэг тул edge урсгал нь DB latency-ээр хэзээ ч
блоклогддоггүй; queue дүүрэн үед log мөр хаягддаг (best-effort). Хэвийн зогсолт
(graceful shutdown) нь HTTP-г асгаж, rate limiter-үүдийг зогсоож, pool, Redis,
tracer-ыг хаадаг.

## Өгөгдлийн давхарга — pgx, ORM ашиглаагүй

- **Driver:** `pgx/v5` + `pgxpool`, `drivers/driver_pgx.go` дотор тохируулагдсан.
  Pool нь per-statement span-д зориулж `otelpgx` tracer хавсаргаж,
  `DB_MAX_OPEN_CONNS` (default 25), `DB_MAX_IDLE_CONNS` (default 5),
  `DB_CONN_MAX_LIFE_MINS` (default 15)-ээс өөрийгөө хэмжинэ. DSN нь env-ээр
  сонгогддог (development эсэх production).
- **AutoMigrate байхгүй.** Схем нь зөвхөн `*.up.sql` файлуудаас гарна.

### Struct ↔ table зураглал

Мөрүүд нь `datasources/records/` доторх энгийн struct-ууд бөгөөд нэрээр scan
хийгддэг. `records.Users` нь `pgx.RowToStructByName` баганатай тааруулдаг
`db:"…"` snake_case tag-уудыг агуулна. Nullable багана нь pointer төрлүүд юм
(`*string`, `*time.Time`) тул SQL `NULL` нь `nil` руу зураглагдана — eID
хэрэглэгчид имэйл/нууц үггүй. Ганц `records.UserColumns` тогтмол нь
SELECT/RETURNING жагсаалтыг төвлөрүүлдэг тул нэр дээр суурилсан scan тогтвортой
хэвээр байна.

Record ↔ domain хөрвүүлэлт нь reflective биш, **тодорхой (explicit)** хийгддэг.
Mapper-ууд нь нэг талд nullable pointer-үүдийг dereference (`derefStr`) хийж,
нөгөө талд хоосон string-ийг `NULL` (`ptrOrNil`) руу хөрвүүлдэг — энэ л eID
хэрэглэгчийн хоосон `national_id`-г SQL `NULL` болгож, partial-unique index
хэрэглэгчид хооронд мөргөлдөхгүй байх аргачлал юм.

### Query idiom-ууд

`users` repository нь эталон хэрэгжүүлэлт юм:

- Ганц round-trip `INSERT … RETURNING <UserColumns>`,
  `pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[records.Users])`-ээр scan
  хийгдсэн.
- Давхардсан түлхүүр: `pgconn.PgError` код `23505` → `apperror.Conflict`.
- Soft delete: query бүр `deleted_at IS NULL`-ыг тодорхойгоор нэмдэг; `SoftDelete`
  нь `deleted_at = NOW()`-ыг тохируулна.
- Зөвхөн параметржүүлсэн query (`$1, $2, …`) — SQL-injection-ийн хамгаалалт.
- Method-per-file зохион байгуулалт нь PR diff-ыг нарийхан байлгадаг.

## Алдааны боловсруулалт { #error-handling }

`apperror.DomainError{ Type, Message, Cause }` нь зургаан `ErrorType` утгатай.
`Unwrap()` нь `Cause`-ыг ил гаргадаг тул `errors.Is`/`errors.As` нь боож
өнхрүүлсэн шалтгаан руу тэдгээрийн текстийг задруулалгүйгээр нэвтэрч чадна.

`mapDomainErrorToHTTP` (`handler_base_response.go` дотор) дараах байдлаар
зураглана:

| ErrorType | HTTP |
|-----------|------|
| NotFound | 404 |
| Unauthorized | 401 |
| Forbidden | 403 |
| Conflict | 409 |
| BadRequest | 400 |
| (validation) | 422, per-field `data.errors`-тэй |
| default / Internal | 500 |

`RespondWithError` нь төвийн sink юм: статус ≥ 500 үед энэ нь хүсэлтийн зам болон
`request_id`-тай хамт үндсэн шалтгааныг лог хийж, дараа нь body мессежийг ерөнхий
`"internal server error"`-ээр солино. Handler-ууд нь `v1.Wrap`-ээр боож
өнхрүүлсэн `func(w, r) error` юм; хариултууд нь ганц `BaseResponse{ status,
message, data, request_id }` дугтуйг ашигладаг.
