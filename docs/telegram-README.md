# Telegram Integration - Documentation Index

**Quick Start**: New to this feature? Start here! 👇

---

## 📚 Documentation Map

### For Product/Design Review
**Start here to understand the user experience**

1. **[User Journey](./telegram-user-journey.md)** ⭐ **START HERE**
   - Visual step-by-step flows
   - Desktop vs mobile experiences
   - Error handling examples
   - Screenshots (ASCII art)
   - Time estimates for each step

   **Best for**: Product managers, designers, stakeholders

---

### For Implementation
**Start here to build the feature**

2. **[Implementation Plan](./telegram-integration-implementation-plan.md)** 🔧
   - Architecture diagrams
   - Complete API contracts
   - Database schema changes
   - Backend/Frontend/DevOps tasks
   - Deployment sequence
   - Risk mitigation
   - **8 critical questions to answer before coding**

   **Best for**: Developers, tech leads

---

## 🎯 Quick Navigation

### I want to understand...

| Topic | Go to |
|-------|-------|
| **How users will experience this** | [User Journey](./telegram-user-journey.md) → Journey 1-6 |
| **What we need to build** | [Implementation Plan](./telegram-integration-implementation-plan.md) → Backend/Frontend/DevOps sections |
| **How the bot conversation works** | [User Journey](./telegram-user-journey.md) → Journey 3 |
| **Database changes needed** | [Implementation Plan](./telegram-integration-implementation-plan.md) → Database Schema |
| **API endpoints to create** | [Implementation Plan](./telegram-integration-implementation-plan.md) → API Contracts |
| **How to deploy this** | [Implementation Plan](./telegram-integration-implementation-plan.md) → Deployment Sequence |
| **What questions to answer** | [Implementation Plan](./telegram-integration-implementation-plan.md) → Questions for Product Owner |
| **Time estimates** | [User Journey](./telegram-user-journey.md) → Summary table |
| **Mobile vs desktop UX** | [User Journey](./telegram-user-journey.md) → Journey 1 vs Journey 2 |
| **Error handling** | [User Journey](./telegram-user-journey.md) → Journey 5 |

---

## 🚀 Get Started (3 Steps)

### Step 1: Review User Experience (15 minutes)
```bash
Read: docs/telegram-user-journey.md
```
- See the complete user flow
- Understand desktop vs mobile differences
- Review error handling
- Confirm this matches your vision

### Step 2: Answer Critical Questions (10 minutes)
```bash
Read: docs/telegram-integration-implementation-plan.md
      → Section: "Questions for Product Owner"
```
- Bot username
- Conversation timeout
- Multi-language support
- Privacy settings
- Future features priority

### Step 3: Review Technical Plan (30 minutes)
```bash
Read: docs/telegram-integration-implementation-plan.md
      → All sections
```
- Architecture decision
- Database changes
- API contracts
- Deployment plan
- Approve or request changes

---

## 📋 Key Decisions Summary

### Architecture: INTEGRATED

**Decision**: Add Telegram bot directly into partner-service (not separate container)

**Why**:
- Resource efficient (CPX11 has limited RAM)
- Operational simplicity (one service to manage)
- Shared business logic (momentBus, userBus)
- Zero infrastructure changes for MVP

**Trade-offs**:
- ✅ Simpler deployment
- ✅ Lower memory usage (+30MB vs +50MB)
- ✅ Faster development
- ❌ Less isolation (bot crash = service crash)
- ❌ Harder to scale bot independently

---

### Communication: POLLING → WEBHOOK

**Phase 1 (MVP)**: Long polling
- Simple, works immediately
- No nginx config changes
- 30-second max latency

**Phase 2 (Production)**: Webhook
- Instant delivery
- Zero idle resource usage
- Recommended by Telegram

---

### Conversation UX: GUIDED 7-STEP

**Flow**:
1. Situation
2. Thoughts
3. Physical symptoms
4. Behavior
5. Consequences
6. Values reflection
7. Intensity (0-10)

**Why**:
- Best mobile UX (no complex formatting)
- Easy validation (one field at a time)
- Natural conversation flow
- Matches existing moment structure

**Alternative considered**: Structured single message (e.g., `/moment situation:... thoughts:...`)
- **Rejected**: Too complex for mobile typing

---

## ⏱️ Timeline

### Phase 1: MVP (2-3 weeks)
**Goal**: Basic working integration

**Backend**:
- [ ] Database migration (1 day)
- [ ] Telegram bot package (3-4 days)
- [ ] API endpoints (2 days)
- [ ] Testing (2 days)

**Frontend**:
- [ ] Settings page (2-3 days)
- [ ] Telegram badge on moments (1 day)
- [ ] API client updates (1 day)
- [ ] Testing (1 day)

**DevOps**:
- [ ] Environment variables (1 hour)
- [ ] Deployment testing (1 day)

### Phase 2: Production Polish (1-2 weeks)
- [ ] Webhook mode
- [ ] Enhanced error handling
- [ ] Import history UI
- [ ] Grafana monitoring

### Phase 3: Advanced Features (Future)
- [ ] Natural language parsing (AI)
- [ ] Voice message support
- [ ] Photo upload
- [ ] Quick templates

---

## 🎨 User Experience Summary

### Desktop User Flow
```
Web App → Generate QR code → Scan with phone → Telegram opens
→ Send /start → Linked! → Send /moment → Answer 7 questions
→ Moment created → View in web app (30s later)

Time: ~2-3 minutes setup, ~60 seconds per moment
```

