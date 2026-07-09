# CLAUDE.md

Government Template Platform V3.0 (eID based, AI enabled) — production-ready full-stack template: Go backend (chi ·
net/http + pgx + PostgreSQL + Redis) + Next.js 15 BFF frontend + Gemini AI
pipeline. Docs index is in [README.md](README.md#documentation); deep dives in
`backend/docs/` (EN/MN pairs) and `docs/DEPLOYMENT.md`.

## Commands

```bash
# Backend (run from backend/)
go build ./...            # build
go test ./...             # unit tests (mocks only, fast)
make test-integration     # testcontainers (needs Docker)
make swag                 # regenerate swagger after touching handler annotations
make pre-push             # mirror CI: lint + test + swag drift + build

# Frontend (run from frontend/)
npm run dev               # local dev
npm run build             # build + lint + typecheck (CI runs this)

# Full stack
docker compose up -d --build   # db + redis + migrate (one-off) + api + web
```

## CI gates (push to main runs .github/workflows/ci.yml)

- **gofmt** — `gofmt -l .` must be empty; run `gofmt -w .` before committing Go code
- **swag drift** — if you add/change swagger annotations, run `make swag` and commit `backend/docs/` output, or CI fails
- go vet + `go test -race`, integration compile check, binary builds
- frontend `npm run lint` + `npm run build`
- gitleaks secrets scan

## Conventions

- **Language:** code identifiers and commit messages in English; comments and
  UI strings in Mongolian. Every source file starts with the two-line
  `Government Template Platform V3.0` header (copy from any existing file).
- **Commits:** conventional commits (`feat:`, `fix:`, `chore:`, `docs:`…).
- **EN/MN doc pairs:** when you touch `backend/docs/X.md`, update `X_MN.md`
  too (same for READMEs and `frontend/src/lib/i18n.ts` — every key exists in
  both `mn` and `en`).

## Backend architecture rules

- Clean Architecture: `handler → usecase → repository → domain`; usecases
  depend only on `repositories/interface` (package `_interface`), never on
  postgres adapters; domain imports nothing internal.
- **No ORM** — hand-written SQL via pgx; records are plain structs scanned
  with `pgx.RowToStructByName`. Parameterized queries only.
- Errors: usecases return `apperror.*` (mapped to HTTP status in
  `handler_base_response.go`); wrap internal causes with
  `apperror.InternalCause` so library errors never reach clients.
- Handlers: `func(w, r) error` wrapped by `v1.Wrap`; decode with
  `v1.DecodeBody`, validate DTOs with `validators.ValidatePayloads` (struct
  tags), respond via `v1.NewSuccessResponse` / `v1.RespondWithError`. Carry
  swagger annotations.
- Wiring is manual DI in `cmd/api/server/server.go` (repo → usecase → route).
- Migrations: numbered SQL files in `backend/migrations/` (`N_name.up.sql` +
  `.down.sql`); the `migrate` compose service applies them on every `up`.
- **RLS:** the api must connect as a non-superuser role (boot guard enforces
  this in production). `users`-table queries go through `withRLS`
  transactions; new per-user tables need their own policies.
- Add-a-feature walkthrough: `backend/docs/DEVELOPMENT.md`.

## AI pipeline (backend/docs/AI_PIPELINE.md)

- `pkg/gemini` is SDK-free REST; usecase layer is `usecases/ai`.
- System prompt = hardcoded guardrails + DB-configurable `scope`/
  `instructions` (`ai_prompts` table, admin API). Never make the guardrail
  layer configurable.
- Tools (`ai.ToolDef`) run server-side with the request context; register in
  `server.go`. Knowledge base lives in `ai_knowledge`.
- Chat degrades to a Mongolian fallback reply (`degraded: true`) on transient
  Gemini failures — don't turn that into a 5xx.

## Frontend rules

- BFF model: browser → same-origin `/api/*` route handlers only; tokens live
  in httpOnly cookies and never reach client JS. Backend errors are proxied
  via `proxyResult`/`toClientResponse` (never leak tokens).
- All mutating browser calls go through `lib/client.ts` `sendJSON`/`postJSON`
  (adds the `x-dgov-csrf` header that `lib/bff.ts checkOrigin` requires).
  New mutating BFF routes must call `checkOrigin` first.
- Server data fetching in components uses TanStack Query (`getJSON` +
  `useQuery`, invalidate on mutations); provider is in
  `components/Providers.tsx`.
- Don't call backend refresh in RSC context — `tryRefresh` probes cookie
  writability first because refresh **rotates** the token (see `lib/api.ts`).
- UI strings via `useT()` + `lib/i18n.ts` keys (mn + en).

## Gotchas

- `backend/internal/config/.env*` and root `.env`/`backend.env` are
  gitignored secrets — never commit; document new env vars in the READMEs.
- `/ai/*` rate limit is ~20 req/min per IP (live translation streams ~8
  chunks/min); auth endpoints ~5/min with a 4 KiB body cap.
- The compose stack runs `ENVIRONMENT=development` on purpose (internal DB
  has no TLS; the production guard requires `sslmode=verify-full`).
