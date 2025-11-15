# Authentication Feature - Comprehensive Deployment Guide

**Status:** Backend Deployed | Frontend Ready for Deployment
**Backend Server:** Hetzner CPX11 (178.156.170.37)
**Frontend Platform:** Vercel
**Last Updated:** 2025-11-15

---

## 1. Executive Summary

### Current Deployment Status

| Component | Status | URL | Notes |
|-----------|--------|-----|-------|
| Backend API | ✅ **DEPLOYED** | https://api.rafiki.lat | Running on Hetzner with SSL |
| Authentication | ✅ **DEPLOYED** | `/v1/auth/token/:kid` | JWT-based auth active |
| Database | ✅ **DEPLOYED** | Internal (PostgreSQL 18) | Users table created |
| CORS | ✅ **CONFIGURED** | - | Allows app.rafiki.lat + Vercel |
| SSL/TLS | ✅ **CONFIGURED** | Let's Encrypt | Auto-renewal enabled |
| Frontend | ⏳ **PENDING** | app.rafiki.lat (pending) | Ready to deploy to Vercel |

### What Needs to Be Done

**High Priority (Before Public Launch):**
1. Deploy frontend to Vercel production
2. Configure custom domain (app.rafiki.lat)
3. Verify end-to-end authentication flow
4. Test CORS from production frontend
5. Set up basic monitoring
6. Configure rate limiting (already done in nginx)
7. Create backup strategy

**Medium Priority (First Week):**
1. Set up error tracking (Sentry/LogRocket)
2. Configure uptime monitoring (UptimeRobot)
3. Implement frontend error boundaries
4. Add request logging
5. Document rollback procedures

**Low Priority (Nice to Have):**
1. Advanced analytics
2. Performance monitoring
3. Automated testing pipeline
4. Database backups automation

---

## 2. Pre-Deployment Checklist

### Backend Verification ✅

Run these commands to verify backend is ready:

```bash
# SSH to production server
ssh root@178.156.170.37

# Check services are running
cd /opt/rafiki
docker compose --profile production ps

# Expected output: All services "Up" and "healthy"
```

```bash
# Verify health checks
curl https://api.rafiki.lat/v1/readiness
curl https://api.rafiki.lat/v1/liveness

# Expected: Both return 200 OK with JSON
```

```bash
# Verify JWT keys exist
ls -la /opt/rafiki/keys/

# Expected: At least one .pem file (e.g., d8859c9d-f7f8-4f25-98e1-7b8951281d1c.pem)
```

```bash
# Check admin user exists
docker exec -it rafiki-postgres psql -U rafiki -d rafiki -c \
  "SELECT user_id, email, roles, enabled FROM users WHERE 'ADMIN' = ANY(roles);"

# Expected: At least one admin user record
```

```bash
# Verify database migrations
docker exec -it rafiki-postgres psql -U rafiki -d rafiki -c \
  "SELECT tablename FROM pg_tables WHERE schemaname='public';"

# Expected: users, thinks, darwin_migrations tables
```

### Frontend Verification ⏳

```bash
# Test local build
cd /Users/francowini/Documents/rafiki/frontend
npm run build

# Expected: Build succeeds with no errors
```

```bash
# Test TypeScript
npm run type-check  # Or: npx tsc --noEmit

# Expected: No type errors
```

```bash
# Verify environment variables template
cat .env.example  # Should list all required variables

# Expected variables:
# - NEXT_PUBLIC_API_URL
# - NEXT_PUBLIC_ENV
```

### Infrastructure Verification ✅

```bash
# Test CORS from Vercel preview domain
curl -i -X OPTIONS https://api.rafiki.lat/v1/thinks \
  -H "Origin: https://rafiki.vercel.app" \
  -H "Access-Control-Request-Method: GET"

# Expected headers:
# Access-Control-Allow-Origin: https://rafiki.vercel.app
# Access-Control-Allow-Methods: GET, POST, PUT, DELETE, OPTIONS
```

```bash
# Test SSL certificate
curl -vI https://api.rafiki.lat/v1/liveness 2>&1 | grep -A5 "SSL certificate"

# Expected: Valid Let's Encrypt certificate
```

```bash
# Verify firewall (from local machine - should timeout)
timeout 5 curl http://178.156.170.37:3000 || echo "Good - internal port blocked"

# Expected: Timeout (internal ports not exposed)
```

---

## 3. Architecture Diagram

### Request Flow

