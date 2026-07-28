# Release Notes

This documentation tracks the `main` branch of the platform. Notable recent
changes (most recent first):

- **Rebrand to `sso.dgov.mn`** — the SSO *consumer* surface was removed and the
  platform focused on being an identity **provider**; branding moved from
  "DAN-Government SSO / dan-dgov-mn" toward `sso.dgov.mn`.
- **Migration consolidation** — the former per-feature migrations (1–37) were
  squashed into a single `1_init_schema` pair, with the originals archived under
  `migrations/old/`. The consolidated init sets `search_path=public`.
- **eID activity grouped by type** in the UI; **super admins see only their own
  system**.
- **Super Admin system** promoted to its own top-level area; the landing-themes
  menu moved under it.
- **Super-admin satellite table + MFA onboarding** — super admins are stored in a
  dedicated `superadmin_accounts` table and onboarded through the invite →
  Google → eID → email OTP → TOTP wizard.

!!! note
    For the forward-looking roadmap (what's shipped vs. planned), see the
    repository's `ROADMAP.md`. This page summarizes shipped changes visible in the
    codebase; it is not a substitute for the git history.

## Documentation gaps flagged during review

The code review that produced this site surfaced a few places where the shipped
docs/spec lag the code. They are noted inline where relevant, and summarized here:

- The generated `swagger.json` is missing the RBAC, Provider, and admin-users
  groups, and still lists removed password/OTP/registration auth routes.
- `backend/docs/` references individual migration files by number; those are now
  consolidated into `1_init_schema`.
- Some doc comments (e.g. a middleware referencing "GORM") predate the pgx / no-ORM
  design and are cosmetic-only stale.
- No dedicated `SIGNING.md` doc pair exists yet for the PAdES signing subsystem —
  it is documented here in [Document Signing](signing/index.md).
