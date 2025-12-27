# Deployment Guide: Deliverable 5a + 5b
# Telegram Moment Tracking - Job Worker Core + AI Integration

**Version:** 1.0
**Date:** 2025-12-27
**Scope:** Core job worker with Anthropic AI integration (NO moment creation)

---

## Table of Contents

1. [Pre-Deployment Checklist](#pre-deployment-checklist)
2. [Deployment Steps](#deployment-steps)
3. [Post-Deployment Validation](#post-deployment-validation)
4. [Monitoring Setup (Logs Only)](#monitoring-setup-logs-only)
5. [Rollback Plan](#rollback-plan)
6. [Rate Limit Management](#rate-limit-management)

---

## Pre-Deployment Checklist

### 1. Anthropic API Key Setup

**Required Tier:** Build Tier 1
- Rate Limit: 50 requests/minute
- Daily Limit: 2,500 messages/day
- Cost: ~$0.01 per conversation (6 steps x ~500 tokens)

**Get API Key:**
1. Go to https://console.anthropic.com/
2. Navigate to **API Keys** → **Create Key**
3. Copy the key (starts with `sk-ant-api03-...`)
4. Save securely (you won't see it again)

### 2. Environment Variables Needed

Add to `/opt/rafiki/.env` on production server:

```bash
# Anthropic Configuration (REQUIRED)
PARTNER_ANTHROPIC_APIKEY=sk-ant-api03-xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
PARTNER_ANTHROPIC_MODEL=claude-sonnet-4-5-20251029
PARTNER_ANTHROPIC_MAXTOKENS=500
PARTNER_ANTHROPIC_TEMPERATURE=0.7

# Telegram Configuration (ALREADY PRESENT from Deliverable 4)
PARTNER_TELEGRAM_BOTTOKEN=<existing_token>
PARTNER_TELEGRAM_WEBHOOKSECRET=<existing_secret>
```

**Verification Commands:**
```bash
# On server: Check .env file
cat /opt/rafiki/.env | grep ANTHROPIC

# Expected output:
# PARTNER_ANTHROPIC_APIKEY=sk-ant-api03-...
# PARTNER_ANTHROPIC_MODEL=claude-sonnet-4-5-20251029
```

### 3. Code Changes Required

**Files to be deployed (already implemented):**
- `app/jobs/telegrammessage/telegrammessage.go` - Worker registration
- `app/jobs/telegrammessage/handler.go` - Message processing
- `app/jobs/telegrammessage/prompts.go` - Prompt loading
- `app/jobs/telegrammessage/anthropic.go` - AI integration
- `app/jobs/telegrammessage/prompts/*.yaml` - 6 YAML prompt files
- `api/services/partners/main.go` - Worker registration in main service

**Database Changes:**
- NONE (telegram_sessions table already exists from Deliverable 3)

**Configuration Changes:**
- Add Anthropic config struct to `main.go` (already done)
- Initialize Anthropic client (already done)
- Register telegrammessage worker (already done)

---

## Deployment Steps

### Step 1: Verify Code is Ready

```bash
# On local machine
cd /Users/francowini/Documents/rafiki

# Check that telegrammessage job exists
ls -la app/jobs/telegrammessage/

# Expected files:
# - telegrammessage.go
# - handler.go
# - prompts.go
# - anthropic.go
# - prompts/ (directory with 6 YAML files)

# Verify main.go has Anthropic config
grep -A 5 "Anthropic struct" api/services/partners/main.go

# Build locally to verify no compilation errors
make build
```

### Step 2: Update Production Environment

**Option A: Via SSH (Manual)**

```bash
# SSH to production server
ssh root@178.156.170.37

# Navigate to project directory
cd /opt/rafiki

# Edit .env file
nano .env
```

Add these lines:
```bash
PARTNER_ANTHROPIC_APIKEY=sk-ant-api03-YOUR_KEY_HERE
PARTNER_ANTHROPIC_MODEL=claude-sonnet-4-5-20251029
PARTNER_ANTHROPIC_MAXTOKENS=500
PARTNER_ANTHROPIC_TEMPERATURE=0.7
```

Save and exit (Ctrl+X, Y, Enter)

**Option B: Using scp (Recommended)**

```bash
# On local machine - create temporary env snippet
cat > /tmp/anthropic_env << 'EOF'
PARTNER_ANTHROPIC_APIKEY=sk-ant-api03-YOUR_KEY_HERE
PARTNER_ANTHROPIC_MODEL=claude-sonnet-4-5-20251029
PARTNER_ANTHROPIC_MAXTOKENS=500
PARTNER_ANTHROPIC_TEMPERATURE=0.7
EOF

# SSH and append to .env
ssh root@178.156.170.37 "cat >> /opt/rafiki/.env" < /tmp/anthropic_env

# Verify
ssh root@178.156.170.37 "grep ANTHROPIC /opt/rafiki/.env"

# Clean up local temp file
rm /tmp/anthropic_env
```

### Step 3: Deploy Code

**From Local Machine (Recommended):**

```bash
# Ensure code is committed and pushed
git add .
git commit -m "feat: add telegrammessage job worker (Deliverable 5a + 5b)"
git push origin main

# Deploy to production (ONE COMMAND)
make deploy
```

**On Production Server (Alternative):**

```bash
# SSH to server
ssh root@178.156.170.37

# Navigate to project
cd /opt/rafiki

# Pull latest code
git pull origin main

# Run deployment script
sudo ./devops/deploy.sh
```

### Step 4: Verify Deployment

```bash
# Check service started successfully
docker compose ps partner-service

# Expected: STATUS = Up (healthy)

# Check logs for Anthropic client initialization
docker compose logs partner-service | grep -i anthropic

# Expected output:
# "anthropic client initialized"
# "registered worker" kind="telegrammessage"
```

**Deployment Time Estimate:** 3-5 minutes

---

## Post-Deployment Validation

### 1. Health Checks

```bash
# Basic service health
curl https://api.rafiki.lat/v1/readiness
# Expected: {"status":"ok"}

# Check job queue is running
docker compose logs partner-service | grep "job queue started"
# Expected: "job queue started" workers=3 queues=1
```

### 2. Verify Job Worker Registration

```bash
# SSH to production
ssh root@178.156.170.37

# Check worker logs
docker compose logs -f partner-service | grep telegrammessage

# Expected logs on startup:
# "registered worker" kind="telegrammessage"
# "job queue started" workers=3
```

### 3. Manual Testing (End-to-End)

**Prerequisites:**
- Test user linked to Telegram (telegram_chat_id set in database)
- Telegram bot token configured
- Webhook already registered (from Deliverable 4)

**Test Steps:**

```bash
# 1. Start conversation in Telegram
# Send to bot: /momento

# Expected response:
# "*Paso 1 de 6* 📝
#
# Contame qué pasó. Describí la situación..."

# 2. Send a test response
# "Estaba en casa solo, pensando que no estoy haciendo nada productivo"

# 3. Check logs for job processing
docker compose logs -f partner-service | grep telegrammessage

# Expected logs:
# "processing_telegram_message" session_id=xxx step=1
# "calling_anthropic_api" step=step_1
# "ai_response_received" is_valid=true
# "session_advanced" step=2
# "telegram_message_sent" chat_id=xxx
```

**Expected Telegram Response:**
```
✅ Perfecto. Sigamos.

*Paso 2 de 6* 🧠

Ahora contame qué sentiste físicamente...
```

### 4. Verify AI Integration

```bash
# Check Anthropic API calls in logs
docker compose logs partner-service | grep -E "anthropic|ai_response"

# Expected patterns:
# "calling_anthropic_api" step=step_1 user_message="..."
# "ai_response_received" is_valid=true tokens_in=X tokens_out=Y
# "session_data_updated" step=1 field=situation
```

### 5. Database Verification

```bash
# Connect to database
docker exec -it rafiki-postgres psql -U rafiki -d rafiki

# Check active sessions
SELECT session_id, current_step, retry_count,
       parsed_data->>'situation' as situation,
       last_activity
FROM telegram_sessions
ORDER BY last_activity DESC
LIMIT 5;

# Expected: Session with current_step advancing (1 → 2 → 3...)

# Check job queue
SELECT id, kind, state, attempt, max_attempts,
       args->>'session_id' as session_id
FROM river_job
WHERE kind = 'telegrammessage'
ORDER BY created_at DESC
LIMIT 10;

# Expected: Jobs in 'completed' state (or 'running' if processing)
```

### 6. Error Scenario Testing

**Test Invalid Response (Retry Logic):**

```bash
# 1. Start new session: /momento

# 2. Send gibberish response
# "asdf"

# 3. Check logs for retry
docker compose logs partner-service | grep retry_count

# Expected:
# "ai_validation_failed" retry_count=1
# "session_retry_incremented" retry_count=1

# 4. Expected Telegram response:
# "Entiendo, pero necesito que me cuentes un poco más sobre..."
```

**Test Max Retries:**

```bash
# Send invalid responses 3 times in a row

# Expected after 2nd retry:
# "Parece que te cuesta responder esta pregunta.
#  ¿Querés pasar a la siguiente? Enviá cualquier mensaje para continuar."
```

### 7. Success Criteria Checklist

- [ ] Service starts without errors
- [ ] Anthropic client initialized (log message present)
- [ ] telegrammessage worker registered
- [ ] Test session advances through Step 1 → Step 2
- [ ] AI response parsing works (is_valid=true)
- [ ] Retry logic works (invalid response → retry_count++)
- [ ] Max retries handled (2 retries → ask to continue)
- [ ] Session data stored in parsed_data JSONB
- [ ] No memory leaks or crashes after 10 test messages

---

## Monitoring Setup (Logs Only)

### Key Log Patterns to Monitor

**1. Job Processing:**
```bash
# Real-time job monitoring
docker compose logs -f partner-service | grep telegrammessage

# Key patterns:
# "processing_telegram_message"  - Job started
# "session_loaded"               - Session fetched
# "calling_anthropic_api"        - AI request sent
# "ai_response_received"         - AI response parsed
# "session_advanced"             - Step completed
# "telegram_message_sent"        - Reply sent
```

**2. Error Monitoring:**
```bash
# Monitor errors
docker compose logs -f partner-service | grep -i error

# Critical errors:
# "anthropic_api_error"          - API failure (check rate limits)
# "session_not_found"            - Session expired or deleted
# "malformed_json_response"      - AI response parsing failed
# "telegram_send_failed"         - Message delivery failed
```

**3. Rate Limit Monitoring:**
```bash
# Check for rate limit errors
docker compose logs partner-service | grep -E "rate_limit|429"

# Expected during rate limit:
# "anthropic_api_error" status_code=429 retryable=true
# Job will retry automatically (River exponential backoff)
```

**4. Health Check Queries (SSH to server):**

```bash
# Job queue health (last 1 hour)
docker exec -it rafiki-postgres psql -U rafiki -d rafiki -c "
SELECT
    state,
    COUNT(*) as count,
    AVG(EXTRACT(EPOCH FROM (finalized_at - attempted_at))) as avg_duration_sec
FROM river_job
WHERE kind = 'telegrammessage'
  AND created_at > NOW() - INTERVAL '1 hour'
GROUP BY state;
"

# Expected states:
# completed | 45 | 2.3
# running   |  2 | NULL
# failed    |  0 | NULL

# Active sessions
docker exec -it rafiki-postgres psql -U rafiki -d rafiki -c "
SELECT
    current_step,
    COUNT(*) as count,
    MAX(last_activity) as most_recent
FROM telegram_sessions
GROUP BY current_step;
"

# Session age distribution
docker exec -it rafiki-postgres psql -U rafiki -d rafiki -c "
SELECT
    session_id,
    current_step,
    retry_count,
    EXTRACT(EPOCH FROM (NOW() - last_activity)) / 60 as age_minutes
FROM telegram_sessions
WHERE last_activity > NOW() - INTERVAL '1 hour'
ORDER BY last_activity DESC;
"
```

### Grep Patterns for Common Issues

```bash
# Check for stuck sessions (>30 min old)
docker exec -it rafiki-postgres psql -U rafiki -d rafiki -c "
SELECT COUNT(*) FROM telegram_sessions
WHERE last_activity < NOW() - INTERVAL '30 minutes';
"

# If count > 0: Session cleanup job will handle it

# Check for failed AI calls
docker compose logs --since 1h partner-service | grep "anthropic_api_error" | wc -l

# Check for job retries
docker compose logs --since 1h partner-service | grep "job_retry" | grep telegrammessage
```

### Daily Monitoring Routine (Manual)

```bash
# 1. Check service health
curl -s https://api.rafiki.lat/v1/readiness | jq

# 2. Check job completion rate (last 24h)
ssh root@178.156.170.37 "docker exec -it rafiki-postgres psql -U rafiki -d rafiki -c \"
SELECT
    state,
    COUNT(*) as count
FROM river_job
WHERE kind = 'telegrammessage'
  AND created_at > NOW() - INTERVAL '24 hours'
GROUP BY state;
\""

# 3. Check for rate limit errors
ssh root@178.156.170.37 "docker compose logs --since 24h partner-service | grep -c '429'"

# 4. Check active sessions
ssh root@178.156.170.37 "docker exec -it rafiki-postgres psql -U rafiki -d rafiki -c \"
SELECT COUNT(*) FROM telegram_sessions;
\""
```

---

## Rollback Plan

### Scenario 1: AI Calls Fail (Anthropic API Issues)

**Symptoms:**
- Logs show "anthropic_api_error" repeatedly
- Jobs stuck in "retrying" state
- Users not receiving responses

**Diagnosis:**
```bash
# Check error rate
docker compose logs --since 10m partner-service | grep anthropic_api_error

# Check API key validity
docker compose logs partner-service | grep "x-api-key.*invalid"

# Check rate limits
docker compose logs partner-service | grep "429"
```

**Fix Options:**

**Option A: Temporary API Key Issue**
```bash
# Verify API key in .env
ssh root@178.156.170.37 "grep ANTHROPIC_APIKEY /opt/rafiki/.env"

# If key is wrong, fix and restart
ssh root@178.156.170.37 "cd /opt/rafiki && docker compose restart partner-service"

# Monitor recovery
docker compose logs -f partner-service | grep anthropic
```

**Option B: Rate Limit Exceeded**
- Jobs will retry automatically (River exponential backoff)
- No action needed - wait for rate limit reset (1 minute)
- If persistent, reduce concurrent users (block /momento temporarily)

**Option C: Rollback to Deliverable 4**
```bash
# SSH to server
ssh root@178.156.170.37
cd /opt/rafiki

# Rollback to previous commit (before D5a+5b)
git log --oneline | head -5
git reset --hard <commit_hash_before_d5>

# Redeploy
sudo ./devops/deploy.sh

# Time: 3-5 minutes
# Impact: /momento command stops working, webhook still receives messages
```

### Scenario 2: Worker Crashes or Memory Leak

**Symptoms:**
- Service restarts frequently (check `docker compose ps`)
- Memory usage climbing (check `docker stats rafiki-service`)
- OOM killer logs in system

**Diagnosis:**
```bash
# Check restart count
docker inspect rafiki-service | jq '.[0].RestartCount'

# Check memory usage
docker stats rafiki-service --no-stream

# Check for OOM kills
ssh root@178.156.170.37 "dmesg | grep -i 'out of memory'"
```

**Fix:**

**Quick Fix: Restart Service**
```bash
ssh root@178.156.170.37 "cd /opt/rafiki && docker compose restart partner-service"
```

**Permanent Fix: Rollback**
```bash
# Same as Scenario 1 Option C
git reset --hard <previous_commit>
sudo ./devops/deploy.sh
```

### Scenario 3: Database Sessions Corrupted

**Symptoms:**
- "malformed_json_response" errors
- Session state inconsistent (current_step wrong)
- parsed_data JSONB invalid

**Diagnosis:**
```bash
# Check for NULL parsed_data
docker exec -it rafiki-postgres psql -U rafiki -d rafiki -c "
SELECT session_id, current_step, parsed_data
FROM telegram_sessions
WHERE parsed_data IS NULL OR parsed_data = '{}';
"

# Check for invalid JSON
docker exec -it rafiki-postgres psql -U rafiki -d rafiki -c "
SELECT session_id, parsed_data
FROM telegram_sessions
WHERE NOT (parsed_data::text ~ '^{.*}$');
"
```

**Fix:**

**Option A: Clean Up Bad Sessions**
```bash
# Delete corrupted sessions
docker exec -it rafiki-postgres psql -U rafiki -d rafiki -c "
DELETE FROM telegram_sessions
WHERE parsed_data IS NULL
   OR parsed_data = '{}'
   OR last_activity < NOW() - INTERVAL '1 hour';
"

# Users affected: Must restart /momento
```

**Option B: Full Session Reset**
```bash
# Nuclear option: Delete all sessions
docker exec -it rafiki-postgres psql -U rafiki -d rafiki -c "
TRUNCATE TABLE telegram_sessions;
"

# Impact: All users lose in-progress sessions
# Notification: Not needed (sessions expire after 30 min anyway)
```

### Scenario 4: Webhook Still Works, Worker Doesn't Process

**Symptoms:**
- Webhook receives messages (200 OK responses)
- Jobs enqueued but never processed
- river_job table fills up with 'available' jobs

**Diagnosis:**
```bash
# Check job queue state
docker exec -it rafiki-postgres psql -U rafiki -d rafiki -c "
SELECT state, COUNT(*)
FROM river_job
WHERE kind = 'telegrammessage'
GROUP BY state;
"

# Expected problem: All jobs in 'available' state (not processing)

# Check worker logs
docker compose logs partner-service | grep "job queue started"
docker compose logs partner-service | grep "telegrammessage worker"
```

**Fix:**

**Option A: Worker Not Registered**
```bash
# Check main.go for worker registration
grep -A 10 "telegrammessage.NewWorker" api/services/partners/main.go

# If missing, worker wasn't added - rollback
```

**Option B: Restart Service**
```bash
docker compose restart partner-service

# Monitor recovery
docker compose logs -f partner-service | grep "job queue started"
```

### Rollback Time Estimates

| Scenario | Detection | Fix Time | User Impact |
|----------|-----------|----------|-------------|
| Wrong API Key | 1 min | 2 min (restart) | Low - jobs retry |
| Rate Limit | Immediate | 1 min (auto) | Low - delays only |
| Worker Crash | 5 min | 5 min (rollback) | Medium - /momento offline |
| DB Corruption | 10 min | 2 min (cleanup) | Low - restart sessions |
| Full Rollback | 15 min | 5 min (git reset) | High - feature offline |

---

## Rate Limit Management

### Build Tier 1 Limits (Anthropic)

- **Rate Limit:** 50 requests/minute
- **Daily Limit:** 2,500 messages/day
- **Burst:** Allowed (up to 50 concurrent)

### Expected Usage (10 Users)

**Best Case (All Sessions Complete):**
- 10 users × 6 steps = 60 API calls/day
- Well under daily limit (2,500)

**Worst Case (Many Retries):**
- 10 users × 6 steps × 3 retries = 180 API calls/day
- Still comfortable (7% of daily limit)

### Rate Limit Handling

**Built-in (River Job Queue):**
- Automatic retry with exponential backoff
- Max 3 attempts per job
- Jobs spread over time (no burst spikes)

**Manual Monitoring:**
```bash
# Check API usage (Anthropic Console)
# https://console.anthropic.com/settings/usage

# Expected for 10 users/day:
# - Requests: 60-180
# - Input tokens: ~30,000 (500 tokens × 60 requests)
# - Output tokens: ~12,000 (200 tokens × 60 requests)
# - Cost: ~$0.60/day ($18/month)
```

### What to Do If Rate Limit Exceeded

**Immediate Action: None Required**
- Jobs automatically retry after 1 minute
- Users experience slight delay (acceptable)

**If Persistent (>10 users active simultaneously):**

1. **Check actual usage:**
```bash
docker compose logs --since 1h partner-service | grep "anthropic_api_error" | grep 429 | wc -l
```

2. **Temporary mitigation:**
```bash
# Disable /momento command temporarily (via webhook handler)
# Edit app/domain/telegramapp/webhook.go
# Comment out /momento handler
# Redeploy

# Or: Return "Sistema en mantenimiento" message
```

3. **Long-term fix:**
- Upgrade to Scale Tier (10,000 req/min)
- Cost: $40/month + usage
- Setup: Contact Anthropic support

---

## Summary

### Pre-Deployment Checklist Summary

- [ ] Anthropic API key obtained (Build Tier 1)
- [ ] `.env` file updated with PARTNER_ANTHROPIC_APIKEY
- [ ] Code built locally without errors
- [ ] Git committed and pushed to main

### Deployment Summary

1. **Update .env:** Add Anthropic config (2 min)
2. **Deploy code:** `make deploy` or `sudo ./devops/deploy.sh` (3-5 min)
3. **Verify logs:** Check for "anthropic client initialized" (1 min)
4. **Test end-to-end:** Send /momento, verify AI response (5 min)

**Total Time:** 15-20 minutes

### Post-Deployment Validation Summary

- [ ] Service healthy (readiness check passes)
- [ ] Worker registered (logs confirm)
- [ ] Test session advances Step 1 → Step 2
- [ ] AI integration works (logs show API calls)
- [ ] Retry logic tested (invalid response handled)
- [ ] Database sessions updated correctly

### Monitoring Summary (Daily)

- Check service health (curl readiness)
- Check job completion rate (SQL query)
- Check for rate limit errors (grep 429)
- Verify no stuck sessions (SQL query)

### Rollback Summary

- **Quick:** Restart service (2 min)
- **Full:** Git reset to previous commit (5 min)
- **Impact:** /momento offline, webhook still works

---

**Questions or Issues?**

- Check logs: `docker compose logs -f partner-service | grep telegrammessage`
- Check database: `psql` queries for sessions and jobs
- Rollback: `git reset --hard <previous_commit> && sudo ./devops/deploy.sh`

**End of Deployment Guide**
