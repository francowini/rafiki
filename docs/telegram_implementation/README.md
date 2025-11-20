# Telegram Integration - Implementation Roadmap

**Last Updated**: 2025-11-20
**Project Status**: 📋 Planning → Implementation

---

## Quick Start

This folder contains the complete implementation plan for Telegram integration in Rafiki. Each document contains small, testable tasks that can be completed independently.

**For LLMs/Developers**: Follow the tasks in numerical order. Each document is self-contained with:
- Clear prerequisites
- Step-by-step instructions
- Test cases
- Checklists

---

## Implementation Order

### Phase 1: Backend Database & Foundation (Week 1)

| # | Document | Description | Est. Time | Status |
|---|----------|-------------|-----------|--------|
| 00 | [Architecture Overview](./00-architecture-overview.md) | System architecture, decisions, directory structure | - | ✅ Approved |
| 01 | [Database Migrations](./01-backend-database.md) | Create tables: telegram_users, user_telegram_links, conversation_states, link_codes | 4-6h | ⏭️ Next |
| 02 | [Database Access Layer](./02-backend-telegramdb.md) | Implement CRUD operations for all Telegram tables | 6-8h | ⏸️ Waiting |
| 03 | [Foundation Package](./03-backend-foundation.md) | Create `foundation/telegram`: Bot, Client, Router, ConversationManager | 12-14h | ⏸️ Waiting |
| 04 | [Business Layer - UserBus](./04-backend-userbus.md) | Add Telegram linking methods to userbus | 3-4h | ⏸️ Waiting |

### Phase 2: Backend Business Logic & API (Week 2)

| # | Document | Description | Est. Time | Status |
|---|----------|-------------|-----------|--------|
| 05 | [Moment Plugin](./05-backend-moment-plugin.md) | Implement moment conversation flow (7-step state machine) | 8-10h | ⏸️ Waiting |
| 06 | [Link & Help Plugins](./06-backend-other-plugins.md) | Implement `/link` and `/help` command handlers | 3-4h | ⏸️ Waiting |
| 07 | [HTTP API Endpoints](./07-backend-api.md) | Create telegramapp: /v1/telegram/status, /link, /disconnect | 6-8h | ⏸️ Waiting |
| 08 | [Main Service Integration](./08-backend-main-integration.md) | Initialize bot in main.go, register plugins, webhook setup | 4-6h | ⏸️ Waiting |

### Phase 3: Frontend (Week 3)

| # | Document | Description | Est. Time | Status |
|---|----------|-------------|-----------|--------|
| 09 | [Frontend Types & API Client](./09-frontend-types-api.md) | Add Telegram types to lib/types.ts, extend api.ts | 3-4h | ⏸️ Waiting |
| 10 | [Generic Telegram Components](./10-frontend-components.md) | TelegramConnectionCard, LinkingFlow, StatusBadge | 8-10h | ⏸️ Waiting |
| 11 | [Settings Page](./11-frontend-settings.md) | Add Telegram section to settings page | 4-6h | ⏸️ Waiting |
| 12 | [Moment Card Updates](./12-frontend-moment-updates.md) | Add source badges to moment cards | 2-3h | ⏸️ Waiting |

### Phase 4: DevOps & Deployment (Week 4)

| # | Document | Description | Est. Time | Status |
|---|----------|-------------|-----------|--------|
| 13 | [Bot Creation & Config](./13-devops-bot-setup.md) | Create bot via @BotFather, configure environment | 1-2h | ⏸️ Waiting |
| 14 | [Nginx Configuration](./14-devops-nginx.md) | Add webhook endpoint, rate limiting, SSL | 2-3h | ⏸️ Waiting |
| 15 | [Deployment & Testing](./15-devops-deployment.md) | Deploy to production, end-to-end testing | 4-6h | ⏸️ Waiting |
| 16 | [Monitoring & Alerts](./16-devops-monitoring.md) | Set up Grafana dashboard, alerts, metrics | 4-6h | ⏸️ Waiting |

### Appendix

| # | Document | Description |
|---|----------|-------------|
| 17 | [Testing Checklist](./17-testing-checklist.md) | Complete testing guide (unit, integration, E2E) |
| 18 | [Rollback Procedures](./18-rollback-procedures.md) | Emergency rollback steps for all components |
| 19 | [Troubleshooting Guide](./19-troubleshooting.md) | Common issues and solutions |

---

## Total Estimated Time

- **Backend**: 41-56 hours (5-7 days)
- **Frontend**: 17-23 hours (2-3 days)
- **DevOps**: 11-17 hours (1.5-2 days)
- **Testing**: 8-12 hours (1-1.5 days)

**Total**: 77-108 hours (9.5-13.5 days) ≈ **2-3 weeks for one full-time developer**

---

## Status Legend

