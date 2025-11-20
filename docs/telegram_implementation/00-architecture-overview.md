# Telegram Integration - Architecture Overview

**Last Updated**: 2025-11-20
**Status**: Architecture Approved
**Team**: Backend + Frontend + DevOps

---

## Executive Summary

This document describes the architecture for Telegram integration in Rafiki. The integration is designed as a **reusable foundation** that supports moments initially, and can be extended to habits, goals, and other features without major refactoring.

---

## Key Architecture Decisions

| Decision | Choice | Rationale |
|----------|--------|-----------|
| **Architecture Pattern** | Foundation-first (reusable) | Support multiple features (moments, habits, goals) without refactoring |
| **Service Integration** | Integrated in partner-service | Resource efficiency, operational simplicity for MVP |
| **Communication Mode** | Webhook (from day 1) | <1s latency, scales to multiple features, recommended by Telegram |
| **Database Schema** | Generic tables (telegram_users, conversation_states) | Feature-agnostic, supports moments/habits/goals |
| **Frontend Settings** | Unified Telegram connection | Link once, enable/disable features individually |
| **Plugin System** | Yes (from Phase 1) | Easy to add habits/goals without modifying foundation |
| **Bot Username** | @rafiki_bot | Generic name (not moment-specific) |
| **Conversation Timeout** | 5 minutes | Balance UX and memory usage |
| **Cancellation** | Not allowed | Simplify state management for MVP |
| **Error Notifications** | Web app only | Reduce notification fatigue |
| **Language** | English only | Simplify MVP |
| **Message Logging** | None (privacy-first) | GDPR compliance |

---

