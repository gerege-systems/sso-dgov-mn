# DAN-Government SSO

<div class="dan-hero" markdown>
<span class="dan-badge">eID based · AI enabled</span>

# Монгол Улсын төрийн Single Sign-On

**DAN-Government SSO**-ийн (`sso.dgov.mn`) техникийн баримт бичиг — **Government
Template Platform V3.0** стек дээр бүтээгдсэн, продакшнд бэлэн болтол
бэхжүүлсэн identity платформ: Clean Architecture зарчмаар зохион байгуулсан Go
backend, Next.js Backend-for-Frontend болон Gemini AI pipeline.
</div>

DAN нь төрийн үйлчилгээнд зориулсан **eID-д суурилсан** Single Sign-On юм.
Хэрэглэгчийн цорын ганц интерактив нэвтрэлт нь **eID-ээр нэвтрэх** (энэ платформ
нь eID Mongolia-ийн Relying Party) бөгөөд түүн дээр Google холболт, DAN-ы өөрийн
OIDC provider ("Sign in with DAN"), PAdES баримт бичгийн гарын үсэг болон AI
туслах нэмэгдэн ажиллана.

## Энэ баримт бичиг юуг хамрах вэ

<div class="grid cards" markdown>

-   :material-rocket-launch: **[Getting Started](getting-started/overview.md)**

    Платформын ерөнхий тойм, monorepo бүтэц болон бүх стекийг локал орчинд хэрхэн ажиллуулах.

-   :material-layers-triple: **[Architecture](architecture/index.md)**

    Clean Architecture давхаргууд, pgx өгөгдлийн давхарга, Postgres Row-Level Security
    болон Next.js BFF загвар.

-   :material-fingerprint: **[Authentication](authentication/index.md)**

    eID нэвтрэлтийн бүрэн урсгал, rotation бүхий JWT session болон
    MFA-аар хамгаалагдсан super-admin онбординг wizard.

-   :material-login-variant: **[Sign in with DAN](sso/index.md)**

    Ory Hydra-аар дамжуулан DAN-ыг OIDC provider болгох нь — relying party-уудын
    login, consent болон logout.

-   :material-api: **[API Reference](api/index.md)**

    Домэйн бүрээр ангилсан бүх REST endpoint, түүнчлэн интерактив OpenAPI explorer.

-   :material-robot: **[AI Pipeline](ai/index.md)**

    SDK-гүй Gemini client, давхаргалсан guardrail prompt, function-calling
    tool-ууд болон дуу/орчуулга.

-   :material-file-sign: **[Document Signing](signing/index.md)**

    eID Mongolia `/v3` болон гуравдагч талын sign-relay-ээр дамжуулсан сервер талын
    PAdES гарын үсэг.

-   :material-shield-lock: **[Security](security/index.md)**

    Security header-үүд, CORS, rate limit, RLS, hash-chain хэлбэрийн audit log болон
    продакшны бэхжүүлэлтийн хамгаалалтууд.

-   :material-server-network: **[Operations](operations/index.md)**

    docker-compose стек, nginx + TLS, migration, CI/CD болон observability.

</div>

## Платформ товчхондоо

| Layer | Technology |
|-------|------------|
| Backend | Go 1.26 · chi (net/http) · pgx v5 (pgxpool) · **ORM ашиглаагүй** |
| Datastore-ууд | PostgreSQL 16 (Row-Level Security) · Redis 7 |
| Frontend | Next.js 15 (App Router) · Backend-for-Frontend · TanStack Query |
| Identity | eID Mongolia (RP · ACSP_V2) · Google OAuth · Ory Hydra (OIDC provider) |
| AI | Google Gemini (SDK-гүй REST) — chat, STT, TTS, шууд орчуулга |
| Signing | eID Mongolia `/v3`-ээр PAdES + `digitorus/pdfsign` fallback |
| Observability | OpenTelemetry · Prometheus · Zap |

!!! note "Эх сурвалжийн үнэн"
    Энэ баримт бичиг нь хувийн `gerege-systems/sso-dgov-mn` репозиторийн
    код түвшний хяналт шалгалтаас үүсгэгдсэн. Аливаа мэдэгдэл кодыг иш татаж
    буй бол файлын замыг зааж өгсөн; хэрэв нийлүүлэгдсэн API spec болон код
    зөрчилдвөл кодыг эрх бүхий эх сурвалж гэж үзэж, зөрүүг тэмдэглэсэн болно.

## License

Энэ платформ нь **MIT лицензтэй** бөгөөд нээлттэй эх
[snykk/go-rest-boilerplate](https://github.com/snykk/go-rest-boilerplate)-оос
гаралтай (HTTP давхаргыг Gin → chi, өгөгдлийн давхаргыг sqlx → pgx руу шилжүүлсэн).
**Gerege Systems Development Team** болон **Claude AI** хамтран хөгжүүлэв, 2026.
