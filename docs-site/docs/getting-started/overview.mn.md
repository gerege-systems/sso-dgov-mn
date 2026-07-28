# Ерөнхий тойм

DAN-Government SSO нь төрийн үйлчилгээнд зориулсан бүрэн стектэй identity
платформ бөгөөд [sso.dgov.mn](https://sso.dgov.mn) хаяг дээр байрлуулсан. Энэ нь
дахин ашиглах боломжтой **Government Template Platform V3.0** дээр бүтээгдэж,
DAN нэрээр брэндлэгдэн өргөтгөгдсөн.

## Дизайны зарчмууд

- **eID-first.** Хэрэглэгчийн цорын ганц интерактив нэвтрэлт нь *eID-ээр нэвтрэх*
  юм. Нууц үг, имэйл, эсвэл OTP нэвтрэлтийн зам байхгүй — хуучин usecase-ууд
  байгаа ч хүрэх боломжгүй. Google нь бие даасан нэвтрэлт биш, харин бүртгэлийн
  **холбоос** юм.
- **Clean Architecture.** Хамаарал зөвхөн дотогшоо чиглэнэ:
  `handler → usecase → repository → domain`. Бизнесийн цөм нь вэб framework буюу
  өгөгдлийн сангийн driver-ыг хэзээ ч import хийхгүй тул алийг нь ч солих боломжтой.
- **ORM ашиглаагүй.** Бүх SQL нь pgx-ээр гараар бичигдэж, параметржүүлсэн байдаг;
  бичлэгүүд нь нэрээр scan хийгддэг энгийн struct-ууд юм. [Backend
  Layers](../architecture/backend-layers.md)-ийг үз.
- **Гүнзгий хамгаалалт (Defense in depth).** Postgres Row-Level Security нь
  репозиторийн аль хэдийн бичсэн `WHERE user_id = …` нөхцлийн *доор* байрлах тул
  query-ийн алдаа гарлаа ч өөр хэрэглэгчийн мөрийг задруулж чадахгүй.
  [Security](../security/index.md)-ийг үз.
- **BFF frontend.** Хөтөч нь зөвхөн ижил-эх (same-origin) Next.js route-уудтай
  харьцдаг; JWT-ууд нь httpOnly cookie дотор амьдардаг бөгөөд client JavaScript
  руу хэзээ ч хүрдэггүй.

## Monorepo бүтэц

```
sso-dgov-mn/
├── backend/           # Go · chi (net/http) · pgx · PostgreSQL · Redis
│   ├── cmd/           # api (server + DI), migration, seed CLIs
│   ├── internal/      # business (domain + usecases), datasources, http
│   ├── pkg/           # framework-agnostic clients (eid, google, hydra, gemini…)
│   ├── migrations/    # numbered SQL (consolidated 1_init_schema)
│   └── docs/          # ARCHITECTURE · DEVELOPMENT · API_CONTRACT · SECURITY (EN/MN)
├── frontend/          # Next.js 15 BFF (server-side proxy; cookie sessions)
├── deploy/            # nginx, initdb, deploy scripts
└── docker-compose.yml # db · redis · migrate · api · web · hydra
```

## Функцийн зураглал

| Хэсэг | Товч тайлбар | Баримт бичиг |
|------|---------|------|
| Authentication | eID нэвтрэлт (QR / deep-link / иргэний бүртгэлийн push), Google холболт, JWT session | [Authentication](../authentication/index.md) |
| SSO provider | Ory Hydra-аар дамжуулан DAN-ыг OIDC provider болгох ("Sign in with DAN") | [Sign in with DAN](../sso/index.md) |
| eID PKI profile | Гэрчилгээ, төхөөрөмж, идэвх, холбогдсон байгууллага болон гарын үсэг зурагчид | [eID Login](../authentication/eid-login.md) |
| Байгууллагууд | Байгууллага үүсгэх/хайх (улсын бүртгэлийн лавлагаа), гишүүд ба роль, RLS-ээр хамрагдсан | [API Reference](../api/index.md) |
| Төрийн үйлчилгээний портал | Өргөдөл, лавлагаа, мэдэгдэл, төлбөр, цаг захиалга | [API Reference](../api/index.md) |
| API gateway | Хүсэлтийн telemetry бүхий, админаар удирддаг service/route/consumer | [API Reference](../api/index.md) |
| Баримт бичгийн гарын үсэг | eID `/v3`-ээр сервер талын PAdES; гуравдагч талын sign-relay | [Document Signing](../signing/index.md) |
| AI туслах | Давхаргалсан guardrail болон мэдлэгийн сан бүхий Gemini chat/voice/орчуулга | [AI Pipeline](../ai/index.md) |
| RBAC ба super admin | Динамик роль + эрхийн каталог; MFA-аар хамгаалагдсан super admin-ууд | [Super Admin MFA](../authentication/super-admin-mfa.md) |
| Audit | Hash-chain хэлбэрийн, зөвхөн нэмдэг (append-only) audit trail, бүрэн бүтэн байдлын баталгаажуулалттай | [Security](../security/index.md) |

## Дараагийн алхмууд

- Локал орчинд ажиллуулах → [Quick Start](quick-start.md)
- Ажиллах орчнуудыг ойлгох → [Environments](environments.md)
