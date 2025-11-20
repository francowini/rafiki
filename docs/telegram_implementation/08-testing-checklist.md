# Telegram Integration - Complete Testing Checklist

**Last Updated**: 2025-11-20
**Purpose**: Comprehensive testing guide for all Telegram integration components

---

## Testing Overview

This checklist covers:
- **Unit Tests**: Individual functions and components
- **Integration Tests**: Multiple components working together
- **E2E Tests**: Complete user flows
- **Performance Tests**: Load and stress testing
- **Security Tests**: Vulnerabilities and edge cases

**Coverage Goal**: >80% for all new code

---

## Phase 1: Backend Unit Tests

### Database Layer Tests (`business/sdk/sqldb/telegramdb/`)

- [ ] **TelegramUser CRUD**
  - [ ] Create Telegram user
  - [ ] Query Telegram user by ID
  - [ ] Update Telegram user (via UPSERT)
  - [ ] Delete Telegram user
  - [ ] Handle duplicate telegram_user_id

- [ ] **LinkCode CRUD**
  - [ ] Create link code
  - [ ] Query link code by code string
  - [ ] Consume link code
  - [ ] Delete expired link codes
  - [ ] Count recent link codes (rate limiting)
  - [ ] Handle duplicate link code (unlikely but possible)

- [ ] **ConversationState CRUD**
  - [ ] Create conversation state
  - [ ] Query by telegram_user_id
  - [ ] Update conversation state (advance step)
  - [ ] Delete conversation state
  - [ ] Delete abandoned conversations (>5 min old)
  - [ ] Handle multiple active conversations (should error)

- [ ] **UserTelegramLink CRUD**
  - [ ] Create user-telegram link
  - [ ] Query by user_id
  - [ ] Query by telegram_user_id
  - [ ] Delete link
  - [ ] Handle duplicate links (1:1 constraint)

**Run Tests**:
```bash
go test ./business/sdk/sqldb/telegramdb/... -v -cover
# Expected: All pass, >80% coverage
```

---

### Foundation Layer Tests (`foundation/telegram/`)

- [ ] **Types Validation**
  - [ ] ChatID: Accept valid IDs, reject zero
  - [ ] TelegramUserID: Accept positive, reject zero/negative
  - [ ] MessageID: Accept positive, reject zero/negative
  - [ ] String formatting for all types

- [ ] **Client Tests**
  - [ ] Create client with valid token
  - [ ] Create client with invalid token (should error)
  - [ ] Send message to valid chat
  - [ ] Set webhook with URL
  - [ ] Delete webhook
  - [ ] GetUpdates (polling mode)

- [ ] **Router Tests**
  - [ ] Register plugin
  - [ ] Route command to correct plugin
  - [ ] Route message to correct plugin
  - [ ] Handle unknown command (no crash)
  - [ ] List all registered plugins

- [ ] **ConversationManager Tests**
  - [ ] Start conversation
  - [ ] Get active conversation
  - [ ] Update conversation (advance step)
  - [ ] Complete conversation (delete)
  - [ ] Cleanup abandoned conversations
  - [ ] Handle no active conversation (return error)

- [ ] **Bot Tests**
  - [ ] Create bot with valid config
  - [ ] Start bot in polling mode
  - [ ] Start bot in webhook mode
  - [ ] Stop bot gracefully
  - [ ] Handle update
  - [ ] Register webhook on start

**Run Tests**:
```bash
go test ./foundation/telegram/... -v -cover
# Expected: All pass, >80% coverage
```

---

### Business Layer Tests (`business/domain/userbus/`)

- [ ] **GenerateLinkCode**
  - [ ] Generate valid link code
  - [ ] Link code has correct format (RAFIKI-xxx)
  - [ ] Link code expires in 5 minutes
  - [ ] Rate limit: max 5 codes per hour
  - [ ] Error if user already linked

- [ ] **LinkTelegramAccount**
  - [ ] Link with valid code
  - [ ] Error with invalid code
  - [ ] Error with expired code
  - [ ] Error with consumed code
  - [ ] Error if telegram_user_id already linked to different account
  - [ ] Idempotent (linking same account twice is OK)

- [ ] **QueryTelegramLinkStatus**
  - [ ] Return linked=true when linked
  - [ ] Return linked=false when not linked
  - [ ] Include telegram_user_id and username when linked

- [ ] **QueryUserIDByTelegramUserID**
  - [ ] Return correct user_id when linked
  - [ ] Error when not linked