## System Architecture

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
│  │  - /v1/telegram/webhook → Partner-Service (webhook)    │ │
│  │  - /v1/telegram/* → Partner-Service (user endpoints)   │ │
│  └────────────────────────────────────────────────────────┘ │
│                           │                                  │
│                           ▼                                  │
│  ┌────────────────────────────────────────────────────────┐ │
│  │ Partner-Service (Port 3000)                            │ │
│  │                                                         │ │
│  │  ┌──────────────────────────────────────────────────┐ │ │
│  │  │ foundation/telegram (NEW)                        │ │ │
│  │  │  - Bot (lifecycle management)                    │ │ │
│  │  │  - Client (Telegram API wrapper)                 │ │ │
│  │  │  - Router (command routing)                      │ │ │
│  │  │  - ConversationManager (state machine)           │ │ │
│  │  │  - WebhookServer (handles Telegram updates)      │ │ │
│  │  └──────────────────────────────────────────────────┘ │ │
│  │                                                         │ │
│  │  ┌──────────────────────────────────────────────────┐ │ │
│  │  │ Plugins (Business Logic Handlers)                │ │ │
│  │  │  - MomentPlugin (/moment command)                │ │ │
│  │  │  - HabitPlugin (future - /habit command)         │ │ │
│  │  │  - GoalPlugin (future - /goal command)           │ │ │
│  │  └──────────────────────────────────────────────────┘ │ │
│  │                                                         │ │
│  │  ┌──────────────────────────────────────────────────┐ │ │
│  │  │ Business Layer                                    │ │ │
│  │  │  - momentbus (existing)                          │ │ │
│  │  │  - userbus (UPDATED - add LinkTelegram)         │ │ │
│  │  └──────────────────────────────────────────────────┘ │ │
│  └────────────────────────────────────────────────────────┘ │
│                           │                                  │
│                           ▼                                  │
│  ┌────────────────────────────────────────────────────────┐ │
│  │ PostgreSQL Database                                     │ │
│  │  - telegram_users (NEW - separate from users)          │ │
│  │  - user_telegram_links (NEW - linking table)           │ │
│  │  - conversation_states (NEW - generic state machine)   │ │
│  │  - moments (UPDATED - add source field)               │ │
│  └────────────────────────────────────────────────────────┘ │
��                                                               │
└─────────────────────────────────────────────────────────────┘
                           ▲
                           │ Webhook (HTTPS POST)
                           │
                   ┌───────┴────────┐
                   │ Telegram API   │
                   │ (Bot API)      │
                   └────────────────┘
```

---

## Directory Structure

```
/Users/francowini/Documents/rafiki/
├── foundation/
│   └── telegram/              # NEW - Reusable Telegram infrastructure
│       ├── bot.go             # Bot lifecycle management
│       ├── client.go          # Telegram API client wrapper
│       ├── router.go          # Command routing
│       ├── conversation.go    # Generic conversation state machine
│       ├── webhook.go         # Webhook server
│       ├── plugin.go          # Plugin interface
│       ├── errors.go          # Telegram-specific errors
│       └── types/
│           ├── chatid.go      # Validated ChatID type
│           ├── messageid.go   # Validated MessageID type
│           └── userid.go      # Validated TelegramUserID type
│
├── business/
│   ├── domain/
│   │   ├── momentbus/         # Existing
│   │   └── userbus/           # UPDATED - add Telegram linking methods
│   │
│   └── sdk/
│       ├── migrate/sql/
│       │   └── migrate.sql    # UPDATED - add version 1.04
│       └── sqldb/
│           └── telegramdb/    # NEW - Database implementation
│               ├── telegram_users.go
│               ├── conversation_states.go
│               └── filters.go
│
├── app/
│   └── domain/
│       ├── momentapp/         # Existing
│       └── telegramapp/       # NEW - HTTP handlers for Telegram endpoints
│           ├── telegramapp.go
│           ├── model.go       # Request/response models
│           └── route.go
│
├── api/
│   └── services/
│       └── partners/
│           ├── main.go        # UPDATED - initialize Telegram bot
│           ├── mux/
│           │   └── mux.go     # UPDATED - register Telegram routes
│           └── plugins/       # NEW - Business-specific Telegram handlers
│               ├── moment.go  # Moment conversation plugin
│               ├── link.go    # Account linking plugin
│               └── help.go    # Help command plugin
│
└── frontend/
    ├── app/
    │   └── (dashboard)/
    │       ├── settings/
    │       │   └── page.tsx   # UPDATED - add Telegram section
    │       └── momentos/
    │           └── page.tsx   # UPDATED - show Telegram badge
    │
    ├── components/
    │   ├── integrations/
    │   │   └── telegram/      # NEW - Generic Telegram components
    │   │       ├── TelegramConnectionCard.tsx
    │   │       ├── TelegramLinkingFlow.tsx
    │   │       ├── TelegramStatusBadge.tsx
    │   │       └── hooks/
    │   │           ├── useTelegramStatus.ts
    │   │           └── useTelegramLinking.ts
    │   │
    │   └── ui/
    │       └── source-badge.tsx  # NEW - Generic source indicator
    │
    └── lib/
        ├── api.ts             # UPDATED - add telegram namespace
        ├── types.ts           # UPDATED - add Telegram types
        └── telegram-context.tsx  # NEW - Global Telegram state
```

---

## Data Flow: Creating a Moment via Telegram

```
1. User opens Telegram → Sends "/moment"
2. Telegram servers → POST to https://api.rafiki.lat/v1/telegram/webhook
3. Nginx → Validates IP whitelist → Forwards to Partner-Service
4. WebhookServer → Parses update → Routes to Router
5. Router → Detects "/moment" → Calls MomentPlugin
6. MomentPlugin → Checks user linkage → Starts conversation
7. ConversationManager → Creates state in database (state="awaiting_situation")
8. MomentPlugin → Sends "Describe the situation:" via Client
9. Client → Calls Telegram API SendMessage
10. User → Replies "Had panic attack at work"
11. Telegram → POST to webhook with user reply
12. Router → Loads conversation state → Routes to MomentPlugin
13. MomentPlugin → Validates input → Advances state to "awaiting_thoughts"
14. ... (repeats for all 7 fields)
15. MomentPlugin → All fields collected → Calls momentBus.Create()
16. Database → Inserts moment with source="telegram"
17. MomentPlugin → Sends success message → Clears conversation state
18. Frontend → Polls /v1/moments → Detects new moment → Shows badge
```

**Latency**: <2 seconds end-to-end (webhook mode)

---

## Component Responsibilities

### Foundation Layer (foundation/telegram)

**Responsibilities**:
- ✅ Telegram API communication
- ✅ Webhook server implementation
- ✅ Command routing (generic)
- ✅ Conversation state management (generic)
- ✅ Bot lifecycle (start, stop, graceful shutdown)

**NOT Responsible For**:
- ❌ Business logic (moments, habits, goals)
- ❌ Database schema (only interfaces)
- ❌ HTTP API endpoints (app layer handles this)

### Plugin Layer (api/services/partners/plugins)

**Responsibilities**:
- ✅ Feature-specific conversation flows
- ✅ Validation of user input
- ✅ Calling business layer (momentBus, habitBus)
- ✅ Error messages to users

**NOT Responsible For**:
- ❌ Telegram API calls (foundation handles this)
- ❌ State persistence (ConversationManager handles this)
- ❌ HTTP endpoints (app layer handles this)

### Business Layer (business/domain/*)

**Responsibilities**:
- ✅ Moment creation (existing)
- ✅ User-Telegram linking (new method in userBus)
- ✅ Business validation (intensity 0-10, content non-empty)

**NOT Responsible For**:
- ❌ Telegram message parsing (plugin handles this)
- ❌ Conversation state (foundation handles this)

### App Layer (app/domain/telegramapp)

**Responsibilities**:
- ✅ HTTP endpoints (/v1/telegram/status, /v1/telegram/link, etc.)
- ✅ Request/response models
- ✅ Authentication (JWT)
- ✅ Rate limiting

**NOT Responsible For**:
- ❌ Telegram bot logic (foundation handles this)
- ❌ Conversation flows (plugins handle this)

### Frontend Layer

**Responsibilities**:
- ✅ Settings page (Telegram connection UI)
- ✅ Link code generation and display
- ✅ Source badges on moments
- ✅ API client for Telegram endpoints

**NOT Responsible For**:
- ❌ Direct Telegram API calls (backend handles this)
- ❌ Conversation logic (backend handles this)

---

## Security Model

### Authentication & Authorization

**User Linking Flow**:
1. User requests link code via `/v1/telegram/link` (requires JWT)
2. Backend generates random code (16 bytes), stores in database with expiry
3. User sends `/link CODE` in Telegram
4. Bot validates code, links `telegram_user_id` to `user_id`
5. Link code is single-use and expires after 5 minutes

**Webhook Security**:
- IP whitelist (Telegram servers: 149.154.160.0/20, 91.108.4.0/22)
- Secret token validation (X-Telegram-Bot-Api-Secret-Token header)
- SSL/TLS required (Let's Encrypt on api.rafiki.lat)
- Rate limiting (30 req/s)

**Bot Token Security**:
- Stored in `.env` file (chmod 600, not in git)
- Rotation every 90 days
- Sanitized in logs (never log full token)

### Privacy & GDPR

**Data Minimization**:
- ❌ DON'T store message content in logs
- ✅ DO store conversation state (for resume)
- ❌ DON'T store user's phone number or profile photo
- ✅ DO store telegram_user_id (required for linking)

**User Rights**:
- Right to disconnect (DELETE /v1/telegram/disconnect)
- Right to erasure (delete all Telegram data)
- Right to export (future - /v1/telegram/export)

**Data Retention**:
- Conversation states: 5 minutes (auto-cleanup if abandoned)
- Link codes: 5 minutes (auto-expire)
- User linkage: Until user disconnects
- Message logs: NONE (privacy-first)

---

## Scaling Plan

### Current Capacity (Integrated Architecture)

**Server**: Hetzner CPX11 (2 vCPU, 2GB RAM)

**Limits**:
- Max concurrent conversations: ~200
- Max daily active users: ~300
- Max messages/day: ~20,000 (webhook)
- Max features: 2-3 (moments + habits + maybe goals)

### Scaling Triggers

| Metric | Threshold | Action |
|--------|-----------|--------|
| Memory usage | >1.3GB | Extract telegram-service to separate container |
| CPU usage | >75% sustained | Optimize or extract |
| Active conversations | >100 concurrent | Extract + load balancing |
| Features | ≥3 (moments + habits + goals) | Consider extraction |

### Future Extraction Plan (Phase 3+)

If limits are reached, extract to separate `telegram-service`:

```yaml
# docker-compose.yml
telegram-service:
  build: ./telegram-service
  environment:
    - TELEGRAM_BOT_TOKEN=${TELEGRAM_BOT_TOKEN}
    - DATABASE_URL=${DATABASE_URL}
  depends_on:
    - postgres
```

**Benefits**:
- Independent scaling
- Isolation (bot crash ≠ API crash)
- Dedicated resources

**Effort**: ~2-3 days extraction work

---

## Success Metrics

### Phase 1 (MVP - Moments Only)

- [ ] ≥5 beta users link Telegram accounts
- [ ] ≥20 moments created via Telegram in first week
- [ ] Zero critical bugs in first 48 hours
- [ ] Average conversation completion rate >80%
- [ ] Average latency <2 seconds (webhook)

### Phase 2 (Habits - Future)

- [ ] Add habits plugin in <4 hours (proves reusability)
- [ ] Memory usage <1.2GB
- [ ] CPU usage <70%
- [ ] Zero refactoring of foundation layer needed

---

## Timeline

### Phase 1: Foundation + Moments (3-4 weeks)

**Week 1**: Backend Foundation
- [ ] Database migrations (1.04)
- [ ] `foundation/telegram` package
- [ ] Plugin interface

**Week 2**: Backend Business Logic
- [ ] MomentPlugin (conversation flow)
- [ ] UserBus Telegram methods (LinkTelegram, UnlinkTelegram)
- [ ] HTTP endpoints (telegramapp)

**Week 3**: Frontend
- [ ] Settings page with Telegram section
- [ ] Link code generation UI
- [ ] Source badges on moment cards
- [ ] API client integration

**Week 4**: Testing & Deployment
- [ ] End-to-end testing
- [ ] Deploy to production
- [ ] Beta user testing
- [ ] Monitor and fix bugs

---

## Dependencies

### Backend
- `github.com/go-telegram-bot-api/telegram-bot-api/v5` - Telegram SDK
- Existing: `slog`, `uuid`, `sqlx`, `conf`

### Frontend
- Existing: Next.js 14+, shadcn/ui, Tailwind CSS
- QR code library (for link code display)

### Infrastructure
- Telegram bot token (from @BotFather)
- Webhook endpoint (already have: api.rafiki.lat)
- SSL certificate (already have: Let's Encrypt)

---

## Risk Mitigation

| Risk | Impact | Mitigation |
|------|--------|------------|
| Bot token leak | CRITICAL | Store in .env (chmod 600), rotate every 90 days, sanitize logs |
| Webhook fails | HIGH | Implement retry logic, log failures, alerting |
| Database migration fails | CRITICAL | Idempotent migrations, test in staging, rollback plan |
| Memory leak | MEDIUM | Set timeout (5 min), cleanup abandoned conversations |
| User confusion | MEDIUM | Clear instructions, example messages, help command |

---

## Next Steps

1. ✅ Review and approve architecture
2. ⏭️ Begin implementation (see task documents 01-09)
3. ⏭️ Create bot via @BotFather
4. ⏭️ Set up development environment
5. ⏭️ Start with database migrations (01-backend-database.md)

---

**Document Status**: ✅ APPROVED
**Next Document**: [01-backend-foundation.md](./01-backend-foundation.md)
