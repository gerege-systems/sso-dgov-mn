# eID-ээр нэвтрэх

## Relying Party загвар

Платформ нь eID Mongolia цахим танилтын үйлчилгээ үзүүлэгчийн
(`https://eidmongolia.mn/v3`) итгэмжлэгдсэн тал (Relying Party буюу RP) юм. RP
клиент нь `backend/pkg/eid/eid.go`. Энэ нь Smart-ID-тэй нийцтэй v3 API-тай
(ACSP_V2) харилцдаг бөгөөд `Authorization: Bearer <rp_sk_…>` толгой болон
биед `relyingPartyUUID` / `relyingPartyName`-ийг ашиглан баталгаажуулдаг.

Хүсэх сертификатын түвшний доод хязгаар нь `ADVANCED` — үүнийг зориудаар хамгийн
багаар нь тавьсан, учир нь `QUALIFIED` болгож өсгөвөл нэвтрэх сертификат нь зөвхөн
ADVANCED түвшинтэй иргэдийг хаана.

Дамжуулалтын протокол:

| Дуудлага | Зорилго |
|------|---------|
| `POST {base}/authentication/device-link/anonymous` | QR / deep-link нэвтрэлт |
| `POST {base}/authentication/notification/etsi/PNOMN-{civil}` | иргэний үнэмлэхийн дугаараар push нэвтрэлт |
| `GET {base}/session/{sessionID}?timeoutMs=25000` | long-poll төлөв |

## Эхлүүлэх гурван арга

| Endpoint | Usecase | Горим |
|----------|---------|------|
| `POST /api/v1/auth/eid/start` | `EIDStart` | QR / мобайл deep-link |
| `POST /api/v1/auth/eid/start-id` | `EIDStartByNationalID` | иргэний үнэмлэхийн дугаараар push |
| `POST /api/v1/auth/eid/poll` | `EIDPoll` | дуустал нь long-poll хийх |

**QR ба deep-link.** `EIDStart` нь `eid.QRInitiate`-ийг дуудна. Хүсэлт нь
сонголтоор `callbackUrl` агуулж болно:

- хоосон → **төхөөрөмж хооронд** (десктоп QR: браузер poll хийнэ);
- байгаа (`<origin>/auth/eid/callback`) → **нэг төхөөрөмж дээр** (мобайл App2App:
  утас зөвшөөрсний дараа браузер тухайн URL руу буцна).

Хариу нь `session_id`, `device_link_url`, `verification_code`, `expires_at`-ийг
буцаана.

**Иргэний үнэмлэхийн дугаараар push.** `EIDStartByNationalID` нь
`eid.Initiate(nationalID, …)`-ийг дуудна — eID нь уг иргэний үнэмлэхийн
дугаарт бүртгэлтэй төхөөрөмж(үүд) рүү зөвшөөрөл хүссэн мэдэгдэл шууд илгээдэг тул
QR шаардлагагүй. Иргэний үнэмлэхийн дугаарыг хэзээ ч логлодоггүй (зөвхөн
`has_national_id` boolean хадгална). Хүсэлт тус бүрийн 16 байтын криптограф
санамсаргүй **nonce** нь IdP-ийн replay халдлагаас хамгаалдаг.

## Long-poll session-ийн амьдралын мөчлөг

Клиент `session_id`-г `/auth/eid/poll` руу post хийнэ. Usecase нь
`uc.eid.Session(ctx, sessionID, 25000)`-г дуудна — энэ нь **25 секундын** барилт
бөгөөд eID клиентийн 30 секундын HTTP timeout-оос доогуур байлгасан. Платформын
дөрвөн төлөв нь `RUNNING`, `COMPLETE`, `EXPIRED`, `REFUSED`.

`COMPLETE` үед:

1. **Субьектийг гаргаж авах.** Нээлттэй RP нь `national_id`-г хүлээж авдаггүй тул
   `civil_id` нь тогтвортой түлхүүр болно (эрх бүхий RP-ийн хувьд `national_id`
   руу шилжинэ). Хоёулаа хоосон бол → татгалзана.
2. `domain.NewEIDUser(...)`-ээр **хэрэглэгчийг бүрдүүлэх** ба нэвтрэх
   сертификатаас уншсан PKI мэдээллийг хавсаргана: `DocumentNumber`,
   `CertSerial`, `CertNotBefore`, `CertNotAfter`, `CertIssuer`, `CertKeyType`.
3. `users.UpsertFromEID`-ээр **upsert** хийх — Postgres upsert нь хэсэгчилсэн
   давхардалгүй индекс дээр түлхүүрлэдэг: `ON CONFLICT (lower(civil_id)) WHERE
   civil_id IS NOT NULL`; хэрэглэгчийн нэр нь `eid_<civil_id>`.
4. `google_link_token` дамжуулагдсан бол **хүлээгдэж буй Google акаунтыг холбох**
   (амин чухал бус).
5. **MFA хаалт** — хэрэглэгч супер админ бол session олгодоггүй; супер-админ MFA
   сорилт эхэлж, `{ MFARequired: true, MFAToken }` буцаана. [Super Admin
   MFA](super-admin-mfa.md)-г үзнэ үү.