### Mobile User Flow
```
Web App → Tap "Open Telegram" → Telegram opens (deep link)
→ Tap START → Linked! → Send /moment → Answer 7 questions
→ Moment created → View in web app (30s later)

Time: ~2 minutes setup, ~60 seconds per moment
```

---

## 💾 Data Changes

### Tables Modified
```sql
users:
  + telegram_user_id BIGINT
  + telegram_username TEXT
  + telegram_connected_at TIMESTAMP
  + telegram_link_code TEXT
  + telegram_link_code_expiry TIMESTAMP

moments:
  + source TEXT DEFAULT 'web'
  + telegram_message_id INTEGER
```

### New Table
```sql
telegram_conversations:
  - telegram_user_id (PK)
  - user_id (FK)
  - state (current conversation step)
  - partial_data (JSONB - accumulated answers)
  - last_activity (for cleanup)
```

---

## 🔌 API Endpoints

### New Endpoints
| Method | Endpoint | Purpose |
|--------|----------|---------|
| GET | `/v1/telegram/status` | Check if user linked |
| POST | `/v1/telegram/link-code` | Generate link code |
| DELETE | `/v1/telegram/disconnect` | Unlink account |
| GET | `/v1/telegram/imports` | Get import history |
| POST | `/v1/telegram/webhook` | Receive Telegram updates (Phase 2) |

### Modified Endpoints
- `GET /v1/moments` - Now returns `source` and `telegram_message_id` fields

---

## 🛡️ Security

### Link Code Security
- Cryptographically random (16 bytes)
- 30-minute expiry
- Single-use (marked as consumed)
- Rate limited (5 per hour per user)

### Bot Token Security
- Stored in `.env` file (not in git)
- Permissions: `chmod 600`
- Rotation: Update `.env`, restart service

### Webhook Security (Phase 2)
- IP whitelist (Telegram servers only)
- Secret token validation
- SSL/TLS required (already configured)

---

## 📊 Success Metrics

### MVP Success (Week 1)
- [ ] 5+ beta users link Telegram
- [ ] 20+ moments created via Telegram
- [ ] Zero critical bugs
- [ ] 80%+ conversation completion rate

### Production Success (Month 1)
- [ ] 50+ active Telegram users
- [ ] <2 second message-to-moment latency
- [ ] <5% parse error rate
- [ ] 4/5+ user satisfaction

### Growth Success (Month 3)
- [ ] 30% of users use Telegram
- [ ] 40% of moments from Telegram
- [ ] 10+ moments per Telegram user/month

---

## ❓ FAQ

### Q: Do existing moments work with this feature?
**A**: Yes! Existing moments are unaffected. They'll show as `source: 'web'`. Only new Telegram moments show the Telegram badge.

### Q: Can users create moments in BOTH web and Telegram?
**A**: Absolutely! That's the point. Use web for detailed reflection, Telegram for quick on-the-go logging.

### Q: What happens if the bot is down?
**A**: Users can still create moments via web app. Telegram is an optional add-on, not a replacement.

### Q: Can users edit Telegram moments in the web app?
**A**: Yes! Full editing capabilities. The source badge just shows where it was created, not where it can be edited.

### Q: What if a user loses their Telegram account?
**A**: They can disconnect in Settings and all their moments remain. Or they can link a new Telegram account.

### Q: Does this work on WhatsApp/Signal/other messengers?
**A**: Not yet. This is Telegram-only for MVP. Other platforms could be added later using the same architecture.

### Q: Is this secure?
**A**: Yes! Telegram uses end-to-end encryption, link codes expire, and the bot only has access to messages sent to it (not all Telegram messages).

### Q: Will this slow down the app?
**A**: No! The bot runs in a background goroutine. It adds ~30MB RAM and <5% CPU. Web app performance is unaffected.

---

## 🔗 External Resources

- [Telegram Bot API Documentation](https://core.telegram.org/bots/api)
- [Go Telegram Bot Library](https://github.com/go-telegram-bot-api/telegram-bot-api)
- [@BotFather](https://t.me/botfather) - Create your bot here

---

## 📝 Next Steps

### Before Coding
1. ✅ Read [User Journey](./telegram-user-journey.md)
2. ✅ Read [Implementation Plan](./telegram-integration-implementation-plan.md)
3. ⬜ Answer 8 critical questions (in Implementation Plan)
4. ⬜ Create Telegram bot via @BotFather
5. ⬜ Get backend, frontend, and DevOps team aligned

### Start Coding
1. ⬜ Backend: Database migration (Version 1.04)
2. ⬜ Backend: Telegram bot package
3. ⬜ Frontend: Settings page
4. ⬜ Frontend: Telegram badge on moments
5. ⬜ DevOps: Add environment variables
6. ⬜ Test: End-to-end flow
7. ⬜ Deploy: Production

---

## 📞 Questions?

If you have questions about:
- **User experience**: See [User Journey](./telegram-user-journey.md)
- **Technical implementation**: See [Implementation Plan](./telegram-integration-implementation-plan.md)
- **Specific code examples**: Check Implementation Plan → Backend/Frontend sections
- **Deployment process**: Check Implementation Plan → Deployment Sequence

---

**Last Updated**: 2025-11-17
**Version**: 1.0

🤖 *Generated with Claude Code - Multi-Mind Analysis*
