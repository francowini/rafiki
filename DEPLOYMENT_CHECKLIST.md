# Rafiki Deployment Checklist

Quick reference for deployment operations.

---

## ✅ ONE-TIME SETUP (Do Once)

```bash
# On Hetzner server (178.156.170.37)

# 1. Install Docker (ONCE)
curl -fsSL https://get.docker.com | sh
apt install docker-compose-plugin -y

# 2. Clone repo (ONCE)
mkdir -p /opt/rafiki
cd /opt
git clone <repo-url> rafiki

# 3. Generate JWT keys (ONCE - NEVER REPEAT!)
cd /opt/rafiki
sudo ./devops/setup-prod-keys.sh
# ⚠️ SAVE THE KEY ID OUTPUT!

# 4. Create .env file (ONCE)
cat > /opt/rafiki/.env << 'EOF'
POSTGRES_DB=rafiki
POSTGRES_USER=rafiki
POSTGRES_PASSWORD=<STRONG_PASSWORD_HERE>
EOF
chmod 600 /opt/rafiki/.env

# 5. First deployment (ONCE)
cd /opt/rafiki
sudo ./devops/deploy.sh

# 6. Create ADMIN user (ONCE - DO NOT REPEAT!)
cd /opt/rafiki
./zarf/create-user.sh admin@rafiki.lat <password> ADMIN "Admin Name"
```

---

## 🚀 REGULAR DEPLOYMENT (Every Update)

### From Your Local Machine:
```bash
git push origin main
make deploy
```

### Or On Server:
```bash
ssh root@178.156.170.37
cd /opt/rafiki
git pull origin main
sudo ./devops/deploy.sh
```

**That's it!** Deployment script handles everything:
- ✅ Stops containers
- ✅ Builds new images
- ✅ Runs migrations (safe, idempotent)
- ✅ Starts services
- ✅ Health checks

---

## 🔍 MONITORING & DEBUGGING

```bash
# View logs
docker compose logs -f partner-service

# Check status
docker compose ps

# Health check
curl http://localhost:3000/v1/readiness

# Restart service
docker compose restart partner-service
```

---

## ⚠️ CRITICAL: DO NOT DO THESE

- ❌ **NEVER** run `setup-prod-keys.sh` again (invalidates all tokens!)
- ❌ **NEVER** run `docker compose down -v` (deletes database!)
- ❌ **NEVER** recreate ADMIN user (email is unique)
- ❌ **NEVER** commit JWT keys to git
- ❌ **NEVER** share .env file

---

## 🆘 EMERGENCY PROCEDURES

### Service Won't Start
```bash
docker compose logs partner-service
docker compose down
docker compose up -d
```

### Database Issues
```bash
docker compose logs postgres
docker exec -it rafiki-postgres psql -U rafiki -d rafiki
```

### Reset Everything (LAST RESORT - LOSES DATA!)
```bash
docker compose down -v  # ⚠️ DELETES DATABASE!
sudo ./devops/deploy.sh
./zarf/create-user.sh admin@rafiki.lat <password> ADMIN "Admin"
```

---

## 📞 QUICK COMMANDS

| Task | Command |
|------|---------|
| Deploy | `cd /opt/rafiki && sudo ./devops/deploy.sh` |
| Logs | `docker compose logs -f partner-service` |
| Status | `docker compose ps` |
| Health | `curl http://localhost:3000/v1/readiness` |
| Restart | `docker compose restart partner-service` |
| Add User | `./zarf/create-user.sh <email> <pwd> <ROLE> <name>` |
| DB Shell | `docker exec -it rafiki-postgres psql -U rafiki -d rafiki` |

---

**Server:** 178.156.170.37
**Path:** /opt/rafiki
**Keep this file on server for quick reference!**
