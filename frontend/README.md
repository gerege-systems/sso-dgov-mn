# gerege-web

[`gerege-backend-template-v27`](../gerege-backend-template-v27)-д тохирсон **Next.js** frontend.
Дизайныг [`me.dgov.mn`](../me.dgov.mn)-ээс хуулбарласан **gerege theme** дээр суурилсан.

## Архитектур — BFF (Backend-for-Frontend)

```
Browser ──(адил origin)──► Next.js (route handlers /api/auth/*) ──(server→server)──► Go API /api/v1
   ▲                              │
   └── httpOnly cookie (токен) ◄──┘
```

- **Токен browser-т хэзээ ч ил гарахгүй.** Access/refresh JWT-г `httpOnly` cookie-д
  (`dgov_access`, `dgov_refresh`) хадгална → XSS-д тэсвэртэй.
- **Browser↔Go хооронд CORS хэрэггүй.** Browser зөвхөн Next.js рүү (адил origin)
  хандана; Go API руу зөвхөн Next.js server прокси хийнэ.
- **Reactive refresh.** Хамгаалагдсан дуудлага `401` авбал refresh токеноор нэг
  удаа автоматаар шинэчилж, дахин оролдоно (`src/lib/api.ts`). Refresh нь
  rotation хийдэг тул cookie бичих боломжгүй (RSC render) контекстод refresh
  огт хийгдэхгүй — хүчинтэй сесс шатахаас сэргийлнэ.
- **Давхар CSRF хамгаалалт.** Бүх state-changing BFF route `x-dgov-csrf`
  custom header + Origin шалгалт шаардана (`src/lib/bff.ts`); header-ийг
  `src/lib/client.ts`-ийн `sendJSON/postJSON` нэг газраас тавьдаг.
- **TanStack Query.** GET өгөгдөл (`/api/rbac/me`, admin жагсаалтууд, AI prompt
  тохиргоо) кэш + deduplication + mutation-ы дараах invalidation-тэй.

## Хуудаснууд

| Зам | Тайлбар | Backend endpoint |
|-----|---------|------------------|
| `/` | Landing (анон) / Хяналтын самбар (нэвтэрсэн) | `GET /users/me` |
| `/login` | Нэвтрэх | `POST /auth/login` |
| `/register` | Бүртгүүлэх | `POST /auth/register` |
| `/verify-otp` | OTP-аар бүртгэл идэвхжүүлэх | `POST /auth/send-otp`, `/auth/verify-otp` |
| `/forgot-password` | Нууц үг сэргээх хүсэлт | `POST /auth/password/forgot` |
| `/reset-password` | Токеноор нууц үг шинэчлэх | `POST /auth/password/reset` |
| `/profile` 🔒 | Профайл (read-only) | `GET /users/me` |
| `/settings` 🔒 | Нууц үг солих + гарах | `PUT /auth/password/change`, `POST /auth/logout` |
| `/me/ai` 🔒 | AI туслах — текст/дуут чат (🎤 дуут мессеж, 🔊 TTS) | `POST /ai/chat`, `/ai/tts` |
| `/me/translate` 🔒 | Шууд орчуулга — микрофоны сегментүүдийг live орчуулна | `POST /ai/translate` |
| `/admin/*` 🔒 | Хэрэглэгч/RBAC удирдлага + AI prompt тохиргоо | `/admin/users*`, `/rbac/*`, `/admin/ai/prompts*` |

🔒 = `src/middleware.ts`-аар хамгаалагдсан (refresh cookie байхгүй бол `/login` руу).

AI боломжуудын дотоод бүтцийг [backend/docs/AI_PIPELINE_MN.md](../backend/docs/AI_PIPELINE_MN.md)-аас үз.

## Идэвхжүүлэлтийн урсгал

Backend нь бүртгүүлсэн хэрэглэгчийг **идэвхгүй** үүсгэдэг:
`register` → `send-otp` → `verify-otp` (идэвхжүүлнэ) → `login`.
`/register` амжилттай бол `/verify-otp` руу шилжиж кодыг автоматаар илгээнэ.

## Ажиллуулах

```bash
# 1) Backend-ийг асаа (өөр терминал дээр)
cd ../gerege-backend-template-v27 && make run    # http://localhost:8080

# 2) Орчны хувьсагч
cp .env.example .env.local       # шаардлагатай бол BACKEND_URL-ийг засна

# 3) Frontend
npm install
npm run dev                      # http://localhost:3001
```

| Хувьсагч | Анхдагч | Тайлбар |
|----------|---------|---------|
| `BACKEND_URL` | `http://localhost:8080` | Go API-ийн суурь (api/v1 угтваргүй). Зөвхөн server тал уншина. |
| `COOKIE_SECURE` | `false` | Production (HTTPS) дээр `true` болго. |

## gerege theme

Дизайн систем `src/app/globals.css` дотор — OKLCH токен (DAN blue `#1767E7`),
гэгээн/харанхуй загвар, Inter + JetBrains Mono фонт. Загвар/хэлийн сонголт
`localStorage` (`gerege.theme` / `gerege.lang`)-д хадгалагдаж, FOUC-аас сэргийлэх
`public/theme-bootstrap.js`-ээр render-ийн өмнө тусгагдана.

## Бүтэц

```
src/
  app/
    api/auth/*/route.ts       # BFF прокси (login, register, otp, logout, …)
    api/ai/{chat,stt,tts,translate}/route.ts  # AI BFF прокси (Gemini pipeline)
    api/admin/ai/prompts/     # AI prompt давхаргын тохиргоо (settings.manage)
    me/ai/page.tsx            # AI чат — текст + дуут мессеж + TTS
    me/translate/page.tsx     # Шууд орчуулга (live, mic сегментүүд)
    (pages)/page.tsx          # хуудас бүр server component + client form
    layout.tsx, globals.css
  components/                 # AppShell, Providers (TanStack Query), admin/*, me/*
  lib/
    api.ts                    # server→Go fetch + reactive refresh
    session.ts, cookies.ts    # httpOnly cookie менежмент
    client.ts                 # browser→BFF fetch (CSRF header нэг газраас)
    aiBff.ts                  # AI route-уудын audio whitelist/validation
    audio.ts                  # MediaRecorder сегмент бичлэг + base64 + playback
    password.ts               # нууц үгийн хүч (тохируулж болно)
    format.ts, preferences.ts, types.ts
  middleware.ts               # route хамгаалалт
```