- [ ] **UnlinkTelegramAccount**
  - [ ] Successfully unlink
  - [ ] Error when not linked
  - [ ] Verify link deleted from database

**Run Tests**:
```bash
go test ./business/domain/userbus/... -v -cover
# Expected: All pass, >80% coverage
```

---

### Plugin Tests (`api/services/partners/plugins/`)

- [ ] **MomentPlugin: HandleCommand**
  - [ ] Start conversation with /moment
  - [ ] Error if user not linked
  - [ ] Error if user already has active conversation
  - [ ] Send first question (situation)

- [ ] **MomentPlugin: Conversation Flow**
  - [ ] Handle all 7 steps in order
  - [ ] Validate each input
  - [ ] Advance state after each step
  - [ ] Create moment after final step
  - [ ] Delete conversation after completion

- [ ] **MomentPlugin: Validation**
  - [ ] Reject empty strings
  - [ ] Reject intensity <0 or >10
  - [ ] Reject non-numeric intensity
  - [ ] Stay on same step when validation fails

- [ ] **MomentPlugin: Error Handling**
  - [ ] Handle database errors gracefully
  - [ ] Handle moment creation failure
  - [ ] Log errors with context

**Run Tests**:
```bash
go test ./api/services/partners/plugins/... -v -cover
# Expected: All pass, >80% coverage
```

---

## Phase 2: Frontend Unit Tests

### Component Tests (Jest + React Testing Library)

- [ ] **SourceBadge Component**
  - [ ] Render "Telegram" badge for source="telegram"
  - [ ] Render nothing for source="web"
  - [ ] Different variants (full, compact, icon-only)

- [ ] **TelegramConnectionCard Component**
  - [ ] Show "Not connected" state initially
  - [ ] Show "Setup Integration" button when disconnected
  - [ ] Show connected status when linked
  - [ ] Show disconnect button when connected
  - [ ] Trigger linking flow on button click

- [ ] **TelegramLinkingFlow Component**
  - [ ] Generate link code on mount
  - [ ] Display QR code on desktop
  - [ ] Display deep link button on mobile
  - [ ] Show countdown timer
  - [ ] Timer decrements every second
  - [ ] Show error when code expired
  - [ ] Copy link code to clipboard

- [ ] **MomentCard Component**
  - [ ] Render Telegram badge for telegram-sourced moments
  - [ ] Render without badge for web-sourced moments
  - [ ] Show subtle blue border for Telegram moments

### Hook Tests

- [ ] **useTelegram Hook**
  - [ ] Fetch status on mount
  - [ ] Update status after linking
  - [ ] Update status after disconnecting
  - [ ] Handle API errors gracefully
  - [ ] isConnected derived correctly

**Run Tests**:
```bash
cd frontend
npm run test
# Expected: All pass, >80% coverage
```

---

## Phase 3: Integration Tests

### Backend Integration Tests

- [ ] **Link Code Flow**
  - [ ] Generate code → consume → verify linked
  - [ ] Generate code → wait 6 min → consume → error expired
  - [ ] Generate code → consume → try again → error already consumed

- [ ] **Moment Creation Flow**
  - [ ] Start conversation → answer 7 questions → verify moment created
  - [ ] Verify moment has source="telegram"
  - [ ] Verify conversation deleted after completion

- [ ] **Conversation Timeout**
  - [ ] Start conversation → wait 6 min → verify auto-deleted
  - [ ] Start conversation → wait 4 min → still active

**Run Tests**:
```bash
go test ./api/services/partners/tests/... -v
```

### Frontend Integration Tests

- [ ] **Settings Page Flow**
  - [ ] Navigate to /settings
  - [ ] Click "Setup Integration"
  - [ ] Linking flow displays
  - [ ] Cancel linking
  - [ ] Back to settings page

- [ ] **Moment List with Mixed Sources**
  - [ ] Fetch moments (mix of web and telegram)
  - [ ] Verify badges render correctly
  - [ ] Filter by source (future)

**Run Tests**:
```bash
cd frontend
npm run test:integration
```

---

## Phase 4: End-to-End Tests (Manual)

### E2E Test 1: Complete Linking Flow

**Prerequisites**: Clean database, no existing links