```
┌─────────────────────────────────────────────────────────────────────┐
│                          END USER (Browser)                          │
└────────────────────────────┬────────────────────────────────────────┘
                             │
                             │ HTTPS
                             ▼
┌─────────────────────────────────────────────────────────────────────┐
│                    VERCEL EDGE NETWORK (CDN)                         │
│                                                                       │
│  Domain: app.rafiki.lat                                              │
│  Service: Next.js 14+ Frontend                                       │
│  Features:                                                            │
│    - Server-side rendering (SSR)                                     │
│    - Static optimization                                              │
│    - Edge caching                                                     │
│    - Automatic HTTPS                                                  │
└────────────────────────────┬────────────────────────────────────────┘
                             │
                             │ HTTPS API Calls
                             │ (CORS-enabled)
                             ▼
┌─────────────────────────────────────────────────────────────────────┐
│                NGINX REVERSE PROXY (Hetzner Server)                  │
│                                                                       │
│  Domain: api.rafiki.lat                                              │
│  IP: 178.156.170.37                                                  │
│  Ports: 80 (HTTP) → 443 (HTTPS redirect)                             │
│         443 (HTTPS/TLS)                                               │
│  Features:                                                            │
│    - SSL/TLS termination (Let's Encrypt)                             │
│    - Rate limiting (10 req/s + burst 20)                             │
│    - Security headers (HSTS, CSP, etc.)                              │
│    - Automatic SSL renewal (certbot)                                 │
└────────────────────────────┬────────────────────────────────────────┘
                             │
                             │ HTTP (internal network)
                             ▼
┌─────────────────────────────────────────────────────────────────────┐
│                GO BACKEND API (Docker Container)                     │
│                                                                       │
│  Service: rafiki-service                                             │
│  Port: 3000 (internal only)                                          │
│  Features:                                                            │
│    - JWT authentication                                               │
│    - CORS middleware                                                  │
│    - OpenTelemetry tracing                                            │
│    - Health checks                                                    │
│                                                                       │
│  Endpoints:                                                           │
│    - GET  /v1/auth/token/:kid (login)                                │
│    - GET  /v1/thinks (list - requires auth)                          │
│    - POST /v1/thinks (create - requires auth)                        │
└────────────────────────────┬────────────────────────────────────────┘
                             │
                             │ PostgreSQL protocol
                             ▼
┌─────────────────────────────────────────────────────────────────────┐
│                POSTGRESQL 18 DATABASE (Docker Volume)                │
│                                                                       │
│  Container: rafiki-postgres                                          │
│  Port: 5432 (internal only)                                          │
│  Tables:                                                              │
│    - users (authentication)                                           │
│    - thinks (user data)                                               │
│    - darwin_migrations (schema versioning)                            │
└─────────────────────────────────────────────────────────────────────┘
```

### Authentication Flow (JWT)

```
┌──────────┐                                              ┌──────────┐
│ Frontend │                                              │ Backend  │
│ (Vercel) │                                              │(Hetzner) │
└────┬─────┘                                              └────┬─────┘
     │                                                          │
     │  1. Login Request (Basic Auth)                          │
     │  GET /v1/auth/token/:kid                                │
     │  Authorization: Basic base64(email:password)            │
     ├────────────────────────────────────────────────────────>│
     │                                                          │
     │                                     2. Validate credentials
     │                                        (bcrypt password hash)
     │                                                          │
     │                                     3. Generate JWT token
     │                                        (signed with RSA key)
     │                                                          │
     │  4. Return JWT token                                    │
     │  { "token": "eyJhbGci...", "kid": "..." }               │
     │<────────────────────────────────────────────────────────┤
     │                                                          │
     │  5. Store token (localStorage/sessionStorage)           │
     │                                                          │
     │  6. Authenticated Request                               │
     │  GET /v1/thinks                                         │
     │  Authorization: Bearer eyJhbGci...                      │
     ├────────────────────────────────────────────────────────>│
     │                                                          │
     │                                     7. Validate JWT signature
     │                                        (verify with RSA public key)
     │                                                          │
     │                                     8. Extract user claims
     │                                        (user_id, roles, etc.)
     │                                                          │
     │  9. Return protected data                               │
     │  { "thinks": [...] }                                    │
     │<────────────────────────────────────────────────────────┤
     │                                                          │
```

### CORS Configuration Flow

```
Browser (app.rafiki.lat)
    │
    │ Preflight: OPTIONS /v1/thinks
    │ Origin: https://app.rafiki.lat
    │ Access-Control-Request-Method: GET
    │ Access-Control-Request-Headers: Authorization
    │
    ▼
NGINX (api.rafiki.lat)
    │
    │ (No CORS headers - proxies to backend)
    │
    ▼
Go Backend API
    │
    │ CORS Middleware checks:
    │   - Is origin in allowed list?
    │     ✓ https://app.rafiki.lat
    │     ✓ https://*.vercel.app
    │
    │ Returns headers:
    │   Access-Control-Allow-Origin: https://app.rafiki.lat
    │   Access-Control-Allow-Methods: GET, POST, PUT, DELETE, OPTIONS
    │   Access-Control-Allow-Headers: Content-Type, Authorization
    │   Access-Control-Allow-Credentials: true
    │   Access-Control-Max-Age: 3600
    │
    ▼
Browser receives CORS approval → Makes actual request
```

---

## 4. Backend Deployment (Already Done ✅)

### How It Was Deployed

```bash
# 1. One-time setup (already completed)
ssh root@178.156.170.37
cd /opt/rafiki

# Generated JWT keys
sudo ./devops/setup-prod-keys.sh
# Output: Key ID (kid): d8859c9d-f7f8-4f25-98e1-7b8951281d1c

# Created .env file with database password
cat > .env << 'EOF'
POSTGRES_DB=rafiki
POSTGRES_USER=rafiki
POSTGRES_PASSWORD=<STRONG_PASSWORD>
EOF

# 2. Deployed services
sudo ./devops/deploy.sh

# 3. Created admin user
./zarf/create-user.sh francoolsson1995@gmail.com Sake2021 ADMIN "Franco Winiarski"
```

### Verification Commands

```bash
# Test authentication endpoint
curl -i -u francoolsson1995@gmail.com:Sake2021 \
  https://api.rafiki.lat/v1/auth/token/d8859c9d-f7f8-4f25-98e1-7b8951281d1c

# Expected response (200 OK):
{
  "token": "eyJhbGciOiJSUzI1NiIsImtpZCI6ImQ4ODU5YzlkLWY3ZjgtNGYyNS05OGUxLTdiODk1MTI4MWQxYyIsInR5cCI6IkpXVCJ9...",
  "kid": "d8859c9d-f7f8-4f25-98e1-7b8951281d1c"
}
```

