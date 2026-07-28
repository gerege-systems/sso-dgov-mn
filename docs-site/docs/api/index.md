# REST API Reference

Source of truth: the Go route files (`internal/http/routes/route_*.go`),
cross-referenced against `swagger.json` and `API_CONTRACT.md`. For an interactive
view see the [API Explorer](explorer.md).

## Global conventions

### Base path & versioning

Every business endpoint lives under **`/api`** and each domain group mounts a
**`/v1/...`** subtree, so the effective path is `/api/v1/<group>/<path>` — e.g.
`POST /api/v1/auth/eid/start`. Two groups deviate in mount prefix (still under
`/api/v1`): personal assets at `/api/v1/me/*` and eID profile at
`/api/v1/users/me/eid/*`.

### Authentication

Bearer JWT in `Authorization: Bearer <token>`, validated by the auth middleware
(which also checks a Redis revocation list). Permission gates layer on top:
`RequirePermission(<perm>)`, `RequireAdmin()`, `RequireSuperAdmin()`.

### Response envelope

```json
{ "status": true, "message": "…", "data": {…}, "request_id": "…" }
```

Errors map `apperror` types to status (404/401/403/409/400, else 500);
validation errors return **422** with a `data.errors` field map. 5xx causes are
logged and replaced with `"internal server error"`.

### Global middleware chain

`Tracing → RequestID → Recoverer → Metrics → SecurityHeaders → CORS →
BodySizeLimit(26 MiB) → AccessLog → Timeout(30 s)`. The `/api` subtree adds
async gateway request logging.

### Body-size caps

| Scope | Cap |
|-------|-----|
| Global | 26 MiB (sign uploads need it) |
| Auth groups (`/auth`, `/auth/superadmin`, `/provider`) | 4 KiB |
| Normal JSON (`DecodeBody`) | 1 MiB |

### Rate limits (per client IP)

| Limiter | Rate / burst | Applied to |
|---------|--------------|------------|
| auth | ~5/min, burst 5 | `/auth/*`, `/auth/superadmin/*` |
| ai | ~20/min, burst 10 | `/ai/*` |
| poll | ~60/min, burst 30 | eID long-poll endpoints |
| gov-write | ~30/min, burst 15 | mutations in `/gov`, `/me`, `/gspace`, eID |

### Infrastructure endpoints (outside `/api`)

| Method | Path | Auth | Purpose |
|--------|------|------|---------|
| GET | `/health` | none | Liveness (DB + Redis) |
| GET | `/ready` | none | Readiness |
| GET | `/metrics` | ObservabilityGate | Prometheus (404 in prod without token) |
| GET | `/swagger/doc.json` | ObservabilityGate | OpenAPI document |

---

## Auth

Group `/api/v1/auth`; body cap 4 KiB + `ServiceRLSContext`. eID is the only login
method. See [Authentication](../authentication/index.md).

| Method | Path | Auth | Purpose |
|--------|------|------|---------|
| POST | `/auth/eid/start` | none · 5/min | Start eID login (QR / deep-link) |
| POST | `/auth/eid/start-id` | none · 5/min | Start eID login by national ID (push) |
| POST | `/auth/eid/poll` | none · ~60/min | Long-poll eID login status |
| POST | `/auth/google` | none · 5/min | Google OAuth callback → link/login |
| POST | `/auth/refresh` | none · 5/min | Rotate session (single-use refresh) |
| POST | `/auth/logout` | none · 5/min | Invalidate session |
| DELETE | `/auth/google/link` | Bearer | Unlink connected Google account |

## Users & eID profile

| Method | Path | Auth | Purpose |
|--------|------|------|---------|
| GET | `/users/me` | Bearer | Current user profile |
| GET | `/users/me/eid/summary` | Bearer | eID account summary |
| GET | `/users/me/eid/certificates` | Bearer | eID certificates |
| GET | `/users/me/eid/devices` | Bearer | Registered devices |
| GET | `/users/me/eid/activity` | Bearer | eID activity log |
| GET/POST | `/users/me/eid/organizations` | Bearer | List / link organizations |
| DELETE | `/users/me/eid/organizations/{regNo}` | Bearer | Unlink organization |
| GET/POST/DELETE | `/users/me/eid/organizations/{regNo}/signers` | Bearer | Manage org signers |