| Icon | Status | Meaning |
|------|--------|---------|
| ✅ | Approved | Architecture/design approved, ready to implement |
| ⏭️ | Next | Ready to start, all prerequisites met |
| 🔄 | In Progress | Currently being worked on |
| ✔️ | Complete | Implemented and tested |
| ⏸️ | Waiting | Blocked by prerequisites |
| ❌ | Blocked | Issue preventing progress |

---

## How to Use This Guide

### For Developers

1. **Start with 00-architecture-overview.md** to understand the system
2. **Follow the order** (01 → 02 → 03...) - each document builds on previous ones
3. **Check prerequisites** before starting a task
4. **Update status** when you complete a task (edit this README)
5. **Run tests** after each task (checklist in each document)

### For Project Managers

- Track progress by updating the status column
- Estimated times are per-developer (assumes full-time focus)
- Documents can be assigned to different team members if dependencies allow:
  - Backend tasks (01-08) must be sequential
  - Frontend tasks (09-12) can start after 07 is complete
  - DevOps tasks (13-16) can start after 08 is complete

### For LLMs/AI Assistants

When asked to implement Telegram integration:

1. Read `00-architecture-overview.md` first
2. Find the next task with status "⏭️ Next"
3. Open that document and follow the implementation steps
4. After completing, update this README with status "✔️ Complete"
5. Move to the next task

**Example prompt for LLMs**:
```
I'm implementing Telegram integration for Rafiki. I've completed task 01 (database migrations).
What should I work on next? Please read the next task document and help me implement it.
```

---

## Key Architecture Decisions (Quick Reference)

From `00-architecture-overview.md`:

- **Foundation-first**: Build reusable `foundation/telegram` package
- **Webhook mode**: Use webhooks from day 1 (<1s latency)
- **Generic schema**: Separate `telegram_users` table, support multiple features
- **Integrated service**: Bot runs inside partner-service (MVP)
- **Plugin architecture**: Easy to add habits/goals without refactoring
- **No cancellation**: Conversations timeout after 5 minutes (no `/cancel`)
- **English only**: Simplified MVP (multi-language later)
- **Privacy-first**: Don't log message content

---

## Testing Strategy

Each document includes test cases, but here's the overall approach:

### Unit Tests
- Database layer: `business/sdk/sqldb/telegramdb/*_test.go`
- Foundation: `foundation/telegram/*_test.go`
- Plugins: `api/services/partners/plugins/*_test.go`

### Integration Tests
- API endpoints: `app/domain/telegramapp/*_test.go`
- Full bot flow: `api/services/partners/tests/telegram_integration_test.go`

### End-to-End Tests
- Manual testing checklist in `17-testing-checklist.md`
- Beta user testing (Phase 4)

**Coverage Goal**: >80% for all new code

---

## Common Issues & Solutions

### Database Migration Fails
- **Solution**: Check `01-backend-database.md` for rollback script
- **Prevention**: Test migration on local DB first

### Bot Not Responding
- **Solution**: Check `19-troubleshooting.md` → Bot Health section
- **Check**: Token valid? Webhook registered? Nginx routing correct?

### Conversation State Lost
- **Solution**: Check `last_activity` in `conversation_states` table
- **Likely Cause**: Timeout (5 minutes) or service restart

### Frontend Can't Connect
- **Solution**: Check CORS in nginx, verify JWT token valid
- **Check**: `curl -H "Authorization: Bearer TOKEN" https://api.rafiki.lat/v1/telegram/status`

---

## Quick Commands

```bash
# Run all backend tests
go test ./business/sdk/sqldb/telegramdb/... -v
go test ./foundation/telegram/... -v
go test ./app/domain/telegramapp/... -v

# Run frontend tests
cd frontend
npm run test

# Deploy to production
make deploy

# Check production logs
make deploy-logs | grep telegram

# Rollback if needed
ssh root@178.156.170.37
cd /opt/rafiki
git checkout HEAD~1
./devops/deploy.sh
```

---

## Contributing

When adding new features (habits, goals, etc.):

1. No database schema changes needed (already generic!)
2. Add new plugin in `api/services/partners/plugins/habit.go`
3. Register plugin in `main.go`
4. Add frontend components (follow existing patterns)
5. Update this README with new tasks if needed

**Time to add habits**: ~4-6 hours (proves reusability!)

---

## Questions?

- **Architecture questions**: See `00-architecture-overview.md`
- **Database questions**: See `01-backend-database.md` and `02-backend-telegramdb.md`
- **Deployment questions**: See `13-devops-bot-setup.md` through `16-devops-monitoring.md`
- **Testing questions**: See `17-testing-checklist.md`

---

**Last Updated**: 2025-11-20
**Next Review**: After Phase 1 completion
**Document Maintainer**: Update this file as tasks are completed