```bash
# Test protected endpoint (without token - should fail)
curl -i https://api.rafiki.lat/v1/thinks

# Expected response (401 Unauthorized):
{
  "error": "unauthorized - missing or invalid authorization header"
}
```

```bash
# Test protected endpoint (with token - should succeed)
TOKEN="<paste_token_from_login>"

curl -i -H "Authorization: Bearer $TOKEN" \
  https://api.rafiki.lat/v1/thinks

# Expected response (200 OK):
{
  "thinks": []
}
```

### Log Verification

```bash
# Check authentication logs
ssh root@178.156.170.37
cd /opt/rafiki
docker compose logs partner-service | grep -i "auth\|keys_loaded"

# Expected log entries:
# - "keys_loaded": 1 (confirms JWT keys loaded)
# - "authenticate: BEGIN" (on login requests)
# - "authorize: BEGIN" (on protected requests)
```

---

## 5. Frontend Deployment Steps

### Prerequisites

**One-time setup:**

```bash
# Install Vercel CLI (if not already installed)
npm install -g vercel

# Login to Vercel
vercel login

# Link project (if not already linked)
cd /Users/francowini/Documents/rafiki/frontend
vercel link
# Project name: rafiki (already linked: prj_v0uCdqkUQD1DJ9mvrqqlePJGgHdJ)
```

### Step 1: Configure Environment Variables

**Add production environment variables in Vercel:**

```bash
# Production environment
vercel env add NEXT_PUBLIC_API_URL production
# Enter: https://api.rafiki.lat

vercel env add NEXT_PUBLIC_ENV production
# Enter: production

# Preview environment (for feature branches)
vercel env add NEXT_PUBLIC_API_URL preview
# Enter: https://api.rafiki.lat

vercel env add NEXT_PUBLIC_ENV preview
# Enter: preview

# Verify variables
vercel env ls
```

**Or configure via Vercel Dashboard:**

1. Go to https://vercel.com/dashboard
2. Select project: **rafiki**
3. Settings → Environment Variables
4. Add variables for both Production and Preview:

| Variable | Production Value | Preview Value |
|----------|-----------------|---------------|
| `NEXT_PUBLIC_API_URL` | `https://api.rafiki.lat` | `https://api.rafiki.lat` |
| `NEXT_PUBLIC_ENV` | `production` | `preview` |

### Step 2: Test Build Locally

```bash
cd /Users/francowini/Documents/rafiki/frontend

# Set production API URL for local build test
export NEXT_PUBLIC_API_URL=https://api.rafiki.lat
export NEXT_PUBLIC_ENV=production

# Run production build
npm run build

# Expected output:
# ✓ Compiled successfully
# ✓ Collecting page data
# ✓ Generating static pages
# ✓ Finalizing page optimization
```

### Step 3: Deploy to Preview (Optional)

```bash
cd /Users/francowini/Documents/rafiki/frontend

# Deploy to preview environment
vercel

# Expected output:
# 🔍 Inspect: https://vercel.com/...
# ✅ Preview: https://rafiki-xxx.vercel.app
```

**Test preview deployment:**

```bash
# Get preview URL from output (e.g., https://rafiki-abc123.vercel.app)
PREVIEW_URL="https://rafiki-abc123.vercel.app"

# Open in browser
open $PREVIEW_URL

# Or test with curl
curl -I $PREVIEW_URL
```

### Step 4: Deploy to Production

```bash
cd /Users/francowini/Documents/rafiki/frontend

# Deploy to production
vercel --prod

# Expected output:
# 🔍 Inspect: https://vercel.com/...
# ✅ Production: https://rafiki.vercel.app (default)
```

### Step 5: Configure Custom Domain

**Via Vercel Dashboard:**

1. Go to **Project → Settings → Domains**
2. Click **"Add Domain"**
3. Enter: `app.rafiki.lat`
4. Vercel provides DNS instructions

**Configure DNS (at your DNS provider):**

```
Type:  CNAME
Name:  app
Value: cname.vercel-dns.com
TTL:   3600 (or Auto)
```

**Wait for DNS propagation (5-30 minutes):**

```bash
# Check DNS propagation
dig app.rafiki.lat

# Expected output should show CNAME to cname.vercel-dns.com
```

**Verify domain in Vercel:**

1. Go back to Vercel dashboard
2. Click **"Refresh"** to verify domain
3. Vercel automatically provisions SSL certificate

### Step 6: Testing

**Test production URL:**

```bash
# Test HTTPS
curl -I https://app.rafiki.lat

# Expected: 200 OK with Vercel headers
```

**Test authentication flow:**

1. Open https://app.rafiki.lat in browser
2. Navigate to login page
3. Enter credentials: francoolsson1995@gmail.com / Sake2021
4. Verify JWT token received and stored
5. Navigate to protected page (e.g., /thinks)
6. Verify data loads correctly

---

## 6. Environment Variables Reference

### Backend Environment Variables

**Location:** `/opt/rafiki/docker-compose.yml` and `/opt/rafiki/docker-compose.prod.yml`