1. [ ] Open https://app.rafiki.lat
2. [ ] Log in with test account
3. [ ] Navigate to Settings page
4. [ ] Verify "Telegram Bot" section visible
5. [ ] Verify status shows "Not connected"
6. [ ] Click "Setup Telegram Integration"
7. [ ] Verify link code displayed (format: RAFIKI-xxxx)
8. [ ] Verify QR code displayed (desktop) OR deep link button (mobile)
9. [ ] Verify countdown timer shows ~5:00
10. [ ] Open Telegram on phone
11. [ ] Search for @rafiki_bot
12. [ ] Bot sends welcome message
13. [ ] Send `/start <CODE>` or scan QR code
14. [ ] Verify bot responds: "✅ Successfully linked!"
15. [ ] Return to web app
16. [ ] Click "I've Completed Setup"
17. [ ] Verify status shows "Connected as @username"
18. [ ] Verify connection timestamp displayed
19. [ ] Verify "Disconnect" button visible

**Expected Result**: ✅ Full linking flow completes successfully

---

### E2E Test 2: Create Moment via Telegram

**Prerequisites**: Telegram account linked

1. [ ] Open Telegram
2. [ ] Navigate to @rafiki_bot
3. [ ] Send `/moment`
4. [ ] Verify bot asks: "1/7: Describe the situation"
5. [ ] Reply: "Had panic attack during work presentation"
6. [ ] Verify bot asks: "2/7: What were your thoughts?"
7. [ ] Reply: "Everyone thinks I'm incompetent"
8. [ ] Verify bot asks: "3/7: Physical symptoms?"
9. [ ] Reply: "Racing heart, sweating, shaking hands"
10. [ ] Verify bot asks: "4/7: How did you behave?"
11. [ ] Reply: "Rushed through slides, avoided eye contact"
12. [ ] Verify bot asks: "5/7: Consequences?"
13. [ ] Reply: "Presentation went poorly, boss disappointed"
14. [ ] Verify bot asks: "6/7: Values reflection?"
15. [ ] Reply: "Value competence but this moved me away from it"
16. [ ] Verify bot asks: "7/7: Intensity (0-10)?"
17. [ ] Reply: "8"
18. [ ] Verify bot responds: "✅ Moment created successfully!"
19. [ ] Verify bot shows intensity and timestamp
20. [ ] Open web app: https://app.rafiki.lat/momentos
21. [ ] Verify new moment appears at top of list
22. [ ] Verify moment has Telegram badge
23. [ ] Verify moment has subtle blue left border
24. [ ] Click "View Details"
25. [ ] Verify all 7 fields saved correctly
26. [ ] Verify moment shows "From Telegram" in header

**Expected Result**: ✅ Complete moment creation flow works end-to-end

---

### E2E Test 3: Error Handling

**Test 3a: Invalid Intensity**

1. [ ] Start `/moment` conversation
2. [ ] Answer first 6 questions correctly
3. [ ] For intensity, reply: "very intense" (invalid)
4. [ ] Verify bot responds: "❌ Please enter a number between 0 and 10"
5. [ ] Verify conversation NOT completed
6. [ ] Reply: "15" (out of range)
7. [ ] Verify bot responds: "❌ Intensity must be between 0 and 10"
8. [ ] Reply: "7" (valid)
9. [ ] Verify moment created successfully

**Test 3b: Conversation Timeout**

1. [ ] Send `/moment`
2. [ ] Answer first 2 questions
3. [ ] Wait 6 minutes (timeout is 5 minutes)
4. [ ] Try to continue: Send answer to question 3
5. [ ] Verify bot doesn't respond OR says "no active conversation"
6. [ ] Send `/moment` again
7. [ ] Verify starts fresh from question 1

**Test 3c: User Not Linked**

1. [ ] Disconnect Telegram account (via web app)
2. [ ] In Telegram, send `/moment`
3. [ ] Verify bot responds: "⚠️ Your Telegram account is not linked"
4. [ ] Verify bot provides link to settings page

---

### E2E Test 4: Disconnection Flow

**Prerequisites**: Telegram account linked

1. [ ] Open https://app.rafiki.lat/settings
2. [ ] Verify status shows "Connected"
3. [ ] Click "Disconnect" button
4. [ ] Verify confirmation dialog appears
5. [ ] Click "Yes, Disconnect"
6. [ ] Verify status changes to "Not connected"
7. [ ] Verify previous Telegram moments still visible (with badges)
8. [ ] In Telegram, send `/moment`
9. [ ] Verify bot responds: "⚠️ Your Telegram account is not linked"
10. [ ] Re-link account (generate new code)
11. [ ] Verify can create moments again

