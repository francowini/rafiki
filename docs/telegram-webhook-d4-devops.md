# Deliverable 4: Webhook Handler - DevOps Guide

## Overview

Deployment guide for the Telegram webhook handler. Infrastructure is **ready** - only needs environment variables and one-time webhook registration.

## Infrastructure Status

| Component | Status | Action Required |
|-----------|--------|-----------------|
| Nginx | ✅ Ready | None |
| SSL/TLS | ✅ Ready | None |
| Firewall | ✅ Ready | Port 443 already open |
| Database | ✅ Ready | Schema already migrated |
| Docker | ✅ Ready | None |
| Rate Limiting | ✅ Sufficient | 10 req/s + 20 burst |

---

## 1. Environment Variables

### 1.1 Add to `.env.example`

```bash
# ==============================================================================
# Telegram Webhook Configuration (Deliverable 4)
# ==============================================================================

# Bot token from @BotFather (REQUIRED for webhook)
# Get from: https://t.me/BotFather -> /newbot or /token
PARTNER_TELEGRAM_BOTTOKEN=

# Webhook signature verification secret (REQUIRED)
# Generate with: openssl rand -hex 32
PARTNER_TELEGRAM_WEBHOOKSECRET=
```

### 1.2 Add to `docker-compose.yml`

In the `partner-service` environment section:
```yaml
environment:
  # ... existing vars ...
  - PARTNER_TELEGRAM_BOTTOKEN=${PARTNER_TELEGRAM_BOTTOKEN:-}
  - PARTNER_TELEGRAM_WEBHOOKSECRET=${PARTNER_TELEGRAM_WEBHOOKSECRET:-}  # ADD
```

### 1.3 Production `.env` (on server)

```bash
# On server: /opt/rafiki/.env
PARTNER_TELEGRAM_BOTTOKEN=1234567890:ABCdefGHIjklMNOpqrsTUVwxyz
PARTNER_TELEGRAM_WEBHOOKSECRET=a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4a5b6c7d8e9f0a1b2
```

**Security**: `chmod 600 /opt/rafiki/.env`

---

## 2. One-Time Setup Checklist

### Step 1: Create Bot (if new)

```bash
# In Telegram, message @BotFather:
/newbot
# Follow prompts, save the token
```

### Step 2: Generate Webhook Secret

```bash
# On any machine
openssl rand -hex 32
# Copy output for next step
```

### Step 3: Add to Production Environment

```bash
ssh root@178.156.170.37
cd /opt/rafiki
nano .env

# Add these lines:
PARTNER_TELEGRAM_BOTTOKEN=<token_from_step_1>
PARTNER_TELEGRAM_WEBHOOKSECRET=<secret_from_step_2>

# Save and exit (Ctrl+X, Y, Enter)
chmod 600 .env
```

### Step 4: Deploy Code

```bash
# From local machine
make deploy

# OR on server
cd /opt/rafiki
git pull origin main
sudo ./devops/deploy.sh
```

### Step 5: Verify Service Health

```bash
curl https://api.rafiki.lat/v1/readiness
# Expected: {"status":"ok"}
```

### Step 6: Register Webhook with Telegram

```bash
# Replace <TOKEN> and <SECRET> with actual values
curl -X POST "https://api.telegram.org/bot<TOKEN>/setWebhook" \
  -H "Content-Type: application/json" \
  -d '{
    "url": "https://api.rafiki.lat/v1/telegram/webhook",
    "secret_token": "<SECRET>",
    "allowed_updates": ["message"]
  }'

# Expected response:
# {"ok":true,"result":true,"description":"Webhook was set"}
```

### Step 7: Verify Webhook Registration

```bash
curl "https://api.telegram.org/bot<TOKEN>/getWebhookInfo" | jq

# Expected:
# {
#   "ok": true,
#   "result": {
#     "url": "https://api.rafiki.lat/v1/telegram/webhook",
#     "has_custom_certificate": false,
#     "pending_update_count": 0
#   }
# }
```