| Variable | Value | Description | Configured In |
|----------|-------|-------------|---------------|
| `POSTGRES_DB` | `rafiki` | Database name | `.env` file |
| `POSTGRES_USER` | `rafiki` | Database user | `.env` file |
| `POSTGRES_PASSWORD` | `<SECRET>` | Database password | `.env` file (not in git) |
| `PARTNER_DB_HOST` | `postgres` | Database host (docker network) | docker-compose.yml |
| `PARTNER_DB_PORT` | `5432` | Database port | docker-compose.yml |
| `PARTNER_WEB_APIHOST` | `0.0.0.0:3000` | API server bind address | docker-compose.yml |
| `PARTNER_WEB_DEBUGHOST` | `0.0.0.0:3010` | Debug server bind address | docker-compose.yml |
| `PARTNER_WEB_CORSALLOWEDORIGINS` | `https://app.rafiki.lat,https://*.vercel.app` | CORS allowed origins | docker-compose.yml |
| `PARTNER_TEMPO_HOST` | `tempo:4317` | Tracing endpoint | docker-compose.yml |

**JWT Keys:** Stored in `/opt/rafiki/keys/*.pem` (NOT in environment variables)

### Frontend Environment Variables

**Location:** Vercel Dashboard (Settings → Environment Variables)

| Variable | Production | Preview | Development | Description |
|----------|-----------|---------|-------------|-------------|
| `NEXT_PUBLIC_API_URL` | `https://api.rafiki.lat` | `https://api.rafiki.lat` | `http://localhost:3000` | Backend API base URL |
| `NEXT_PUBLIC_ENV` | `production` | `preview` | `development` | Environment identifier |

**Note:** All frontend variables MUST have `NEXT_PUBLIC_` prefix to be accessible in browser.

---

## 7. Integration Testing

### End-to-End Authentication Flow

**Test 1: Login Flow**

```bash
# 1. Get JWT token
curl -i -u francoolsson1995@gmail.com:Sake2021 \
  https://api.rafiki.lat/v1/auth/token/d8859c9d-f7f8-4f25-98e1-7b8951281d1c

# Expected: 200 OK with JWT token

# 2. Extract token from response
TOKEN="<paste_token_here>"

# 3. Test authenticated request
curl -i -H "Authorization: Bearer $TOKEN" \
  https://api.rafiki.lat/v1/thinks

# Expected: 200 OK with thinks array
```

**Test 2: CORS Verification**

```bash
# Preflight request from production domain
curl -i -X OPTIONS https://api.rafiki.lat/v1/thinks \
  -H "Origin: https://app.rafiki.lat" \
  -H "Access-Control-Request-Method: GET" \
  -H "Access-Control-Request-Headers: Authorization"

# Expected headers:
# Access-Control-Allow-Origin: https://app.rafiki.lat
# Access-Control-Allow-Methods: GET, POST, PUT, DELETE, OPTIONS
# Access-Control-Allow-Headers: Content-Type, Authorization
# Access-Control-Allow-Credentials: true
```

```bash
# Preflight request from Vercel preview
curl -i -X OPTIONS https://api.rafiki.lat/v1/thinks \
  -H "Origin: https://rafiki-abc123.vercel.app" \
  -H "Access-Control-Request-Method: GET"

# Expected: Same CORS headers (*.vercel.app is allowed)
```

**Test 3: Browser Console Test**

```javascript
// Open browser console at https://app.rafiki.lat
// Run this code:

// 1. Login
const loginResponse = await fetch('https://api.rafiki.lat/v1/auth/token/d8859c9d-f7f8-4f25-98e1-7b8951281d1c', {
  method: 'GET',
  headers: {
    'Authorization': 'Basic ' + btoa('francoolsson1995@gmail.com:Sake2021')
  }
});
const { token } = await loginResponse.json();
console.log('Token:', token);

// 2. Fetch protected data
const thinksResponse = await fetch('https://api.rafiki.lat/v1/thinks', {
  headers: {
    'Authorization': `Bearer ${token}`
  }
});
const thinks = await thinksResponse.json();
console.log('Thinks:', thinks);

// Expected: No CORS errors, data returned successfully
```

### Error Scenarios

**Test 4: Invalid Credentials**

```bash
curl -i -u wrong@email.com:wrongpass \
  https://api.rafiki.lat/v1/auth/token/d8859c9d-f7f8-4f25-98e1-7b8951281d1c

# Expected: 401 Unauthorized
```

**Test 5: Missing Token**

```bash
curl -i https://api.rafiki.lat/v1/thinks

# Expected: 401 Unauthorized
# Error: "unauthorized - missing or invalid authorization header"
```

**Test 6: Invalid Token**

```bash
curl -i -H "Authorization: Bearer invalid.token.here" \
  https://api.rafiki.lat/v1/thinks

# Expected: 401 Unauthorized
# Error: "unauthorized - token signature invalid"
```

**Test 7: Expired Token**

```bash
# Wait 48 hours (token expiration) or manually create expired token
curl -i -H "Authorization: Bearer <expired_token>" \
  https://api.rafiki.lat/v1/thinks

# Expected: 401 Unauthorized
# Error: "unauthorized - token has expired"
```

---

## 8. Critical Infrastructure Gaps

### High Priority (Fix Before Public Launch) 🔴

**1. Certbot Directories Not Initialized**

```bash
# Issue: Certbot directories exist but may not have SSL cert yet
ssh root@178.156.170.37
ls -la /opt/rafiki/certbot/conf/live/api.rafiki.lat/

# If missing, run:
docker compose --profile production run --rm certbot certonly \
  --webroot \
  --webroot-path=/var/www/certbot \
  -d api.rafiki.lat \
  --email your-email@example.com \
  --agree-tos \
  --no-eff-email

# Verify auto-renewal works:
docker compose --profile production exec certbot certbot renew --dry-run
```

