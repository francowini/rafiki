# Field-Level Encryption: Final Deployment Configuration

**Document Version:** 1.0
**Created:** 2025-11-21
**Status:** FINAL - Ready for Implementation
**Target Platform:** Hetzner CPX11 (178.156.170.37)

---

## Table of Contents

1. [Architecture Summary](#architecture-summary)
2. [Environment Configuration](#environment-configuration)
3. [main.go Configuration](#maingo-configuration)
4. [Deployment Procedure](#deployment-procedure)
5. [Verification Commands](#verification-commands)
6. [Rollback Procedure](#rollback-procedure)
7. [Monitoring and Troubleshooting](#monitoring-and-troubleshooting)

---

## Architecture Summary

**Agreed Design**:
- Encryptor created in `main.go` from `PARTNER_ENCRYPTION_KEY` environment variable
- Encryptor injected into database stores (no global state)
- `EncryptedContent` business type guarantees encryption at compile-time
- AES-256-GCM algorithm (authenticated encryption)

**Data Flow**:
```
API Request (plaintext)
    → App Layer (parses to EncryptedContent type)
    → Business Layer (works with EncryptedContent - plaintext in memory)
    → DB Store (encrypts before INSERT/UPDATE, decrypts after SELECT)
    → Database (stores base64-encoded ciphertext)
```

---

## Environment Configuration

### 1. Update `.env.example`

Add the following section to `/Users/francowini/Documents/rafiki/.env.example`:

```bash
# ==============================================================================
# Field-Level Encryption Configuration
# ==============================================================================

# Encryption key for sensitive user data (AES-256-GCM)
# REQUIRED for production deployment
#
# Key requirements:
# - Must be exactly 32 bytes (64 hex characters)
# - Generate with: openssl rand -hex 32
# - Store securely in password manager (key loss = permanent data loss!)
# - Never commit this key to version control
#
# Encrypted fields:
# - Moments: situation, thoughts, physicalSymptoms, behavior, consequences, valuesReflection
# - Thinks: content
#
# Example key generation:
# $ openssl rand -hex 32
# a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4a5b6c7d8e9f0a1b2
#
PARTNER_ENCRYPTION_KEY=
```

### 2. Update `docker-compose.yml`

Add the encryption key environment variable to the `partner-service` section.

In `/Users/francowini/Documents/rafiki/docker-compose.yml`, find the `partner-service` environment section and add:

```yaml
      # Encryption Configuration
      - PARTNER_ENCRYPTION_KEY=${PARTNER_ENCRYPTION_KEY}
```

**Full context** (add after the Go Runtime Configuration section, around line 149):

```yaml
      # Go Runtime Configuration
      - GOMAXPROCS=2
      - GOMEMLIMIT=128MiB
      - GOGC=100

      # Encryption Configuration (ADD THIS)
      - PARTNER_ENCRYPTION_KEY=${PARTNER_ENCRYPTION_KEY}
```

### 3. Key Generation Command

```bash
# Generate a cryptographically secure 32-byte key (64 hex characters)
openssl rand -hex 32

# Example output (DO NOT USE - generate your own!):
# a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4a5b6c7d8e9f0a1b2
```

---

## main.go Configuration

### Complete Code Changes for `/Users/francowini/Documents/rafiki/api/services/partners/main.go`

#### Step 1: Add Imports

Add to the import block:

```go
import (
    // ... existing imports ...
    "encoding/hex"

    "github.com/francowini/rafiki/business/sdk/encrypt"
)
```

#### Step 2: Add Encryption to Config Struct

Find the `cfg` struct definition (around line 73) and add the Encryption section:

```go
cfg := struct {
    conf.Version
    Web struct {
        ReadTimeout        time.Duration `conf:"default:5s"`
        WriteTimeout       time.Duration `conf:"default:10s"`
        IdleTimeout        time.Duration `conf:"default:120s"`
        ShutdownTimeout    time.Duration `conf:"default:20s"`
        APIHost            string        `conf:"default:0.0.0.0:3000"`
        DebugHost          string        `conf:"default:0.0.0.0:3010"`
        CORSAllowedOrigins []string      `conf:"default:*"`
    }
    DB struct {
        User         string `conf:"default:postgres"`
        Password     string `conf:"default:postgres,mask"`
        Host         string `conf:"default:database-service"`
        Name         string `conf:"default:postgres"`
        MaxIdleConns int    `conf:"default:0"`
        MaxOpenConns int    `conf:"default:0"`
        DisableTLS   bool   `conf:"default:true"`
    }
    Encryption struct {                      // ADD THIS SECTION
        Key string `conf:"required,mask"`    // PARTNER_ENCRYPTION_KEY
    }
    Tempo struct {
        Host        string  `conf:"default:tempo:4317"`
        ServiceName string  `conf:"default:rafiki-service"`
        Probability float64 `conf:"default:1.0"`
    }
}{
    Version: conf.Version{
        Build: build,
        Desc:  "Partner",
    },
}
```

#### Step 3: Add Encryption Initialization

Add this section AFTER database migrations and BEFORE "Create Business Packages" (around line 167):

```go
    log.Info(ctx, "startup", "status", "database migrations completed")

    // -------------------------------------------------------------------------
    // Initialize Encryption Support
    // -------------------------------------------------------------------------

    log.Info(ctx, "startup", "status", "initializing encryption")

    // Validate key format (must be 64 hex characters = 32 bytes)
    if len(cfg.Encryption.Key) != 64 {
        return fmt.Errorf("encryption key must be 64 hex characters (32 bytes), got %d characters", len(cfg.Encryption.Key))
    }

    // Decode hex string to bytes
    encryptionKey, err := hex.DecodeString(cfg.Encryption.Key)
    if err != nil {
        return fmt.Errorf("decode encryption key (must be valid hex): %w", err)
    }

    // Create AES-256-GCM encryptor
    encryptor, err := encrypt.NewAESEncryptor(encryptionKey)
    if err != nil {
        return fmt.Errorf("create encryptor: %w", err)
    }

    // Clear key from memory after encryptor is created
    for i := range encryptionKey {
        encryptionKey[i] = 0
    }

    log.Info(ctx, "startup", "status", "encryption initialized", "algorithm", "AES-256-GCM")

    // -------------------------------------------------------------------------
    // Create Business Packages
    // -------------------------------------------------------------------------
```

#### Step 4: Update Store Creation

Find the existing store creation code (around line 172) and update to inject the encryptor:

```go
    // -------------------------------------------------------------------------
    // Create Business Packages
    // -------------------------------------------------------------------------

    // OLD CODE (replace this):
    // thinkStore := thinkdb.NewStore(log, db)
    // momentStore := momentdb.NewStore(log, db)

    // NEW CODE (with encryptor injection):
    thinkStore := thinkdb.NewStore(log, db, encryptor)
    thinkBus := thinkbus.NewBusiness(log, thinkStore)

    momentStore := momentdb.NewStore(log, db, encryptor)
    momentBus := momentbus.NewBusiness(log, momentStore)
```

#### Complete Diff Summary

```diff
 import (
     "context"
+    "encoding/hex"
     "errors"
     "expvar"
     "fmt"
@@ ... existing imports ...
     "github.com/francowini/rafiki/business/domain/userbus/stores/userdb"
+    "github.com/francowini/rafiki/business/sdk/encrypt"
     "github.com/francowini/rafiki/business/sdk/migrate"
 )

@@ ... in cfg struct ...
         DisableTLS   bool   `conf:"default:true"`
     }
+    Encryption struct {
+        Key string `conf:"required,mask"`
+    }
     Tempo struct {

@@ ... after database migrations ...
     log.Info(ctx, "startup", "status", "database migrations completed")

+    // -------------------------------------------------------------------------
+    // Initialize Encryption Support
+    // -------------------------------------------------------------------------
+
+    log.Info(ctx, "startup", "status", "initializing encryption")
+
+    if len(cfg.Encryption.Key) != 64 {
+        return fmt.Errorf("encryption key must be 64 hex characters (32 bytes), got %d characters", len(cfg.Encryption.Key))
+    }
+
+    encryptionKey, err := hex.DecodeString(cfg.Encryption.Key)
+    if err != nil {
+        return fmt.Errorf("decode encryption key (must be valid hex): %w", err)
+    }
+
+    encryptor, err := encrypt.NewAESEncryptor(encryptionKey)
+    if err != nil {
+        return fmt.Errorf("create encryptor: %w", err)
+    }
+
+    for i := range encryptionKey {
+        encryptionKey[i] = 0
+    }
+
+    log.Info(ctx, "startup", "status", "encryption initialized", "algorithm", "AES-256-GCM")

@@ ... store creation ...
-    thinkStore := thinkdb.NewStore(log, db)
+    thinkStore := thinkdb.NewStore(log, db, encryptor)
     thinkBus := thinkbus.NewBusiness(log, thinkStore)

-    momentStore := momentdb.NewStore(log, db)
+    momentStore := momentdb.NewStore(log, db, encryptor)
     momentBus := momentbus.NewBusiness(log, momentStore)
```

---

## Deployment Procedure

### Pre-Deployment Checklist

- [ ] All encryption code implemented and compiles: `go build ./...`
- [ ] golangci-lint passes: `golangci-lint run`
- [ ] Local testing complete
- [ ] Code merged to `main` branch
- [ ] Encryption key generated and backed up in password manager

### Step-by-Step Production Deployment

#### Step 1: Generate Production Encryption Key

```bash
# On your LOCAL machine (not server)
openssl rand -hex 32

# IMMEDIATELY save output to password manager!
# Example: a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4a5b6c7d8e9f0a1b2
```

**CRITICAL**: Store the key in multiple locations:
1. Password manager (primary)
2. Encrypted offline backup (secondary)
3. Write down and store in physical safe (tertiary)

**Key loss = permanent data loss. There is NO recovery.**

#### Step 2: SSH to Production Server

```bash
ssh root@178.156.170.37
cd /opt/rafiki
```

#### Step 3: Create Database Backup

```bash
# Backup current database
docker exec rafiki-postgres pg_dump -U rafiki rafiki > backup-pre-encryption-$(date +%Y%m%d-%H%M%S).sql

# Verify backup
ls -lh backup-*.sql
```

#### Step 4: Stop Services

```bash
docker compose --profile production down
```

#### Step 5: Add Encryption Key to .env

```bash
# Edit .env file
nano .env
```

Add this line (replace with YOUR generated key):

```bash
# Field-Level Encryption Key (AES-256-GCM)
# Generated: YYYY-MM-DD
# CRITICAL: Back up in password manager! Key loss = permanent data loss!
PARTNER_ENCRYPTION_KEY=your_64_character_hex_key_here
```

Save and exit: `Ctrl+X`, `Y`, `Enter`

#### Step 6: Secure .env File

```bash
# Set restrictive permissions
chmod 600 .env
chown root:root .env

# Verify
ls -la .env
# Expected: -rw------- 1 root root
```

#### Step 7: Pull Latest Code

```bash
git pull origin main
```

#### Step 8: Deploy

```bash
./devops/deploy.sh
```

#### Step 9: Verify Deployment

```bash
# Check service is running
docker compose ps

# Check encryption initialized
docker compose logs partner-service 2>&1 | grep -i encryption

# Expected output:
# {"level":"INFO","msg":"startup","status":"initializing encryption"}
# {"level":"INFO","msg":"startup","status":"encryption initialized","algorithm":"AES-256-GCM"}
```

### Key Management Best Practices

| Action | Command/Location |
|--------|-----------------|
| Generate key | `openssl rand -hex 32` |
| Primary backup | Password manager (1Password, Bitwarden, etc.) |
| Secondary backup | Encrypted USB drive in physical safe |
| Production storage | `/opt/rafiki/.env` with `chmod 600` |
| Never store in | Git, Slack, email, cloud notes, plain text files |

---

## Verification Commands

### 1. Check Service Started with Encryption

```bash
# Check logs for encryption initialization
docker compose logs partner-service 2>&1 | grep -i encryption

# Expected:
# {"level":"INFO","msg":"startup","status":"initializing encryption"}
# {"level":"INFO","msg":"startup","status":"encryption initialized","algorithm":"AES-256-GCM"}
```

### 2. Check Key is Loaded

```bash
# Verify key is set (without revealing value)
docker exec rafiki-service sh -c 'test -n "$PARTNER_ENCRYPTION_KEY" && echo "OK: Key loaded (length: ${#PARTNER_ENCRYPTION_KEY} chars)" || echo "ERROR: Key NOT loaded"'

# Expected: OK: Key loaded (length: 64 chars)
```

### 3. Check Data is Encrypted in DB

```bash
# Connect to database
docker exec -it rafiki-postgres psql -U rafiki -d rafiki

# Check moments table (should see base64 gibberish)
SELECT
    moment_id,
    LEFT(situation, 50) as situation_preview,
    intensity
FROM moments
LIMIT 3;

# Check thinks table
SELECT
    think_id,
    LEFT(content, 50) as content_preview,
    category
FROM thinks
LIMIT 3;

# Exit psql
\q
```

**Expected output**: `situation_preview` and `content_preview` should show base64-encoded strings (random characters ending in `=` or `==`), NOT readable text.

### 4. Check API Returns Decrypted Data

```bash
# Get auth token first
TOKEN="your-jwt-token"

# Query moments via API
curl -s https://api.rafiki.lat/v1/moments \
  -H "Authorization: Bearer $TOKEN" | jq '.items[0].situation'

# Should return readable plaintext, e.g.: "I felt anxious about..."
```

### 5. Full End-to-End Test

```bash
# Create a test moment
curl -X POST https://api.rafiki.lat/v1/moments \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "momentDate": "2025-11-21T10:00:00Z",
    "situation": "ENCRYPTION_TEST_PLAINTEXT_STRING",
    "thoughts": "Test thoughts",
    "physicalSymptoms": "Test symptoms",
    "behavior": "Test behavior",
    "consequences": "Test consequences",
    "valuesReflection": "Test reflection",
    "intensity": 5
  }' | jq

# Save the returned moment_id

# Check database (should NOT contain "ENCRYPTION_TEST_PLAINTEXT_STRING")
docker exec -it rafiki-postgres psql -U rafiki -d rafiki -c \
  "SELECT situation FROM moments WHERE situation LIKE '%ENCRYPTION_TEST%';"

# Expected: 0 rows (because the plaintext is encrypted)

# Check API (should contain plaintext)
curl -s https://api.rafiki.lat/v1/moments/<moment-id> \
  -H "Authorization: Bearer $TOKEN" | jq '.situation'

# Expected: "ENCRYPTION_TEST_PLAINTEXT_STRING"
```

---

## Rollback Procedure

### If Encryption Causes Issues

**Scenario**: Service won't start or encryption is causing errors.

#### Option A: Disable Encryption Temporarily (No Data Loss)

**Note**: Only works if you haven't created new encrypted data yet.

```bash
# 1. SSH to server
ssh root@178.156.170.37
cd /opt/rafiki

# 2. Stop services
docker compose --profile production down

# 3. Checkout pre-encryption commit
git log --oneline  # Find last commit before encryption
git checkout <commit-sha-before-encryption>

# 4. Comment out encryption key
nano .env
# Add # before: PARTNER_ENCRYPTION_KEY=...

# 5. Redeploy
./devops/deploy.sh
```

#### Option B: Quick Fix for Common Errors

**Error: "encryption key must be 64 hex characters"**
```bash
# Verify key length
grep PARTNER_ENCRYPTION_KEY .env | wc -c
# Should be 84 (64 chars + "PARTNER_ENCRYPTION_KEY=" + newline)

# Regenerate if wrong length
openssl rand -hex 32
# Update .env with new key
```

**Error: "decode encryption key: invalid hex"**
```bash
# Check for non-hex characters
grep PARTNER_ENCRYPTION_KEY .env
# Key must only contain: 0-9, a-f, A-F

# Regenerate if invalid
openssl rand -hex 32
```

**Error: "decrypt: cipher: message authentication failed"**
```bash
# Wrong key used. Data was encrypted with different key.
# Options:
# 1. Find original key from backup
# 2. Restore database from pre-encryption backup
```

#### Option C: Full Rollback with Database Restore

**Use when**: Encrypted data exists but key is lost or corrupted.

```bash
# 1. SSH to server
ssh root@178.156.170.37
cd /opt/rafiki

# 2. Stop services
docker compose --profile production down

# 3. Remove encrypted database
docker volume rm rafiki_postgres_data

# 4. Checkout pre-encryption code
git checkout <commit-before-encryption>

# 5. Remove encryption key from .env
nano .env
# Delete or comment out PARTNER_ENCRYPTION_KEY line

# 6. Restore database from backup (if available)
docker compose up -d postgres
sleep 10
docker exec -i rafiki-postgres psql -U rafiki -d rafiki < backup-pre-encryption-YYYYMMDD-HHMMSS.sql

# 7. Deploy
./devops/deploy.sh
```

---

## Monitoring and Troubleshooting

### Log Messages to Look For

#### Successful Startup
```json
{"level":"INFO","msg":"startup","status":"initializing encryption"}
{"level":"INFO","msg":"startup","status":"encryption initialized","algorithm":"AES-256-GCM"}
```

#### Error: Missing Key
```
Error: configuration: field Encryption.Key is required
```
**Fix**: Add `PARTNER_ENCRYPTION_KEY` to `.env`

#### Error: Invalid Key Length
```
Error: encryption key must be 64 hex characters (32 bytes), got X characters
```
**Fix**: Generate new key with `openssl rand -hex 32`

#### Error: Invalid Hex Characters
```
Error: decode encryption key (must be valid hex): encoding/hex: invalid byte
```
**Fix**: Key contains non-hex characters. Regenerate key.

#### Error: Decryption Failed
```
ERROR: decrypt situation: decrypt: cipher: message authentication failed
```
**Cause**: Wrong encryption key or corrupted data
**Fix**:
1. Verify correct key in `.env`
2. Check if key was changed after data was encrypted
3. Restore from backup if key is lost

### Common Errors and Fixes

| Error Message | Cause | Fix |
|--------------|-------|-----|
| `field Encryption.Key is required` | Key not in `.env` | Add `PARTNER_ENCRYPTION_KEY=...` to `.env` |
| `must be 64 hex characters` | Key too short/long | Use `openssl rand -hex 32` |
| `invalid byte` | Non-hex characters in key | Regenerate key, use only 0-9, a-f |
| `message authentication failed` | Wrong key or corrupted data | Verify key, restore backup |
| `ciphertext too short` | Corrupted encrypted data | Restore from backup |

### Health Check Commands

```bash
# Check service is running
docker compose ps partner-service

# Check service health
curl -s http://localhost:3000/v1/readiness | jq

# Check for errors in last hour
docker compose logs --since 1h partner-service 2>&1 | grep -i error

# Check memory usage (encryption adds minimal overhead)
docker stats rafiki-service --no-stream

# Check encryption-related logs
docker compose logs partner-service 2>&1 | grep -iE "encrypt|decrypt|cipher"
```

### Performance Monitoring

```bash
# Time API response (should be <200ms)
time curl -s https://api.rafiki.lat/v1/moments \
  -H "Authorization: Bearer $TOKEN" > /dev/null

# Check database query performance
docker exec -it rafiki-postgres psql -U rafiki -d rafiki -c \
  "SELECT count(*) FROM moments;"
```

**Expected**: Encryption adds <1ms overhead per field. If you see >10ms, investigate.

---

## Quick Reference

### Generate Key
```bash
openssl rand -hex 32
```

### Check Encryption Status
```bash
docker compose logs partner-service 2>&1 | grep -i encryption
```

### Check Key Loaded
```bash
docker exec rafiki-service sh -c 'test -n "$PARTNER_ENCRYPTION_KEY" && echo "OK" || echo "MISSING"'
```

### View Encrypted Data in DB
```bash
docker exec -it rafiki-postgres psql -U rafiki -d rafiki -c \
  "SELECT LEFT(situation, 50) FROM moments LIMIT 1;"
```

### View Decrypted Data via API
```bash
curl -s https://api.rafiki.lat/v1/moments -H "Authorization: Bearer $TOKEN" | jq
```

### Emergency Stop
```bash
docker compose --profile production down
```

### Emergency Rollback
```bash
git checkout <previous-commit>
./devops/deploy.sh
```

---

**Document Status**: FINAL - Ready for Implementation
**Author**: DevOps Engineering
**Last Updated**: 2025-11-21