### Step 8: Test Bot

1. Open Telegram
2. Search for your bot
3. Send `/start` or `/ayuda`
4. Verify response received

---

## 3. @BotFather Configuration

### Set Commands (for autocomplete)

```
# Message @BotFather:
/setcommands

# Select your bot, then send:
momento - Comenzar a registrar un momento 📝
cancel - Pausar el registro actual ⏸️
ayuda - Ver ayuda y cómo funciona ❓
ejemplo - Ver un ejemplo completo 💡
```

### Set Description

```
# Message @BotFather:
/setdescription

# Select your bot, then send:
Tu compañero para procesar momentos difíciles. Seis preguntas, reflexión profunda. 🌟
```

### Set About Text

```
# Message @BotFather:
/setabouttext

# Select your bot, then send:
Rafiki te acompaña a procesar momentos de malestar.

A través de seis preguntas simples, exploramos juntos qué pasó, cómo te afectó, y qué podés aprender.

¿Listo? Escribí /momento para empezar.
```

---

## 4. Webhook Management Commands

### Check Status
```bash
curl "https://api.telegram.org/bot<TOKEN>/getWebhookInfo" | jq
```

### Delete Webhook (emergency)
```bash
curl -X POST "https://api.telegram.org/bot<TOKEN>/deleteWebhook" \
  -d '{"drop_pending_updates": true}'
```

### Re-register Webhook
```bash
curl -X POST "https://api.telegram.org/bot<TOKEN>/setWebhook" \
  -d '{"url":"https://api.rafiki.lat/v1/telegram/webhook","secret_token":"<SECRET>"}'
```

---

## 5. Monitoring Commands

### View Logs
```bash
# From local
make deploy-logs | grep -i telegram

# On server
docker compose logs -f partner-service | grep -i telegram
```

### Check Active Sessions
```sql
-- Connect to database
docker exec -it rafiki-postgres psql -U rafiki -d rafiki

-- Active sessions
SELECT session_id, user_id, session_type, current_step, last_activity
FROM telegram_sessions
ORDER BY last_activity DESC;

-- Session count by step
SELECT current_step, COUNT(*) as count
FROM telegram_sessions
GROUP BY current_step;

-- Expired sessions (should be 0 if cleanup job running)
SELECT COUNT(*) FROM telegram_sessions
WHERE last_activity < NOW() - INTERVAL '15 minutes';
```

### Check Job Queue
```sql
-- Telegram message jobs
SELECT state, COUNT(*)
FROM river_job
WHERE kind = 'telegram_message'
GROUP BY state;

-- Recent failures
SELECT * FROM river_job
WHERE kind = 'telegram_message'
AND state = 'discarded'
ORDER BY created_at DESC
LIMIT 10;
```

---

## 6. Rollback Procedures

### Scenario 1: Code Issue (rollback deployment)

```bash
ssh root@178.156.170.37
cd /opt/rafiki
git reset --hard HEAD~1
sudo ./devops/deploy.sh
```

### Scenario 2: Configuration Error

```bash
# Fix .env
nano /opt/rafiki/.env

# Restart service
docker compose restart partner-service
```

### Scenario 3: Disable Webhook Temporarily

```bash
# Delete webhook (instant effect)
curl -X POST "https://api.telegram.org/bot<TOKEN>/deleteWebhook"

# Fix issue...

# Re-register webhook
curl -X POST "https://api.telegram.org/bot<TOKEN>/setWebhook" \
  -d '{"url":"https://api.rafiki.lat/v1/telegram/webhook","secret_token":"<SECRET>"}'
```

### Scenario 4: Database Migration Failed

```bash
# Check migration status
docker exec -it rafiki-postgres psql -U rafiki -d rafiki \
  -c "SELECT * FROM darwin_migrations ORDER BY version DESC LIMIT 5;"

# Migrations are idempotent - just redeploy
sudo ./devops/deploy.sh
```

---

## 7. Troubleshooting

