# DAN-Government SSO

<div class="dan-hero" markdown>
<span class="dan-badge">eID based · AI enabled</span>

# Government Single Sign-On for Mongolia

The technical documentation for **DAN-Government SSO** (`sso.dgov.mn`) — a
production-hardened identity platform built on the **Government Template
Platform V3.0** stack: a Clean-Architecture Go backend, a Next.js Backend-for-Frontend, and a Gemini AI pipeline.
</div>

DAN is an **eID-based** Single Sign-On for government services. The only
interactive login is **Login with eID** (the platform is a Relying Party of eID
Mongolia); on top of that sit Google account-linking, DAN's own OIDC provider
("Sign in with DAN"), PAdES document signing, and an AI assistant.

## What this documentation covers

<div class="grid cards" markdown>

-   :material-rocket-launch: **[Getting Started](getting-started/overview.md)**

    Platform overview, the monorepo, and how to bring up the full stack locally.

-   :material-layers-triple: **[Architecture](architecture/index.md)**

    Clean Architecture layers, the pgx data layer, Postgres Row-Level Security,
    and the Next.js BFF model.

-   :material-fingerprint: **[Authentication](authentication/index.md)**

    The eID login flow end-to-end, JWT sessions with rotation, and the
    MFA-gated super-admin onboarding wizard.

-   :material-login-variant: **[Sign in with DAN](sso/index.md)**

    DAN as an OIDC provider via Ory Hydra — login, consent, and logout for
    relying parties.

-   :material-api: **[API Reference](api/index.md)**

    Every REST endpoint by domain, plus an interactive OpenAPI explorer.

-   :material-robot: **[AI Pipeline](ai/index.md)**

    The SDK-free Gemini client, layered guardrail prompt, function-calling
    tools, and voice/translation.

-   :material-file-sign: **[Document Signing](signing/index.md)**

    Server-side PAdES signing through eID Mongolia `/v3` and the third-party
    sign-relay.

-   :material-shield-lock: **[Security](security/index.md)**

    Security headers, CORS, rate limits, RLS, the hash-chained audit log, and
    production hardening guards.

-   :material-server-network: **[Operations](operations/index.md)**

    The docker-compose stack, nginx + TLS, migrations, CI/CD, and observability.

</div>

## Platform at a glance

| Layer | Technology |
|-------|------------|
| Backend | Go 1.26 · chi (net/http) · pgx v5 (pgxpool) · **no ORM** |
| Datastores | PostgreSQL 16 (Row-Level Security) · Redis 7 |
| Frontend | Next.js 15 (App Router) · Backend-for-Frontend · TanStack Query |
| Identity | eID Mongolia (RP · ACSP_V2) · Google OAuth · Ory Hydra (OIDC provider) |
| AI | Google Gemini (SDK-free REST) — chat, STT, TTS, live translation |
| Signing | PAdES via eID Mongolia `/v3` + `digitorus/pdfsign` fallback |
| Observability | OpenTelemetry · Prometheus · Zap |

!!! note "Source of truth"
    This documentation is generated from a code-level review of the private
    `gerege-systems/sso-dgov-mn` repository. Where a claim references code it
    cites the file path; where the shipped API spec and the code diverge, the
    code is authoritative and the gap is flagged.

## License

The platform is **MIT-licensed**, derived from the open-source
[snykk/go-rest-boilerplate](https://github.com/snykk/go-rest-boilerplate)
(HTTP layer ported Gin → chi, data layer sqlx → pgx). Co-developed by the
**Gerege Systems Development Team** and **Claude AI**, 2026.
