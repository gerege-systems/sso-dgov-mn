# Тохиргоо

Тохиргоог Viper (`backend/internal/config/config.go`) уншдаг. Нууц мэдээлэл нь хоёр
**gitignore хийсэн** env файлд байдаг; зөвхөн `*.env.example` commit хийгддэг.

- **`./.env`** — зөвхөн compose interpolation (`POSTGRES_*`, `APP_DB_*`,
  `REDIS_PASS`, `APP_ORIGIN`, `WEB_PORT`, `API_RELAY_PORT`, `HYDRA_*` port/нууцууд,
  web талын OAuth client ID-ууд). Hydra нууцууд `${VAR:?}` ашигладаг тул тохируулаагүй
  бол compose эхлэхээс татгалздаг.
- **`./backend.env`** — `api` / `migrate` руу `/app/.env`-д read-only байдлаар
  mount хийгддэг.

!!! danger "Нууцыг хэзээ ч бүү commit хий"
    `.env`, `backend.env`, болон `backend/internal/config/.env*` нь gitignore
    хийгдсэн. Шинэ хувьсагчдыг repo-ийн commit хийсэн config-д биш, README-д
    баримтжуул. Production-д flat `.env` файлаас илүү бодит secret store / KMS-ийг
    сонго.

## Гол backend хувьсагчид

| Хувьсагч | Зорилго |
|----------|---------|
| `ENVIRONMENT` | `development` \| `production` — HSTS, sslmode, RLS guard, observability gate-ийг хянана |
| `JWT_SECRET` | HMAC гарын үсгийн түлхүүр (≥ 32 тэмдэгт) |
| `JWT_EXPIRED` / `JWT_REFRESH_EXPIRED` | access TTL (цаг) / refresh TTL (өдөр, өгөгдмөл 7) |
| `DB_POSTGRE_DSN` / `DB_POSTGRE_URL` | dev DSN / prod URL (prod-д `sslmode=verify-full` шаардлагатай) |
| `DB_MAX_OPEN_CONNS` / `DB_MAX_IDLE_CONNS` / `DB_CONN_MAX_LIFE_MINS` | pool хэмжээ (25 / 5 / 15) |
| `REDIS_*` | cache / session-цуцлалтын store |
| `TRUSTED_PROXIES` | `X-Forwarded-For`-т нь итгэдэг CIDR-ууд |
| `EID_RP_UUID` / `EID_RP_NAME` / `EID_RP_SECRET` / `EID_BASE_URL` | eID Relying Party identity |
| `SIGN_SIGNER_CERT_FILE` / `SIGN_SIGNER_KEY_FILE` | байнгын Document-Signer PEM (prod-д шаардлагатай) |
| `SIGN_RELAY_TOKEN` | гуравдагч этгээдийн sign-relay-г идэвхжүүлнэ (`EID_RP_SECRET`-тэй хамт) |
| `INTEGRATION_ENC_KEY` | integration token + super-admin TOTP-ийн AES-256-GCM түлхүүр; super-admin onboarding-ийг идэвхжүүлнэ |
| `HYDRA_ADMIN_URL` / `HYDRA_PUBLIC_URL` / `SSO_STATE_KEY` (≥32) | OIDC provider-ийг идэвхжүүлнэ |
| `SSO_FIRSTPARTY_CLIENTS` | consent UI-г алгасдаг client ID-ууд |
| `GEMINI_API_KEY` / `GEMINI_MODEL` / `GEMINI_TTS_MODEL` / `GEMINI_VOICE` | AI pipeline |
| `AI_SCOPE_PROMPT` | AI guardrail давхаргын нөөц scope |
| `OBSERVABILITY_TOKEN` | prod-д `/metrics` + `/swagger/doc.json`-ийн bearer |
| `OTEL_EXPORTER_OTLP_ENDPOINT` | OTLP tracing эндпойнт (prod) |

## Гол frontend хувьсагчид

| Хувьсагч | Зорилго |
|----------|---------|
| `BACKEND_URL` | BFF-ийн серверийн талын суурь (`+ /api/v1`); compose-д `http://api:8080` |
| `APP_ORIGIN` | CSRF origin шалгалтад хүлээгдэж буй Origin |
| `COOKIE_SECURE` | production-д `true` (зөвхөн HTTPS cookie) |
| `NEXT_PUBLIC_DOCS_URL` | "Docs" холбоосын зорилтыг дарж бичнэ (өгөгдмөлөөр энэ баримт бичгийн сайт) |

!!! note "Feature flag нь config-ийн оршин байх эсэх"
    Хэд хэдэн дэд систем зөвхөн тохируулагдсан үед mount хийгддэг: OIDC provider-т
    `HYDRA_*` + `SSO_STATE_KEY` гурвал хэрэгтэй; super-admin onboarding-т
    `INTEGRATION_ENC_KEY` хэрэгтэй; sign-relay-д `SIGN_RELAY_TOKEN` + `EID_RP_SECRET`
    хэрэгтэй.
