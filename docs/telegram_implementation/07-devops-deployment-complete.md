# DevOps & Deployment - Complete Guide

**Task Category**: DevOps
**Estimated Time**: 11-17 hours total
**Prerequisites**: All backend and frontend tasks completed
**Dependencies**: Telegram Bot Token

---

## Overview

This document combines all DevOps tasks:
1. Bot Creation & Configuration (1-2h)
2. Nginx Configuration (2-3h)
3. Deployment (4-6h)
4. Monitoring & Alerts (4-6h)

---

## Part 1: Bot Creation & Configuration (1-2h)

### Task 1.1: Create Bot via @BotFather

**Steps**:
1. Open Telegram
2. Search for **@BotFather**
3. Send `/newbot`
4. Follow prompts:
   - **Name**: Rafiki Bot
   - **Username**: rafiki_bot (must end with 'bot')
5. Save the bot token: `1234567890:ABCdefGHIjklMNOpqrsTUVwxyz`

**Set bot commands** (for Telegram UI):
```
/setcommands
/moment - Create a new moment
/help - Show help information
```

**Set bot description**:
```
/setdescription
I help you track difficult moments for emotional well-being.
```

### Task 1.2: Configure Environment Variables

**File**: `/opt/rafiki/.env` (on production server)

Add to the end:

```bash
# Telegram Bot Configuration
TELEGRAM_BOT_TOKEN=1234567890:ABCdefGHIjklMNOpqrsTUVwxyz
TELEGRAM_BOT_USERNAME=rafiki_bot
TELEGRAM_WEBHOOK_URL=https://api.rafiki.lat/v1/telegram/webhook
TELEGRAM_WEBHOOK_SECRET=<generate-random-32-char-string>
```

**Generate secret token**:
```bash
openssl rand -hex 16
```

**Set file permissions**:
```bash
chmod 600 /opt/rafiki/.env
chown root:root /opt/rafiki/.env
```

**Checklist Part 1**:
- [ ] Create bot via @BotFather
- [ ] Save bot token securely
- [ ] Set bot commands and description
- [ ] Add environment variables to `.env`
- [ ] Generate webhook secret token
- [ ] Set proper file permissions
- [ ] Verify token not committed to git

---

## Part 2: Nginx Configuration (2-3h)

### Task 2.1: Update Nginx Config

**File**: `nginx/nginx.conf`

Add webhook endpoint:

```nginx
# Telegram Bot Webhook (add to server block)
location /v1/telegram/webhook {
    # Rate limiting: 30 req/s with burst of 50
    limit_req zone=telegram_webhook burst=50 nodelay;

    # IP whitelist (Telegram servers only)
    allow 149.154.160.0/20;
    allow 91.108.4.0/22;
    deny all;

    # Custom logging
    access_log /var/log/nginx/telegram_webhook.log combined;
    error_log /var/log/nginx/telegram_error.log warn;

    proxy_pass http://partner-service:3000;
    proxy_set_header Host $host;
    proxy_set_header X-Real-IP $remote_addr;
    proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    proxy_set_header X-Forwarded-Proto $scheme;
    proxy_set_header X-Telegram-Bot-Api-Secret-Token $http_x_telegram_bot_api_secret_token;

    proxy_read_timeout 60s;
}

# Telegram User Endpoints (add after webhook)
location /v1/telegram/ {
    limit_req zone=api_limit burst=5 nodelay;

    proxy_pass http://partner-service:3000;
    proxy_set_header Host $host;
    proxy_set_header X-Real-IP $remote_addr;
    proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    proxy_set_header X-Forwarded-Proto $scheme;
}
```

Add rate limit zone (in http block):

```nginx
http {
    # ... existing config ...

    # Telegram webhook rate limiting
    limit_req_zone $binary_remote_addr zone=telegram_webhook:10m rate=30r/s;

    # ... rest of config ...
}
```

### Task 2.2: Verify Nginx Config

```bash
# Test config syntax
nginx -t

# Reload nginx (if syntax OK)
nginx -s reload

# Or via Docker
docker exec rafiki-nginx nginx -t
docker exec rafiki-nginx nginx -s reload
```

