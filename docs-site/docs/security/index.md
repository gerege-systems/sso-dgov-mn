# Security

This section documents the **enforced** security controls, condensed from
`backend/docs/SECURITY.md`. The platform maps its controls to OWASP ASVS / API
Top 10, NIST SP 800-63B / 800-218, and CIS Controls.

## HTTP edge controls

### Security response headers

Set on **every** response by `SecurityHeadersMiddleware`:

| Header | Value |
|--------|-------|
| `X-Content-Type-Options` | `nosniff` |
| `X-Frame-Options` | `DENY` |
| `Referrer-Policy` | `strict-origin-when-cross-origin` |
| `Content-Security-Policy` | `default-src 'none'; frame-ancestors 'none'` (API is JSON-only) |
| `Permissions-Policy` | deny-all for camera/geolocation/microphone/payment/usb/… |
| `Cross-Origin-Opener-Policy` | `same-origin` |
| `Cross-Origin-Resource-Policy` | `same-site` |
| `Cross-Origin-Embedder-Policy` | `require-corp` |
| `Strict-Transport-Security` | `max-age=31536000; includeSubDomains` — **production only** |

HSTS is withheld in development so a localhost dev server doesn't pin
plain-HTTP refusal for a year.

### CORS allow-list

`CORSMiddleware` reflects an Origin only if it is on the exact allow-list.
**Credentials and wildcard are mutually exclusive**:
`Access-Control-Allow-Credentials: true` is sent only in allow-list mode, never
with `*`. Production rejects an empty allow-list at config validation.

### Rate limiting

Per-IP token buckets:

| Limiter | Rate | Burst |
|---------|------|-------|
| `/auth` | ~5/min | 5 |
| `/ai` | ~20/min | 10 |
| long-poll | ~1/s | 30 |
| gov-write | ~30/min | 15 |

On exhaustion → `429` with `Retry-After` and `X-RateLimit-*` headers. The client
IP is resolved trusting `X-Forwarded-For` **only** when the peer is in
`TRUSTED_PROXIES`, so a spoofed XFF cannot poison rate-limit or audit
attribution.

### Body-size limits & server timeouts

- Global ceiling 26 MiB (for the multipart PDF-sign route); normal JSON 1 MiB;
  auth routes **4 KiB**. Chi parent middleware can only *tighten* child limits, so
  the 4 KiB auth cap composes over the global net.
- `http.Server`: `ReadHeaderTimeout` 10 s, `ReadTimeout` 30 s, `WriteTimeout`
  60 s, `IdleTimeout` 120 s, `MaxHeaderBytes` 16 KiB. A per-request 30 s context
  deadline propagates into pgx queries so stuck queries are cancelled.

## Postgres Row-Level Security

- **A non-superuser DB role is mandatory** — RLS (even `FORCE`d) is silently
  bypassed by superuser / `BYPASSRLS` roles. The app connects as a `NOSUPERUSER
  NOBYPASSRLS` role; the `migrate` container keeps the superuser for DDL.
- **Boot-time enforceability guard** — `guardRLSEnforceable` queries `pg_roles`
  at startup; if the app's role is superuser/`BYPASSRLS` it **fails boot in
  production** (fail-closed) and warns in development.
- `ENABLE` + `FORCE` RLS on every per-user table; `withRLS` transactions publish
  identity via `SET LOCAL app.user_id / app.user_role`. **No identity ⇒ zero rows.**

See [Data Model & RLS](../architecture/data-model.md) for the full policy table.

## Hash-chained append-only audit log

- **Design** (`pkg/audit/chain.go`): each row stores `chain_hash = SHA-256(
  prevHash ‖ canonical-json(entry) )`. `canonicalJSON` is deterministic
  (`occurred_at` as UTC unix-nanos, sorted `metadata` keys, fixed field order) —
  **adding a field is a chain-breaking change**.
- **Serialized writers** — a `pg_advisory_xact_lock` before each append so
  concurrent appends can't fork the chain.
- **Integrity verification** — `VerifyChain` recomputes from genesis under an
  admin GUC, exposed as `GET /api/v1/audit/verify` (admin-only, alongside
  `GET /api/v1/audit/`).

## Secrets handling

- Gitignored env files (`.env`, `backend.env`, `internal/config/.env*`); only
  `*.env.example` is kept.
- CI runs **gitleaks** (`detect --no-git --redact`) on the working tree.
- **AES-256-GCM at rest** for integration tokens and super-admin TOTP secrets —
  key derived via `SHA-256(INTEGRATION_ENC_KEY)`, fresh random nonce per encrypt,
  stored as `base64(nonce ‖ ciphertext)`.

## Production hardening guards

- **`sslmode=verify-full` required** — in production, `DB_POSTGRE_URL` is
  rejected unless its sslmode is `verify-full` (or `verify-ca` internal).
- **`/metrics` + `/swagger/doc.json` bearer-gated** — `ObservabilityGate` compares
  `Authorization: Bearer <OBSERVABILITY_TOKEN>` with `crypto/subtle`; **any miss
  returns 404** (not 401) to hide the endpoint from recon. `/health` + `/ready`
  stay public.
- **Compose runs `ENVIRONMENT=development`** on purpose — the internal DB has no
  TLS, so the HSTS / sslmode / observability-gate production paths engage only
  behind nginx in production.

## RBAC & super-admin MFA (security angle)

Dynamic roles **SuperAdmin / Admin / Manager / User** with a permission
catalogue; route gates `RequirePermission` / `RequireAdmin` / `RequireSuperAdmin`
(a plain admin cannot pass the super-admin gate). Super-admin login is
structurally MFA-gated — a session is issued only after TOTP/recovery
verification, so any super-admin session is inherently MFA-verified. See
[Super Admin MFA](../authentication/super-admin-mfa.md).

## ASVS roadmap

| Phase | Status |
|-------|--------|
| **Phase 1 (ASVS L1)** | ✅ HTTPS + HSTS, eID-only login, parameterized queries, security headers, strict CORS, input validation, no committed secrets · ⏳ container scan / `govulncheck` in CI |
| **Phase 2 (ASVS L2)** | ✅ rate limiting, refresh rotation, phishing-resistant eID device-link, request timeout, encrypted integration tokens, hash-chained audit · ⏳ WAF, SIEM, backup-restore test, IR plan |
| **Phase 3 (ASVS L3)** | ◻ field-level PII encryption (KMS), mTLS, SLSA L3 provenance, external pentest |

## Reporting a vulnerability

Coordinated disclosure is handled per the repository-root `SECURITY.md`. Please
do not open public issues for security reports.
