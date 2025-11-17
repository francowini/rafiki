# Telegram Integration Implementation Plan for Rafiki

**Generated**: 2025-11-17
**Team**: Backend Engineer + Frontend Engineer + DevOps Engineer
**Methodology**: Multi-mind collaborative analysis

---

## Executive Summary

This document provides a comprehensive implementation plan for integrating Telegram bot functionality into Rafiki, allowing users to create moments directly from Telegram messages. The plan represents consensus across backend, frontend, and DevOps perspectives.

### Key Decisions

| Decision | Choice | Rationale |
|----------|--------|-----------|
| **Architecture** | Integrated into partner-service | Resource efficiency on CPX11, operational simplicity |
| **Communication Mode** | Long polling (MVP) → Webhook (Production) | Simple start, scalable future |
| **Conversation UX** | Guided conversation (7 steps) | Best mobile UX, easy validation |
| **State Management** | Database-backed (telegram_conversations table) | Stateless webhooks, horizontal scaling ready |
| **Frontend Updates** | Polling (30s) + Page Visibility API | Battery-efficient, user-controlled |
| **Deployment** | Zero infrastructure changes for MVP | Risk mitigation, fast iteration |

### Timeline

- **Phase 1 (MVP)**: 2-3 weeks - Guided conversation, polling mode, basic UI
- **Phase 2 (Production)**: 1-2 weeks - Webhook mode, enhanced UI, monitoring
- **Phase 3 (Advanced)**: Future - NLP parsing, voice messages, analytics

---

## Table of Contents

