# Нэвтрэлт (Authentication)

DAN-Government SSO нь eID-д суурилсан цахим танилтын платформ юм. **Цорын ганц**
интерактив нэвтрэх арга бол **eID-ээр нэвтрэх** (платформ нь eID Mongolia-ийн
итгэмжлэгдсэн тал буюу Relying Party) юм; үүн дээр нэмэлтээр **Google OAuth
акаунт холбох**, **["Sign in with DAN"](../sso/index.md)** (платформ нь Ory
Hydra-ээр дамжуулан өөрөө OIDC provider болж ажиллах), мөн бэхжүүлэлт хийсэн
**[супер-админ MFA бүртгэлийн](super-admin-mfa.md)** визард зэрэг бий.

Нууц үг, имэйл, эсвэл OTP-ээр нэвтрэх зам байхгүй — хуучин (legacy) usecase-ууд
хэдийгээр код дотор байгаа ч хүрч ажиллуулах боломжгүй.

## Бүлэг

Нэвтрэлтийн бүх endpoint нь `/api/v1/auth` дор байрладаг (`route_auth.go`).
Бүхэл бүлэг нь хоёр middleware-ийн ард ажилладаг:

- **4 KiB** биеийн хэмжээний хязгаар (`AuthBodyMaxBytes`), болон
- `ServiceRLSContext()` — нэвтрэхээс өмнөх урсгалууд нь нэвтрээгүй хэрэглэгчдийн
  мөрүүдэд хандах тул (eID upsert, refresh identity хайлт) тэдгээр нь `service`
  RLS цахим танилтаар ажилладаг.

Эхлүүлэх / амьдралын мөчлөгийн endpoint-ууд нь IP тус бүрд **~5 хүсэлт/минут**
хатуу хурдны хязгаартай; `/auth/eid/poll` нь **тусдаа, илүү сул** хязгаартай
(~60/минут, burst 30) тул ~2.5 секундын long-poll давталт хэзээ ч 429 алдаа өгөхгүй.

## Дэлгэрэнгүйг унших

- **[eID Login](eid-login.md)** — Relying Party загвар, эхлүүлэх гурван горим,
  long-poll амьдралын мөчлөг, болон буцаан уншсан PKI өгөгдөл.
- **[Sessions & JWT](sessions.md)** — access + refresh token, нэг удаагийн
  эргэлт (rotation), болон гарах үед token хүчингүй болгох.
- **[Super Admin MFA](super-admin-mfa.md)** — урилгаар хамгаалагдсан бүртгэлийн
  визард болон заавал шаардлагатай хоёр дахь баталгаажуулалт.

## Нэвтрэлтийн дөрвөн гадаргуу

| Гадаргуу | Үүрэг | Хаана |
|---------|------|-------|
| eID login | Цорын ганц интерактив нэвтрэлт (eID Mongolia-ийн RP) | [eID Login](eid-login.md) |
| Google | eID-ээр баталгаажсан хүнд акаунт **холбох** | [eID Login](eid-login.md#google-account-linking) |
| Sign in with DAN | DAN нь OIDC **provider** болох (Hydra) | [Sign in with DAN](../sso/index.md) |
| Super-admin MFA | Урилга → Google → eID → имэйл OTP → TOTP | [Super Admin MFA](super-admin-mfa.md) |
