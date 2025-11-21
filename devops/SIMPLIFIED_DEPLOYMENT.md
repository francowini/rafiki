# Simplified Rafiki Deployment Guide

## ✅ What Changed

### Removed (Saves 512MB RAM!)
- ❌ **Grafana** - Removed completely
- ❌ **Tempo** - Removed completely
- 🎉 **Result**: Minimal production stack uses only ~320MB RAM instead of 832MB!

### Simplified
- ✅ Clearer .env configuration
- ✅ Diagnostic script for troubleshooting
- ✅ Comprehensive troubleshooting guide
- ✅ Same 3 docker-compose files (but simpler)

---

## 🚀 Quick Start

### Local Development (PostgreSQL in Docker)
```bash
# 1. Start services
docker compose up -d

# 2. Check health
curl http://localhost:3000/v1/readiness

# 3. View logs
docker compose logs -f partner-service
```

**Services Running**:
- `postgres` (256MB) - Database
- `partner-service` (192MB) - API
- **Total**: ~450MB RAM

---

### Production (Hetzner with PlanetScale)

#### One-Time Setup

**1. Configure PlanetScale Credentials**
```bash
ssh root@178.156.170.37
cd /opt/rafiki
sudo nano .env
```

Add these lines (replace with your actual PlanetScale values):
```bash
# External Database (PlanetScale)
PARTNER_DB_HOST=aws.connect.psdb.cloud
PARTNER_DB_PORT=3306
PARTNER_DB_USER=your_username
PARTNER_DB_PASSWORD=pscale_pw_xxxxxxxxxx
PARTNER_DB_NAME=your_database
PARTNER_DB_SSLMODE=require
PARTNER_DB_DISABLETLS=false

# Encryption Key (MUST be set, generate with: openssl rand -hex 32)
PARTNER_ENCRYPTION_KEY=your_64_char_hex_key_here

# CORS
PARTNER_WEB_CORSALLOWEDORIGINS=https://app.rafiki.lat,https://*.vercel.app
```

**2. Verify JWT Keys Exist**
```bash
ls -la /var/lib/rafiki/keys/

# Should show .pem files
# If not, run: sudo ./devops/setup-prod-keys.sh
```

**3. Test Direct Connection (Important!)**
```bash
# Replace USER, PASS, HOST, PORT, DB with your values
psql "postgresql://USER:PASS@HOST:PORT/DB?sslmode=require"

# If this works, proceed to deployment
# If this fails, fix PlanetScale config first
```

**4. Run Diagnostic**
```bash
sudo ./devops/diagnose-planetscale.sh
```

Fix any issues reported before deploying.

**5. Deploy**
```bash
sudo ./devops/deploy.sh
```

**Services Running** (Production):
- `partner-service` (192MB) - API
- `nginx` (64MB) - Reverse proxy
- `certbot` (~0MB) - SSL renewal (periodic)
- **Total**: ~256MB RAM (saves 576MB!)

---

## 🐛 Troubleshooting "Context Deadline Exceeded"

This is the error you're experiencing. Here's the fix:

### Most Likely Cause: Environment Variables Not Loaded

**Verify .env is configured**:
```bash
ssh root@178.156.170.37
cat /opt/rafiki/.env | grep PARTNER_DB_

# Should show:
# PARTNER_DB_HOST=aws.connect.psdb.cloud (not empty!)
# PARTNER_DB_PORT=3306 (or 5432)
# etc.
```

**If variables are missing**, add them as shown in step 1 above.

### Second Most Likely: Docker Not Reading .env

**Check what Docker sees**:
```bash
cd /opt/rafiki
docker compose --env-file .env config | grep PARTNER_DB_HOST

# Should show your PlanetScale host
# If it shows "postgres", env vars aren't loaded
```

**Fix**: The deploy.sh script should handle this automatically (line 47):
```bash
export $(grep -v '^#' "$ENV_FILE" | xargs)
```

### Third Most Likely: IP Not Whitelisted

**Get your Hetzner IP**:
```bash
curl ifconfig.me
# Example: 178.156.170.37
```

**Add to PlanetScale**:
1. Go to PlanetScale dashboard
2. Settings → Security
3. Add Hetzner IP: `178.156.170.37`
4. Save and wait 1-2 minutes

### Complete Diagnostic

Run the diagnostic script:
```bash
cd /opt/rafiki
sudo ./devops/diagnose-planetscale.sh
```

This will test:
- ✅ Environment variables set
- ✅ DNS resolution
- ✅ Network connectivity
- ✅ Direct psql connection
- ✅ Docker environment
- ✅ Container networking

**See**: [PLANETSCALE_TROUBLESHOOTING.md](PLANETSCALE_TROUBLESHOOTING.md) for complete guide.

---

## 📊 Resource Comparison

| Configuration | RAM | Services |
|--------------|-----|----------|
| **Old (Local DB + Observability)** | 1024MB | postgres, tempo, grafana, app, nginx |
| **Old (PlanetScale + Observability)** | 768MB | tempo, grafana, app, nginx |
| **New (Local DB)** | 450MB | postgres, app |
| **New (PlanetScale) ⭐** | 256MB | app, nginx |

