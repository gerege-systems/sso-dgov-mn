# Authentication

DAN-Government SSO is an eID-based identity platform. The **only** interactive
login method is **Login with eID** (the platform is a Relying Party of eID
Mongolia); on top of that sit **Google OAuth account-linking**,
**["Sign in with DAN"](../sso/index.md)** (the platform acting as its own OIDC
provider via Ory Hydra), and a hardened
**[super-admin MFA onboarding](super-admin-mfa.md)** wizard.

There is no password, email, or OTP login path — the legacy usecases exist but
are unreachable.

## The group

All auth endpoints live under `/api/v1/auth` (`route_auth.go`). The whole group
runs behind two middlewares:

- a **4 KiB** body cap (`AuthBodyMaxBytes`), and
- `ServiceRLSContext()` — pre-login flows touch unauthenticated users' rows (eID
  upsert, refresh identity lookup), so they run under a `service` RLS identity.

Start/lifecycle endpoints get a strict **~5 req/min** per-IP rate limiter;
`/auth/eid/poll` gets a **separate, looser** limiter (~60/min, burst 30) so a
~2.5 s long-poll loop never 429s.

## Read the details

- **[eID Login](eid-login.md)** — the Relying Party model, the three start modes,
  the long-poll lifecycle, and the PKI data read back.
- **[Sessions & JWT](sessions.md)** — access + refresh tokens, single-use
  rotation, and logout revocation.
- **[Super Admin MFA](super-admin-mfa.md)** — the invite-gated onboarding wizard
  and the mandatory second factor.

## The four auth surfaces

| Surface | Role | Where |
|---------|------|-------|
| eID login | Sole interactive login (RP of eID Mongolia) | [eID Login](eid-login.md) |
| Google | Account **link** to an eID-verified person | [eID Login](eid-login.md#google-account-linking) |
| Sign in with DAN | DAN as an OIDC **provider** (Hydra) | [Sign in with DAN](../sso/index.md) |
| Super-admin MFA | Invite → Google → eID → email OTP → TOTP | [Super Admin MFA](super-admin-mfa.md) |
