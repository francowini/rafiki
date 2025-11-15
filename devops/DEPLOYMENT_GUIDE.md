# Rafiki Deployment Guide

This guide separates **one-time setup** from **regular deployments** to avoid repeating initialization steps.

---

## 🏁 ONE-TIME SETUP (Do Once Per Server)

These steps should ONLY be done once when setting up a new server or after a complete reset.

### 1. Server Preparation

```bash
# On your Hetzner server (SSH as root)
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

# Clone repository (FIRST TIME ONLY)
cd /opt
git clone <your-repo-url> rafiki

# Or if you prefer to deploy via rsync/scp initially
cd /opt/rafiki
```

### 3. Generate Production JWT Keys (CRITICAL - DO ONCE)

```bash
cd /opt/rafiki

# Generate production keys
sudo ./devops/setup-prod-keys.sh

# IMPORTANT: Save the output!
# You'll see: "Key ID (kid): xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"
# Keep this key ID safe - you'll need it for user creation

# Verify keys were created
ls -la /opt/rafiki/keys/
# You should see a .pem file
```

**⚠️ WARNING:**
- JWT keys should NEVER be regenerated unless absolutely necessary
- Regenerating keys will invalidate ALL existing user tokens
- Backup your keys: `sudo tar -czf rafiki-keys-backup.tar.gz /opt/rafiki/keys/`

### 4. Create Environment File

```bash
cd /opt/rafiki

# Create .env file (DO ONCE)
cat > .env << 'EOF'
# Database Configuration
POSTGRES_DB=rafiki
POSTGRES_USER=rafiki
POSTGRES_PASSWORD=<CHANGE_ME_TO_STRONG_PASSWORD>

# Note: Other environment variables are in docker-compose.yml
# Only database credentials need to be in .env file
EOF

# Secure the .env file
chmod 600 .env
```

**⚠️ IMPORTANT:** Change `<CHANGE_ME_TO_STRONG_PASSWORD>` to a strong password!

### 5. First Deployment

```bash
cd /opt/rafiki

# Run first deployment (this will create database, run migrations)
sudo ./devops/deploy.sh
```

This will:
- ✅ Create database for the first time
- ✅ Run all migrations automatically (idempotent - safe to run multiple times)
- ✅ Start all services
- ✅ Verify health checks

### 6. Create ADMIN User (DO ONCE)

```bash
cd /opt/rafiki

# Create your first admin user
./zarf/create-user.sh admin@rafiki.lat <password> ADMIN "Admin User"

# Example:
# ./zarf/create-user.sh admin@rafiki.lat MySecurePassword123 ADMIN "Franco Winiarski"
```

**⚠️ IMPORTANT:**
- Only create the admin user ONCE
- Running this command multiple times with the same email will fail (email is unique)
- Keep the admin credentials safe!

### 7. Configure Firewall (Optional but Recommended)

```bash
# Allow SSH
ufw allow 22/tcp

# Allow HTTP (for Let's Encrypt challenges)
ufw allow 80/tcp

# Allow HTTPS
ufw allow 443/tcp

# Enable firewall
ufw enable
```

See [FIREWALL_GUIDE.md](FIREWALL_GUIDE.md) for details.

### 8. SSL Certificate Setup (Optional - Production Only)

See [SSL_CERTIFICATE_SETUP.md](SSL_CERTIFICATE_SETUP.md) for Let's Encrypt configuration.

---

## 🚀 REGULAR DEPLOYMENT (Every Update)

These steps are for deploying code changes, updates, or bug fixes.

### Option A: From Your Local Machine (Recommended)

```bash
# From your local rafiki directory

# 1. Commit and push your changes
git add .
git commit -m "Your changes"
git push origin main

# 2. Deploy to production (ONE COMMAND!)
make deploy
```

The `make deploy` command will:
1. SSH to the server
2. Pull latest code (`git pull origin main`)
3. Run deployment script
4. Build and restart services
5. Verify health checks
6. Show deployment status

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

**That's it!** The deployment script handles everything else.

---

## 🔍 What Happens During Regular Deployment

The `deploy.sh` script automatically:

1. ✅ Stops existing containers
2. ✅ Pulls latest git changes
3. ✅ Builds new Docker images
4. ✅ Starts services with production profile
5. ✅ **Runs database migrations** (idempotent - safe every time)
6. ✅ Waits for health checks (max 60 seconds)
7. ✅ Verifies authentication (checks JWT keys loaded)
8. ✅ Shows deployment summary

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