**Checklist Part 2**:
- [ ] Add webhook endpoint to nginx.conf
- [ ] Add user endpoints location block
- [ ] Add rate limit zone
- [ ] Test nginx config syntax
- [ ] Reload nginx (no restart needed)
- [ ] Verify webhook endpoint accessible (should return 403 from non-Telegram IP)

---

## Part 3: Deployment (4-6h)

### Task 3.1: Update Docker Compose

**File**: `docker-compose.yml`

Ensure partner-service has Telegram env vars:

```yaml
partner-service:
  environment:
    # ... existing vars ...

    # Telegram Bot
    - PARTNER_TELEGRAM_BOTTOKEN=${TELEGRAM_BOT_TOKEN:-}
    - PARTNER_TELEGRAM_BOTUSERNAME=${TELEGRAM_BOT_USERNAME:-}
    - PARTNER_TELEGRAM_WEBHOOKURL=${TELEGRAM_WEBHOOK_URL:-}
    - PARTNER_TELEGRAM_WEBHOOKSECRET=${TELEGRAM_WEBHOOK_SECRET:-}
```

### Task 3.2: Deploy Backend

**From local machine**:

```bash
# 1. Commit all changes
git add .
git commit -m "feat: add telegram integration"
git push origin main

# 2. Deploy to production
make deploy

# Expected output:
# ✓ Pulling latest changes
# ✓ Building containers
# ✓ Running migrations (1.04 will be applied)
# ✓ Health checks passed
# ✓ Deployment complete
```

**From server** (manual deployment):

```bash
ssh root@178.156.170.37
cd /opt/rafiki
git pull origin main
./devops/deploy.sh
```

### Task 3.3: Verify Deployment

**Check logs**:
```bash
make deploy-logs | grep telegram

# Expected output:
# {"level":"INFO","msg":"startup","status":"initializing telegram bot"}
# {"level":"INFO","msg":"telegram client created","username":"rafiki_bot"}
# {"level":"INFO","msg":"webhook set","url":"https://api.rafiki.lat/v1/telegram/webhook"}
# {"level":"INFO","msg":"bot starting"}
# {"level":"INFO","msg":"bot started in webhook mode"}
```

**Check webhook status**:
```bash
# Via Telegram API
curl https://api.telegram.org/bot<TOKEN>/getWebhookInfo

# Should return:
# {
#   "ok": true,
#   "result": {
#     "url": "https://api.rafiki.lat/v1/telegram/webhook",
#     "has_custom_certificate": false,
#     "pending_update_count": 0
#   }
# }
```

**Test health endpoint**:
```bash
curl https://api.rafiki.lat/v1/readiness

# Should return 200 OK
```

### Task 3.4: Deploy Frontend

Frontend automatically deploys via Vercel on push to main:

```bash
git push origin main
# Vercel auto-deploys within 2-3 minutes
```

Verify deployment:
- Visit https://app.rafiki.lat/settings
- Check if Telegram section appears
- Test link code generation

**Checklist Part 3**:
- [ ] Update docker-compose.yml with Telegram env vars
- [ ] Commit and push all changes
- [ ] Run `make deploy` or manual deployment
- [ ] Verify migrations ran (check logs for version 1.04)
- [ ] Verify bot initialized (check logs)
- [ ] Verify webhook set (check Telegram API)
- [ ] Test health endpoint
- [ ] Verify frontend deployed (Vercel)
- [ ] Test settings page loads

---

## Part 4: Monitoring & Alerts (4-6h)

### Task 4.1: Add Telegram Metrics

**File**: `api/services/partners/telegram/metrics.go`

```go
package telegram

import "expvar"

var telegramMetrics = expvar.NewMap("telegram")

func init() {
    telegramMetrics.Add("messages_received", 0)
    telegramMetrics.Add("messages_sent", 0)
    telegramMetrics.Add("conversations_started", 0)
    telegramMetrics.Add("conversations_completed", 0)
    telegramMetrics.Add("moments_created", 0)
    telegramMetrics.Add("errors_total", 0)
    telegramMetrics.Add("active_conversations", 0)
}

// Track metrics in bot handlers
func (b *Bot) handleUpdate(ctx context.Context, update tgbotapi.Update) {
    telegramMetrics.Add("messages_received", 1)

    // ... handle update ...

    if err != nil {
        telegramMetrics.Add("errors_total", 1)
    }
}
```

