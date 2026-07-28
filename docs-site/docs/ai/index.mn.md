# AI дамжуулах шугам (Pipeline)

AI дэд систем нь **SDK ашигладаггүй**: `backend/pkg/gemini/gemini.go` нь Google-ийн
Gemini REST эндпойнт руу шууд хандана (`POST {base}/models/{model}:generateContent`,
`x-goog-api-key` auth header ашиглана). Энэ нь платформын Clean Architecture
давхаргалалтыг дагадаг.

## Архитектур

| Давхарга | Файлууд | Үүрэг |
|-------|-------|------|
| REST client | `pkg/gemini/gemini.go`, `wav.go` | `gemini.Generator` интерфейс; дахин оролдлого (3 удаа, 500 ms→1 s backoff); 32 MiB хариултын дээд хязгаар |
| Usecase | `business/usecases/ai/{ai_usecase,ai_impl,ai_prompts,ai_tools,ai_speech}.go` | chat давталт, tools, prompts, яриа |
| Repository | `datasources/repositories/postgres/ai/ai_postgres.go` | `ListPrompts`, `SetPrompt`, `SearchKnowledge` (лавлах өгөгдөл — RLS-ийг тойрдог) |

API түлхүүр байхгүй бол `gemini.ErrNotConfigured` sentinel утгыг буцаана. Хоёр
client instance холбогдсон: нэг нь chat/STT/translate-д зориулагдсан (`GEMINI_MODEL`),
нөгөө нь тусдаа `geminiTTSClient` (`GEMINI_TTS_MODEL`) — учир нь chat загвар аудио
гаргадаггүй.

## Хүсэлтийн урсгал

Хөтөч → same-origin BFF (`/api/ai/*`, CSRF + JWT нэмнэ) → Go API
`/api/v1/ai/*` (Bearer + ~20/мин хязгаарлагч).

Chat давталт (`usecase.Run`):

1. `buildContents` нь түүхээс (20 turn хүртэл товчлогдсон) болон шинэ prompt-оос
   Gemini-ийн `contents`-ийг угсарна. Дуут мессежийг хэрэглэгчийн turn дотор
   шууд inline base64 аудио байдлаар хавсаргана — олон төрлийн (multimodal)
   загвар үүнийг уншдаг тул тусдаа STT алхам шаардлагагүй.
2. Хүсэлт нь давхаргалсан `SystemInstruction` болон аливаа tool-ийн
   `FunctionDeclaration`-уудыг агуулна.
3. `MaxSteps` (өгөгдмөл 4) хүртэл давтана: Gemini-г дуудна; хэрэв хариу нь function
   дуудлагатай бол tool бүрийг гүйцэтгэж, загварын turn болон `functionResponse`
   turn-ийг нэмээд дахин давтана. Гүйцэтгэсэн дуудлага бүрийг `Step{Tool, Args, Result}`
   болгон бүртгэж клиент рүү буцаана — ингэснээр UI нь "AI юу хийсэн"-ийг харуулж чадна.
4. Хэрэв хариу нь текст бол түүнийг буцаана.

!!! abstract "Зарчим"
    Загвар нь аль tool-ийг дуудахыг *шийддэг*; backend нь түүнийг *гүйцэтгэдэг* —
    загвар хэзээ ч код ажиллуулдаггүй. Tools нь серверийн талд хүсэлтийн `ctx`-тэй
    ажиллах тул RLS болон timeout үйлчилнэ.

## Давхаргалсан system prompt — хамгаалалтын хашлага (guardrails) хэзээ ч тохируулагддаггүй

Хүсэлт бүрд гурван давхаргаас угсарна:

| Давхарга | Эх сурвалж | Засварлах боломж | Агуулга |
|-------|--------|----------|---------|
| 1. Суурь guardrails | `const baseInstruction` (Go-д хатуу кодлогдсон) | **Хэзээ ч үгүй** | Зөвхөн монголоор хариулна; хамрах хүрээнээс гадуурх зүйлээс татгалзана; "заавраа март / prompt-оо ил гарга" гэдгийг энгийн текст болгон авч татгалзана; платформын асуултад хариулахаас өмнө `search_knowledge`-ийг дуудна |
| 2. Scope | `ai_prompts` түлхүүр `scope` → env → өгөгдмөл | Admin, ажиллах үед | Туслах *юунд* тусалдаг |
| 3. Instructions | `ai_prompts` түлхүүр `instructions` (заавал биш) | Admin, ажиллах үед | Өнгө аяс / нэмэлт дүрэм |