---

## 🔧 Common Operations

### View Production Logs

```bash
# From local machine
make deploy-logs

# Or on server
cd /opt/rafiki
docker compose logs -f partner-service
```

### Check Production Status

```bash
# From local machine
make deploy-status

# Or on server
cd /opt/rafiki
docker compose ps
```

### Health Check

```bash
# From local machine
make deploy-health

# Or on server
curl http://localhost:3000/v1/readiness
```

### Restart Service (Without Full Deployment)

```bash
# From local machine
make deploy-restart

# Or on server
cd /opt/rafiki
docker compose restart partner-service
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
```

---

## 📋 Deployment Checklist

### First Time Deployment (Complete Server Setup)
- [ ] Server is provisioned (Hetzner CPX11)
- [ ] Docker and Docker Compose installed
- [ ] Repository cloned to `/opt/rafiki`
- [ ] JWT keys generated (`setup-prod-keys.sh`) - **ONCE ONLY**
- [ ] `.env` file created with database password - **ONCE ONLY**
- [ ] First deployment run (`./devops/deploy.sh`)
- [ ] ADMIN user created (`create-user.sh`) - **ONCE ONLY**
- [ ] Firewall configured (optional)
- [ ] SSL certificate obtained (optional)

### Regular Deployment (Code Updates)
- [ ] Code committed and pushed to `main` branch
- [ ] Run `make deploy` from local machine
- [ ] Verify health checks pass
- [ ] Test API functionality
- [ ] Check logs for errors

---

## ⚠️ IMPORTANT: What NOT to Do

### ❌ DO NOT Regenerate JWT Keys Unless Necessary
```bash
# DON'T run this again after initial setup!
# ./devops/setup-prod-keys.sh
```
**Why:** Regenerating keys invalidates all existing user sessions.

### ❌ DO NOT Recreate ADMIN User
```bash
# DON'T run this multiple times with same email!
# ./zarf/create-user.sh admin@rafiki.lat ...
```
**Why:** Email is unique - command will fail.

### ❌ DO NOT Delete Docker Volumes Accidentally
```bash
# NEVER run this on production!
# docker compose down -v  # The -v flag deletes volumes (DATABASE!)
```
**Why:** This deletes your entire database!

Safe shutdown:
```bash
docker compose down  # Safe - keeps database intact
```

### ❌ DO NOT Edit .env During Deployment
**Why:** Changes won't take effect until containers restart. Edit .env only between deployments.

---

## 🆘 Troubleshooting

### Deployment Health Check Fails

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

# If missing, keys weren't created - run setup-prod-keys.sh
```

### Database Connection Fails

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

### Container Won't Start

```bash
# Check logs
docker compose logs partner-service

# Check for port conflicts
netstat -tulpn | grep -E '3000|3010'

# Rebuild from scratch
docker compose down
docker compose build --no-cache
docker compose up -d
```

---

## 📞 Quick Reference

| Task | Command |
|------|---------|
| Deploy from local | `make deploy` |
| Deploy on server | `cd /opt/rafiki && sudo ./devops/deploy.sh` |
| View logs | `make deploy-logs` or `docker compose logs -f partner-service` |
| Check status | `make deploy-status` or `docker compose ps` |
| Health check | `make deploy-health` or `curl http://localhost:3000/v1/readiness` |
| Restart service | `make deploy-restart` or `docker compose restart partner-service` |
| SSH to server | `make ssh` or `ssh root@178.156.170.37` |
| Create user | `./zarf/create-user.sh <email> <password> <ROLE> <name>` |
| Database shell | `make db-shell-prod` or `docker exec -it rafiki-postgres psql -U rafiki -d rafiki` |

---

## 🔐 Security Notes

1. **JWT Keys:** Stored in `/opt/rafiki/keys/` - backup regularly, never commit to git
2. **Database Password:** Stored in `/opt/rafiki/.env` - use strong password
3. **Admin Credentials:** Keep admin user credentials secure
4. **CORS:** Already configured for `app.rafiki.lat` and Vercel previews
5. **Firewall:** Only ports 22, 80, 443 should be open to internet
6. **SSL:** Use Let's Encrypt for HTTPS (see SSL_CERTIFICATE_SETUP.md)

---

**Last Updated:** 2025-11-15
**Server:** Hetzner CPX11 @ 178.156.170.37
**Repository:** /opt/rafiki
