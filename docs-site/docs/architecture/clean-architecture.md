# Clean Architecture

The backend follows a four-ring Clean Architecture where **dependencies point
inward only**: `handler → usecase → repository → domain`.

## The dependency rule

| Ring | Package root | May import | Must not import |
|------|--------------|-----------|-----------------|
| HTTP (delivery) | `internal/http/` | usecase interfaces (`auth.Usecase`, `users.Usecase`, …) | — |
| Usecase | `internal/business/usecases/` | `repositories/interface` + domain | postgres adapters, chi/net-http |
| Repository | `internal/datasources/repositories/postgres/` | domain, `records`, `rls` | — |
| Domain | `internal/business/domain/` | stdlib + `golang.org/x/crypto/bcrypt` | anything in `internal/` or `pkg/` |

The domain package documents its own constraint at
`internal/business/domain/domain_users.go` — it imports only the standard
library and `bcrypt`.

### The one sanctioned exception

"Domain imports nothing internal" has a single deliberate exception: the leaf
package `internal/datasources/rls/rls.go`. It depends only on stdlib `context`
and is shared across all three outer rings so it can carry per-request RLS
identity without creating an import cycle.

## The `_interface` package

The gateway abstractions live in package `_interface` at
`internal/datasources/repositories/interface/interface.go`. The package is named
`_interface` because `interface` is a Go reserved word; the leading underscore
keeps it a valid identifier.

Usecases depend **only** on these interfaces — for example
`_interface.UserRepository` — never on the concrete `postgres` adapters. The
concrete adapter is constructed in the composition root and injected:

```go
// server.go — repo returns the interface type, not the concrete struct
userRepo := userspostgres.NewUserRepository(pool) // → _interface.UserRepository
usersUC  := users.NewUsecase(userRepo, cache, users.Config{...})
```

This is what makes the storage engine swappable without touching business code;
`interface.go` even names a hypothetical future `mongo/` adapter as a sibling
implementation.

!!! note "Interfaces are split across files"
    Most gateways live in `interface.go`, but `OrgStampRepository` and
    `UserIntegrationsRepository` live in `interface_org_stamp.go` /
    `interface_user_integrations.go`. Grepping only `interface.go` will miss them.

## Errors cross the rings as typed values

Usecases return `apperror.*` values (`apperror.NotFound`, `Unauthorized`,
`Forbidden`, `Conflict`, `BadRequest`, `Internal`). The HTTP layer maps the
`ErrorType` to a status code in `handler_base_response.go`. Library/internal
errors are wrapped with `apperror.InternalCause(cause)` — which fixes the
client-facing message to `"internal server error"` while retaining the real
cause for logging — so a wrapped `pgx`/`sql` error never reaches a client. See
[Backend Layers](backend-layers.md#error-handling).
