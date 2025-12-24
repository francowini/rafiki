# Telegram Moment Tracking - Complete Overview

## Summary

Enable moment tracking via Telegram bot using AI-guided conversation based on ACT (Acceptance and Commitment Therapy) functional analysis framework.

**Scope**: Negative moments only (positive moments in Phase 2)

---

## Flow Diagram

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                         TELEGRAM MOMENT TRACKING                             │
└─────────────────────────────────────────────────────────────────────────────┘

1. USER STARTS
   User sends /momento → @RafikiBot
   (any other text outside session → "Usá /momento para empezar")

2. WEBHOOK RECEIVES (immediate response)
   POST /v1/telegram/webhook → partner-service
   • Verify Telegram signature (secret token)
   • Lookup user by telegram_chat_id
   • Load existing session OR create new session (15 min TTL)
   • Enqueue job (River queue, 3 retries, 30s deduplication)
   • Return 200 OK immediately

3. JOB PROCESSES (async)
   app/jobs/telegrammessage/ worker:
   • Load session from PostgreSQL
   • Build prompt with session context + previous steps
   • Call Anthropic API (500 tokens, temp 0.7)
   • Parse AI response (status, feedback, parsed_data)
   • Update session fields
   • Send reply to Telegram

4. 6-STEP CONVERSATION
   ┌────────────────────────────────────────────────────────────────┐
   │ Step 1: Situación + Pensamientos  → "¿Qué pasó?"              │
   │ Step 2: Síntomas + Emociones      → "¿Qué sentiste?"          │
   │ Step 3: Conducta                  → "¿Qué hiciste?"           │
   │ Step 4: Consecuencias             → "¿Qué pasó después?"      │
   │ Step 5: Evitación/Aproximación    → "¿Te acercó o alejó?"     │
   │ Step 6: Intensidad                → "Del 0 al 10..."          │
   └────────────────────────────────────────────────────────────────┘
   • AI validates each step, may ask for refinement (max 2 retries/step)
   • Pure free-text only (no inline keyboards)
   • Therapeutic, non-judgmental tone

5. COMPLETION (auto-save)
   Job worker:
   • Parse intensity (0-10 validated)
   • Call momentbus.Create() with all extracted data
   • Send "✓ Momento guardado" to user
   • Delete session

6. SESSION MANAGEMENT
   • TTL: 15 minutes from last activity
   • /cancel → Hard discard, "Momento descartado"
   • Timeout → Session expires silently, user can restart
```

---

## Commands

| Command | Behavior |
|---------|----------|
| `/momento` | Start new moment tracking session |
| `/cancel` | Discard current session (hard delete, no partial save) |
| `/ayuda` | Show brief explanation + command list |
| `/ejemplo` | Show one complete example conversation |

### Command Responses

**/ayuda**:
```
Rafiki te ayuda a registrar momentos difíciles usando un análisis funcional.

Comandos:
• /momento - Empezar un nuevo registro
• /cancel - Descartar el registro actual
• /ejemplo - Ver un ejemplo de registro completo

El registro tiene 6 pasos y dura unos 5 minutos.
```

**/ejemplo**:
```
Ejemplo de registro:

Paso 1 - ¿Qué pasó?
"Estaba en casa después de almorzar, solo en el sillón. Me apareció el pensamiento de que estoy perdiendo el tiempo."

Paso 2 - ¿Qué sentiste?
"Sentí palpitaciones y las manos transpiradas. Me sentía ansioso."

Paso 3 - ¿Qué hiciste?
"Agarré el celular y me puse a scrollear."

Paso 4 - ¿Qué pasó después?
"Me distraje un rato pero después volvió la ansiedad."

Paso 5 - ¿Te acercó o alejó?
"Evité estar conmigo mismo. Me alejé de lo que quería hacer."

Paso 6 - Intensidad
"7"

✓ Momento guardado
```

---

## Edge Cases

| Scenario | Response |
|----------|----------|
| Text outside session | "Usá /momento para empezar un nuevo registro." |
| Session timeout (15 min) | Session expires silently. Next message: "Tu sesión expiró. Usá /momento para empezar de nuevo." |
| /momento while in session | "Ya tenés un momento en curso. Usá /cancel para descartarlo o seguí respondiendo." |
| Invalid intensity ("mucho") | "Necesito un número del 0 al 10. ¿Qué intensidad le pondrías?" |
| Empty response | "Parece que no escribiste nada. ¿Podés contarme un poco más?" |
| Anthropic API failure | "Hubo un problema técnico. Intentá de nuevo en un momento." (retry up to 3 times) |
| Telegram send failure | Job retries up to 3 times with exponential backoff |

---

## Session State Machine

```
                    ┌──────────────────────────────────────────────────────────┐
                    │                    SESSION STATES                         │
                    └──────────────────────────────────────────────────────────┘

    /momento
        │
        ▼