**Savings**: 75% reduction in RAM usage! 🎉

---

## 🔍 Verification

After deploying, verify everything works:

```bash
# 1. Health check
curl http://localhost:3000/v1/readiness
# Expected: {"status":"ok"}

# 2. Check logs
docker compose logs -f partner-service
# Look for:
# - "database migrations completed"
# - "api router started"
# - NO errors about connection timeouts

# 3. Verify external DB is used
docker compose ps | grep postgres
# Should show: Exit 0 (not running)
# OR not appear at all

# 4. Check container environment
docker exec rafiki-service env | grep PARTNER_DB_HOST
# Should show your PlanetScale host

# 5. Create test user
./zarf/create-user.sh test@example.com password123 USER "Test User"
```

---

## 🚨 Emergency Rollback

If PlanetScale deployment fails, quickly rollback to local PostgreSQL:

```bash
ssh root@178.156.170.37
cd /opt/rafiki
sudo nano .env

# Comment out PARTNER_DB_* variables
# PARTNER_DB_HOST=...  →  # PARTNER_DB_HOST=...

# Or delete them entirely

# Redeploy
sudo ./devops/deploy.sh

# This will use local Docker PostgreSQL
```

---

## 📝 .env File Examples

### Local Development
```bash
# .env (local)
POSTGRES_DB=rafiki
POSTGRES_USER=rafiki
POSTGRES_PASSWORD=db
PARTNER_ENCRYPTION_KEY=6c6cb068bd296243d90d4ed387849c6670d4c6a8f02641a1939e7f5177666697
PARTNER_WEB_CORSALLOWEDORIGINS=http://localhost:3000,*
```

### Production (Hetzner with PlanetScale)
```bash
# .env (production - /opt/rafiki/.env on Hetzner)

# External Database (PlanetScale)
PARTNER_DB_HOST=aws.connect.psdb.cloud
PARTNER_DB_PORT=3306
PARTNER_DB_USER=rafiki_prod
PARTNER_DB_PASSWORD=pscale_pw_xxxxxxxxxxxxxxxxxxxxxxxxxx
PARTNER_DB_NAME=rafiki_main
PARTNER_DB_SSLMODE=require
PARTNER_DB_DISABLETLS=false

# Application
PARTNER_ENCRYPTION_KEY=YOUR_64_CHAR_HEX_KEY_FROM_PASSWORD_MANAGER
PARTNER_WEB_CORSALLOWEDORIGINS=https://app.rafiki.lat,https://*.vercel.app

# Optional: Fallback local DB (not used when PARTNER_DB_HOST is set)
POSTGRES_DB=rafiki
POSTGRES_USER=rafiki
POSTGRES_PASSWORD=secure_password_here
```

---

## 🎯 Next Steps

### To Fix Your Current Issue:

1. **SSH to Hetzner**
   ```bash
   ssh root@178.156.170.37
   cd /opt/rafiki
   ```

2. **Verify .env has PARTNER_DB_* variables**
   ```bash
   cat .env | grep PARTNER_DB_
   ```

3. **If missing, add them** (see Production example above)

4. **Run diagnostic**
   ```bash
   sudo ./devops/diagnose-planetscale.sh
   ```

5. **Deploy**
   ```bash
   sudo ./devops/deploy.sh
   ```

6. **Watch logs**
   ```bash
   docker compose logs -f partner-service
   ```

7. **If still fails**, check [PLANETSCALE_TROUBLESHOOTING.md](PLANETSCALE_TROUBLESHOOTING.md)

---

## 📚 Documentation

- **Troubleshooting**: [PLANETSCALE_TROUBLESHOOTING.md](PLANETSCALE_TROUBLESHOOTING.md)
- **Complete Deployment Guide**: [DEPLOYMENT_GUIDE.md](DEPLOYMENT_GUIDE.md)
- **External Database Setup**: [EXTERNAL_DATABASE_GUIDE.md](EXTERNAL_DATABASE_GUIDE.md)
- **Backend Development**: [BACKEND_DEVELOPMENT.md](BACKEND_DEVELOPMENT.md)

---

## 🆘 Getting Help

If still stuck:

1. Run diagnostic and share output:
   ```bash
   sudo ./devops/diagnose-planetscale.sh > diagnostic.log 2>&1
   cat diagnostic.log
   ```

2. Share container logs:
   ```bash
   docker compose logs --tail=50 partner-service
   ```

3. Share .env (with passwords redacted):
   ```bash
   cat .env | sed 's/PASSWORD=.*/PASSWORD=***REDACTED***/'
   ```

---

## ✅ Success Indicators

You'll know it's working when:

- ✅ Health check returns `{"status":"ok"}`
- ✅ Logs show "database migrations completed"
- ✅ No "context deadline exceeded" errors
- ✅ Local postgres NOT running (`docker compose ps`)
- ✅ Can create users: `./zarf/create-user.sh ...`
- ✅ API accessible via https://api.rafiki.lat

**RAM usage** on Hetzner should drop from ~1GB to ~256MB! 🎉