**2. Database Backups Not Automated**

```bash
# Issue: No automated database backups configured

# Create backup script: /opt/rafiki/scripts/backup-db.sh
#!/bin/bash
BACKUP_DIR="/opt/rafiki/backups"
DATE=$(date +%Y%m%d_%H%M%S)
mkdir -p $BACKUP_DIR

docker exec rafiki-postgres pg_dump -U rafiki -d rafiki | \
  gzip > $BACKUP_DIR/rafiki_backup_$DATE.sql.gz

# Keep only last 7 days
find $BACKUP_DIR -name "rafiki_backup_*.sql.gz" -mtime +7 -delete

echo "Backup completed: rafiki_backup_$DATE.sql.gz"

# Make executable
chmod +x /opt/rafiki/scripts/backup-db.sh

# Add to crontab (daily at 2 AM)
crontab -e
# Add: 0 2 * * * /opt/rafiki/scripts/backup-db.sh >> /var/log/rafiki-backup.log 2>&1
```

**3. No Monitoring/Alerting**

```bash
# Issue: No uptime monitoring, no error alerts

# Solution: Set up UptimeRobot (free tier)
# 1. Go to https://uptimerobot.com
# 2. Add monitor:
#    - Type: HTTPS
#    - URL: https://api.rafiki.lat/v1/liveness
#    - Interval: 5 minutes
#    - Alert contacts: your-email@example.com
```

**4. Rate Limiting Not Tested Under Load**

```bash
# Issue: Rate limiting configured but not stress tested

# Test rate limiting
ab -n 200 -c 20 https://api.rafiki.lat/v1/liveness

# Expected: After 10 req/s, should see 503 (rate limit exceeded)
# If not working, check nginx logs:
ssh root@178.156.170.37
docker compose logs nginx | grep "limit"
```

### Medium Priority (First Week) 🟡

**5. No Log Aggregation**

```bash
# Issue: Logs only accessible via SSH + docker logs

# Quick solution: Set up log rotation
ssh root@178.156.170.37

# Configure docker log rotation in /etc/docker/daemon.json
cat > /etc/docker/daemon.json << 'EOF'
{
  "log-driver": "json-file",
  "log-opts": {
    "max-size": "10m",
    "max-file": "3"
  }
}
EOF

# Restart docker daemon
systemctl restart docker
docker compose --profile production up -d
```

**6. No Error Tracking on Frontend**

```bash
# Issue: Frontend errors not captured

# Solution: Add Sentry (optional)
cd /Users/francowini/Documents/rafiki/frontend

# Install Sentry
npm install @sentry/nextjs

# Initialize (follow prompts)
npx @sentry/wizard -i nextjs

# Configure environment variables in Vercel:
# NEXT_PUBLIC_SENTRY_DSN=<your_dsn>
# SENTRY_AUTH_TOKEN=<your_token>
```

**7. No Health Check for Frontend**

```bash
# Issue: Only backend has health checks

# Solution: Add health endpoint to Next.js
# Create: frontend/app/api/health/route.ts
export async function GET() {
  return Response.json({ status: 'ok', timestamp: Date.now() });
}

# Add to UptimeRobot:
# URL: https://app.rafiki.lat/api/health
```

### Low Priority (Nice to Have) 🟢

**8. No CI/CD Pipeline**

```bash
# Current: Manual deployments
# Future: GitHub Actions for automated testing

# Example: .github/workflows/test.yml
name: Test
on: [push, pull_request]
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v4
        with:
          go-version: '1.23'
      - run: go test ./...
```

**9. No Performance Monitoring**

```bash
# Consider adding:
# - Vercel Analytics (built-in, just enable)
# - Web Vitals tracking
# - API response time monitoring
```

**10. No Automated Security Scanning**

```bash
# Consider adding:
# - Dependabot for dependency updates
# - Snyk for vulnerability scanning
# - OWASP ZAP for penetration testing
```

---

## 9. Security Hardening

### High Priority (Before Deployment) 🔴

**1. Secure Database Credentials**

```bash
# ✅ Already done: Database password in .env (not in git)
# ✅ Already done: .env file permissions 600

# Verify:
ssh root@178.156.170.37
ls -la /opt/rafiki/.env
# Expected: -rw------- (600) root root
```

**2. Secure JWT Keys**

```bash
# ✅ Already done: Keys in /opt/rafiki/keys (not in git)
# ✅ Already done: Keys mounted read-only in docker

# Verify:
ssh root@178.156.170.37
ls -la /opt/rafiki/keys/
# Expected: -rw------- (600) root root *.pem

# Create encrypted backup:
cd /opt/rafiki
tar -czf - keys/ | gpg --symmetric --cipher-algo AES256 > keys-backup-$(date +%Y%m%d).tar.gz.gpg
# Store backup securely (not on same server)
```

**3. Firewall Configuration**

```bash
# ✅ Hetzner Cloud Firewall should be configured
# Verify firewall rules allow only:
# - Port 22 (SSH) - restricted to your IP
# - Port 80 (HTTP) - public (for Let's Encrypt)
# - Port 443 (HTTPS) - public

# Test internal ports are blocked:
timeout 5 curl http://178.156.170.37:3000 || echo "✅ Port 3000 blocked"
timeout 5 curl http://178.156.170.37:3010 || echo "✅ Port 3010 blocked"
```

**4. HTTPS Enforcement**

