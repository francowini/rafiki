# Production Keys Setup

## Overview

**Local Development**: Keys are copied from `zarf/keys/` into Docker image
**Production**: Keys are mounted from external directory (NOT in image)

This approach ensures:
- ✅ Easy local development (keys in image)
- ✅ Secure production (keys external, not in image)
- ✅ Different keys per environment
- ✅ Keys can be rotated without rebuilding image

---

## Local Development (Current Setup)

Keys are **included in the Docker image** from `zarf/keys/`:

```bash
# Keys are in the repo
ls zarf/keys/
# 54bb2165-71e1-41a6-af3e-7da4a0e1e2c1.pem

# Build includes them automatically
docker compose up -d --build
```

**How it works:**
- Dockerfile copies `zarf/keys/` into image at build time
- Easy for development - just build and run
- Dev key is committed to Git (safe for local only)

---

## Production Setup

Keys are **mounted from external directory** (NOT in image):

### Step 1: Create Keys Directory on Server

```bash
# SSH into production server
ssh user@your-server

# Create keys directory
sudo mkdir -p /opt/rafiki/keys
sudo chmod 700 /opt/rafiki/keys
```

### Step 2: Generate Production Keys

```bash
# Generate new production key (different from dev!)
sudo openssl genrsa -out /opt/rafiki/keys/prod-key-$(uuidgen).pem 4096

# Set permissions
sudo chmod 600 /opt/rafiki/keys/*.pem

# Verify
sudo ls -la /opt/rafiki/keys/
```

**IMPORTANT:** Use a **different key** for production. Never use the dev key in production!

### Step 3: Configure Docker Compose

The `docker-compose.prod.yml` mounts the external keys directory:

```yaml
services:
  partner-service:
    volumes:
      - /opt/rafiki/keys:/app/zarf/keys:ro  # Read-only mount
```

### Step 4: Deploy

```bash
# Deploy using both compose files
docker compose -f docker-compose.yml -f docker-compose.prod.yml up -d --build
```

**How it works:**
- Docker image is built (zarf/keys copied from build context)
- Production volume mount **overrides** the image keys
- Container sees keys from `/opt/rafiki/keys` instead
- Keys are read-only inside container (`:ro`)

---

## Key Rotation (Production)

To rotate keys without downtime:

### Step 1: Generate New Key

```bash
# Generate new key with new ID
NEW_KID=$(uuidgen)
sudo openssl genrsa -out /opt/rafiki/keys/${NEW_KID}.pem 4096
sudo chmod 600 /opt/rafiki/keys/${NEW_KID}.pem
```

### Step 2: Keep Old Key Active

Don't delete the old key yet! Keep both:

```bash
sudo ls /opt/rafiki/keys/
# old-key-abc123.pem  <- Keep this
# new-key-def456.pem  <- New one
```

### Step 3: Restart Service

```bash
docker compose -f docker-compose.yml -f docker-compose.prod.yml restart partner-service
```

Service now loads **both keys**:
- Old tokens still validate (using old key)
- New tokens issued with new key

### Step 4: Update Clients

Update your login endpoint calls to use new kid:

```bash
# Old (still works)
GET /v1/auth/token/old-key-abc123

# New
GET /v1/auth/token/new-key-def456
```

### Step 5: Remove Old Key (After Expiry)

Wait for all old tokens to expire (48 hours), then:

```bash
sudo rm /opt/rafiki/keys/old-key-abc123.pem
docker compose -f docker-compose.yml -f docker-compose.prod.yml restart partner-service
```

---

## Security Best Practices

### File Permissions

```bash
# Keys directory
sudo chmod 700 /opt/rafiki/keys

# Key files
sudo chmod 600 /opt/rafiki/keys/*.pem

# Owned by root
sudo chown -R root:root /opt/rafiki/keys
```

### Backup Keys Securely

```bash
# Encrypt backup
sudo tar -czf - /opt/rafiki/keys | \
  gpg --symmetric --cipher-algo AES256 > rafiki-keys-backup.tar.gz.gpg

# Store in secure location (off-server)
# Use password manager for encryption password
```

### Never Commit Production Keys

```bash
# .gitignore already protects this
zarf/keys/*.pem
!zarf/keys/54bb2165-71e1-41a6-af3e-7da4a0e1e2c1.pem  # Only dev key allowed
```

---

## Verification

### Check Keys are Mounted

```bash
# List keys in running container
docker exec rafiki-service ls -la /app/zarf/keys/

# Should show keys from /opt/rafiki/keys (production)
# NOT from image's zarf/keys (development)
```

### Check Keys Loaded

```bash
# Check service logs
docker compose logs partner-service | grep "keys_loaded"

# Should see:
# "keys_loaded": 1 (or more if multiple keys)
```

### Test Authentication

```bash
# Get kid from filename
KID=$(sudo ls /opt/rafiki/keys/*.pem | head -1 | xargs basename | sed 's/.pem//')

# Test login
curl -X GET \
  -H "Authorization: Basic $(echo -n 'admin@rafiki.com:password' | base64)" \
  https://your-domain.com/v1/auth/token/${KID}

# Should return JWT token
```

---

## Troubleshooting

### Issue: "no authentication keys loaded"

**Cause:** Volume mount not working or keys directory empty

**Solution:**
```bash
# Check volume mount
docker inspect rafiki-service | grep -A 5 Mounts

# Check keys exist
sudo ls -la /opt/rafiki/keys/

# Verify .pem extension
```

### Issue: "key not found" for specific kid

**Cause:** kid in request doesn't match filename

**Solution:**
```bash
# List available kids
sudo ls /opt/rafiki/keys/*.pem | xargs basename | sed 's/.pem//'

# Use correct kid in request
GET /v1/auth/token/{correct-kid}
```

### Issue: Old tokens still work after key rotation

**Expected behavior!** Old key is still present. Wait for token expiry (48h) or remove old key.

---

## Architecture Comparison

### Local Development
```
Docker Build
  ↓
Copy zarf/keys/ → Image
  ↓
Container runs with keys IN image
```

### Production
```
Docker Build
  ↓
Copy zarf/keys/ → Image (ignored)
  ↓
Mount /opt/rafiki/keys → Container
  ↓
Container runs with keys FROM MOUNT (overrides image)
```

---

## Quick Reference

```bash
# LOCAL (development)
docker compose up -d --build

# PRODUCTION (with external keys)
docker compose -f docker-compose.yml -f docker-compose.prod.yml up -d --build

# Generate prod key
sudo openssl genrsa -out /opt/rafiki/keys/$(uuidgen).pem 4096

# Check loaded keys
docker compose logs partner-service | grep keys_loaded

# Rotate keys (keep both)
# 1. Generate new key
# 2. Restart service
# 3. Update clients
# 4. Remove old after 48h
```
