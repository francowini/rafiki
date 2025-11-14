# Rafiki Deployment Guide - Hetzner CPX11

## Prerequisites

- Hetzner CPX11 server with Ubuntu/Debian
- Docker and Docker Compose installed
- Git installed
- Go 1.23+ installed (required for user creation script)
- Root or sudo access

## Credentials Configuration

### PostgreSQL Credentials
- **Database Name:** rafiki
- **Database User:** rafiki
- **Database Password:** db (⚠️ Change this in production!)
- **Database Host:** postgres (internal docker network)
- **Database Port:** 5432

### Server Access
- **SSH Host:** 178.156.170.37
- **SSH User:** root
- **SSH Port:** 22

## Initial Server Setup

### 1. Connect to Hetzner Server
```bash
ssh root@178.156.170.37
```

### 2. Install Docker and Docker Compose
```bash
# Update system
apt update && apt upgrade -y

# Install Docker
curl -fsSL https://get.docker.com -o get-docker.sh
sh get-docker.sh

# Install Docker Compose
apt install docker-compose -y

# Verify installation
docker --version
docker-compose --version
```

### 2.5. Install Go (required for user creation)
```bash
# Download and install Go 1.23
wget https://go.dev/dl/go1.23.0.linux-amd64.tar.gz
rm -rf /usr/local/go && tar -C /usr/local -xzf go1.23.0.linux-amd64.tar.gz
export PATH=$PATH:/usr/local/go/bin
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc

# Verify installation
go version
```

### 3. Create deployment directory
```bash
mkdir -p /opt/rafiki
cd /opt/rafiki
```

### 4. Clone or upload your code
```bash
# Option 1: Using git
git clone YOUR_REPO_URL .

# Option 2: Manual upload
# Use scp from your local machine:
# scp -r /path/to/rafiki/* root@YOUR_SERVER_IP:/opt/rafiki/
```

### 5. Create .env file
```bash
# Copy example file
cp .env.example .env

# Edit credentials
nano .env
```

Set these variables:
```
POSTGRES_DB=rafiki
POSTGRES_USER=rafiki
POSTGRES_PASSWORD=db  # ⚠️ Change this in production!
```

### 6. Configure firewall
```bash
# Allow SSH, HTTP, and your API ports
ufw allow 22/tcp
ufw allow 80/tcp
ufw allow 443/tcp
ufw allow 3000/tcp  # API port
ufw --force enable
```

### 7. Setup JWT Authentication Keys ⚠️ IMPORTANT
```bash
# Generate production JWT keys (REQUIRED for authentication)
cd /opt/rafiki
./devops/setup-prod-keys.sh
```

This script will:
- Create `/opt/rafiki/keys` directory
- Generate RSA 4096-bit key for JWT signing
- Set secure permissions (600)
- Display the Key ID (kid) - **save this value!**
- Optionally create encrypted backup

**IMPORTANT:**
- Save the Key ID (kid) - you'll need it for login requests
- Keys are stored in `/opt/rafiki/keys` (NOT in Docker image)
- Different from development keys (more secure)
- Backup the keys securely!

## Deployment

### From Local Machine to Server

#### Step 1: Push code changes
```bash
git push
```

#### Step 2: Copy .env file to server
```bash
./devops/copy-env.sh
# Enter SSH password when prompted
```

### On Server

#### Step 3: Setup JWT keys (First deployment only)
```bash
# SSH into server
ssh root@178.156.170.37

# Setup production keys
cd /opt/rafiki
./devops/setup-prod-keys.sh

# Save the Key ID displayed - you'll need it later!
```

#### Step 4: Run deployment script
```bash
# SSH into server
ssh root@178.156.170.37

# Run deployment (script auto-detects project root)
/opt/rafiki/devops/deploy.sh
```

The deployment script will:
1. Load environment variables from `.env`
2. Check for JWT keys (warns if missing)
3. Stop existing containers
4. Pull latest git changes
5. Build and start all services with production config
6. Wait for services to become healthy
7. Verify authentication is enabled
8. Show deployment status

**Note:** The script uses both `docker-compose.yml` and `docker-compose.prod.yml` to mount production keys.

