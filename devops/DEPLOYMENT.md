# Rafiki Deployment Guide - Hetzner CPX11

## Prerequisites

- Hetzner CPX11 server with Ubuntu/Debian
- Docker and Docker Compose installed
- Git installed
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

## Deployment

### First Deployment
```bash
cd /opt/rafiki
chmod +x deploy.sh
./deploy.sh
```

### Subsequent Deployments
```bash
cd /opt/rafiki
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
docker-compose exec postgres psql -U rafiki -d rafiki
# Password: db
```

## Service Endpoints

- **API:** http://YOUR_SERVER_IP:3000
- **Debug/Metrics:** http://YOUR_SERVER_IP:3010
- **Health Check:** http://YOUR_SERVER_IP:3000/v1/readiness
- **Liveness:** http://YOUR_SERVER_IP:3000/v1/liveness

## Backup PostgreSQL

```bash
# Create backup
docker-compose exec postgres pg_dump -U rafiki rafiki > backup_$(date +%Y%m%d_%H%M%S).sql

# Restore backup
cat backup_file.sql | docker-compose exec -T postgres psql -U rafiki rafiki
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
docker-compose exec postgres pg_isready -U rafiki
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

1. ⚠️ **Change default password "db" in production!**
2. Use SSH keys instead of passwords
3. Keep Docker and system packages updated
4. Configure HTTPS with Let's Encrypt (nginx/traefik)
5. Limit exposed ports via firewall
6. Regular backups of PostgreSQL data
7. Monitor logs for suspicious activity
8. Store production credentials securely (not in version control)

## Next Steps

- [ ] Set up SSL/TLS with Let's Encrypt
- [ ] Configure reverse proxy (nginx/traefik)
- [ ] Set up automated backups
- [ ] Configure log rotation
- [ ] Set up monitoring (Prometheus/Grafana)
- [ ] Implement CI/CD pipeline
