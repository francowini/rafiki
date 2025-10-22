# Topifier Deployment Guide - Hetzner CPX11

## Prerequisites

- Hetzner CPX11 server with Ubuntu/Debian
- Docker and Docker Compose installed
- Git installed
- Root or sudo access

## Bitwarden Credentials Structure

Store the following items in Bitwarden for easy DevOps access:

### 1. Hetzner Server Access
**Bitwarden Item Name:** `topifier-hetzner-server`
- **SSH Host:** Your server IP address
- **SSH Port:** 22 (or custom)
- **SSH User:** root or your user
- **SSH Key:** Private key or password

### 2. PostgreSQL Credentials
**Bitwarden Item Name:** `topifier-hetzner-postgres`
- **Database Name:** topifier
- **Database User:** topifier
- **Database Password:** [secure password]
- **Database Host:** postgres (internal docker network)
- **Database Port:** 5432

### 3. Service Environment Variables
**Bitwarden Item Name:** `topifier-hetzner-service`
- Add any API keys, tokens, or secrets needed by your service
- Store the `.env` file contents as a secure note

## Initial Server Setup

### 1. Connect to Hetzner Server
```bash
# Get credentials from Bitwarden: topifier-hetzner-server
ssh root@YOUR_SERVER_IP
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

### 3. Create deployment directory
```bash
mkdir -p /opt/topifier
cd /opt/topifier
```

### 4. Clone or upload your code
```bash
# Option 1: Using git
git clone YOUR_REPO_URL .

# Option 2: Manual upload
# Use scp from your local machine:
# scp -r /path/to/topifier/* root@YOUR_SERVER_IP:/opt/topifier/
```

### 5. Create .env file
```bash
# Copy example file
cp .env.example .env

# Edit with credentials from Bitwarden: topifier-hetzner-postgres
nano .env
```

Set these variables (get values from Bitwarden):
```
POSTGRES_DB=topifier
POSTGRES_USER=topifier
POSTGRES_PASSWORD=YOUR_SECURE_PASSWORD_FROM_BITWARDEN
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

## Deployment

### First Deployment
```bash
cd /opt/topifier
chmod +x deploy.sh
./deploy.sh
```

### Subsequent Deployments
```bash
cd /opt/topifier
git pull  # If using git
./deploy.sh
```

## Management Commands

### View logs
```bash
# All services
docker-compose logs -f

# Specific service
docker-compose logs -f partner-service
docker-compose logs -f postgres
```

### Restart services
```bash
docker-compose restart
```

### Stop services
```bash
docker-compose down
```

### Check service status
```bash
docker-compose ps
```

### Access PostgreSQL
```bash
# Get password from Bitwarden: topifier-hetzner-postgres
docker-compose exec postgres psql -U topifier -d topifier
```

## Service Endpoints

- **API:** http://YOUR_SERVER_IP:3000
- **Debug/Metrics:** http://YOUR_SERVER_IP:3010
- **Health Check:** http://YOUR_SERVER_IP:3010/debug/readiness

## Backup PostgreSQL

```bash
# Create backup
docker-compose exec postgres pg_dump -U topifier topifier > backup_$(date +%Y%m%d_%H%M%S).sql

# Restore backup
cat backup_file.sql | docker-compose exec -T postgres psql -U topifier topifier
```

## Monitoring

### Check service health
```bash
curl http://localhost:3010/debug/readiness
```

### View metrics
```bash
curl http://localhost:3010/debug/vars
```

## Troubleshooting

### Service won't start
```bash
# Check logs
docker-compose logs

# Check if ports are available
netstat -tulpn | grep -E '3000|3010|5432'
```

### Database connection issues
```bash
# Verify PostgreSQL is running
docker-compose ps postgres

# Check PostgreSQL logs
docker-compose logs postgres

# Test connection
docker-compose exec postgres pg_isready -U topifier
```

### Clean restart
```bash
# Stop everything
docker-compose down

# Remove volumes (WARNING: This deletes data!)
docker-compose down -v

# Rebuild and start
docker-compose up -d --build
```

## Security Recommendations

1. Change default passwords (stored in Bitwarden)
2. Use SSH keys instead of passwords
3. Keep Docker and system packages updated
4. Configure HTTPS with Let's Encrypt (nginx/traefik)
5. Limit exposed ports via firewall
6. Regular backups of PostgreSQL data
7. Monitor logs for suspicious activity

## Accessing Credentials via Bitwarden CLI

Install Bitwarden CLI for automated deployments:

```bash
# Install bw CLI
npm install -g @bitwarden/cli

# Login
bw login

# Get credentials
bw get item topifier-hetzner-postgres
bw get password topifier-hetzner-postgres
```

## Next Steps

- [ ] Set up SSL/TLS with Let's Encrypt
- [ ] Configure reverse proxy (nginx/traefik)
- [ ] Set up automated backups
- [ ] Configure log rotation
- [ ] Set up monitoring (Prometheus/Grafana)
- [ ] Implement CI/CD pipeline