**Access metrics**:
```bash
curl http://178.156.170.37:3010/debug/vars | jq .telegram

# Returns:
# {
#   "messages_received": 42,
#   "conversations_started": 5,
#   "moments_created": 4,
#   ...
# }
```

### Task 4.2: Set Up Grafana Dashboard

**Create Telegram Dashboard**:

1. Open Grafana: http://178.156.170.37:3100
2. Create new dashboard: "Telegram Bot Metrics"
3. Add panels:

**Panel 1: Message Volume**
- Query: `telegram.messages_received`
- Type: Graph
- Time range: Last 24 hours

**Panel 2: Conversation Funnel**
- Query: `telegram.conversations_started` vs `telegram.conversations_completed`
- Type: Bar chart
- Calculation: Completion rate = completed / started * 100

**Panel 3: Active Conversations**
- Query: `telegram.active_conversations`
- Type: Gauge
- Threshold: Yellow at 50, Red at 100

**Panel 4: Error Rate**
- Query: `telegram.errors_total`
- Type: Graph
- Time range: Last 1 hour

### Task 4.3: Set Up Alerts

**Alert 1: Bot Not Responding**

```yaml
# Grafana alert config
name: "Telegram Bot Not Responding"
condition: "telegram.messages_received (5-min rate) == 0"
duration: "5m"
action: "Send email to on-call engineer"
```

**Alert 2: High Error Rate**

```yaml
name: "Telegram High Error Rate"
condition: "telegram.errors_total (1-min rate) > 10"
duration: "1m"
action: "Send Slack notification"
```

**Alert 3: Memory Warning**

```yaml
name: "Service Memory High"
condition: "container_memory_usage > 1.8GB"
duration: "5m"
action: "Send email + Slack"
```

### Task 4.4: Log Monitoring

**View Telegram logs in real-time**:
```bash
make deploy-logs | grep telegram | tail -f
```

**Search for errors**:
```bash
make deploy-logs | grep -i "telegram.*error"
```

**Count messages per hour**:
```bash
make deploy-logs | grep "telegram.*message_received" | \
  awk '{print $1}' | cut -d'T' -f2 | cut -d':' -f1 | \
  sort | uniq -c
```

**Checklist Part 4**:
- [ ] Add expvar metrics to bot handlers
- [ ] Create Grafana dashboard
- [ ] Add 4 key panels (messages, funnel, active, errors)
- [ ] Set up alerts for critical metrics
- [ ] Test metrics endpoint (`/debug/vars`)
- [ ] Verify dashboard updates in real-time
- [ ] Test alert notifications

---

## Part 5: End-to-End Testing (2-3h)

### Test 1: Account Linking

**Steps**:
1. ✅ Visit https://app.rafiki.lat/settings
2. ✅ Click "Setup Telegram Integration"
3. ✅ Scan QR code (or use deep link on mobile)
4. ✅ Telegram opens with @rafiki_bot
5. ✅ Bot sends welcome message with /link command
6. ✅ Send `/start <CODE>` in Telegram
7. ✅ Bot responds: "✅ Successfully linked!"
8. ✅ Return to web app
9. ✅ Settings page shows "Connected as @username"

### Test 2: Moment Creation

**Steps**:
1. ✅ Open Telegram, navigate to @rafiki_bot
2. ✅ Send `/moment`
3. ✅ Bot asks: "1/7: Describe the situation"
4. ✅ Reply: "Had panic attack at work"
5. ✅ Bot asks: "2/7: What were your thoughts?"
6. ✅ Reply: "Everyone is judging me"
7. ✅ Bot asks: "3/7: Physical symptoms?"
8. ✅ Reply: "Racing heart, tight chest"
9. ✅ Bot asks: "4/7: How did you behave?"
10. ✅ Reply: "Left the room"
11. ✅ Bot asks: "5/7: Consequences?"
12. ✅ Reply: "Missed important decision"
13. ✅ Bot asks: "6/7: Values reflection?"
14. ✅ Reply: "Struggled with professionalism vs self-care"
15. ✅ Bot asks: "7/7: Intensity (0-10)?"
16. ✅ Reply: "8"
17. ✅ Bot responds: "✅ Moment created!"
18. ✅ Open https://app.rafiki.lat/momentos
19. ✅ New moment appears with Telegram badge
20. ✅ Verify all 7 fields saved correctly

### Test 3: Error Handling