┌───────────────┐    user responds    ┌───────────────┐    user responds
│  step_1       │ ──────────────────► │  step_2       │ ────────────────►
│ (situacion)   │                     │ (sintomas)    │
└───────────────┘                     └───────────────┘
        │                                     │
        │ needs_refinement                    │ needs_refinement
        │ (max 2 retries)                     │ (max 2 retries)
        ▼                                     ▼
   [re-ask step 1]                       [re-ask step 2]


┌───────────────┐    user responds    ┌───────────────┐    user responds
│  step_3       │ ──────────────────► │  step_4       │ ────────────────►
│ (conducta)    │                     │ (consecuencias)│
└───────────────┘                     └───────────────┘


┌───────────────┐    user responds    ┌───────────────┐    valid 0-10
│  step_5       │ ──────────────────► │  step_6       │ ────────────────►
│ (valores)     │                     │ (intensidad)  │
└───────────────┘                     └───────────────┘
                                              │
                                              ▼
                                      ┌───────────────┐
                                      │  completed    │
                                      │ (auto-save)   │
                                      └───────────────┘

    At any step:
    ─────────────
    /cancel ──────► expired (hard discard)
    15 min TTL ───► expired (silent timeout)
```

---

## Architecture

### Domain Relationships

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                              APP LAYER                                       │
│                    app/jobs/telegrammessage/                                 │
│                                                                              │
│    ┌─────────────────────┐              ┌─────────────────────┐             │
│    │ telegramsessionbus  │              │    momentbus        │             │
│    │   (Child Domain)    │              │  (Child Domain)     │             │
│    └─────────────────────┘              └─────────────────────┘             │
│              │                                    │                          │
│              └────────────────┬───────────────────┘                          │
│                               │                                              │
│                      App Layer Orchestration                                 │
│                    (NO domain-to-domain imports)                             │
└─────────────────────────────────────────────────────────────────────────────┘
                                │
                                ▼
                        ┌───────────────┐
                        │    userbus    │
                        │ (Root Domain) │
                        └───────────────┘
```

**Key Decision**: App Layer Orchestration
- `telegramsessionbus` and `momentbus` are BOTH Child Domains of `userbus`
- The job worker (`app/jobs/telegrammessage/`) imports both domains
- No domain-to-domain coupling (no architecture violation)
- Moment creation happens in job worker after session completion

### Component Diagram

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                           TELEGRAM                                           │
│                        (External Service)                                    │
└─────────────────────────────────────────────────────────────────────────────┘
                                │
                                │ POST /v1/telegram/webhook
                                ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                         NGINX (api.rafiki.lat)                               │
│                    (SSL termination, rate limiting)                          │
└─────────────────────────────────────────────────────────────────────────────┘
                                │
                                ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                         PARTNER-SERVICE                                      │