```bash
# ✅ Already done: nginx redirects HTTP to HTTPS
# ✅ Already done: HSTS header configured

# Verify:
curl -I http://api.rafiki.lat
# Expected: 301 redirect to https://

curl -I https://api.rafiki.lat
# Expected header: Strict-Transport-Security: max-age=31536000
```

**5. Security Headers**

```bash
# ✅ Already configured in nginx.conf:
# - Strict-Transport-Security (HSTS)
# - X-Content-Type-Options: nosniff
# - X-Frame-Options: DENY
# - X-XSS-Protection: 1; mode=block

# Verify:
curl -I https://api.rafiki.lat/v1/liveness | grep -E "X-|Strict"
```

### Medium Priority (First Week) 🟡

**6. Disable SSH Password Authentication**

```bash
ssh root@178.156.170.37

# Ensure SSH key is working first!
# Then disable password auth:
sed -i 's/#PasswordAuthentication yes/PasswordAuthentication no/' /etc/ssh/sshd_config
systemctl restart sshd
```

**7. Install Fail2Ban**

```bash
ssh root@178.156.170.37

apt update
apt install fail2ban -y
systemctl enable fail2ban
systemctl start fail2ban

# Check status
fail2ban-client status sshd
```

**8. Enable Automated Security Updates**

```bash
ssh root@178.156.170.37

apt install unattended-upgrades -y
dpkg-reconfigure -plow unattended-upgrades
# Select: Yes
```

**9. Add Content Security Policy (CSP)**

```bash
# Add to nginx.conf (after deployment):
add_header Content-Security-Policy "default-src 'self'; script-src 'self' 'unsafe-inline' 'unsafe-eval'; style-src 'self' 'unsafe-inline'; img-src 'self' data: https:; font-src 'self' data:; connect-src 'self' https://api.rafiki.lat;" always;
```

### Low Priority (Nice to Have) 🟢

**10. Regular Security Audits**

```bash
# Run periodically:

# Check for vulnerable dependencies (backend)
cd /Users/francowini/Documents/rafiki
go list -m all | nancy sleuth

# Check for vulnerable dependencies (frontend)
cd frontend
npm audit

# Run security scanner (example: Nikto)
nikto -h https://api.rafiki.lat
```

---

## 10. Rollback Strategy

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
# SSH to server
ssh root@178.156.170.37
cd /opt/rafiki

# Check migration status
docker exec -it rafiki-postgres psql -U rafiki -d rafiki -c \
  "SELECT version, description, applied_at FROM darwin_migrations ORDER BY applied_at DESC LIMIT 5;"

# Migrations are idempotent, but if needed, manual rollback:
# 1. Restore database backup
gunzip < /opt/rafiki/backups/rafiki_backup_YYYYMMDD_HHMMSS.sql.gz | \
  docker exec -i rafiki-postgres psql -U rafiki -d rafiki

# 2. Restart services
docker compose --profile production restart partner-service
```

**Scenario 3: JWT keys corrupted**

```bash
# Restore from backup
ssh root@178.156.170.37
cd /opt/rafiki

# Decrypt backup
gpg -d keys-backup-YYYYMMDD.tar.gz.gpg | tar -xzf -

# Restart service
docker compose --profile production restart partner-service

# Time estimate: 2-3 minutes
# Impact: All users will need to re-login (tokens invalidated)
```

### Frontend Rollback

**Scenario 1: Recent deployment broke frontend**

**Option A: Vercel Dashboard (Easiest)**

1. Go to https://vercel.com/dashboard
2. Select project: **rafiki**
3. Go to **Deployments** tab
4. Find last working deployment
5. Click **⋯** (three dots) → **Promote to Production**
6. Time estimate: 30 seconds
7. Impact: Instant rollback, no downtime

**Option B: Vercel CLI**

```bash
cd /Users/francowini/Documents/rafiki/frontend

# Rollback to previous deployment
vercel rollback

# Time estimate: 30 seconds
```

**Option C: Git Revert**

```bash
cd /Users/francowini/Documents/rafiki

# Revert last commit
git revert HEAD
git push origin main

# Vercel auto-deploys reverted version
# Time estimate: 2-3 minutes
```

**Scenario 2: Environment variable misconfiguration**

```bash
# Fix via Vercel Dashboard
# 1. Settings → Environment Variables
# 2. Update incorrect variable
# 3. Redeploy (Deployments → ⋯ → Redeploy)

# Or via CLI:
vercel env rm NEXT_PUBLIC_API_URL production
vercel env add NEXT_PUBLIC_API_URL production
# Enter correct value: https://api.rafiki.lat

# Redeploy
vercel --prod
```

### When to Rollback

**Critical (Rollback Immediately):**
- Authentication completely broken (can't login)
- Database corruption detected
- Security vulnerability discovered
- 500 errors affecting all users
- Data loss occurring

**High Priority (Rollback Within 1 Hour):**
- Major features broken (e.g., can't create thinks)
- CORS errors preventing frontend access
- Performance degradation (>5s response times)
- SSL certificate issues

**Medium Priority (Investigate First, Maybe Rollback):**
- Minor UI bugs
- Non-critical features broken
- Slow performance on specific endpoints
- High error rate on specific features

**Low Priority (Fix Forward):**
- Cosmetic issues
- Minor UX improvements needed
- Log noise (non-functional errors)

---

## 11. Monitoring (Recommended)

### What to Monitor

**Backend Health:**
```bash
# Liveness check (is service running?)
https://api.rafiki.lat/v1/liveness

# Readiness check (is service ready to accept traffic?)
https://api.rafiki.lat/v1/readiness