### Webhook Not Receiving Messages

1. **Check registration**: `curl .../getWebhookInfo`
2. **Check SSL**: `openssl s_client -connect api.rafiki.lat:443`
3. **Check service health**: `curl https://api.rafiki.lat/v1/readiness`
4. **Check logs**: `docker compose logs partner-service | tail -50`

### Signature Verification Failing (401)

1. **Verify secret matches**: Compare `.env` with registered secret
2. **Re-register webhook** with correct secret
3. **Check header**: Telegram sends `X-Telegram-Bot-Api-Secret-Token`

### High Pending Update Count

1. **Check service**: Is it running? `docker compose ps`
2. **Check logs for errors**: `docker compose logs partner-service | grep ERROR`
3. **Restart if needed**: `docker compose restart partner-service`

### Job Queue Backing Up

```sql
-- Check queue status
SELECT state, COUNT(*) FROM river_job
WHERE kind = 'telegram_message'
GROUP BY state;

-- If many 'available' (not processing):
-- Check if worker is running
docker compose logs partner-service | grep "telegram_message"
```

---

## 8. Quick Reference

### Endpoints
| Endpoint | Purpose |
|----------|---------|
| `POST /v1/telegram/webhook` | Telegram webhook receiver |
| `GET /v1/readiness` | Health check |

### Environment Variables
| Variable | Required | Description |
|----------|----------|-------------|
| `PARTNER_TELEGRAM_BOTTOKEN` | Yes | Bot token from @BotFather |
| `PARTNER_TELEGRAM_WEBHOOKSECRET` | Yes | Webhook signature secret |

### Database Tables
| Table | Purpose |
|-------|---------|
| `telegram_sessions` | Active conversation sessions |
| `river_job` | Job queue (telegram_message jobs) |

### Useful Commands
```bash
# Deploy
make deploy

# Logs
make deploy-logs | grep telegram

# Webhook status
curl "https://api.telegram.org/bot<TOKEN>/getWebhookInfo"

# Service restart
docker compose restart partner-service
```

---

## 9. Externalized Messages

### Message File Location

Bot messages are stored in YAML and embedded at compile time:

```
app/domain/telegramapp/
└── messages/
    └── es.yaml    # Spanish user-facing messages
```

### How It Works

- Uses Go's `go:embed` directive (same pattern as OPA policies)
- Messages are compiled INTO the binary at build time
- **No runtime file I/O** - no external files needed in Docker container
- **No environment variables** for messages

### Updating Messages

**Any message change requires a rebuild and redeploy:**

```bash
# 1. Edit the YAML file
vim app/domain/telegramapp/messages/es.yaml

# 2. Commit the change
git add -A && git commit -m "Update Telegram bot messages"

# 3. Deploy (rebuild required)
make deploy
```

### Validation

- YAML parsing happens at application startup
- Malformed YAML will **panic** immediately (caught during deploy)
- Check logs for: `telegramapp: failed to parse messages/es.yaml`

### Benefits Over Environment Variables

| Approach | Pros | Cons |
|----------|------|------|
| **YAML + embed** | Version controlled, type-safe, no runtime deps | Requires rebuild |
| **Env vars** | Hot-reload possible | Hard to version, no validation |
| **Database** | Hot-reload, UI editing | Adds complexity, runtime deps |

We chose YAML + embed for MVP simplicity and safety.

---

## 10. Regular Deployment (After Initial Setup)

Once webhook is registered, regular deployments are simple:

```bash
# From local machine
make deploy

# That's it! Webhook persists, no re-registration needed.
```

Only re-register webhook if:
- URL changes (e.g., new domain)
- Secret changes (security rotation)
- Switching to different bot

---

## Success Criteria

- [ ] Webhook receiving messages (check logs)
- [ ] `getWebhookInfo` shows `pending_update_count: 0`
- [ ] Bot responds to `/ayuda` command
- [ ] No errors in logs for signature verification
- [ ] Sessions created when user sends `/momento`