1. [Architecture Overview](#architecture-overview)
2. [Backend Implementation](#backend-implementation)
3. [Frontend Implementation](#frontend-implementation)
4. [DevOps & Deployment](#devops--deployment)
5. [API Contracts](#api-contracts)
6. [Database Schema](#database-schema)
7. [User Flows](#user-flows)
8. [Testing Strategy](#testing-strategy)
9. [Deployment Sequence](#deployment-sequence)
10. [Monitoring & Observability](#monitoring--observability)
11. [Risk Mitigation](#risk-mitigation)
12. [Questions for Product Owner](#questions-for-product-owner)

---

## Architecture Overview

### System Architecture (Phase 1 - MVP)

```
┌─────────────────────────────────────────────────────────────┐
│                     User Journey                             │
└─────────────────────────────────────────────────────────────┘
                            │
        ┌───────────────────┴───────────────────┐
        │                                       │
        ▼                                       ▼
┌──────────────────┐                   ┌──────────────────┐
│  Rafiki Web App  │                   │  Telegram App    │
│  (Vercel)        │                   │  (Mobile)        │
└────────┬─────────┘                   └────────┬─────────┘
         │                                      │
         │ HTTPS                                │
         │                                      │
         ▼                                      ▼
┌─────────────────────────────────────────────────────────────┐
│              Hetzner CPX11 Server (178.156.170.37)          │
├─────────────────────────────────────────────────────────────┤
│                                                               │
│  ┌────────────────────────────────────────────────────────┐ │
│  │ Nginx (Port 443 - SSL Termination)                     │ │
│  │  - CORS: Vercel origins allowed                        │ │
│  │  - Rate Limiting: 10 req/s (API), 30 req/s (webhook)   │ │
│  └────────────────────────────────────────────────────────┘ │
│                           │                                  │
│                           ▼                                  │
│  ┌────────────────────────────────────────────────────────┐ │
│  │ Partner-Service (Port 3000)                            │ │
│  │                                                         │ │
│  │  ┌──────────────────────────────────────────────────┐ │ │
│  │  │ REST API Layer                                    │ │ │
│  │  │  - /v1/moments (existing)                        │ │ │
│  │  │  - /v1/telegram/status                           │ │ │
│  │  │  - /v1/telegram/link-code                        │ │ │
│  │  │  - /v1/telegram/disconnect                       │ │ │
│  │  │  - /v1/telegram/webhook (Phase 2)                │ │ │
│  │  └──────────────────────────────────────────────────┘ │ │
│  │                                                         │ │
│  │  ┌──────────────────────────────────────────────────┐ │ │
│  │  │ Telegram Bot Service (Background Goroutine)      │ │ │
│  │  │  - Long Polling (GetUpdates every 30s)          │ │ │
│  │  │  - Conversation State Machine                    │ │ │
│  │  │  - Message Parsing & Validation                  │ │ │
│  │  └──────────────────────────────────────────────────┘ │ │
│  │                                                         │ │
│  │  ┌──────────────────────────────────────────────────┐ │ │
│  │  │ Business Layer                                    │ │ │
│  │  │  - momentbus (existing)                          │ │ │
│  │  │  - userbus (existing)                            │ │ │
│  │  │  - telegrambus (NEW)                             │ │ │
│  │  └──────────────────────────────────────────────────┘ │ │
│  └────────────────────────────────────────────────────────┘ │
│                           │                                  │
│                           ▼                                  │
│  ┌────────────────────────────────────────────────────────┐ │
│  │ PostgreSQL Database                                     │ │
│  │  - users (extended with telegram_user_id, link_code)   │ │
│  │  - moments (extended with source, telegram_message_id) │ │
│  │  - telegram_conversations (NEW)                        │ │
│  └────────────────────────────────────────────────────────┘ │
│                                                               │
└─────────────────────────────────────────────────────────────┘
                           ▲
                           │ Long Polling (GetUpdates)
                           │ SendMessage (responses)
                           │
                   ┌───────┴────────┐
                   │ Telegram API   │
                   │ (Bot API)      │
                   └────────────────┘
```

### Data Flow: Creating a Moment via Telegram

```
1. User opens Telegram app → Sends "/moment" command
2. Telegram servers → Store message
3. Partner-service bot → Polls GetUpdates (max 30s delay)
4. Bot receives update → Loads conversation state from DB
5. Bot detects "/moment" → Creates new conversation state
6. Bot sends "Describe the situation:" → via SendMessage API
7. User replies "Had panic attack at work"
8. Bot receives reply → Validates, stores in partial_data
9. Bot advances state → "What were your thoughts?"
10. ... (repeats for all 7 fields)
11. Bot receives final field (intensity) → Validates all data
12. Bot calls momentBus.Create() → Moment saved to DB
13. Bot sends success message → "✅ Moment created!"
14. Frontend polls /v1/moments → Detects new moment (source=telegram)
15. Frontend shows notification → "New moment from Telegram"
```

---

## Backend Implementation

### Directory Structure

```
business/
├── domain/
│   └── telegrambus/           # NEW: Telegram business logic
│       ├── model.go           # TelegramConnection, ConversationState
│       ├── telegrambus.go     # Business layer interface
│       ├── filter.go          # Query filters
│       ├── order.go           # Ordering
│       └── stores/
│           └── telegramdb/
│               ├── model.go   # DB models
│               ├── telegramdb.go
│               └── filter.go
├── sdk/
│   ├── migrate/sql/
│   │   └── migrate.sql        # UPDATED: Add version 1.04
│   └── telegram/              # NEW: Telegram API client
│       ├── client.go          # Bot API wrapper
│       └── types.go           # Telegram types
└── types/
    └── telegram/              # NEW: Business types (optional)
        └── chatid.go          # ChatID type with validation

app/
└── domain/
    └── telegramapp/           # NEW: HTTP handlers
        ├── model.go           # API request/response models
        ├── telegramapp.go     # Handler implementations
        └── route.go           # Route registration

api/services/partners/
├── main.go                    # UPDATED: Initialize telegram bot
├── telegram/                  # NEW: Bot service
│   ├── bot.go                # Bot lifecycle, polling
│   ├── handlers.go           # Command handlers
│   ├── conversation.go       # State machine
│   └── state.go              # In-memory state store
└── mux/
    └── mux.go                # UPDATED: Register telegram routes
```

### Key Components

#### 1. Database Migration (Version 1.04)

**File**: `business/sdk/migrate/sql/migrate.sql`

```sql
-- Version: 1.04
-- Description: Telegram integration tables

-- Extend users table for Telegram linking
ALTER TABLE users
  ADD COLUMN telegram_user_id BIGINT NULL UNIQUE,
  ADD COLUMN telegram_username TEXT NULL,
  ADD COLUMN telegram_connected_at TIMESTAMP NULL,
  ADD COLUMN telegram_link_code TEXT NULL UNIQUE,
  ADD COLUMN telegram_link_code_expiry TIMESTAMP NULL;

CREATE INDEX users_telegram_link_code_idx
  ON users(telegram_link_code)
  WHERE telegram_link_code IS NOT NULL;

-- Extend moments table for source tracking
ALTER TABLE moments
  ADD COLUMN source TEXT NOT NULL DEFAULT 'web',
  ADD COLUMN telegram_message_id INTEGER NULL;

CREATE INDEX moments_source_idx
  ON moments(user_id, source, date_created DESC);

-- Conversation state table
CREATE TABLE telegram_conversations (
    telegram_user_id   BIGINT      NOT NULL,
    user_id            UUID        NOT NULL,
    state              TEXT        NOT NULL,
    partial_data       JSONB       NOT NULL DEFAULT '{}',
    last_message_id    INTEGER     NULL,
    last_activity      TIMESTAMP   NOT NULL,
    date_created       TIMESTAMP   NOT NULL,
    date_updated       TIMESTAMP   NOT NULL,

    PRIMARY KEY (telegram_user_id),
    FOREIGN KEY (user_id) REFERENCES users(user_id) ON DELETE CASCADE
);

CREATE INDEX telegram_conversations_last_activity_idx
  ON telegram_conversations(last_activity);

COMMENT ON TABLE telegram_conversations IS 'Active Telegram conversation states';
COMMENT ON COLUMN telegram_conversations.state IS 'Current state: idle, awaiting_situation, etc.';
COMMENT ON COLUMN telegram_conversations.partial_data IS 'Accumulated moment data as JSON';
```

#### 2. Conversation State Machine

**States**:
- `idle`: No active conversation
- `awaiting_situation`: Waiting for situation description
- `awaiting_thoughts`: Waiting for thoughts
- `awaiting_physical_symptoms`: Waiting for physical symptoms
- `awaiting_behavior`: Waiting for behavior description
- `awaiting_consequences`: Waiting for consequences
- `awaiting_values_reflection`: Waiting for values reflection
- `awaiting_intensity`: Waiting for intensity (0-10)

**Transitions**:
```
/moment → awaiting_situation
awaiting_situation + text → awaiting_thoughts
awaiting_thoughts + text → awaiting_physical_symptoms
awaiting_physical_symptoms + text → awaiting_behavior
awaiting_behavior + text → awaiting_consequences
awaiting_consequences + text → awaiting_values_reflection
awaiting_values_reflection + text → awaiting_intensity
awaiting_intensity + number → CREATE MOMENT → idle
/cancel (any state) → idle
```

#### 3. Telegram Bot Service

**File**: `api/services/partners/telegram/bot.go`

```go
type Bot struct {
    log       *logger.Logger
    api       *telegram.Client
    momentBus *momentbus.Business
    userBus   *userbus.Business
    stateStore *StateStore
}

func (b *Bot) StartPolling(ctx context.Context) error {
    updateConfig := tgbotapi.NewUpdate(0)
    updateConfig.Timeout = 30

    updates := b.api.GetUpdatesChan(updateConfig)

    for {
        select {
        case <-ctx.Done():
            return ctx.Err()
        case update := <-updates:
            go b.handleUpdate(ctx, update)
        }
    }
}

func (b *Bot) handleUpdate(ctx context.Context, update tgbotapi.Update) {
    if update.Message == nil {
        return
    }

    chatID := update.Message.Chat.ID
    text := update.Message.Text

    // Load conversation state
    state, err := b.stateStore.Get(chatID)
    if err != nil {
        b.log.Error(ctx, "get conversation state", "err", err)
        return
    }

    // Route based on state and message
    if strings.HasPrefix(text, "/") {
        b.handleCommand(ctx, chatID, text)
        return
    }

    b.handleConversation(ctx, chatID, text, state)
}
```

#### 4. User Linking Flow

**Backend endpoints**:
1. `POST /v1/telegram/link-code` - Generate link code (30-minute TTL)
2. User sends `/link CODE` in Telegram
3. Bot validates code, links telegram_user_id to user_id
4. `GET /v1/telegram/status` - Frontend checks status

**Security**:
- Link codes are cryptographically random (16 bytes)
- Stored hashed in database
- Single-use (marked as consumed after linking)
- 30-minute expiry
- Rate-limited: Max 5 codes per user per hour

---

## Frontend Implementation

### Components to Create

#### 1. Settings Page

**File**: `frontend/app/(dashboard)/settings/page.tsx`

Features:
- Telegram connection status card
- Link code generation with countdown timer
- QR code display (for desktop users)
- Deep link button (for mobile users)
- Disconnect button with confirmation
- Import history table

#### 2. Telegram Status Card

**File**: `frontend/components/features/TelegramStatusCard.tsx`

States:
- **Not Connected**: Show setup instructions
- **Connecting**: Show link code and countdown
- **Connected**: Show username, connection date, disconnect button
- **Error**: Show error message with retry action

#### 3. Moment Card Update

**File**: `frontend/components/features/MomentCard.tsx` (UPDATED)

Changes:
- Add source badge (Telegram icon for telegram-sourced moments)
- Subtle visual difference (border color or background tint)
- Preserve existing functionality

### API Client Updates

**File**: `frontend/lib/api.ts`

```typescript
export const api = {
  // ... existing methods

  telegram: {
    getStatus: async (): Promise<TelegramStatus> => {
      return fetchAPI<TelegramStatus>("/v1/telegram/status");
    },

    generateLinkCode: async (): Promise<LinkCodeResponse> => {
      return fetchAPI<LinkCodeResponse>("/v1/telegram/link-code", {
        method: "POST",
      });
    },

    disconnect: async (): Promise<void> => {
      return fetchAPI("/v1/telegram/disconnect", {
        method: "DELETE",
      });
    },

    getImportHistory: async (params?: PaginationParams): Promise<ImportHistoryResponse> => {
      const query = new URLSearchParams();
      if (params?.page) query.set("page", params.page.toString());
      if (params?.rows) query.set("rows", params.rows.toString());
      return fetchAPI(`/v1/telegram/imports${query ? `?${query}` : ""}`);
    },
  },
};
```

### Polling Strategy (Page Visibility API)

```typescript
// Only poll when page is visible
useEffect(() => {
  const handleVisibilityChange = () => {
    if (document.visibilityState === 'visible') {
      // Refresh moments when user returns to tab
      refreshMoments();
    }
  };

  document.addEventListener('visibilitychange', handleVisibilityChange);
  return () => {
    document.removeEventListener('visibilitychange', handleVisibilityChange);
  };
}, []);

// Manual refresh button (user-controlled)
<Button onClick={refreshMoments}>
  <RefreshCw className="h-4 w-4" />
  Refresh
</Button>
```

---

## DevOps & Deployment

### Environment Configuration

**Production file**: `/opt/rafiki/.env` (on server)

```bash
# Existing variables...

# Telegram Bot Configuration (ADD ONCE)
TELEGRAM_BOT_TOKEN=1234567890:ABCdefGHIjklMNOpqrsTUVwxyz
TELEGRAM_BOT_USERNAME=rafiki_moments_bot
```

### Docker Compose Updates

**File**: `docker-compose.yml`

```yaml
partner-service:
  environment:
    # ... existing vars ...

    # Telegram Bot
    - PARTNER_TELEGRAM_BOTTOKEN=${TELEGRAM_BOT_TOKEN:-}
    - PARTNER_TELEGRAM_BOTUSERNAME=${TELEGRAM_BOT_USERNAME:-}
```

### Nginx Configuration (Phase 2 - Webhook)

**File**: `nginx/nginx.conf`

```nginx
# Telegram webhook endpoint (Phase 2)
location /v1/telegram/webhook {
    # IP whitelist (Telegram servers only)
    allow 149.154.160.0/20;
    allow 91.108.4.0/22;
    deny all;

    proxy_pass http://partner-service:3000;
    proxy_set_header Host $host;
    proxy_set_header X-Real-IP $remote_addr;
    proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    proxy_read_timeout 60s;
}

# Telegram user endpoints
location /v1/telegram/ {
    limit_req zone=api_limit burst=5 nodelay;
    proxy_pass http://partner-service:3000;
    # ... standard proxy headers
}
```

---

## API Contracts

### Complete Endpoint Specification

#### 1. GET /v1/telegram/status

**Purpose**: Check if user's Telegram account is linked

**Authentication**: Required (Bearer JWT)

**Response** (Connected):
```json
{
  "connected": true,
  "telegramUserId": "123456789",
  "telegramUsername": "john_doe",
  "connectedAt": "2025-11-17T10:30:00Z",
  "lastMessageAt": "2025-11-17T12:45:00Z"
}
```

**Response** (Not Connected):
```json
{
  "connected": false,
  "linkCode": null,
  "linkCodeExpiry": null
}
```

---

#### 2. POST /v1/telegram/link-code

**Purpose**: Generate a new link code for connecting Telegram account

**Authentication**: Required (Bearer JWT)

**Request**: Empty body

**Response**:
```json
{
  "linkCode": "LINK-ABC123XYZ",
  "expiresAt": "2025-11-17T11:00:00Z",
  "botUsername": "rafiki_moments_bot",
  "instructions": "Open Telegram, search for @rafiki_moments_bot, send: /link LINK-ABC123XYZ"
}
```

**Error Codes**:
- `409 Conflict`: Already connected
- `429 Too Many Requests`: Rate limit exceeded (max 5/hour)

---

#### 3. DELETE /v1/telegram/disconnect

**Purpose**: Disconnect user's Telegram account

**Authentication**: Required (Bearer JWT)

**Response**:
```json
{
  "message": "Telegram account disconnected successfully"
}
```

---

#### 4. GET /v1/telegram/imports

**Purpose**: Get history of moments imported via Telegram

**Authentication**: Required (Bearer JWT)

**Query Parameters**:
- `page` (optional): Page number (default: 1)
- `rows` (optional): Rows per page (default: 10)

**Response**:
```json
{
  "items": [
    {
      "momentId": "550e8400-e29b-41d4-a716-446655440000",
      "telegramMessageId": "12345",
      "importedAt": "2025-11-17T12:30:00Z",
      "situation": "Had a panic attack at work...",
      "intensity": 8
    }
  ],
  "total": 25,
  "page": 1,
  "rows": 10
}
```

---

## Database Schema

### Extended Tables

#### users (Extended)

```sql
ALTER TABLE users ADD COLUMN:
- telegram_user_id BIGINT NULL UNIQUE
- telegram_username TEXT NULL
- telegram_connected_at TIMESTAMP NULL
- telegram_link_code TEXT NULL UNIQUE
- telegram_link_code_expiry TIMESTAMP NULL
```

#### moments (Extended)

```sql
ALTER TABLE moments ADD COLUMN:
- source TEXT NOT NULL DEFAULT 'web'  -- 'web' | 'telegram'
- telegram_message_id INTEGER NULL
```

#### telegram_conversations (New)

```sql
CREATE TABLE telegram_conversations (
    telegram_user_id BIGINT PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(user_id),
    state TEXT NOT NULL,
    partial_data JSONB NOT NULL DEFAULT '{}',
    last_message_id INTEGER NULL,
    last_activity TIMESTAMP NOT NULL,
    date_created TIMESTAMP NOT NULL,
    date_updated TIMESTAMP NOT NULL
);
```

---

## User Flows

### Flow 1: First-Time Setup (Desktop)

1. User logs into Rafiki web app
2. Navigates to Settings (new page)
3. Sees "Telegram Bot" section - Status: Not Connected
4. Clicks "Setup Telegram Integration"
5. Frontend calls `POST /v1/telegram/link-code`
6. Frontend displays:
   - QR code (encodes: `https://t.me/rafiki_moments_bot?start=LINK-ABC123`)
   - Link code: `LINK-ABC123`
   - Countdown timer: "Expires in 29:45"
7. User scans QR code with phone → Opens Telegram app
8. Telegram bot starts automatically with `/start LINK-ABC123`
9. Bot validates code, links account
10. Bot sends: "✅ Connected! You can now create moments here."
11. User returns to web app, clicks "I've completed setup"
12. Frontend calls `GET /v1/telegram/status` → `connected: true`
13. Status changes to "Connected as @john_doe"

### Flow 2: Creating a Moment via Telegram

1. User opens Telegram, navigates to @rafiki_moments_bot
2. Sends: `/moment`
3. Bot responds: "Let's create a moment. Describe the situation:"
4. User: "Had a panic attack during a work meeting"
5. Bot: "What were your thoughts during this?"
6. User: "I felt like everyone was judging me"
7. Bot: "What physical symptoms did you notice?"
8. User: "Heart racing, sweating, tight chest"
9. Bot: "How did you behave?"
10. User: "I excused myself and left the room"
11. Bot: "What were the consequences?"
12. User: "Missed important decision, felt embarrassed"
13. Bot: "What values were involved?"
14. User: "Struggled between professionalism and self-care"
15. Bot: "Rate the intensity (0-10):"
16. User: "8"
17. Bot: "✅ Moment created! View in app: https://app.rafiki.lat/moments"
18. [30 seconds later] User opens web app
19. Frontend polls `/v1/moments` → Detects new moment
20. Shows notification: "New moment from Telegram"
21. Moment appears with Telegram badge

### Flow 3: Handling Errors

**Scenario**: User sends invalid intensity

1. User in conversation, at intensity step
2. Sends: "very high"
3. Bot detects non-numeric input
4. Bot responds: "❌ Please enter a number between 0 and 10.\n\n0 = no distress\n5 = moderate\n10 = extreme"
5. Conversation state unchanged (still awaiting_intensity)
6. User sends: "9"
7. Bot accepts, creates moment

---

## Testing Strategy

### Backend Tests

**Unit Tests**:
```go
// business/domain/telegrambus/telegrambus_test.go
func TestCreateLinkCode(t *testing.T)
func TestValidateLinkCode(t *testing.T)
func TestConversationStateTransition(t *testing.T)
func TestParseIntensity(t *testing.T)
```

**Integration Tests**:
```go
// app/domain/telegramapp/telegramapp_test.go
func TestLinkCodeEndpoint(t *testing.T)
func TestStatusEndpoint(t *testing.T)
func TestWebhookHandler(t *testing.T)
```

### Frontend Tests

**Component Tests** (Jest + React Testing Library):
```typescript
// TelegramStatusCard.test.tsx
test('shows not connected state initially')
test('displays link code with countdown timer')
test('shows connected state with username')
test('handles disconnect action')

// MomentCard.test.tsx
test('renders Telegram badge for telegram-sourced moments')
test('does not render badge for web-sourced moments')
```

### End-to-End Tests

**Manual Testing Checklist**:
- [ ] Generate link code from web app
- [ ] Link Telegram account via bot
- [ ] Create moment with all 7 fields
- [ ] Verify moment appears in web app with Telegram badge
- [ ] Test /cancel command mid-conversation
- [ ] Test invalid intensity input
- [ ] Test link code expiry (wait 30 minutes)
- [ ] Test disconnect flow
- [ ] Test mobile responsive design
- [ ] Test QR code scanning (desktop → mobile)

---

## Deployment Sequence

### Pre-Deployment Checklist

- [ ] Backend code merged to `main` branch
- [ ] Frontend code deployed to Vercel
- [ ] Database migration reviewed (version 1.04)
- [ ] Bot token obtained from @BotFather
- [ ] Environment variables documented

### Deployment Steps

#### 1. Create Telegram Bot (One-Time)

```bash
# Open Telegram, message @BotFather
/newbot
# Name: Rafiki Moments Bot
# Username: rafiki_moments_bot (must end in 'bot')

# Save the bot token: 1234567890:ABCdefGHIjklMNOpqrsTUVwxyz
```

#### 2. Configure Production Environment

```bash
# SSH to server
ssh root@178.156.170.37

# Edit .env file
nano /opt/rafiki/.env

# Add:
TELEGRAM_BOT_TOKEN=1234567890:ABCdefGHIjklMNOpqrsTUVwxyz
TELEGRAM_BOT_USERNAME=rafiki_moments_bot

# Save (Ctrl+X, Y, Enter)
exit
```

#### 3. Deploy Backend

```bash
# From local machine
make deploy

# Expected output:
# ✓ Pulling latest changes
# ✓ Building containers
# ✓ Running migrations (1.04 will be applied)
# ✓ Health checks passed
# ✓ Deployment complete
```

#### 4. Verify Telegram Bot Started

```bash
# Check logs
make deploy-logs | grep telegram

# Expected:
# {"level":"INFO","msg":"startup","status":"initializing telegram bot"}
# {"level":"INFO","msg":"telegram","status":"bot initialized","username":"rafiki_moments_bot"}
# {"level":"INFO","msg":"telegram","status":"polling started"}
```

#### 5. Test Linking Flow

```bash
# Open Rafiki web app → Settings → Telegram
# Click "Setup Telegram Integration"
# Copy link code
# Open Telegram → Search @rafiki_moments_bot
# Send: /link <code>
# Expected: "✅ Connected! You can now create moments here."
```

#### 6. Test Moment Creation

```bash
# In Telegram bot:
# Send: /moment
# Follow prompts to create a complete moment
# Verify moment appears in web app with Telegram badge
```

### Rollback Procedure

**Level 1: Disable Telegram** (30 seconds downtime)
```bash
ssh root@178.156.170.37
nano /opt/rafiki/.env
# Comment out: # TELEGRAM_BOT_TOKEN=...
docker compose restart partner-service
```

**Level 2: Code Rollback** (3 minutes downtime)
```bash
ssh root@178.156.170.37
cd /opt/rafiki
git log --oneline -5  # Find previous commit
git checkout <previous_sha>
./devops/deploy.sh
```

**Level 3: Database Rollback** (Use with extreme caution)
```sql
-- Only if migration corrupted data
ALTER TABLE users DROP COLUMN telegram_user_id;
ALTER TABLE moments DROP COLUMN source;
DROP TABLE telegram_conversations;
```

---

## Monitoring & Observability

### Logging

**What to log**:
```go
// Startup
log.Info(ctx, "telegram", "status", "bot initialized")

// User actions
log.Info(ctx, "telegram", "event", "link_code_generated", "user_id", userID)
log.Info(ctx, "telegram", "event", "account_linked", "telegram_user_id", tgUserID)
log.Info(ctx, "telegram", "event", "moment_created", "moment_id", momentID)

// Errors
log.Error(ctx, "telegram", "error", err, "action", "parse_intensity")
```

### Metrics (expvar)

```go
telegramMetrics.Add("messages_received", 1)
telegramMetrics.Add("moments_created", 1)
telegramMetrics.Add("errors_parse", 1)
telegramMetrics.Add("active_conversations", delta)
```

### Grafana Dashboard (Future)

Panels:
- Telegram messages received (24h)
- Moments created via Telegram (7d)
- Active conversations (current)
- Parse errors (1h)
- Average conversation duration

### Tracing (Tempo)

```go
// Trace flow: Telegram message → moment creation
ctx, span := tracer.Start(ctx, "telegram.message")
span.SetAttributes(
    attribute.Int64("telegram.user_id", tgUserID),
    attribute.String("telegram.command", "/moment"),
)
// ... process message ...
moment, err := momentBus.Create(ctx, newMoment)  // Span propagates
```

---

## Risk Mitigation

### Technical Risks

| Risk | Impact | Mitigation |
|------|--------|------------|
| Telegram API downtime | High | Graceful degradation, queue messages locally |
| Database migration failure | Critical | Idempotent migrations, manual rollback script |
| Bot crashes partner-service | High | Isolate bot errors, restart loop protection |
| Memory leak in conversation state | Medium | TTL cleanup (5 minutes), max conversations limit |
| Rate limiting abuse | Medium | Rate limit link code generation (5/hour) |

### Operational Risks

| Risk | Impact | Mitigation |
|------|--------|------------|
| Bot token leak | Critical | Token in .env (not git), rotate immediately if leaked |
| User confusion (UX) | Medium | Clear instructions, example messages, help command |
| Frontend CORS issues | High | Test preflight requests, verify Vercel origin |
| Webhook registration fails | Medium | Fallback to polling, manual registration script |

### Deployment Risks

| Risk | Impact | Mitigation |
|------|--------|------------|
| Breaking existing API | Critical | No changes to existing endpoints, additive only |
| CPX11 resource exhaustion | Medium | Monitor RAM/CPU, set resource limits |
| Frontend deployment before backend | Medium | Optional fields (`source?`), backward compatible types |

---

## Questions for Product Owner

Before starting implementation, please clarify:

### 1. Bot Naming & Branding

**Question**: What should the Telegram bot be named?
- **Options**:
  - `@rafiki_moments_bot` (descriptive)
  - `@rafikibot` (simple)
  - `@rafiki_app_bot` (branded)
- **Impact**: Username cannot be changed after creation

**Your Answer**: _________________________

---

### 2. Conversation Cancellation

**Question**: Should users be able to cancel a moment mid-conversation?
- **Options**:
  - Yes, with `/cancel` command (clears partial data)
  - No, must complete or timeout (5 minutes)
- **Impact**: User experience, state management complexity

**Your Answer**: _________________________

---

### 3. Conversation Timeout

**Question**: How long should incomplete conversations persist?
- **Options**:
  - 5 minutes (strict, saves memory)
  - 30 minutes (relaxed, allows interruptions)
  - 24 hours (very relaxed, high memory usage)
- **Impact**: Memory usage, user experience

**Your Answer**: _________________________

---

### 4. Error Notification Strategy

**Question**: How should users be notified of import errors?
- **Options**:
  - Only in Telegram (bot sends error message)
  - Only in web app (Settings page shows failed imports)
  - Both (dual notification)
- **Impact**: Development effort, user awareness

**Your Answer**: _________________________

---

### 5. Historical Data Import

**Question**: Should we allow users to backdate moments via Telegram?
- **Options**:
  - Yes, ask "When did this happen?" (add date field)
  - No, always use current timestamp
- **Impact**: Conversation length, data accuracy

**Your Answer**: _________________________

---

### 6. Multi-Language Support

**Question**: Should the bot support multiple languages?
- **Options**:
  - English only (MVP)
  - Spanish + English (based on user preference)
  - Auto-detect from Telegram language settings
- **Impact**: Development time, maintenance

**Your Answer**: _________________________

---

### 7. Privacy & Data Retention

**Question**: How long should we keep Telegram conversation logs?
- **Options**:
  - Don't log messages (privacy-first)
  - Log for 7 days (debugging)
  - Log indefinitely (analytics)
- **Impact**: Privacy, debugging capability, storage

**Your Answer**: _________________________

---

### 8. Advanced Features (Future)

**Question**: Which advanced features should we prioritize for Phase 3?
- **Options** (rank 1-5):
  - [ ] Natural language parsing (AI-powered)
  - [ ] Voice message support (speech-to-text)
  - [ ] Photo upload (OCR for journal entries)
  - [ ] Quick templates (save common situations)
  - [ ] Reminders (daily check-in prompts)

**Your Rankings**: _________________________

---

## Success Metrics

### Phase 1 (MVP)

- [ ] Deployment completes without errors
- [ ] At least 5 beta users successfully link Telegram accounts
- [ ] At least 20 moments created via Telegram in first week
- [ ] Zero critical bugs in first 48 hours
- [ ] Average conversation completion rate >80%

### Phase 2 (Production)

- [ ] Webhook migration completes with <5 minutes downtime
- [ ] 50+ active Telegram users within 30 days
- [ ] Average message-to-moment latency <2 seconds
- [ ] Parse error rate <5%
- [ ] User satisfaction score >4/5 (survey)

### Phase 3 (Growth)

- [ ] 30% of active users use Telegram integration
- [ ] Telegram-sourced moments = 40% of total moments
- [ ] Average moments per Telegram user >10/month
- [ ] Feature request volume for Telegram <10/month
- [ ] Zero security incidents

---

## Appendix

### Glossary

- **Telegram User ID**: Unique numeric identifier for Telegram user (int64)
- **Chat ID**: Same as Telegram User ID for private chats
- **Link Code**: One-time code for linking Telegram to Rafiki account
- **Conversation State**: Current step in guided conversation (e.g., awaiting_thoughts)
- **Partial Data**: Accumulated moment fields during conversation (JSONB)
- **GetUpdates**: Telegram API method for long polling
- **SendMessage**: Telegram API method for sending messages to users
- **Webhook**: HTTPS endpoint for receiving Telegram updates (Phase 2)

### References

- [Telegram Bot API Documentation](https://core.telegram.org/bots/api)
- [Go Telegram Bot API Library](https://github.com/go-telegram-bot-api/telegram-bot-api)
- [Rafiki Backend Architecture](../CLAUDE.md)
- [Database Migration Guide](../business/sdk/migrate/README.md)
- [Frontend Implementation Examples](./COMPONENT_IMPLEMENTATION_EXAMPLES.md)

---

**Document Version**: 1.0
**Last Updated**: 2025-11-17
**Next Review**: After Phase 1 completion

---

🤖 *Generated with Claude Code - Multi-Mind Analysis*