# Monitor with UptimeRobot:
# - URL: https://api.rafiki.lat/v1/liveness
# - Interval: 5 minutes
# - Alert on: Down for 2 consecutive checks
```

**Frontend Health:**
```bash
# Create health endpoint (see Section 8.7)
https://app.rafiki.lat/api/health

# Monitor with UptimeRobot
```

**Database Health:**
```bash
# SSH to server and run:
docker exec rafiki-postgres pg_isready -U rafiki -d rafiki

# Expected: "postgres:5432 - accepting connections"

# Add to monitoring script (run every 5 minutes)
```

**SSL Certificate Expiration:**
```bash
# Check certificate expiration
echo | openssl s_client -servername api.rafiki.lat -connect api.rafiki.lat:443 2>/dev/null | \
  openssl x509 -noout -dates

# Monitor with UptimeRobot SSL monitoring
```

### Metrics to Track

**Application Metrics (via statsviz - already available):**
```bash
# Access metrics dashboard
http://178.156.170.37:3010/debug/statsviz/

# Metrics available:
# - Goroutine count
# - Memory usage
# - GC stats
# - HTTP request counts
```

**Business Metrics (to implement later):**
- User registrations per day
- Login attempts (success/failure)
- API response times (p50, p95, p99)
- Error rates by endpoint
- Active users

**Infrastructure Metrics:**
```bash
# Disk usage
ssh root@178.156.170.37
df -h /opt/rafiki

# Docker container stats
docker stats --no-stream

# Database size
docker exec rafiki-postgres psql -U rafiki -d rafiki -c \
  "SELECT pg_size_pretty(pg_database_size('rafiki'));"
```

### Recommended Monitoring Tools

**Free Tier Options:**

1. **UptimeRobot** (Uptime monitoring)
   - URL: https://uptimerobot.com
   - Free tier: 50 monitors, 5-minute checks
   - Set up: https://api.rafiki.lat/v1/liveness

2. **Vercel Analytics** (Frontend performance)
   - Built-in to Vercel
   - Enable in: Project Settings → Analytics

3. **Sentry** (Error tracking - optional)
   - URL: https://sentry.io
   - Free tier: 5,000 events/month

4. **Grafana** (Already running!)
   - Access: http://178.156.170.37:3100
   - Already configured with Tempo for traces

---

## 12. Troubleshooting Guide

### Common Issues and Solutions

**Issue 1: CORS Error in Browser Console**

```
Error: Access to fetch at 'https://api.rafiki.lat/v1/thinks' from origin 'https://app.rafiki.lat'
has been blocked by CORS policy
```

**Diagnosis:**
```bash
# Check CORS configuration
ssh root@178.156.170.37
cd /opt/rafiki
grep CORSALLOWEDORIGINS docker-compose.yml

# Expected:
# PARTNER_WEB_CORSALLOWEDORIGINS=https://app.rafiki.lat,https://*.vercel.app
```

**Solution:**
```bash
# If CORS config is wrong, fix docker-compose.yml
# Then redeploy:
sudo ./devops/deploy.sh
```

**Issue 2: JWT Token Invalid/Expired**

```
Error: unauthorized - token signature invalid
```

**Diagnosis:**
```bash
# Check JWT keys exist
ssh root@178.156.170.37
ls -la /opt/rafiki/keys/

# Check service logs
docker compose logs partner-service | grep "keys_loaded"
# Expected: "keys_loaded": 1
```

**Solution:**
```bash
# If keys missing, regenerate (WARNING: invalidates all tokens)
ssh root@178.156.170.37
cd /opt/rafiki
sudo ./devops/setup-prod-keys.sh
sudo ./devops/deploy.sh

# All users must re-login with new kid
```

**Issue 3: Frontend Build Fails on Vercel**

```
Error: Type error: Cannot find module '@/lib/api' or its corresponding type declarations
```

**Diagnosis:**
```bash
# Test build locally
cd /Users/francowini/Documents/rafiki/frontend
npm run build

# Check TypeScript errors
npx tsc --noEmit
```

**Solution:**
```bash
# Fix TypeScript errors locally
# Test build succeeds
npm run build

# Then commit and push
git add .
git commit -m "fix: TypeScript errors"
git push origin main

# Vercel auto-deploys
```

**Issue 4: Database Connection Fails**

```
Error: dial tcp 10.10.0.2:5432: connect: connection refused
```

**Diagnosis:**
```bash
ssh root@178.156.170.37
cd /opt/rafiki

# Check postgres is running
docker compose ps postgres

# Check postgres logs
docker compose logs postgres | tail -20
```

**Solution:**
```bash
# Restart postgres
docker compose restart postgres

# Wait for health check
sleep 10

# Restart partner-service
docker compose restart partner-service

# Verify health
curl http://localhost:3000/v1/readiness
```

**Issue 5: Nginx 502 Bad Gateway**

```
Error: 502 Bad Gateway
```

**Diagnosis:**
```bash
ssh root@178.156.170.37
cd /opt/rafiki

# Check backend is running
docker compose ps partner-service

# Check nginx logs
docker compose logs nginx | tail -20

# Check backend logs
docker compose logs partner-service | tail -20
```

**Solution:**
```bash
# Restart backend service
docker compose restart partner-service

# Wait for health check
sleep 5

# Test directly (bypass nginx)
curl http://localhost:3000/v1/liveness

# If direct access works, restart nginx
docker compose restart nginx
```

**Issue 6: SSL Certificate Not Renewing**

```
Warning: Certificate expires in 7 days
```

**Diagnosis:**
```bash
ssh root@178.156.170.37