6. Бусад тохиолдолд access + refresh хосыг олгож, `{ User, AccessToken,
   RefreshToken }`-г буцаана.

```mermaid
sequenceDiagram
    participant B as Browser (BFF)
    participant API as DAN backend
    participant IdP as eID Mongolia
    participant R as Redis
    participant DB as Postgres

    B->>API: POST /auth/eid/start {callbackUrl?}
    API->>IdP: QRInitiate(displayText, callbackUrl, nonce)
    IdP-->>API: session_id, device_link_url, verification_code
    API-->>B: show QR / deep-link

    loop every ~2.5s until terminal
        B->>API: POST /auth/eid/poll {session_id}
        API->>IdP: Session(session_id, timeoutMs=25000)
        IdP-->>API: RUNNING | COMPLETE | EXPIRED | REFUSED
    end

    Note over API,IdP: On COMPLETE
    API->>API: subject = civil_id; read cert/PKI details
    API->>DB: UpsertFromEID (ON CONFLICT lower(civil_id))
    alt super admin
        API->>R: SET superadmin_mfa:<token> (5m)
        API-->>B: {COMPLETE, MFARequired:true, MFAToken}
    else regular user
        API->>API: GenerateTokenPair(...)
        API->>R: SET refresh:<jti>
        API-->>B: {COMPLETE, User, AccessToken, RefreshToken}
    end
```

!!! note "Итгэмжлэгдсэн дамжуулалтын загвар"
    IdP нь TLS-ээр хамгаалагдсан эрх бүхий эх сурвалж тул платформ одоогоор
    `COMPLETE` хариуд итгэдэг. ACSP_V2 гарын үсгийг сертификаттай тулган
    баталгаажуулахыг код дотор ирээдүйд хийж болох сонголттой бэхжүүлэлт гэж
    тэмдэглэсэн.

## Буцаан уншсан цахим танилт / PKI өгөгдөл

Нэвтрэлтээс гадна eID usecase нь нэвтэрсэн иргэний баялаг PKI өгөгдлийг
`/api/v1/users/me/eid` дор гаргадаг (бүгд иргэний ETSI id-г `PNOMN-<CIVIL_ID>`
хэлбэрээр шийддэг):

| Endpoint | Өгөгдөл |
|----------|------|
| `GET …/eid/summary` | eID акаунтын хураангуй |
| `GET …/eid/certificates` | сертификатууд |
| `GET …/eid/devices` | бүртгэлтэй төхөөрөмжүүд |
| `GET …/eid/activity` | RP-д хамаарах нэвтрэлт / гарын үсгийн түүх |
| `GET/POST/DELETE …/eid/organizations` | иргэний төлөөлж буй холбоотой байгууллагууд |
| `…/eid/organizations/{regNo}/signers` | байгууллагын гарын үсэг зурагчдыг жагсаах / нэмэх / дахин илгээх / хасах |

Байгууллага бүртгэх үед ХУР системээр гүйцэтгэх захирал/үүсгэн байгуулагч/эзэмшигчдийн
регистрийн дугаарыг хайдаг. Нэмэгдсэн гарын үсэг зурагчид eID-ийн гарын үсгийн
push хүлээн авах ба өөрийн PIN кодоор зөвшөөрөх хүртэл `PENDING` төлөвт үлдэнэ.
`PKI_READ` эрх байхгүй бол `Forbidden` руу буулгана.

## Google акаунт холболт { #google-account-linking }

Google нь **бие даасан** нэвтрэлт биш; энэ нь eID-ээр баталгаажсан хүнд холбогдох
(эсвэл түүгээр нэвтрэх) юм. Usecase: `auth_google.go`.

- **`POST /auth/google`** нь OAuth `code`-г солилцоод, дараа нь `GetByGoogleSub`-г
  хайна:
    - **аль хэдийн холбогдсон** → хадгалагдсан Google профайлыг шинэчилж, (хэрэв
      хэрэглэгч супер админ биш бол) session олгож, `{ Linked: true, Login }`
      буцаана.
    - **анх удаа** → богино хугацаат `LinkToken` (15 минутын TTL) үүсгэж, бүрэн
      Google профайлыг Redis дотор `google_link:<token>` дор хадгалж, `{ Linked:
      false, LinkToken, Email }` буцаана. Дараа нь клиент eID нэвтрэлт хийж, уг
      token-г `google_link_token` болгон дамжуулна.
- **eID дуусах үед холбох** (`linkGoogleIfPending`) нь бүрэн амин чухал бус —
  холболт амжилтгүй болсон ч (жишээ нь тухайн Google акаунт өөр хүнд аль хэдийн
  холбогдсон) eID нэвтрэлт үргэлж амжилттай болно.
- **`DELETE /auth/google/link`** нь холболтыг салгана; энэ нь `authMiddleware`-ийн
  ард байдаг цорын ганц Google endpoint. Дахин холбох нь зөвхөн нэвтрэлтийн
  урсгалаар л боломжтой.