#### Step 5: Create admin user
```bash
# Create admin user for authentication
cd /opt/rafiki
./zarf/create-user.sh admin@yourdomain.com <strong-password> ADMIN "Admin User"

# Save the credentials securely!
```

**Note:** The script will automatically build a bcrypt hash generator (`zarf/genhash/genhash`) on first run. This requires Go to be installed on the server.

### Subsequent Deployments
Same as above - push code, copy env (if changed), run deploy script on server.
**Keys don't need to be regenerated** unless rotating.

## Management Commands

**IMPORTANT**: Use `docker compose` (with space), NOT `docker-compose` (with hyphen)

### View logs
```bash
# All services
docker compose logs -f

# Specific service
docker compose logs -f partner-service
docker compose logs -f postgres
docker compose logs -f tempo
docker compose logs -f grafana
```

### Restart services
```bash
# Restart all
docker compose restart

# Restart specific service
docker compose restart partner-service
```

### Stop services
```bash
docker compose down
```

### Check service status
```bash
docker compose ps
```

### Access PostgreSQL
```bash
docker exec -it rafiki-postgres psql -U rafiki -d rafiki
# Password: db (or value from .env)
```

### Create users (SQL only - no registration endpoint)
```bash
# Use the helper script (auto-builds genhash tool on first run)
./zarf/create-user.sh user@example.com password123 USER "User Name"

# Example with all parameters:
# ./zarf/create-user.sh <email> <password> <role> <name>
# Roles: USER or ADMIN

# Or manually via SQL
docker exec -it rafiki-postgres psql -U rafiki -d rafiki
# Then INSERT INTO users...
```

### Check authentication status
```bash
# Check if keys are loaded
docker compose logs partner-service | grep "keys_loaded"

# Should show: "keys_loaded": 1 (or more)
```

## Service Endpoints

### Main Application
- **API:** http://178.156.170.37:3000
- **Debug/Metrics:** http://178.156.170.37:3010
- **Health Check:** http://178.156.170.37:3000/v1/readiness
- **Liveness:** http://178.156.170.37:3000/v1/liveness

### Authentication Endpoints
- **Login:** `GET /v1/auth/token/{kid}` (Basic Auth)
  - Replace `{kid}` with your production Key ID
  - Use Basic Auth header: `email:password`
  - Returns JWT token (valid 48 hours)
- **Protected Routes:** All `/v1/thinks/*` require Bearer token

### Observability Stack
- **Grafana Dashboard:** http://178.156.170.37:3100 (anonymous access enabled)
- **Tempo Traces:** http://178.156.170.37:3200
- **PostgreSQL:** 178.156.170.37:5432 (external access - consider firewalling)

### Internal Network (container-to-container)
- postgres: `10.10.0.2:5432`
- partner-service: `10.10.0.10:3000`
- tempo: `10.10.0.20:4317` (OTLP gRPC)
- grafana: `10.10.0.21:3100`

## Backup PostgreSQL

```bash
# Create backup
docker compose exec postgres pg_dump -U rafiki rafiki > backup_$(date +%Y%m%d_%H%M%S).sql

# Or with docker exec
docker exec rafiki-postgres pg_dump -U rafiki rafiki > backup_$(date +%Y%m%d_%H%M%S).sql

# Restore backup
cat backup_file.sql | docker compose exec -T postgres psql -U rafiki rafiki
```

## Monitoring

### Check service health
```bash
curl http://localhost:3000/v1/readiness
curl http://localhost:3000/v1/liveness
```

### View metrics
```bash
curl http://localhost:3010/debug/vars
```

## Authentication Testing

### Test login (get JWT token)
```bash
# Get your production Key ID
ls /opt/rafiki/keys/*.pem | xargs basename | sed 's/.pem//'

# Save it
KID="your-production-kid"

# Test login
curl -i -X GET \
  -H "Authorization: Basic $(echo -n 'admin@yourdomain.com:password' | base64)" \
  http://localhost:3000/v1/auth/token/${KID}

# Should return 200 OK with JWT token
```

### Test authenticated endpoint
```bash
# Save the token from above
TOKEN="your-jwt-token"

# Create a think
curl -X POST \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"category":"reflection","content":"Test from production"}' \
  http://localhost:3000/v1/thinks

# Get all thinks
curl -H "Authorization: Bearer $TOKEN" \
  http://localhost:3000/v1/thinks
```

