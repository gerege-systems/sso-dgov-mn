# Backend Layers

## Directory map

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

Repository sub-packages under `repositories/postgres/`: `users`, `rbac`, `org`,
`gov`, `gateway`, `applications`, `audit`, `security`, `site`, `theme`, `ai`,
`orgstamp`, `superadminaccount`, `superadmininvite`, `recovery`,
`userintegrations`.

## Dependency injection — `server.go`

`NewApp()` is the single manual-DI composition root — no DI container, no global
singletons. It runs a fixed sequence:

1. **Infra init** — tracing, the pgx pool (`drivers.SetupPgxPostgres`, which runs
   the RLS boot guard), the JWT service, Redis + Ristretto caches.
2. **External clients** — `verify`, `eid`, `google`, `xyp`, `gemini`, `hydra`,
   `gspace`.
3. **repo → usecase → route, by hand**, per bounded context.

The canonical triple:

```go
userRepo := userspostgres.NewUserRepository(pool)     // repo (returns _interface.UserRepository)
usersUC  := users.NewUsecase(userRepo, ristretto, users.Config{
    BcryptCost: config.AppConfig.BcryptCost,
})                                                    // usecase (takes the interface)
// … inside r.Route("/api", …):
routes.NewUsersRoute(api, usersUC, authMiddleware).Routes()   // route (takes the usecase)
```

Every module follows this shape and mounts under a single `r.Route("/api", …)`
block, each route object registering `/v1/<module>`.

### Conditional wiring

Feature groups are constructed and mounted only when their config is present:
the OIDC-provider usecase and the Applications registry require Hydra
(`ProviderConfigured()`); the super-admin onboarding surface requires
`INTEGRATION_ENC_KEY`; the `/rp/sign` relay requires both a relay token and the
eID RP secret.

### Gateway request logging

The gateway request-log middleware writes to a buffered queue (cap 512) drained
by four worker goroutines, so edge traffic is never blocked by DB latency; on a
full queue the log line is dropped (best-effort). Graceful shutdown drains HTTP,
stops the rate limiters, and closes the pool, Redis, and tracer.

## The data layer — pgx, no ORM

- **Driver:** `pgx/v5` + `pgxpool`, configured in `drivers/driver_pgx.go`. The
  pool attaches an `otelpgx` tracer for per-statement spans and sizes itself from
  `DB_MAX_OPEN_CONNS` (default 25), `DB_MAX_IDLE_CONNS` (default 5),
  `DB_CONN_MAX_LIFE_MINS` (default 15). The DSN is env-selected (development vs
  production).
- **No AutoMigrate.** Schema comes only from `*.up.sql` files.

### Struct ↔ table mapping

Rows are plain structs in `datasources/records/`, scanned by name. `records.Users`
carries `db:"…"` snake_case tags that `pgx.RowToStructByName` matches to columns.
Nullable columns are pointer types (`*string`, `*time.Time`) so SQL `NULL` maps
to `nil` — eID users have no email/password. A single `records.UserColumns`
constant centralizes the SELECT/RETURNING list so name-based scanning stays
stable.

Record ↔ domain conversion is **explicit**, not reflective. Mappers dereference
nullable pointers (`derefStr`) one way and convert empty strings back to `NULL`
(`ptrOrNil`) the other — this is how an eID user's empty `national_id` becomes
SQL `NULL` so a partial-unique index doesn't collide across users.

### Query idioms

The `users` repository is the reference implementation:

- Single round-trip `INSERT … RETURNING <UserColumns>`, scanned with
  `pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[records.Users])`.
- Duplicate keys: `pgconn.PgError` code `23505` → `apperror.Conflict`.
- Soft delete: every query adds `deleted_at IS NULL` explicitly; `SoftDelete`
  sets `deleted_at = NOW()`.
- Parameterized queries only (`$1, $2, …`) — the SQL-injection backstop.
- Method-per-file layout keeps PR diffs narrow.

## Error handling

`apperror.DomainError{ Type, Message, Cause }` has six `ErrorType` values.
`Unwrap()` exposes `Cause` so `errors.Is`/`errors.As` traverse to wrapped causes
without leaking their text.

`mapDomainErrorToHTTP` (in `handler_base_response.go`) maps:

| ErrorType | HTTP |
|-----------|------|
| NotFound | 404 |
| Unauthorized | 401 |
| Forbidden | 403 |
| Conflict | 409 |
| BadRequest | 400 |
| (validation) | 422 with per-field `data.errors` |
| default / Internal | 500 |

`RespondWithError` is the central sink: for status ≥ 500 it logs the underlying
cause with the request path and `request_id`, then replaces the body message
with the generic `"internal server error"`. Handlers are `func(w, r) error`
wrapped by `v1.Wrap`; responses use the single `BaseResponse{ status, message,
data, request_id }` envelope.