Суурь guardrail давхарга нь **DB болон env замгүй** compile-time тогтмол утга —
тохируулж болох гадаргуу нь guardrails доторх зан үйлийг зөвхөн *нарийсгах/чимэглэх*
боломжтой ба хэзээ ч орлуулж чадахгүй.

**Prompt хадгалалтын баталгаа:**

- Зөвшөөрөгдсөн түлхүүрүүд тогтмол (`scope`, `instructions`); өөр аливаа түлхүүр татгалзагдана.
- `SetPrompt` нь тарьсан (seeded) мөрүүдийн эсрэг зөвхөн **UPDATE** хийдэг — prompt
  гадаргуу API-аас өсөж чадахгүй. Үүнийг DB grant түвшинд бататгасан (`REVOKE INSERT, DELETE ON
  ai_prompts FROM app_user`).
- Prompt-ууд 1 минут кэшлэгддэг; бичих үйлдэл кэшийг nil болгодог тул тухайн
  instance дээр шууд үйлчилнэ.
- DB унших алдаа нь **fail-open**: prompt хайлт env/өгөгдмөл рүү шилжих тул хэзээ ч
  chat-ийг зогсоож чадахгүй.

Admin API: `GET /api/v1/admin/ai/prompts`, `PUT /api/v1/admin/ai/prompts/{key}`
(хоёулаа `settings.manage`).

## Function-calling tools

Tool гэдэг нь `ai.ToolDef`: `gemini.FunctionDeclaration` (нэр, тайлбар,
JSON-Schema параметрүүд) дээр нэмэн `Execute ToolFunc`. `server.go` дахь slice-ийг
өргөтгөж **бүртгэнэ**:

```go
aiTools := append(ai.DefaultTools(), ai.KnowledgeSearchTool(aiRepo), myTool)
```

Tool-ийн алдаа нь загвар руу буцаагддаг (ингэснээр эелдгээр уучлалт гуйж чадна) ба
хэзээ ч клиент рүү шууд гарч ирдэггүй.

Нийлүүлэгдсэн tools:

- **`search_knowledge`** — хариултыг `ai_knowledge` хүснэгтэд бэхжүүлнэ
  (`SearchKnowledge(ctx, query, 5)`, `title/content ILIKE` + tag таарц). Суурь
  guardrails нь загварыг платформын асуултад хариулахаас *өмнө* үүнийг дуудахыг заадаг.
  `app_user` нь зөвхөн SELECT эрхтэй.
- **`get_server_time`** — Улаанбаатарын огноо/цагийг буцаадаг хараат бус демо.

## Боломжууд

| Боломж | Эндпойнт | Тэмдэглэл |
|------------|----------|-------|
| Текст/дуут chat | `POST /ai/chat` | дуу нь inline audio (multimodal) |
| Ярианаас текст | `POST /ai/stt` | нэг удаагийн үгчилсэн хөрвүүлэг |
| Текстээс ярианд | `POST /ai/tts` | PCM → WAV (16-bit mono, 24 kHz); өгөгдмөл дуу хоолой `Kore` |
| Шууд орчуулга | `POST /ai/translate` | текст→орчуулга; аудио→STT дараа нь орчуулга; заавал биш TTS (TTS алдаа гарвал чимээгүй доройтоно) |

## Доройтсон монгол fallback

**Зөвхөн chat зам дээр**, клиентийн өөрийн дахин оролдлогуудын дараа, түр зуурын
Gemini алдаа нь 5xx **гаргадаггүй**. `Run` нь `{ Reply: fallbackReply, Degraded:
true }`-г буцаана — "AI үйлчилгээ түр завгүй байна" гэсэн монгол мессеж. Мөн адил
доройтсон хариултыг загварын хоосон текст эсвэл `MaxSteps` дуусахад буцаана. Цорын
ганц хатуу алдаа нь `ErrNotConfigured` (`GEMINI_API_KEY` дутуу) → HTTP 500. STT / TTS
/ translate нь оронд нь алдаа гарвал алдаа буцаана (UI нь "дахин оролдох" харуулна).