# Check certificate expiration
openssl x509 -enddate -noout -in /opt/rafiki/certbot/conf/live/api.rafiki.lat/fullchain.pem

# Check certbot logs
docker compose logs certbot | tail -50

# Test renewal (dry run)
docker compose --profile production exec certbot certbot renew --dry-run
```

**Solution:**
```bash
# Force renewal
docker compose --profile production exec certbot certbot renew --force-renewal

# Restart nginx to load new cert
docker compose restart nginx
```

**Issue 7: High Memory Usage**

```
Warning: Container using 180MB/192MB memory limit
```

**Diagnosis:**
```bash
ssh root@178.156.170.37

# Check container stats
docker stats --no-stream

# Check Go runtime metrics
curl http://localhost:3010/debug/vars | jq '.memstats'
```

**Solution:**
```bash
# Increase memory limit in docker-compose.yml
# Or optimize GOMEMLIMIT environment variable

# Current limit: 128MiB
# Consider increasing to 256MiB if needed

# Edit docker-compose.yml
# Then redeploy:
sudo ./devops/deploy.sh
```

**Issue 8: Rate Limiting Too Aggressive**

```
Error: 503 Service Temporarily Unavailable (nginx rate limit)
```

**Diagnosis:**
```bash
ssh root@178.156.170.37
docker compose logs nginx | grep "limiting"
```

**Solution:**
```bash
# Adjust rate limit in nginx/nginx.conf
# Current: limit_req_zone ... rate=10r/s; burst=20

# Increase to: rate=20r/s; burst=50
# Then redeploy nginx:
docker compose restart nginx
```

---

## 13. Post-Deployment Checklist

### Immediate Verification (Within 5 Minutes)

- [ ] Backend health check passes: `curl https://api.rafiki.lat/v1/liveness`
- [ ] Frontend loads: Open https://app.rafiki.lat in browser
- [ ] Login works: Test with admin credentials
- [ ] Protected routes require authentication
- [ ] CORS works from production domain
- [ ] SSL certificate valid: Check browser padlock icon
- [ ] Database accessible: `ssh root@178.156.170.37 && docker exec -it rafiki-postgres psql -U rafiki -d rafiki -c '\l'`

### First Hour Verification

- [ ] Monitor error logs: `ssh root@178.156.170.37 && cd /opt/rafiki && docker compose logs -f --tail=100`
- [ ] Check for CORS errors in browser console
- [ ] Test all main user flows: Login → Create Think → View Thinks → Logout
- [ ] Verify JWT token expiration (check token claims)
- [ ] Test rate limiting: Make 30 requests in 5 seconds
- [ ] Check Vercel deployment status: https://vercel.com/dashboard
- [ ] Verify environment variables: Check Vercel Settings → Environment Variables

### First Day Verification

- [ ] Monitor uptime: Set up UptimeRobot monitor
- [ ] Check SSL certificate expiration date
- [ ] Review application logs for errors
- [ ] Test from different browsers (Chrome, Firefox, Safari)
- [ ] Test from mobile devices
- [ ] Verify database backup exists: `ls -la /opt/rafiki/backups/`
- [ ] Document any issues encountered

### First Week Verification

- [ ] Review error logs daily
- [ ] Monitor server resources: CPU, memory, disk
- [ ] Check database size growth
- [ ] Test rollback procedure (in staging if possible)
- [ ] Verify certbot auto-renewal: `docker compose --profile production exec certbot certbot renew --dry-run`
- [ ] Review security logs: `fail2ban-client status sshd`
- [ ] Update documentation with any issues found

### Performance Checks

- [ ] API response time < 500ms: Test with curl -w "@curl-format.txt"
- [ ] Frontend load time < 3s: Test with Chrome DevTools
- [ ] Database query performance: Check logs for slow queries
- [ ] Memory usage stable: `docker stats`
- [ ] No goroutine leaks: Check http://178.156.170.37:3010/debug/statsviz/
- [ ] SSL handshake time < 200ms: Test with curl -w "%{time_connect}\n"

### Security Checks

- [ ] Firewall rules verified: Only ports 22, 80, 443 open
- [ ] SSH key authentication only (no passwords)
- [ ] Database credentials secure (not in git)
- [ ] JWT keys backed up and encrypted
- [ ] HTTPS enforced (HTTP redirects to HTTPS)
- [ ] Security headers present: `curl -I https://api.rafiki.lat`
- [ ] No sensitive data in logs: Review logs for passwords, tokens

---

## Additional Resources

- **Backend Deployment Guide:** `/Users/francowini/Documents/rafiki/devops/DEPLOYMENT_GUIDE.md`
- **Frontend Deployment Guide:** `/Users/francowini/Documents/rafiki/devops/FRONTEND_DEPLOYMENT.md`
- **SSL Certificate Setup:** `/Users/francowini/Documents/rafiki/devops/SSL_CERTIFICATE_SETUP.md`
- **Firewall Configuration:** `/Users/francowini/Documents/rafiki/devops/FIREWALL_GUIDE.md`
- **Project Instructions:** `/Users/francowini/Documents/rafiki/CLAUDE.md`

## Support Contacts

- **Hetzner Support:** https://www.hetzner.com/support
- **Vercel Support:** https://vercel.com/support
- **Let's Encrypt Community:** https://community.letsencrypt.org

---

**Document Version:** 1.0
**Last Updated:** 2025-11-15
**Next Review:** After first production deployment
