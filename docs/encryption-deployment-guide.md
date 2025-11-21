# Field-Level Encryption Deployment Guide

**Document Version:** 1.0
**Created:** 2025-11-20
**Target Platform:** Hetzner CPX11 (178.156.170.37)
**Status:** Planning Phase

---

## Table of Contents

1. [Overview](#overview)
2. [Prerequisites](#prerequisites)
3. [Pre-Deployment Checklist](#pre-deployment-checklist)
4. [Production Deployment Procedure](#production-deployment-procedure)
5. [Verification Steps](#verification-steps)
6. [Monitoring](#monitoring)
7. [Troubleshooting](#troubleshooting)
8. [Rollback Procedure](#rollback-procedure)
9. [Security Checklist](#security-checklist)

---

## Overview

### What This Deploys

Field-level encryption for sensitive user data in Rafiki habits tracker:

**Encrypted Fields**:
- **Moments**: situation, thoughts, physicalSymptoms, behavior, consequences, valuesReflection (6 fields)
- **Thinks**: content (1 field)

**Not Encrypted** (needed for queries):
- Moment: intensity, momentDate, userID, IDs, timestamps
- Think: category, userID, IDs, timestamps

### Deployment Strategy

**Approach**: Fresh start (no production data exists)

1. Generate encryption key
2. Add key to server `.env` file
3. Deploy code with encryption enabled
4. Wipe database (fresh start)
5. All new data encrypted automatically

**No schema changes needed** - existing TEXT columns store encrypted base64.

---

## Prerequisites

### Local Machine

**Required**:
```bash
# Verify you have access to server
ssh root@178.156.170.37

# Verify you can deploy
cd /Users/francowini/Documents/rafiki
git status
```

**Encryption key ready**:
- 32-byte key (64 hex characters)
- Backed up in password manager
- Never committed to git

### Server Access

**Verify SSH access**:
```bash
ssh root@178.156.170.37
```

**Verify deployment directory**:
```bash
ls -la /opt/rafiki
```

---

## Pre-Deployment Checklist

### Code Preparation

- [ ] Implementation complete (all phases from [implementation plan](./encryption-implementation-plan.md))
- [ ] All code compiles: `go build ./...`
- [ ] golangci-lint passing: `golangci-lint run`
- [ ] Local testing complete (see implementation plan Phase 7)
- [ ] Code committed to branch
- [ ] Branch pushed to GitHub
- [ ] Ready to merge to `main`

### Key Management

- [ ] Encryption key generated: `openssl rand -hex 32`
- [ ] Key is exactly 64 hex characters
- [ ] Key backed up in password manager
- [ ] Key NOT in git repository
- [ ] Key NOT in any commit history

### Documentation

- [ ] CLAUDE.md updated with encryption details
- [ ] .env.example updated with key documentation
- [ ] Team notified of deployment

---

## Production Deployment Procedure

### Step 1: Merge Code to Main

```bash
# On local machine
cd /Users/francowini/Documents/rafiki

# Ensure all changes committed
git status

# Merge to main (or create PR and merge)
git checkout main
git pull origin main
git merge <your-encryption-branch>
git push origin main
```

### Step 2: Generate Production Encryption Key

```bash
# Generate secure 32-byte key
openssl rand -hex 32

# Example output (DO NOT USE THIS - generate your own!):
# a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4a5b6c7d8e9f0a1b2
```

**CRITICAL**: Save this key immediately in your password manager!

### Step 3: SSH to Production Server

```bash
ssh root@178.156.170.37
```

### Step 4: Navigate to Project Directory

```bash
cd /opt/rafiki
```

### Step 5: Create Backup (Important!)

```bash
# Backup current database (just in case)
docker exec rafiki-postgres pg_dump -U rafiki rafiki > backup-$(date +%Y%m%d-%H%M%S).sql

# Verify backup created
ls -lh backup-*.sql
```

### Step 6: Stop Services

```bash
# Stop all running services
docker compose --profile production down

# Verify stopped
docker compose ps
```

### Step 7: Wipe Database (Fresh Start)

**WARNING**: This deletes ALL existing data!

```bash
# Remove database volume
docker volume rm rafiki_postgres_data

# Verify removed
docker volume ls | grep rafiki
# Should NOT show postgres_data
```

### Step 8: Add Encryption Key to .env

```bash
# Edit .env file
nano .env
```

**Add this line** (replace with your actual key):
```bash
# Field-Level Encryption Key (AES-256-GCM)
# Generated: 2025-11-20
# CRITICAL: Key loss = permanent data loss! Back up in password manager!
PARTNER_ENCRYPTION_KEY=a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4a5b6c7d8e9f0a1b2
```

**Save and exit**: `Ctrl+X`, then `Y`, then `Enter`

### Step 9: Secure .env File

```bash
# Set restrictive permissions (owner read/write only)
chmod 600 .env

# Set ownership to root
chown root:root .env

# Verify permissions
ls -la .env
# Expected: -rw------- 1 root root
```

### Step 10: Pull Latest Code

```bash
# Pull latest code from main branch
git pull origin main

# Verify encryption code is present
ls -la business/sdk/encrypt/
ls -la business/types/encryptedcontent/
```

### Step 11: Deploy

```bash
# Run deployment script
./devops/deploy.sh
```

**Watch for**:
- ✅ "Encryption initialized" log message
- ✅ Database migrations run successfully
- ✅ Service starts without errors
- ✅ Health check passes

### Step 12: Verify Encryption Key Loaded

```bash
# Check if key is loaded (without showing value)
docker exec rafiki-service sh -c 'test -n "$PARTNER_ENCRYPTION_KEY" && echo "✅ Key loaded" || echo "❌ Key NOT loaded"'

# Expected: ✅ Key loaded
```

---

## Verification Steps

### 1. Check Service Health

```bash
# Health check (local)
curl http://localhost:3000/v1/readiness

# Expected: {"status":"ok"}

# Health check (public)
curl https://api.rafiki.lat/v1/readiness

# Expected: {"status":"ok"}
```

### 2. Check Encryption Initialization

```bash
# View service logs
docker compose logs partner-service | grep -i encryption

# Expected output:
# {"level":"INFO","msg":"startup","status":"initializing encryption"}
# {"level":"INFO","msg":"startup","status":"encryption initialized","algorithm":"AES-256-GCM"}
```

### 3. Create Test Moment (via API)

```bash
# Get auth token (replace with your method)
TOKEN="your-test-user-token"

# Create moment with sensitive data
curl -X POST https://api.rafiki.lat/v1/moments \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "momentDate": "2025-11-20T12:00:00Z",
    "situation": "Test - This is sensitive personal information",
    "thoughts": "Test - Very private thoughts",
    "physicalSymptoms": "Test symptoms",
    "behavior": "Test behavior",
    "consequences": "Test consequences",
    "valuesReflection": "Test reflection",
    "intensity": 7
  }' | jq

# Save moment_id from response
```

### 4. Verify Encryption in Database

```bash
# Connect to PostgreSQL
docker exec -it rafiki-postgres psql -U rafiki -d rafiki

# Query moment (should see ENCRYPTED data)
SELECT
    moment_id,
    LEFT(situation, 60) as situation_preview,
    LEFT(thoughts, 60) as thoughts_preview,
    intensity
FROM moments
ORDER BY date_created DESC
LIMIT 1;

# Expected:
# - situation_preview: Base64 gibberish (NOT plaintext)
# - thoughts_preview: Base64 gibberish (NOT plaintext)
# - intensity: 7 (plaintext number - NOT encrypted)

# Exit
\q
```

### 5. Verify Decryption in API

```bash
# Query moment via API (should see DECRYPTED data)
curl https://api.rafiki.lat/v1/moments/<moment-id> \
  -H "Authorization: Bearer $TOKEN" | jq

# Expected:
# {
#   "situation": "Test - This is sensitive personal information",  ← Plaintext!
#   "thoughts": "Test - Very private thoughts",  ← Plaintext!
#   "intensity": 7
# }
```

### 6. Test Think Entity

```bash
# Create think
curl -X POST https://api.rafiki.lat/v1/thinks \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "category": "personal",
    "content": "Test - Private personal content that should be encrypted"
  }' | jq

# Verify in database (content should be encrypted)
docker exec -it rafiki-postgres psql -U rafiki -d rafiki -c \
  "SELECT LEFT(content, 60) as content_preview, category FROM thinks ORDER BY date_created DESC LIMIT 1;"

# Verify in API (content should be decrypted)
curl https://api.rafiki.lat/v1/thinks/<think-id> \
  -H "Authorization: Bearer $TOKEN" | jq
```

---

## Monitoring

### First 24 Hours

**Monitor these metrics**:

#### 1. Application Logs

```bash
# Watch logs in real-time
docker compose logs -f partner-service

# Look for errors related to encryption
docker compose logs partner-service | grep -i "encrypt\|decrypt"

# Should NOT see: "failed to encrypt" or "failed to decrypt"
```

#### 2. Error Rates

```bash
# Check for 500 errors
docker compose logs partner-service | grep "status\":500"

# Should be minimal/none
```

#### 3. Response Times

```bash
# Test API response time
time curl -s https://api.rafiki.lat/v1/moments \
  -H "Authorization: Bearer $TOKEN" > /dev/null

# Should be <200ms (encryption adds <1ms overhead)
```

#### 4. Database Size

```bash
# Check database size
docker exec -it rafiki-postgres psql -U rafiki -d rafiki -c "
  SELECT
    schemaname,
    tablename,
    pg_size_pretty(pg_total_relation_size(schemaname||'.'||tablename)) AS size
  FROM pg_tables
  WHERE schemaname = 'public'
  ORDER BY pg_total_relation_size(schemaname||'.'||tablename) DESC;
"

# Encrypted data is ~2-3x larger than plaintext
# TEXT columns handle this fine
```

#### 5. Memory Usage

```bash
# Check service memory
docker stats rafiki-service --no-stream

# Encryption should add <10MB memory overhead
```

### Ongoing Monitoring

**Weekly**:
- [ ] Review error logs for decryption failures
- [ ] Verify backup includes `.env` file (separately, encrypted)
- [ ] Confirm encryption key backed up in password manager

**Monthly**:
- [ ] Review access logs to `.env` file
- [ ] Audit who has server access
- [ ] Consider key rotation (future enhancement)

---

## Troubleshooting

### Service Won't Start

**Symptom**: Service crashes on startup

**Check logs**:
```bash
docker compose logs partner-service
```

**Common issues**:

1. **Missing encryption key**:
   ```
   Error: encryption key must be 64 hex characters, got 0
   ```
   **Fix**: Add `PARTNER_ENCRYPTION_KEY` to `.env`

2. **Invalid key length**:
   ```
   Error: encryption key must be 64 hex characters, got 32
   ```
   **Fix**: Key must be 64 hex chars (32 bytes). Regenerate with `openssl rand -hex 32`

3. **Invalid key format**:
   ```
   Error: decode encryption key: encoding/hex: invalid byte
   ```
   **Fix**: Key must be valid hex. Only characters 0-9, a-f allowed.

### Decryption Failures

**Symptom**: API returns 500 errors when querying moments/thinks

**Check logs**:
```bash
docker compose logs partner-service | grep -i decrypt
```

**Common issues**:

1. **Wrong encryption key**:
   ```
   ERROR: decrypt situation: decrypt: cipher: message authentication failed
   ```
   **Fix**: Verify encryption key in `.env` matches key used when data was encrypted

2. **Corrupted data**:
   ```
   ERROR: decrypt situation: decode base64: illegal base64 data
   ```
   **Fix**: Database contains corrupted data. May need to restore from backup.

### Performance Issues

**Symptom**: API responses slower than expected

**Check**:
```bash
# Enable slow query logging
docker compose logs partner-service | grep "duration"
```

**Expected**: Encryption adds <1ms per record. If you see >10ms overhead, investigate.

### Database Connection Issues

**Symptom**: Can't connect to database

**Check**:
```bash
# Verify postgres is running
docker compose ps postgres

# Test connection
docker exec -it rafiki-postgres psql -U rafiki -d rafiki -c "SELECT 1;"
```

---

## Rollback Procedure

### If Encryption Causes Issues

**Scenario**: Encryption is causing problems, need to rollback.

#### Step 1: Stop Services

```bash
ssh root@178.156.170.37
cd /opt/rafiki
docker compose --profile production down
```

#### Step 2: Checkout Previous Commit

```bash
# Find last working commit (before encryption)
git log --oneline

# Checkout previous commit
git checkout <previous-commit-sha>
```

#### Step 3: Remove Encryption Key

```bash
# Edit .env
nano .env

# Comment out encryption key
# PARTNER_ENCRYPTION_KEY=...

# Save and exit
```

#### Step 4: Restore Database (if needed)

```bash
# If you need old data, restore from backup
docker volume create rafiki_postgres_data

# Restore (if backup exists)
# docker run --rm -v rafiki_postgres_data:/var/lib/postgresql/data ...
```

#### Step 5: Redeploy

```bash
# Deploy without encryption
./devops/deploy.sh
```

### If Need to Recover Encrypted Data

**Scenario**: Deployed encryption, but need to decrypt all data (emergency).

**Requirement**: Must have original encryption key!

```bash
# This requires a custom migration script (not yet implemented)
# Contact development team for assistance
```

---

## Security Checklist

### Encryption Key Security

- [ ] Key is 32 bytes (64 hex characters)
- [ ] Key generated from cryptographically secure source (`openssl rand`)
- [ ] Key backed up in password manager (tested recovery)
- [ ] Key stored in `/opt/rafiki/.env` only
- [ ] `.env` file has 600 permissions (`-rw-------`)
- [ ] `.env` file owned by root (`root:root`)
- [ ] Key NOT in git repository
- [ ] Key NOT in any commit history
- [ ] Key NOT logged anywhere
- [ ] Key NOT sent via email/Slack/etc.

### Server Security

- [ ] SSH access restricted to authorized users
- [ ] Server firewall configured (only ports 80, 443, 22 open)
- [ ] TLS enabled (HTTPS only)
- [ ] Database not exposed publicly (internal Docker network)
- [ ] Regular security updates applied

### Backup Security

- [ ] Database backups encrypted separately
- [ ] Encryption key backed up separately (not with database)
- [ ] Backup restoration tested
- [ ] Disaster recovery plan documented

### Compliance

- [ ] Document which fields contain PII (Personal Identifiable Information)
- [ ] GDPR: Encryption at rest ✅
- [ ] HIPAA: ePHI encrypted ✅ (if applicable)
- [ ] Data retention policy defined
- [ ] Incident response plan prepared

---

## Post-Deployment

### Immediate (Within 1 Hour)

- [ ] Verify service health
- [ ] Verify encryption working (database check)
- [ ] Verify decryption working (API check)
- [ ] Monitor error logs
- [ ] Notify team of successful deployment

### First 24 Hours

- [ ] Monitor application logs
- [ ] Monitor error rates
- [ ] Monitor response times
- [ ] Test all API endpoints
- [ ] Verify frontend works correctly

### First Week

- [ ] Review accumulated logs
- [ ] Verify no decryption failures
- [ ] Check database size growth
- [ ] Confirm backups working
- [ ] Update runbook with any issues encountered

---

## Emergency Contacts

**If deployment fails**:
1. Check troubleshooting section above
2. Review logs: `docker compose logs partner-service`
3. Contact development team
4. Rollback if necessary (see Rollback Procedure)

**If encryption key lost**:
- **WARNING**: Data is permanently unrecoverable!
- Check password manager (primary backup)
- Check encrypted offline backup (secondary backup)
- If truly lost, must deploy with new key (lose all encrypted data)

---

## Appendix: Quick Reference

### Generate Encryption Key
```bash
openssl rand -hex 32
```

### Check Encryption Status
```bash
# On server
docker exec rafiki-service sh -c 'test -n "$PARTNER_ENCRYPTION_KEY" && echo "✅ Encrypted" || echo "❌ Not encrypted"'
```

### View Encrypted Data
```bash
# Connect to database
docker exec -it rafiki-postgres psql -U rafiki -d rafiki

# Query encrypted field
SELECT LEFT(situation, 60) FROM moments LIMIT 1;

# Should see base64 gibberish
```

### Verify Decryption
```bash
# Query via API (needs auth token)
curl https://api.rafiki.lat/v1/moments/<id> \
  -H "Authorization: Bearer $TOKEN" | jq .situation

# Should see plaintext
```

---

**Document Status**: ✅ Ready for Deployment
**Prerequisites**: [Implementation Plan](./encryption-implementation-plan.md) completed
**Related Docs**: [Architecture](./encryption-architecture.md)
