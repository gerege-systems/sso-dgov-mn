# Data Model & Row-Level Security

The live schema is a single consolidated migration,
`backend/migrations/1_init_schema.up.sql` (a `pg_dump` of the former migrations
1–37; the originals are archived under `migrations/old/`). The `uuid-ossp`
extension provides `uuid_generate_v4()` defaults.

## Row-Level Security (RLS)

RLS is the load-bearing per-user isolation boundary. It sits **beneath** the
`WHERE user_id = …` clauses the repositories already write, so even a query bug
cannot leak another user's rows.

### Identity on the context

`internal/datasources/rls/rls.go` is a leaf package (stdlib `context` only)
carrying `Identity{ UserID string, Role Role }`. `Role` is one of three string
constants that **must exactly match the SQL policy literals**:

| Role | Constant | Meaning |
|------|----------|---------|
| `service` | `RoleService` | trusted pre-auth / system flows (eID upsert, refresh lookup) — full access |
| `admin` | `RoleAdmin` | full access to all rows |
| `user` | `RoleUser` | only the caller's own rows |

Helpers `WithService` / `WithUser` / `WithAdmin` stamp the context. The auth
middleware sets `WithAdmin` for admin JWTs and `WithUser` otherwise;
`ServiceRLSContext` installs the `service` role on the anonymous `/auth` group.
`FromContext` returns a zero `Identity` when none is set → empty GUCs →
fail-closed.

### The `withRLS` transaction pattern

Each per-user query is wrapped in a transaction that first runs:

```sql
SELECT set_config('app.user_id', $1, true),
       set_config('app.user_role', $2, true);
```

The third arg `true` = `is_local`, i.e. `SET LOCAL` semantics — the GUCs are
scoped to the transaction and cannot leak across pooled connections. A missing
identity yields empty GUCs, which match no policy → every row hidden, every
write rejected.

### Per-table policies

Every RLS table uses `ENABLE ROW LEVEL SECURITY` **plus** `FORCE ROW LEVEL
SECURITY` (FORCE applies RLS even to the table owner). The self policy pattern:

```sql
CREATE POLICY users_self ON public.users
  USING     ((current_setting('app.user_role', true) = 'user')
             AND (id = (NULLIF(current_setting('app.user_id', true), ''))::uuid))
  WITH CHECK (…same…);
```

`NULLIF(…, '')` turns an empty GUC into `NULL` so the `::uuid` cast never errors
and the row is simply excluded.

| Table(s) | Policies |
|----------|----------|
| `users` | service / admin / self |
| `organizations`, `organization_memberships` | admin / service / member (via `SECURITY DEFINER` `app_is_org_member()` to break recursion) |
| `gov_applications`, `gov_appointments`, `gov_notifications`, `gov_payments`, `gov_references` | admin / service / self |
| `user_integrations` | admin / service / self |
| `user_recovery_codes` | admin / service / self |
| `security_events` | admin / service + `user`-role INSERT-only self |
| `superadmin_accounts` | service / admin only (no self policy) |
| `audit_log` | admin / service only |

Global/admin-config tables (`gov_services`, `gateway_*`, `applications`,
`themes`, `site_appearance`, `roles`/`permissions`, `developer_apps`,
`admin_api_keys`, `login_events`) are deliberately **not** RLS-protected — they
are guarded in the usecase/handler layer, with a DB backstop of table-privilege
`REVOKE`s against the `app_user` role.

!!! danger "The boot-time enforceability guard"
    RLS is silently bypassed by Postgres superusers and `BYPASSRLS` roles. At
    startup, `guardRLSEnforceable` (`drivers/driver_pgx.go`) queries `pg_roles`
    for `current_user`. If the role has `rolsuper` or `rolbypassrls`:
    **production fails boot** (the pool closes and the process aborts);
    development logs a warning and continues. The API must therefore connect as
    a least-privilege non-superuser role in production.

## The tables by domain

### Identity & users

- **`users`** — the account. UUID PK; `username`, optional `email`/`password`
  (nullable for eID users), `role_id` (smallint FK → `roles`), Mongolian + Latin
  name pairs, eID identity columns (`national_id`, `civil_id`, `kyc_level`,
  `document_number`, `cert_*`), Google-link columns (`google_sub`,
  `google_email`, …), soft-delete `deleted_at`, `password_changed_at`
  (token-revocation cutoff). Partial-unique indexes on `civil_id`, `national_id`,
  `google_sub`, `email`, `username`.
- **`superadmin_accounts`** — satellite table keyed by `user_id`. Holds
  super-admin credentials the `users` row deliberately omits: `civil_id`,
  `email_verified`, `mfa_enabled`, AES-GCM-encrypted `totp_secret`, `invited_by`,
  `onboarded_at`. This separation lets one person be an eID admin *and* a Google
  super-admin without colliding on the `civil_id` index. See
  [Super Admin MFA](../authentication/super-admin-mfa.md).
- **`superadmin_invites`** — email allow-list for the onboarding wizard.
- **`user_recovery_codes`** — 2FA recovery codes, SHA-256 hashes only, per-user RLS.

### RBAC

- **`roles`** — dynamic roles; seeded superadmin(1)/admin(2)/manager(3)/user(4).
- **`permissions`** — permission catalogue (e.g. `users.manage`, `gateway.manage`).
- **`role_permissions`** — role↔permission join.

### Organizations

- **`organizations`** — eID-linked orgs; case-insensitive `reg_no`, RLS by
  membership. **`organization_memberships`** — (org, user) with role text.
  **`org_stamps`** — org stamp image, usecase-guarded.

### Gov services portal

- **`gov_services`** — public service catalogue (no RLS).
  **`gov_applications`**, **`gov_references`**, **`gov_notifications`**,
  **`gov_payments`**, **`gov_appointments`** — per-user rows, each RLS-forced.

### API gateway & OAuth clients

- **`gateway_services`** — upstream backends, each with an OAuth `scope`.
  **`gateway_request_logs`** — request telemetry. **`applications`** — unified
  OAuth2 client registry (`client_id`, `app_type`, `redirect_uris`).
  **`application_services`** — application↔service grants.

### OIDC provider operator surface

- **`developer_apps`** — legacy RP overlay store. **`admin_api_keys`** —
  bootstrap admin keys (SHA-256 hash). **`login_events`** — provider login audit.

### AI pipeline

- **`ai_prompts`** — configurable prompt layers keyed by `key` (`scope`,
  `instructions`); app UPDATE-only. **`ai_knowledge`** — knowledge base for the
  `search_knowledge` tool; app read-only.

### Audit & appearance

- **`audit_log`** — append-only, hash-chained (`prev_hash`/`chain_hash`).
  **`security_events`** — RASP-style ingest. **`site_appearance`** — singleton
  look config. **`themes`** — named landing themes (`config` jsonb).