**Test 3a: Invalid Intensity**
1. ✅ Start `/moment` conversation
2. ✅ Answer first 6 questions
3. ✅ For intensity, reply: "very high" (invalid)
4. ✅ Bot responds: "❌ Please enter a number between 0 and 10"
5. ✅ Reply: "9" (valid)
6. ✅ Moment created successfully

**Test 3b: Conversation Timeout**
1. ✅ Start `/moment` conversation
2. ✅ Answer first 2 questions
3. ✅ Wait 6 minutes (timeout is 5 minutes)
4. ✅ Try to continue conversation
5. ✅ Bot responds: no active conversation (or starts fresh)

### Test 4: Disconnection

**Steps**:
1. ✅ Visit https://app.rafiki.lat/settings
2. ✅ Click "Disconnect" button
3. ✅ Confirm dialog
4. ✅ Status changes to "Not connected"
5. ✅ Try to send `/moment` in Telegram
6. ✅ Bot responds: "⚠️ Your Telegram account is not linked"

---

## Rollback Procedures

### Level 1: Disable Telegram (30 seconds)

```bash
ssh root@178.156.170.37
cd /opt/rafiki
# Comment out bot token
sed -i 's/^TELEGRAM_BOT_TOKEN/#TELEGRAM_BOT_TOKEN/' .env
docker compose restart partner-service
```

**Effect**: Bot stops, API remains operational

### Level 2: Code Rollback (3 minutes)

```bash
ssh root@178.156.170.37
cd /opt/rafiki
git log --oneline -5  # Find previous commit
git checkout <previous-sha>
./devops/deploy.sh
```

**Effect**: Full rollback to previous version

### Level 3: Database Rollback (DANGEROUS)

**File**: `scripts/rollback-telegram-migration.sql`

```sql
BEGIN;
DROP TABLE IF EXISTS conversation_states CASCADE;
DROP TABLE IF EXISTS telegram_link_codes CASCADE;
DROP TABLE IF EXISTS user_telegram_links CASCADE;
DROP TABLE IF EXISTS telegram_users CASCADE;
ALTER TABLE moments DROP COLUMN IF EXISTS source;
ALTER TABLE moments DROP COLUMN IF EXISTS source_metadata;
COMMIT;
```

**Run**:
```bash
psql -U rafiki -d rafiki < scripts/rollback-telegram-migration.sql
```

---

## Troubleshooting

### Bot Not Responding

**Symptoms**: Messages to @rafiki_bot go unanswered

**Diagnosis**:
```bash
# Check if bot is running
make deploy-logs | grep "bot started"

# Check webhook status
curl https://api.telegram.org/bot<TOKEN>/getWebhookInfo

# Check nginx logs
docker exec rafiki-nginx tail -f /var/log/nginx/telegram_webhook.log
```

**Solutions**:
1. Verify bot token in .env
2. Check webhook URL is correct
3. Verify nginx routing
4. Restart partner-service: `docker compose restart partner-service`

### Webhook 403 Forbidden

**Symptoms**: Telegram can't POST to webhook

**Diagnosis**:
```bash
# Check nginx logs
docker exec rafiki-nginx tail -f /var/log/nginx/telegram_error.log
```

**Solutions**:
1. Verify IP whitelist includes Telegram servers
2. Check SSL certificate is valid
3. Verify secret token matching

### Conversation State Lost

**Symptoms**: User's conversation disappears mid-flow

**Diagnosis**:
```sql
-- Check for abandoned conversations
SELECT * FROM conversation_states WHERE last_activity < NOW() - INTERVAL '5 minutes';
```

**Solutions**:
1. Check timeout setting (5 minutes)
2. Verify database connection stable
3. Check service restarts (disrupts conversations)

### Memory Exhaustion

**Symptoms**: Service restarts frequently, slow responses

**Diagnosis**:
```bash
# Check memory usage
docker stats rafiki-partner-service
```

**Solutions**:
1. Check active conversations: `SELECT COUNT(*) FROM conversation_states`
2. Run cleanup manually: Exec SQL `DELETE FROM conversation_states WHERE ...`
3. Consider extracting bot to separate service

---

**Status**: ⏭️ Ready for Deployment
**Total Time**: 11-17 hours
**Next**: [08-testing-checklist.md](./08-testing-checklist.md) - Comprehensive testing guide