│                                                                              │
│  ┌──────────────────────┐    ┌──────────────────────┐                       │
│  │   Webhook Handler    │───►│    River Job Queue   │                       │
│  │ (telegramapp)        │    │                      │                       │
│  └──────────────────────┘    └──────────────────────┘                       │
│                                        │                                     │
│                                        ▼                                     │
│  ┌──────────────────────┐    ┌──────────────────────┐                       │
│  │   Telegram Client    │◄───│   Job Worker         │                       │
│  │ (foundation/telegram)│    │ (telegrammessage)    │                       │
│  └──────────────────────┘    └──────────────────────┘                       │
│                                        │                                     │
│                                        ▼                                     │
│  ┌──────────────────────┐    ┌──────────────────────┐                       │
│  │   Anthropic Client   │◄───│   Session Domain     │                       │
│  │ (foundation/anthropic│    │ (telegramsessionbus) │                       │
│  └──────────────────────┘    └──────────────────────┘                       │
│                                        │                                     │
│                                        ▼                                     │
│                              ┌──────────────────────┐                       │
│                              │   Moment Domain      │                       │
│                              │   (momentbus)        │                       │
│                              └──────────────────────┘                       │
└─────────────────────────────────────────────────────────────────────────────┘
                                │
                                ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                           POSTGRESQL                                         │
│    users | telegram_sessions | moments                                       │
└─────────────────────────────────────────────────────────────────────────────┘
```

---

## Deliverables

### Deliverable 1: Anthropic Client ✅ DONE

**Scope**: `foundation/anthropic/`

| File | Purpose |
|------|---------|
| `anthropic.go` | HTTP client for Anthropic Messages API |
| `errors.go` | Error types (RateLimitError, APIError) and helpers |
| `anthropic_test.go` | Unit tests + integration test |

**Purpose of Foundation Layer**:
The foundation layer provides a **generic, decoupled** HTTP client for the Anthropic API. It:
- Returns raw text content from Claude (no JSON parsing)
- Includes token usage information
- Handles error classification (rate limits, retryable errors)
- Has ZERO business domain dependencies

The **app layer** (job workers) is responsible for:
- Defining domain-specific response structures
- Parsing Claude's raw text into those structures
- Building prompts with business context

This separation ensures the client can be reused for any AI feature, not just moment tracking.

**Environment Variables**:
- `PARTNER_ANTHROPIC_APIKEY` - API key (required)
- `PARTNER_ANTHROPIC_MODEL` - Model (default: `claude-haiku-4-5`)
- `PARTNER_ANTHROPIC_MAXTOKENS` - Max response tokens (default: 500)
- `PARTNER_ANTHROPIC_TEMPERATURE` - Temperature (default: 0.7)

**API**:
```go
// Create client
client, err := anthropic.NewClient(apiKey, model)

// Send message - returns raw content
resp, err := client.SendMessage(ctx, anthropic.MessageRequest{
    SystemPrompt: "You are a helpful assistant",
    UserMessage:  "Hello",
})

// resp.Content = raw text from Claude
// resp.Usage.InputTokens, resp.Usage.OutputTokens
```

**Dependencies**: None

---

### Deliverable 2: Database Schema

**Scope**: `business/sdk/migrate/sql/migrate.sql`

**New Table**: `telegram_sessions`
```
telegram_sessions:
  - session_id (UUID, PK)
  - user_id (UUID, FK → users)
  - chat_id (BIGINT)
  - current_step (INT, 1-6)
  - retry_count (INT, 0-2)
  - parsed_data (JSONB)
  - last_activity (TIMESTAMP)
  - date_created (TIMESTAMP)

Index: (last_activity) for TTL cleanup
```

**Update Table**: `users`
```
Add columns:
  - telegram_chat_id (BIGINT, nullable, unique)
  - telegram_enabled (BOOLEAN, default false)
  - telegram_linked_at (TIMESTAMP, nullable)
```

**Dependencies**: None

---

### Deliverable 3: Session Domain

**Scope**: `business/domain/telegramsessionbus/`

| File | Purpose |
|------|---------|
| `model.go` | Session domain model with strong types |
| `telegramsessionbus.go` | CRUD + state machine transitions |
| `stores/telegramsessiondb/` | PostgreSQL storage |

**State Machine Operations**:
- `Create(userID, chatID)` → step_1
- `AdvanceStep(sessionID)` → next step
- `IncrementRetry(sessionID)` → retry_count++
- `UpdateParsedData(sessionID, data)` → update JSONB
- `Complete(sessionID)` → completed
- `Delete(sessionID)` → hard delete

**Dependencies**: Deliverable 2

---

### Deliverable 4: Webhook Handler

**Scope**: `app/domain/telegramapp/`

| File | Purpose |
|------|---------|
| `webhook.go` | POST /v1/telegram/webhook handler |
| `middleware.go` | Signature verification |

**Flow**:
1. Verify Telegram signature (X-Telegram-Bot-Api-Secret-Token)
2. Parse Update (message, chat_id, text)
3. Lookup user by `telegram_chat_id`
4. Handle commands (/momento, /cancel, /ayuda, /ejemplo)
5. Load/create session
6. Enqueue job (River)
7. Return 200 OK

**Environment Variables**:
- `PARTNER_TELEGRAM_BOTTOKEN` - Bot token from @BotFather
- `PARTNER_TELEGRAM_WEBHOOKSECRET` - Secret for signature verification

**Dependencies**: Deliverables 2, 3

---

### Deliverable 5a: Job Worker - Core Structure

**Scope**: `app/jobs/telegrammessage/`

| File | Purpose |
|------|---------|
| `telegrammessage.go` | River job worker definition |
| `handler.go` | Main message processing logic |

**Flow**:
1. Load session from telegramsessionbus
2. Determine current step
3. Route to step handler
4. Send Telegram reply

**Dependencies**: Deliverables 3, 4

---

### Deliverable 5b: Job Worker - Anthropic Integration

**Scope**: `app/jobs/telegrammessage/`

| File | Purpose |
|------|---------|
| `prompts.go` | Prompt templates for each step |
| `anthropic.go` | Anthropic API call + response parsing |

**Prompt Building**:
- Load system prompt from `moment-tracker.md` JSON
- Build step-specific prompt with context
- Replace template variables ({{SITUACION}}, etc.)
- Call Anthropic API (500 tokens, temp 0.7)
- Parse JSON response (status, feedback, parsed_data)

**Dependencies**: Deliverables 1, 5a

---

### Deliverable 5c: Job Worker - Session Management

**Scope**: `app/jobs/telegrammessage/`

| File | Purpose |
|------|---------|
| `steps.go` | Step transition logic |
| `completion.go` | Moment creation on step 6 completion |

**Step Logic**:
- If `status: approved` → advance to next step
- If `status: needs_refinement` → increment retry, re-ask
- If retry_count >= 2 → auto-approve, advance
- If step 6 completed → call momentbus.Create()

**Dependencies**: Deliverables 3, 5b, momentbus

---

### Deliverable 6: Moment Creation Integration

**Scope**: Integration in `app/jobs/telegrammessage/completion.go`

**Flow**:
1. Session reaches step 6 with valid intensity
2. Map parsed_data to Moment fields:
   - `situation` ← step_1.situacion
   - `thoughts` ← step_1.pensamientos
   - `physicalSymptoms` ← step_2.sintomas_fisicos
   - (emotions stored in thoughts or separate field)
   - `behavior` ← step_3.conducta
   - `consequences` ← step_4.consecuencias
   - `valuesReflection` ← step_5.descripcion
   - `intensity` ← step_6.intensidad
3. Call `momentbus.Create()`
4. Delete session
5. Send "✓ Momento guardado"

**Database Update**:
- Add `source` column to moments table: `"web" | "telegram"`
- Add `ai_insights` JSONB column (for future use)

**Dependencies**: Deliverables 3, 5c

---

### Deliverable 7: Admin Linking + Deployment

**Scope**: Admin endpoint + infrastructure

**Admin Endpoint**: `POST /v1/admin/telegram/link`
```json
Request:
{
  "user_id": "uuid",
  "telegram_chat_id": 123456789
}

Response:
{
  "success": true,
  "linked_at": "2024-01-15T10:30:00Z"
}
```

**@BotFather Setup**:
1. `/newbot` → Create @RafikiMomentosBot
2. `/setcommands` → Set command list
3. `/setdescription` → Set bot description
4. Copy bot token → `PARTNER_TELEGRAM_BOTTOKEN`

**Webhook Registration**:
```bash
curl -X POST "https://api.telegram.org/bot<TOKEN>/setWebhook" \
  -H "Content-Type: application/json" \
  -d '{
    "url": "https://api.rafiki.lat/v1/telegram/webhook",
    "secret_token": "<WEBHOOK_SECRET>"
  }'