## Personal assets — signatures & stamps

Group `/api/v1/me`; writes rate-limited.

| Method | Path | Auth | Purpose |
|--------|------|------|---------|
| GET/PUT/DELETE | `/me/signature` | Bearer | Personal signature image |
| PUT | `/me/latin-name` | Bearer | Correct Latin transliteration of own name |
| PUT | `/me/org-name-latin/{regNo}` | Bearer | Correct org Latin name |
| GET/PUT/DELETE | `/me/orgstamp/{regNo}` | Bearer (PUT: org ADMIN) | Organization stamp image |

## RBAC / Roles

Group `/api/v1/rbac`. Management requires `roles.manage` (admin auto-passes).

| Method | Path | Auth | Purpose |
|--------|------|------|---------|
| GET | `/rbac/me` | Bearer | Current user's permissions (menu filtering) |
| GET | `/rbac/roles` | `roles.manage` | List roles |
| GET | `/rbac/permissions` | `roles.manage` | List permissions |
| POST | `/rbac/roles` | `roles.manage` | Create role |
| PUT | `/rbac/roles/{id}` | `roles.manage` | Update role |
| PUT | `/rbac/roles/{id}/permissions` | `roles.manage` | Set role permissions |
| DELETE | `/rbac/roles/{id}` | `roles.manage` | Delete role |

## Admin — users & AI prompts

Group `/api/v1/admin`.

| Method | Path | Auth | Purpose |
|--------|------|------|---------|
| GET | `/admin/users` | `users.manage` | List users |
| PUT | `/admin/users/{id}/role` | `users.manage` | Change a user's role |
| PUT | `/admin/users/{id}/active` | `users.manage` | Enable/disable a user |
| DELETE | `/admin/users/{id}` | `users.manage` | Delete a user |
| GET | `/admin/ai/prompts` | `settings.manage` | List AI prompt-layer config |
| PUT | `/admin/ai/prompts/{key}` | `settings.manage` | Set AI prompt scope/instructions |

## Organizations

Group `/api/v1/org`.

| Method | Path | Auth | Purpose |
|--------|------|------|---------|
| POST | `/org/` | Bearer | Create organization |
| GET | `/org/` | Bearer | List my organizations |
| GET | `/org/lookup/{regNo}` | Bearer | Lookup by registration number |
| GET | `/org/{id}` | Bearer | Get organization |
| GET/POST | `/org/{id}/members` | Bearer | List / add members |
| PUT/DELETE | `/org/{id}/members/{userID}` | Bearer | Update / remove member |

## Gateway — services & telemetry

Group `/api/v1/gateway`; whole group `gateway.manage`.

| Method | Path | Purpose |
|--------|------|---------|
| GET | `/gateway/overview` | Telemetry overview |
| GET | `/gateway/logs` | Request logs |
| GET/POST | `/gateway/services` | List / create service |
| PUT/DELETE | `/gateway/services/{id}` | Update / delete service |

## Applications — OAuth2 clients

Group `/api/v1/applications`; whole group `gateway.manage` (Hydra-gated).

| Method | Path | Purpose |
|--------|------|---------|
| GET/POST | `/applications/` | List / create application |
| GET/PUT/DELETE | `/applications/{id}` | Get / update / delete |
| POST | `/applications/{id}/rotate-secret` | Rotate client secret |
| PUT | `/applications/{id}/services` | Set allowed services |

## Provider — OIDC login/consent/logout

Group `/api/v1/provider` (Hydra-gated). See [Sign in with DAN](../sso/index.md).

## Gov services portal

Group `/api/v1/gov`; mutations rate-limited.