**Expected Result**: ✅ Disconnection works, data preserved, can re-link

---

## Phase 5: Performance Tests

### Load Test: Concurrent Conversations

**Objective**: Verify bot handles 50 concurrent conversations

**Setup**:
```bash
# Use a load testing tool (e.g., k6, locust)
# Simulate 50 users starting conversations simultaneously
```

**Scenarios**:
1. [ ] 50 users send `/moment` within 10 seconds
2. [ ] Monitor memory usage (should stay <1.5GB)
3. [ ] Monitor CPU usage (should stay <80%)
4. [ ] Verify all conversations started successfully
5. [ ] Verify no conversations lost/corrupted

**Success Criteria**:
- ✅ All 50 conversations started
- ✅ Memory usage <1.5GB
- ✅ CPU usage <80%
- ✅ Average response time <2 seconds

---

### Stress Test: Message Volume

**Objective**: Verify bot handles high message volume

**Setup**:
```bash
# Simulate 100 messages per second for 1 minute
```

**Scenarios**:
1. [ ] Send 6,000 messages in 1 minute
2. [ ] Monitor rate limiting (should throttle at 30 req/s)
3. [ ] Monitor error rate (should be <1%)
4. [ ] Verify no crashes or restarts

**Success Criteria**:
- ✅ Bot stays responsive
- ✅ No crashes or errors
- ✅ Rate limiting works correctly
- ✅ Memory doesn't leak

---

## Phase 6: Security Tests

### Security Test 1: Link Code Security

- [ ] **Randomness**: Generate 1000 link codes, verify no duplicates
- [ ] **Expiry**: Create code, wait 6 min, try to use → should fail
- [ ] **Single-use**: Consume code, try again → should fail
- [ ] **Rate limiting**: Generate 6 codes in 1 minute → 6th should fail

### Security Test 2: Webhook Security

- [ ] **IP Whitelist**: POST to webhook from non-Telegram IP → should get 403
- [ ] **Secret Token**: POST without secret header → should get 403
- [ ] **Secret Token**: POST with wrong secret → should get 403
- [ ] **Rate Limiting**: Send 100 req/s → should throttle at 30 req/s

### Security Test 3: Input Validation

- [ ] **XSS**: Send `<script>alert('xss')</script>` as situation → should be escaped
- [ ] **SQL Injection**: Send `'; DROP TABLE moments; --` → should be safe
- [ ] **Long Input**: Send 10,000 character string → should be truncated or rejected
- [ ] **Unicode**: Send emojis and special characters → should be saved correctly

---

## Phase 7: Regression Tests

### Verify No Regressions in Existing Features

- [ ] **Moments (Web)**
  - [ ] Create moment via web form
  - [ ] Edit existing moment
  - [ ] Delete moment
  - [ ] List moments (pagination works)

- [ ] **Authentication**
  - [ ] Login still works
  - [ ] Logout still works
  - [ ] JWT tokens still valid

- [ ] **API**
  - [ ] All existing endpoints return 200
  - [ ] CORS still configured correctly

---

## Test Summary Report Template

After completing all tests, fill out this report:

```markdown
# Telegram Integration - Test Report

**Date**: ___________
**Tester**: ___________
**Environment**: Production / Staging

## Test Results

| Phase | Tests Run | Passed | Failed | Coverage |
|-------|-----------|--------|--------|----------|
| Backend Unit | ___ | ___ | ___ | ___% |
| Frontend Unit | ___ | ___ | ___ | ___% |
| Integration | ___ | ___ | ___ | N/A |
| E2E | ___ | ___ | ___ | N/A |
| Performance | ___ | ___ | ___ | N/A |
| Security | ___ | ___ | ___ | N/A |
| Regression | ___ | ___ | ___ | N/A |

## Critical Issues Found

1. ___________
2. ___________
3. ___________

## Non-Critical Issues Found

1. ___________
2. ___________

## Performance Metrics

- Average response time: ___ ms
- Peak memory usage: ___ MB
- Peak CPU usage: ___%
- Max concurrent conversations: ___

## Recommendations

1. ___________
2. ___________

## Sign-Off

✅ Approved for production deployment: Yes / No

**Approver**: ___________
**Date**: ___________
```

---

**Status**: ⏭️ Ready for Testing
**Total Estimated Time**: 8-12 hours for complete testing
**Next**: Production deployment after all tests pass!
