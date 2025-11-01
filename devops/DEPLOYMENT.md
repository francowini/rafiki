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

#### Step 3: Run deployment script
```bash
# SSH into server
ssh root@178.156.170.37

# Run deployment (script auto-detects project root)
/opt/rafiki/devops/deploy.sh
```

The deployment script will:
1. Load environment variables from `.env`
2. Stop existing containers
3. Pull latest git changes
4. Build and start all services with Docker Compose
5. Wait for services to become healthy
6. Show deployment status

### Subsequent Deployments
Same as above - push code, copy env (if changed), run deploy script on server.

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

## Service Endpoints

### Main Application
- **API:** http://178.156.170.37:3000
- **Debug/Metrics:** http://178.156.170.37:3010
- **Health Check:** http://178.156.170.37:3000/v1/readiness
- **Liveness:** http://178.156.170.37:3000/v1/liveness

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

## Troubleshooting

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
2. Use SSH keys instead of passwords
3. Keep Docker and system packages updated
4. Configure HTTPS with Let's Encrypt (nginx/traefik)
5. Limit exposed ports via firewall
6. Regular backups of PostgreSQL data
7. Monitor logs for suspicious activity
8. Store production credentials securely (not in version control)

## Docker Compose Services

The application stack includes:

1. **partner-service** - Main Go application (192M RAM, 0.5 CPU)
2. **postgres** - PostgreSQL 18.0 database (256M RAM, 0.5 CPU)
3. **tempo** - Distributed tracing backend (256M RAM, 0.5 CPU)
4. **grafana** - Observability visualization (256M RAM, 0.3 CPU)

Total resource allocation: ~1GB RAM, 1.8 CPU cores (fits comfortably on CPX11)

## Next Steps

- [x] Set up monitoring (Grafana + Tempo configured)
- [ ] Set up SSL/TLS with Let's Encrypt
- [ ] Configure reverse proxy (nginx/traefik)
- [ ] Set up automated backups
- [ ] Configure log rotation
- [ ] Set up SSH key authentication (currently using password)
- [ ] Implement CI/CD pipeline
- [ ] Configure firewall rules for observability ports