| Method | Path | Purpose |
|--------|------|---------|
| GET | `/gov/services` · `/gov/overview` | Catalogue · dashboard |
| GET/POST | `/gov/applications` | List / apply |
| POST | `/gov/applications/{id}/cancel` | Cancel application |
| GET/POST | `/gov/references` | List / request reference |
| GET | `/gov/notifications` | List notifications |
| POST | `/gov/notifications/read-all` · `/{id}/read` | Mark read |
| GET | `/gov/payments` | List payments |
| POST | `/gov/payments/{id}/pay` | Pay |
| GET/POST | `/gov/appointments` | List / book |
| POST | `/gov/appointments/{id}/cancel` | Cancel appointment |

## AI pipeline

Group `/api/v1/ai`; Bearer + ~20/min. See [AI Pipeline](../ai/index.md).

| Method | Path | Purpose |
|--------|------|---------|
| POST | `/ai/chat` | Gemini chat (degrades to Mongolian fallback) |
| POST | `/ai/stt` | Speech-to-text |
| POST | `/ai/tts` | Text-to-speech |
| POST | `/ai/translate` | Live translation |

## eID signing — PAdES

Group `/api/v1/sign`. See [Document Signing](../signing/index.md).

| Method | Path | Purpose |
|--------|------|---------|
| POST | `/sign/init` | Initiate PDF signature (multipart, ≤25 MB) |
| GET | `/sign/{id}` | Poll signing status |
| GET | `/sign/{id}/download` | Download signed PDF |

## Integrations & storage

| Method | Path | Purpose |
|--------|------|---------|
| GET/POST | `/integrations/` | List / connect a provider |
| GET | `/integrations/{provider}/token` | Get access token |
| DELETE | `/integrations/{provider}` | Disconnect |
| GET | `/gspace/` · `/gspace/download` | Gerege Space overview / download |
| POST | `/gspace/upload` | Upload a file |
| DELETE | `/gspace/` | Delete a file |

## Security & audit

| Method | Path | Auth | Purpose |
|--------|------|------|---------|
| POST | `/security/events` | Bearer | Ingest a security event (RLS-scoped) |
| GET | `/security/events` | Admin | List security events |
| GET | `/audit/` | Admin | List audit log entries |
| GET | `/audit/verify` | Admin | Verify audit hash chain |

## Super admin

Group `/api/v1/superadmin`; whole group `RequireSuperAdmin`.

| Method | Path | Purpose |
|--------|------|---------|
| GET/POST | `/superadmin/admins` | List / create admins |
| GET/POST | `/superadmin/admins/by-register` | Lookup / add by register number |
| PUT | `/superadmin/admins/{id}/grant` | Grant admin privileges |
| DELETE | `/superadmin/admins/{id}` | Revoke admin |
| GET/POST | `/superadmin/invites` | List / create invites (allow-list) |
| DELETE | `/superadmin/invites/{email}` | Delete invite |

Onboarding & MFA endpoints (`/api/v1/auth/superadmin/*`) are covered in
[Super Admin MFA](../authentication/super-admin-mfa.md).

## Site & themes

| Method | Path | Auth | Purpose |
|--------|------|------|---------|
| GET | `/site/appearance` | public | Landing/anon appearance |
| PUT | `/site/appearance` | `settings.manage` | Update appearance |
| GET | `/themes/active` | public | Active landing theme |
| GET/POST | `/themes` | `settings.manage` | List / create theme |
| GET/PUT/DELETE | `/themes/{id}` | `settings.manage` | Get / update / delete |
| PUT | `/themes/{id}/active` | `settings.manage` | Activate theme |

---

## Spec vs. code discrepancies

The generated `swagger.json` has known gaps — the route files are authoritative:

- **Missing from the spec:** the entire RBAC and Provider groups, and the
  admin-users endpoints.
- **Stale in the spec:** password/OTP/registration auth routes
  (`/auth/login`, `/auth/register`, `/auth/send-otp`, `/auth/password/*`) — these
  were **removed** from the code.
- Some spec paths carry a doubled `/v1/` prefix; the real paths are a single
  `/v1` under `/api`. The spec `host` is still `localhost:8080`.