```

**Verify Webhook**:
```bash
curl "https://api.telegram.org/bot<TOKEN>/getWebhookInfo"
```

**Dependencies**: Deliverable 4

---

## Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `PARTNER_ANTHROPIC_APIKEY` | Anthropic API key | (required) |
| `PARTNER_ANTHROPIC_MODEL` | Claude model | `claude-haiku-4-5` |
| `PARTNER_ANTHROPIC_MAXTOKENS` | Max response tokens | `500` |
| `PARTNER_ANTHROPIC_TEMPERATURE` | Response temperature | `0.7` |
| `PARTNER_TELEGRAM_BOTTOKEN` | Telegram bot token | (required) |
| `PARTNER_TELEGRAM_WEBHOOKSECRET` | Webhook signature secret | (required) |

---

## Key Decisions

| Area | Decision | Rationale |
|------|----------|-----------|
| Architecture | Extend partner-service | Simpler deployment, shared infrastructure |
| State Storage | PostgreSQL (15 min TTL) | Already available, ACID, cleanup job |
| Telegram | Webhook (not polling) | Production-ready, instant, existing HTTPS |
| Input | Pure free-text only | Therapeutic authenticity |
| Save Mode | Auto-save | Natural flow, no decision fatigue |
| Cancel | Hard discard | No partial saves, clean UX |
| Account Linking | Manual by admin | Simple MVP, no deep link complexity |
| Moment Scope | Negative only (MVP) | Focus on core use case, positive in Phase 2 |
| Frontend | ZERO changes | TypeScript handles extra fields gracefully |

---

## Frontend Impact

**Zero frontend changes required for MVP.**

- TypeScript structural typing accepts extra backend fields (`source`, `aiInsights`)
- Existing `MomentList` and `MomentListItem` display Telegram moments identically
- New fields stored but not displayed until Phase 2

---

## Phase 2 (Future)

Not in scope for MVP, but architecture supports:
- Positive moment tracking (neutral prompts)
- Display AI insights in frontend
- Filter moments by source (web/telegram)
- Weekly AI summaries
- Task suggestions after moment completion
- WhatsApp integration (same architecture)
