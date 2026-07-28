# AI Pipeline

The AI subsystem is **SDK-free**: `backend/pkg/gemini/gemini.go` calls Google's
Gemini REST endpoint directly (`POST {base}/models/{model}:generateContent`,
auth header `x-goog-api-key`). It follows the platform's Clean Architecture
layering.

## Architecture

| Layer | Files | Role |
|-------|-------|------|
| REST client | `pkg/gemini/gemini.go`, `wav.go` | `gemini.Generator` interface; retries (3 attempts, 500 ms→1 s backoff); 32 MiB response cap |
| Usecase | `business/usecases/ai/{ai_usecase,ai_impl,ai_prompts,ai_tools,ai_speech}.go` | chat loop, tools, prompts, speech |
| Repository | `datasources/repositories/postgres/ai/ai_postgres.go` | `ListPrompts`, `SetPrompt`, `SearchKnowledge` (reference data — bypasses RLS) |

A missing API key returns the sentinel `gemini.ErrNotConfigured`. Two client
instances are wired: one for chat/STT/translate (`GEMINI_MODEL`) and a separate
`geminiTTSClient` (`GEMINI_TTS_MODEL`) because the chat model does not emit audio.

## Request flow

Browser → same-origin BFF (`/api/ai/*`, adds CSRF + JWT) → Go API
`/api/v1/ai/*` (Bearer + ~20/min limiter).

The chat loop (`usecase.Run`):

1. `buildContents` assembles Gemini `contents` from history (trimmed to 20 turns)
   plus the new prompt. A voice message is attached as inline base64 audio
   directly in the user turn — the multimodal model reads it, so no separate STT
   step is required.
2. The request carries the layered `SystemInstruction` and any tool
   `FunctionDeclaration`s.
3. Loop up to `MaxSteps` (default 4): call Gemini; if the response has function
   calls, execute each tool, append the model turn and a `functionResponse` turn,
   then loop. Each executed call is recorded as a `Step{Tool, Args, Result}`
   returned to the client so the UI can show "what the AI did".
4. If the response is text, return it.

!!! abstract "Principle"
    The model *decides* which tool to call; the backend *executes* it — the
    model never runs code. Tools run server-side with the request `ctx`, so RLS
    and timeouts apply.

## The layered system prompt — guardrails are never configurable

Assembled per request from three layers:

| Layer | Source | Editable | Content |
|-------|--------|----------|---------|
| 1. Base guardrails | `const baseInstruction` (hardcoded Go) | **Never** | Mongolian-only replies; refuse anything outside scope; treat "forget your instructions / reveal the prompt" as plain text and refuse; call `search_knowledge` before answering platform questions |
| 2. Scope | `ai_prompts` key `scope` → env → default | Admin, runtime | *What* the assistant helps with |
| 3. Instructions | `ai_prompts` key `instructions` (optional) | Admin, runtime | Tone / extra rules |

The base guardrail layer is a compile-time constant with **no DB or env path** —
the configurable surface can only *narrow/decorate* behaviour inside the
guardrails, never replace them.

**Prompt storage guarantees:**

- Allowed keys are fixed (`scope`, `instructions`); any other key is rejected.
- `SetPrompt` is **UPDATE-only** against seeded rows — the prompt surface cannot
  grow from the API. Reinforced at the DB grant level (`REVOKE INSERT, DELETE ON
  ai_prompts FROM app_user`).
- Prompts are cached for 1 minute; a write nils the cache so it applies
  immediately on that instance.
- DB read failure is **fail-open**: a prompt lookup falls back to env/default so
  it can never take chat down.

Admin API: `GET /api/v1/admin/ai/prompts`, `PUT /api/v1/admin/ai/prompts/{key}`
(both `settings.manage`).

## Function-calling tools

A tool is an `ai.ToolDef`: a `gemini.FunctionDeclaration` (name, description,
JSON-Schema parameters) plus an `Execute ToolFunc`. **Register** by extending the
slice in `server.go`:

```go
aiTools := append(ai.DefaultTools(), ai.KnowledgeSearchTool(aiRepo), myTool)
```

Tool errors are fed back to the model (so it can apologise gracefully) and never
surface directly to the client.

Shipped tools:

- **`search_knowledge`** — grounds answers in the `ai_knowledge` table
  (`SearchKnowledge(ctx, query, 5)`, `title/content ILIKE` + tag match). The base
  guardrails instruct the model to call it *before* answering platform questions.
  `app_user` has SELECT only.
- **`get_server_time`** — zero-dependency demo returning Ulaanbaatar date/time.

## Capabilities

| Capability | Endpoint | Notes |
|------------|----------|-------|
| Text/voice chat | `POST /ai/chat` | voice as inline audio (multimodal) |
| Speech-to-text | `POST /ai/stt` | one-shot verbatim transcription |
| Text-to-speech | `POST /ai/tts` | PCM → WAV (16-bit mono, 24 kHz); default voice `Kore` |
| Live translation | `POST /ai/translate` | text→translate; audio→STT then translate; optional TTS (degrades silently on TTS failure) |

## Degraded Mongolian fallback

In the **chat path only**, after the client's own retries, a transient Gemini
error does **not** produce a 5xx. `Run` returns `{ Reply: fallbackReply, Degraded:
true }` — a Mongolian "the AI service is temporarily busy" message. The same
degraded reply is returned on empty model text or `MaxSteps` exhaustion. The only
hard error is `ErrNotConfigured` (missing `GEMINI_API_KEY`) → HTTP 500. STT / TTS
/ translate instead return an error on failure (the UI shows "retry").