## Troubleshooting

### Authentication issues

**Symptom:** "key not found" or "no authentication keys loaded"

**Solution:**
```bash
# Check if keys directory exists and has keys
ls -la /opt/rafiki/keys/

# Should show .pem files
# If not, run: ./devops/setup-prod-keys.sh

# Restart service
docker compose restart partner-service

# Verify keys loaded
docker compose logs partner-service | grep "keys_loaded"
```

**Symptom:** Login returns 401 Unauthorized

**Solution:**
```bash
# Verify user exists and is enabled
docker exec -it rafiki-postgres psql -U rafiki -d rafiki \
  -c "SELECT email, enabled FROM users WHERE email='your@email.com';"

# Check password hash format (should start with $2a$ or $2b$)
# Recreate user if needed
```

### Service won't start
```bash
# Check logs
docker compose logs

# Check if ports are available
netstat -tulpn | grep -E '3000|3010|3100|3200|4317|5432'

# Check container status
docker compose ps
```

### Database connection issues
```bash
# Verify PostgreSQL is running
docker compose ps postgres

# Check PostgreSQL logs
docker compose logs postgres

# Test connection
docker compose exec postgres pg_isready -U rafiki
```

### Observability issues
```bash
# Check Tempo
curl http://localhost:3200/ready

# Check Grafana
curl http://localhost:3100/api/health

# View traces in Grafana
# Navigate to http://YOUR_IP:3100 and explore Tempo datasource
```

### Clean restart
```bash
# Stop everything
docker compose down

# Remove volumes (WARNING: This deletes data!)
docker compose down -v

# Rebuild and start
docker compose up -d --build
```

## Security Recommendations

1. ⚠️ **Change default password "db" in production!**
2. ⚠️ **Setup JWT keys with `setup-prod-keys.sh` - DO NOT use dev keys!**
3. ⚠️ **Backup JWT keys securely - store encryption password in password manager**
4. Use SSH keys instead of passwords
5. Keep Docker and system packages updated
6. Configure HTTPS with Let's Encrypt (nginx/traefik)
7. Limit exposed ports via firewall
8. Regular backups of PostgreSQL data
9. Monitor logs for suspicious activity
10. Store production credentials securely (not in version control)
11. Use strong passwords for admin users
12. Rotate JWT keys periodically (see docs/PRODUCTION_KEYS_SETUP.md)

## Docker Compose Services

The application stack includes:

1. **partner-service** - Main Go application (192M RAM, 0.5 CPU)
2. **postgres** - PostgreSQL 18.0 database (256M RAM, 0.5 CPU)
3. **tempo** - Distributed tracing backend (256M RAM, 0.5 CPU)
4. **grafana** - Observability visualization (256M RAM, 0.3 CPU)

Total resource allocation: ~1GB RAM, 1.8 CPU cores (fits comfortably on CPX11)

## Key Management

### Viewing loaded keys
```bash
# Check which keys are currently loaded
docker compose logs partner-service | grep "keys_loaded"

# List keys on filesystem
ls -la /opt/rafiki/keys/
```

### Rotating keys (zero downtime)
```bash
# 1. Generate NEW key (keeps old one)
./devops/setup-prod-keys.sh

# 2. Restart service (loads both keys)
docker compose restart partner-service

# 3. Update clients to use new kid
# 4. Wait 48 hours (token expiry)
# 5. Remove old key
rm /opt/rafiki/keys/old-key-id.pem
docker compose restart partner-service
```

See [docs/PRODUCTION_KEYS_SETUP.md](../docs/PRODUCTION_KEYS_SETUP.md) for detailed key management.

## Next Steps

- [x] Set up monitoring (Grafana + Tempo configured)
- [x] Set up JWT authentication
- [ ] Set up SSL/TLS with Let's Encrypt
- [ ] Configure reverse proxy (nginx/traefik)
- [ ] Set up automated backups (database + JWT keys)
- [ ] Configure log rotation
- [ ] Set up SSH key authentication (currently using password)
- [ ] Implement CI/CD pipeline
- [ ] Configure firewall rules for observability ports
- [ ] Document frontend authentication integration
