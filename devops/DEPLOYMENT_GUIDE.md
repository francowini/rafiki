# Rafiki Deployment Guide

Complete operational guide for deploying and running the Rafiki service on production infrastructure.

**Target Platform:** Hetzner servers (backend) + Vercel (frontend)
**Architecture:** Go backend API + Next.js frontend + PostgreSQL database

---

## Table of Contents

1. [Quick Start](#quick-start)
2. [One-Time Server Setup](#one-time-server-setup)
3. [Database Configuration](#database-configuration)
4. [Regular Deployment](#regular-deployment)
5. [Frontend Deployment (Vercel)](#frontend-deployment-vercel)
6. [Verification & Health Checks](#verification--health-checks)
7. [Common Operations](#common-operations)
8. [Troubleshooting](#troubleshooting)
9. [Security](#security)
10. [Rollback Procedures](#rollback-procedures)
11. [Quick Reference](#quick-reference)

---

## Quick Start

### Local Development
```bash
# Start all services
docker compose up -d

# Check health
curl http://localhost:3000/v1/readiness

# View logs
docker compose logs -f partner-service
```

### Production Deployment
```bash
# From local machine (recommended)
make deploy

# Or on server
ssh root@178.156.170.37
cd /opt/rafiki
sudo ./devops/deploy.sh
```

---

## One-Time Server Setup

**⚠️ IMPORTANT:** These steps should ONLY be done once when setting up a new server.

### 1. Server Preparation

```bash
# SSH to your Hetzner server
ssh root@178.156.170.37

# Update system
apt update && apt upgrade -y

# Install Docker
curl -fsSL https://get.docker.com | sh

# Install Docker Compose
apt install docker-compose-plugin -y

# Verify installation
docker --version
docker compose version
```

### 2. Clone Repository

```bash
# Create project directory
mkdir -p /opt/rafiki

# Clone repository (first time only)
cd /opt
git clone <your-repo-url> rafiki

cd /opt/rafiki
```

### 3. Generate JWT Keys (CRITICAL)

```bash
cd /opt/rafiki

# Generate production keys
sudo ./devops/setup-prod-keys.sh

# IMPORTANT: Save the output!
# You'll see: "Key ID (kid): xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"
# Keep this key ID safe - you'll need it for user authentication

# Verify keys were created
ls -la /opt/rafiki/keys/
# You should see a .pem file
```

**⚠️ WARNING:**
- JWT keys should NEVER be regenerated unless absolutely necessary
- Regenerating keys will invalidate ALL existing user tokens
- Backup your keys: `sudo tar -czf rafiki-keys-backup.tar.gz /opt/rafiki/keys/`

### 4. Firewall Configuration

**Option A: Hetzner Cloud Firewall (Recommended)**

1. Go to https://console.hetzner.cloud/
2. Select your project → **Firewalls** → **Create Firewall**
3. Name it `rafiki-production`
4. Add inbound rules:

```
SSH:   Port 22  | Source: Your IP only (e.g., 200.123.45.67/32)
HTTP:  Port 80  | Source: 0.0.0.0/0 (for Let's Encrypt)
HTTPS: Port 443 | Source: 0.0.0.0/0 (public API access)
```

5. Apply firewall to your server

**Option B: UFW (Alternative)**

```bash
# Allow SSH (your IP only - replace with your actual IP)
ufw allow from YOUR_IP_ADDRESS to any port 22 proto tcp

# Allow HTTP (for Let's Encrypt challenges)
ufw allow 80/tcp

# Allow HTTPS
ufw allow 443/tcp

# Enable firewall
ufw enable

# Verify
ufw status
```

### 5. SSL Certificate Setup (Production Only)

**Prerequisites:**
- Domain `api.rafiki.lat` must point to your server IP (178.156.170.37)
- Firewall allows ports 80 and 443

**Initial Certificate Acquisition:**

```bash
cd /opt/rafiki

# Start services without nginx first
docker compose up -d

# Run certbot to get certificate
docker compose --profile production run --rm certbot certonly \
  --webroot \
  --webroot-path=/var/www/certbot \
  -d api.rafiki.lat \
  --email your-email@example.com \
  --agree-tos \
  --no-eff-email

# Verify certificate was created
ls -la /opt/rafiki/certbot/conf/live/api.rafiki.lat/

# Stop services and restart with nginx
docker compose down
docker compose --profile production up -d --build
```

**Auto-Renewal:**
- Certbot container runs continuously
- Automatically attempts renewal every 12 hours
- Certificates renew when <30 days until expiration

**Test Renewal:**
```bash
docker compose --profile production exec certbot certbot renew --dry-run
```

---

## Database Configuration

### Option A: Local PostgreSQL (Docker)

**Default configuration** - no additional setup required.

Create `.env` file:

```bash
cd /opt/rafiki

cat > .env << 'EOF'
# Database Configuration
POSTGRES_DB=rafiki
POSTGRES_USER=rafiki
POSTGRES_PASSWORD=<CHANGE_ME_TO_STRONG_PASSWORD>

# CORS Configuration
PARTNER_WEB_CORSALLOWEDORIGINS=https://app.rafiki.lat,https://*.vercel.app
EOF

# Secure the file
chmod 600 .env
```

**⚠️ IMPORTANT:** Change `<CHANGE_ME_TO_STRONG_PASSWORD>` to a strong password!

**Resources:** ~256MB RAM for PostgreSQL container

### Option B: External Database (PlanetScale, Neon, RDS)

**Saves ~256MB RAM** by not running local PostgreSQL.

Add to `.env`:

```bash
# External Database Configuration
PARTNER_DB_HOST=your-db-host.example.com  # Triggers external DB mode
PARTNER_DB_PORT=5432                      # Or 3306 for MySQL-compatible
PARTNER_DB_USER=your_username
PARTNER_DB_PASSWORD=your_password
PARTNER_DB_NAME=your_database_name
PARTNER_DB_SSLMODE=require                # Options: disable, require, verify-ca
PARTNER_DB_DISABLETLS=false

# CORS Configuration
PARTNER_WEB_CORSALLOWEDORIGINS=https://app.rafiki.lat,https://*.vercel.app
```

**Database Permissions Required:**

Your database user needs DDL + DML permissions (migrations run on app startup):

```sql
GRANT CREATE ON SCHEMA public TO your_username;
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO your_username;
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO your_username;
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON TABLES TO your_username;
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON SEQUENCES TO your_username;
```

**Troubleshooting External Database:**

If you encounter "context deadline exceeded" errors:

```bash
# Run diagnostic script
cd /opt/rafiki
sudo ./devops/diagnose-planetscale.sh
```

This checks:
- Environment variables loaded correctly
- DNS resolution works
- Network connectivity to database
- Direct connection with psql
- Docker container can reach database

---

## Regular Deployment

**For deploying code changes, updates, or bug fixes.**

### Option A: From Local Machine (Recommended)

```bash
# From your local rafiki directory

# 1. Commit and push your changes
git add .
git commit -m "Your changes"
git push origin main

# 2. Deploy to production (ONE COMMAND!)
make deploy
```

The `make deploy` command automatically:
1. SSH to the server
2. Pull latest code (`git pull origin main`)
3. Run deployment script
4. Build and restart services
5. Run database migrations
6. Verify health checks
7. Show deployment status

**Total time:** ~2-3 minutes

### Option B: Directly on Server

```bash
# SSH to server
ssh root@178.156.170.37

# Go to project directory
cd /opt/rafiki

# Pull latest changes
git pull origin main

# Run deployment
sudo ./devops/deploy.sh
```

### What Happens During Deployment

The `deploy.sh` script automatically:

1. ✅ Detects database configuration (local vs external)
2. ✅ Stops existing containers
3. ✅ Pulls latest git changes
4. ✅ Builds new Docker images
5. ✅ Starts services with correct profile (production/external-db)
6. ✅ **Runs database migrations** (idempotent - safe every time)
7. ✅ Waits for health checks (max 60 seconds)
8. ✅ Verifies authentication (checks JWT keys loaded)
9. ✅ Shows deployment summary

**Safe operations (run every deployment):**
- ✅ Database migrations (idempotent - won't duplicate data)
- ✅ Container rebuild and restart
- ✅ Health checks
- ✅ Key verification (doesn't regenerate keys, just checks they exist)

**NOT done automatically (manual only):**
- ❌ JWT key generation (manual via `setup-prod-keys.sh`)
- ❌ User creation (manual via `create-user.sh`)
- ❌ Database reset/deletion
- ❌ Environment variable changes (requires manual .env edit)

### First Deployment (After One-Time Setup)

```bash
cd /opt/rafiki

# Run first deployment (creates database, runs migrations)
sudo ./devops/deploy.sh

# Create ADMIN user (DO ONCE)
./zarf/create-user.sh admin@rafiki.lat <password> ADMIN "Admin User"
```

---

## Frontend Deployment (Vercel)

### Prerequisites

- Vercel account (free tier sufficient)
- Vercel CLI installed: `npm install -g vercel`
- Git repository connected to Vercel

### One-Time Setup

```bash
# Install Vercel CLI
npm install -g vercel

# Login to Vercel
vercel login

# Link project
cd /Users/francowini/Documents/rafiki/frontend
vercel link
```

### Configure Environment Variables

**Via Vercel CLI:**

```bash
# Production
vercel env add NEXT_PUBLIC_API_URL production
# Enter: https://api.rafiki.lat

vercel env add NEXT_PUBLIC_ENV production
# Enter: production

# Preview
vercel env add NEXT_PUBLIC_API_URL preview
# Enter: https://api.rafiki.lat

vercel env add NEXT_PUBLIC_ENV preview
# Enter: preview
```

**Or via Vercel Dashboard:**

1. Go to https://vercel.com/dashboard
2. Select project → **Settings** → **Environment Variables**
3. Add:
   - `NEXT_PUBLIC_API_URL` = `https://api.rafiki.lat` (production + preview)
   - `NEXT_PUBLIC_ENV` = `production` (production) / `preview` (preview)

### Deploy

```bash
cd frontend

# Deploy to preview
vercel

# Deploy to production
vercel --prod
```

### Custom Domain

1. Go to **Project → Settings → Domains** in Vercel Dashboard
2. Add domain: `app.rafiki.lat`
3. Configure DNS (at your DNS provider):

```
Type:  CNAME
Name:  app
Value: cname.vercel-dns.com
TTL:   3600
```

4. Wait for DNS propagation (5-30 minutes)
5. Verify in Vercel dashboard - SSL certificate automatically provisioned

### CORS Configuration

Backend is pre-configured to allow:
- `https://app.rafiki.lat` (production frontend)
- `https://*.vercel.app` (Vercel preview deployments)

**No action needed** - CORS already configured in `docker-compose.yml`

---

## Verification & Health Checks

### Backend Health

```bash
# Readiness check (is service ready to accept traffic?)
curl https://api.rafiki.lat/v1/readiness

# Liveness check (is service running?)
curl https://api.rafiki.lat/v1/liveness

# Expected response: {"status":"ok"}
```

### Frontend Health

```bash
# Test frontend loads
curl -I https://app.rafiki.lat

# Expected: 200 OK
```

### Database Health

```bash
# SSH to server
ssh root@178.156.170.37

# Check PostgreSQL (if using local DB)
docker exec -it rafiki-postgres pg_isready -U rafiki -d rafiki

# Check database tables
docker exec -it rafiki-postgres psql -U rafiki -d rafiki -c \
  "SELECT tablename FROM pg_tables WHERE schemaname='public';"

# Expected tables: users, thinks, darwin_migrations
```

### End-to-End Authentication Test

```bash
# 1. Get JWT token
curl -i -u your-email@example.com:your-password \
  https://api.rafiki.lat/v1/auth/token/<your-kid>

# Expected: 200 OK with JWT token

# 2. Extract token from response
TOKEN="<paste_token_here>"

# 3. Test authenticated request
curl -i -H "Authorization: Bearer $TOKEN" \
  https://api.rafiki.lat/v1/thinks

# Expected: 200 OK with thinks array
```

---

## Common Operations

### View Production Logs

```bash
# From local machine
make deploy-logs

# Or on server
cd /opt/rafiki
docker compose logs -f partner-service

# View last 100 lines
docker compose logs --tail=100 partner-service
```

### Check Service Status

```bash
# From local machine
make deploy-status

# Or on server
cd /opt/rafiki
docker compose ps

# Check specific service
docker compose ps partner-service
```

### Restart Service

```bash
# From local machine
make deploy-restart

# Or on server
cd /opt/rafiki
docker compose restart partner-service

# Restart all services
docker compose restart
```

### Create Additional Users

```bash
# On server only
cd /opt/rafiki
./zarf/create-user.sh user@example.com password USER "User Name"

# Roles: ADMIN or USER
```

### Access Database

```bash
# From local machine
make db-shell-prod

# Or on server
cd /opt/rafiki
docker exec -it rafiki-postgres psql -U rafiki -d rafiki

# Run SQL query
docker exec -it rafiki-postgres psql -U rafiki -d rafiki -c "SELECT * FROM users;"
```

### Database Backup

```bash
# On server
cd /opt/rafiki

# Create backup
docker exec rafiki-postgres pg_dump -U rafiki -d rafiki | \
  gzip > backups/rafiki_backup_$(date +%Y%m%d_%H%M%S).sql.gz

# Restore backup
gunzip < backups/rafiki_backup_YYYYMMDD_HHMMSS.sql.gz | \
  docker exec -i rafiki-postgres psql -U rafiki -d rafiki
```

### Docker Cleanup

**Safe cleanup (preserves database):**
```bash
cd /opt/rafiki

# Stop containers (keeps volumes)
docker compose down

# Remove unused networks
docker network prune -f

# Restart
sudo ./devops/deploy.sh
```

**⚠️ DANGER: Nuclear cleanup (DELETES DATABASE!):**
```bash
# DO NOT run this on production unless you want to delete everything!
docker compose down -v  # The -v flag deletes volumes (DATABASE!)
```

---

## Troubleshooting

### Health Check Fails

```bash
# Check container status
docker compose ps

# Check partner-service logs
docker compose logs partner-service

# Check if database is healthy
docker compose logs postgres

# Manual health check
curl http://localhost:3000/v1/readiness
```

### JWT Keys Not Loading

```bash
# Verify keys exist
ls -la /opt/rafiki/keys/

# Check if keys are mounted in container
docker compose exec partner-service ls -la /app/zarf/keys/

# Check service logs for "keys_loaded"
docker compose logs partner-service | grep "keys_loaded"
# Expected: "keys_loaded": 1

# If missing, keys weren't created - run setup-prod-keys.sh
```

### Database Connection Fails

**Local PostgreSQL:**
```bash
# Check .env file exists
cat /opt/rafiki/.env

# Verify postgres container is running
docker compose ps postgres

# Check postgres logs
docker compose logs postgres

# Test database connection
docker exec -it rafiki-postgres psql -U rafiki -d rafiki -c '\l'
```

**External Database (PlanetScale/Neon):**
```bash
# Run diagnostic script
cd /opt/rafiki
sudo ./devops/diagnose-planetscale.sh

# This checks:
# - Environment variables set correctly
# - DNS resolution works
# - Network connectivity to database
# - Direct psql connection works
# - Docker environment configured correctly

# Common issues:
# 1. PARTNER_DB_HOST not set in .env
# 2. IP not whitelisted in database provider
# 3. Incorrect credentials
# 4. SSL/TLS configuration mismatch
```

### Container Won't Start

```bash
# Check logs
docker compose logs partner-service

# Check for port conflicts
netstat -tulpn | grep -E '3000|3010|5432'

# Rebuild from scratch
docker compose down
docker compose build --no-cache
docker compose up -d
```

### CORS Errors

**Symptoms:** "Access to fetch blocked by CORS policy" in browser console

**Diagnosis:**
```bash
# Check CORS configuration
ssh root@178.156.170.37
cd /opt/rafiki
grep CORSALLOWEDORIGINS docker-compose.yml

# Expected: https://app.rafiki.lat,https://*.vercel.app
```

**Solution:**
```bash
# If CORS config is wrong, fix docker-compose.yml
# Then redeploy:
sudo ./devops/deploy.sh
```

### Nginx 502 Bad Gateway

```bash
# Check backend is running
docker compose ps partner-service

# Check nginx logs
docker compose logs nginx

# Check backend logs
docker compose logs partner-service

# Test backend directly (bypass nginx)
curl http://localhost:3000/v1/liveness

# If direct access works, restart nginx
docker compose restart nginx
```

### SSL Certificate Issues

```bash
# Check certificate expiration
openssl x509 -enddate -noout -in /opt/rafiki/certbot/conf/live/api.rafiki.lat/fullchain.pem

# Check certbot logs
docker compose logs certbot

# Test renewal (dry run)
docker compose --profile production exec certbot certbot renew --dry-run

# Force renewal
docker compose --profile production exec certbot certbot renew --force-renewal

# Restart nginx to load new cert
docker compose restart nginx
```

---

## Security

### Key Security Measures

1. **JWT Keys:**
   - Stored in `/opt/rafiki/keys/`
   - Permissions: 600 (owner read/write only)
   - Backup regularly
   - Never commit to git

2. **Database Password:**
   - Stored in `/opt/rafiki/.env`
   - Permissions: 600
   - Use strong password (20+ characters)

3. **Firewall:**
   - Only ports 22, 80, 443 open to internet
   - Port 22 (SSH) restricted to your IP only
   - Internal ports (3000, 3010, 5432) NOT exposed

4. **SSL/TLS:**
   - Let's Encrypt certificates (auto-renewal)
   - TLS 1.2+ only
   - Strong cipher suites
   - HSTS enabled (max-age=31536000)

5. **Security Headers:**
   - Strict-Transport-Security (HSTS)
   - X-Content-Type-Options: nosniff
   - X-Frame-Options: DENY
   - X-XSS-Protection: 1; mode=block

### Additional Hardening (Recommended)

```bash
# Disable SSH password authentication (use keys only)
ssh root@178.156.170.37
sed -i 's/#PasswordAuthentication yes/PasswordAuthentication no/' /etc/ssh/sshd_config
systemctl restart sshd

# Install Fail2Ban
apt install fail2ban -y
systemctl enable fail2ban
systemctl start fail2ban

# Enable automatic security updates
apt install unattended-upgrades -y
dpkg-reconfigure -plow unattended-upgrades
```

### Backup JWT Keys

```bash
# Create encrypted backup
cd /opt/rafiki
tar -czf - keys/ | gpg --symmetric --cipher-algo AES256 > keys-backup-$(date +%Y%m%d).tar.gz.gpg

# Store backup securely (NOT on same server!)
# Transfer to local machine or cloud storage
```

---

## Rollback Procedures

### Backend Rollback

**Scenario 1: Recent deployment broke something**

```bash
# SSH to server
ssh root@178.156.170.37
cd /opt/rafiki

# Check git log
git log --oneline -5

# Rollback to previous commit
git reset --hard HEAD~1

# Redeploy
sudo ./devops/deploy.sh

# Time estimate: 3-5 minutes
```

**Scenario 2: Database migration failed**

```bash
# Check migration status
docker exec -it rafiki-postgres psql -U rafiki -d rafiki -c \
  "SELECT version, description, applied_at FROM darwin_migrations ORDER BY applied_at DESC LIMIT 5;"

# Restore database backup
gunzip < /opt/rafiki/backups/rafiki_backup_YYYYMMDD_HHMMSS.sql.gz | \
  docker exec -i rafiki-postgres psql -U rafiki -d rafiki

# Restart services
docker compose restart partner-service
```

**Scenario 3: JWT keys corrupted**

```bash
# Restore from backup
gpg -d keys-backup-YYYYMMDD.tar.gz.gpg | tar -xzf -

# Restart service
docker compose restart partner-service

# Impact: All users will need to re-login (tokens invalidated)
```

### Frontend Rollback

**Option A: Vercel Dashboard (Easiest)**

1. Go to https://vercel.com/dashboard
2. Select project → **Deployments**
3. Find last working deployment
4. Click **⋯** → **Promote to Production**
5. Time: 30 seconds, no downtime

**Option B: Vercel CLI**

```bash
cd frontend
vercel rollback
```

**Option C: Git Revert**

```bash
cd /Users/francowini/Documents/rafiki

# Revert last commit
git revert HEAD
git push origin main

# Vercel auto-deploys reverted version
```

---

## Quick Reference

### Deployment Commands

| Task | Command |
|------|---------|
| Deploy from local | `make deploy` |
| Deploy on server | `cd /opt/rafiki && sudo ./devops/deploy.sh` |
| View logs | `make deploy-logs` or `docker compose logs -f partner-service` |
| Check status | `make deploy-status` or `docker compose ps` |
| Health check | `curl https://api.rafiki.lat/v1/readiness` |
| Restart service | `docker compose restart partner-service` |
| SSH to server | `ssh root@178.156.170.37` |
| Create user | `./zarf/create-user.sh <email> <password> <ROLE> <name>` |
| Database shell | `docker exec -it rafiki-postgres psql -U rafiki -d rafiki` |

### Service Endpoints

| Endpoint | URL | Purpose |
|----------|-----|---------|
| Backend API (production) | https://api.rafiki.lat | Public API with SSL |
| Backend API (direct) | http://localhost:3000 | Internal access only |
| Frontend (production) | https://app.rafiki.lat | User-facing application |
| Debug/Metrics | http://localhost:3010 | Internal debugging only |

### Environment Variables

**Backend (.env on server):**
```bash
# Local Database
POSTGRES_DB=rafiki
POSTGRES_USER=rafiki
POSTGRES_PASSWORD=<strong_password>

# OR External Database
PARTNER_DB_HOST=your-db-host.example.com
PARTNER_DB_PORT=5432
PARTNER_DB_USER=your_username
PARTNER_DB_PASSWORD=your_password
PARTNER_DB_NAME=your_database_name
PARTNER_DB_SSLMODE=require
PARTNER_DB_DISABLETLS=false

# CORS
PARTNER_WEB_CORSALLOWEDORIGINS=https://app.rafiki.lat,https://*.vercel.app
```

**Frontend (Vercel Dashboard):**
```bash
NEXT_PUBLIC_API_URL=https://api.rafiki.lat
NEXT_PUBLIC_ENV=production
```

### Resource Usage

| Configuration | RAM | Services |
|--------------|-----|----------|
| Local DB + Production | ~512MB | postgres, partner-service, nginx |
| External DB + Production | ~256MB | partner-service, nginx |
| Development | ~450MB | postgres, partner-service |

---

## Additional Notes

### What NOT to Do

**❌ DO NOT regenerate JWT keys** (unless absolutely necessary)
```bash
# DON'T run this again after initial setup!
# sudo ./devops/setup-prod-keys.sh
```
**Why:** Regenerating keys invalidates all existing user sessions.

**❌ DO NOT recreate ADMIN user** (with same email)
```bash
# DON'T run this multiple times with same email!
# ./zarf/create-user.sh admin@rafiki.lat ...
```
**Why:** Email is unique - command will fail.

**❌ DO NOT delete Docker volumes accidentally**
```bash
# NEVER run this on production!
# docker compose down -v  # The -v flag deletes volumes (DATABASE!)
```
**Why:** This deletes your entire database!

Safe shutdown:
```bash
docker compose down  # Safe - keeps database intact
```

---

**Last Updated:** 2025-11-22
**Server:** Hetzner CPX11 @ 178.156.170.37
**Repository:** /opt/rafiki
**Documentation:** See `/docs/` for implementation guides
