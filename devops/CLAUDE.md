# Rafiki DevOps Context

## Project Identity
- **Project Name**: Rafiki
- **Deployment Path**: `/opt/rafiki` (on server)
- **Main Branch**: `main`
- **Server IP**: `178.156.170.37`
- **Server User**: `root`

## Server Infrastructure
- **Provider**: Hetzner CPX11
- **OS**: Debian/Ubuntu
- **Docker Compose**: Uses modern plugin syntax (`docker compose` with space, NOT `docker-compose` with hyphen)

## Important Conventions
- **Branch naming**: `feature/{{linear-issue}}-description`, `fix/{{linear-issue}}-description`, or `internal/{{linear-issue}}-description`
- **NEVER force push to main/master branch**

## Deployment Scripts
All deployment scripts are in the `devops/` folder and auto-detect the project root:

- **`devops/copy-env.sh`**: Copies local `.env` file to server (run from local machine)
  - Auto-finds `.env` in project root regardless of where script is run from
  - Uses `sshpass` for password authentication
  - Target: `/opt/rafiki/.env`

- **`devops/deploy.sh`**: Full deployment on server (run on Hetzner server as root)
  - Auto-finds `.env` and project root
  - Changes to project root before running docker compose
  - Pulls latest git changes
  - Rebuilds and restarts all services

## Docker Compose Configuration

**File**: `docker-compose.yml` (project root)

**Services**:
1. **postgres** (`10.10.0.2:5432`)
   - Image: `postgres:18.0`
   - Container: `rafiki-postgres`
   - Exposed port: `5432`
   - Resources: 256M RAM, 0.5 CPU
   - Health check enabled

2. **tempo** (`10.10.0.20`)
   - Image: `grafana/tempo:2.8.1`
   - Container: `rafiki-tempo`
   - Ports: `3200` (query), `4317` (OTLP gRPC)
   - Resources: 256M RAM, 0.5 CPU
   - Used for distributed tracing

3. **grafana** (`10.10.0.21:3100`)
   - Image: `grafana/grafana:12.2.0`
   - Container: `rafiki-grafana`
   - Port: `3100`
   - Resources: 256M RAM, 0.3 CPU
   - Anonymous auth enabled (dev mode)
   - TraceQL editor enabled

4. **partner-service** (`10.10.0.10`)
   - Build: Local Dockerfile
   - Container: `rafiki-service`
   - Ports: `3000` (API), `3010` (Debug/Metrics)
   - Resources: 192M RAM, 0.5 CPU
   - Depends on: postgres, tempo
   - Go Runtime: GOMAXPROCS=2, GOMEMLIMIT=128MiB

**Network**: `rafiki-network` (10.10.0.0/24)

**Required Environment Variables** (in `.env`):
- `POSTGRES_DB`: Database name
- `POSTGRES_USER`: Database user
- `POSTGRES_PASSWORD`: Database password

## Key Services
- **API Port**: 3000
- **Debug/Metrics Port**: 3010
- **Database**: PostgreSQL (internal Docker network, port 5432)
  - DB Name: `rafiki`
  - DB User: `rafiki`
  - Host: `postgres` (internal Docker network name)

## Credentials Management
All credentials stored in Bitwarden:
- `rafiki-hetzner-server` - SSH access
- `rafiki-hetzner-postgres` - Database credentials
- `rafiki-hetzner-service` - Service environment variables

## Common Operations

### Full Deployment Workflow (Local → Server)

**From local machine:**
```bash
# 1. Push code changes to git
git push

# 2. Copy .env to server
./devops/copy-env.sh
```

**On server:**
```bash
# 3. SSH into server
ssh root@178.156.170.37

# 4. Run deployment script
/opt/rafiki/devops/deploy.sh
```

### Server Management

**View logs:**
```bash
docker compose logs -f                  # All services
docker compose logs -f partner-service  # Just the app
docker compose logs -f postgres         # Just database
```

**Check service health:**
```bash
curl http://localhost:3000/v1/readiness  # App readiness
curl http://localhost:3010/debug/readiness # Debug endpoint
docker compose ps                        # Container status
```

**Restart services:**
```bash
docker compose restart partner-service   # Restart app only
docker compose restart                   # Restart all services
```

**Database access:**
```bash
docker exec -it rafiki-postgres psql -U rafiki -d rafiki
```

### Service URLs (on server)
- **API**: `http://localhost:3000`
- **Debug/Metrics**: `http://localhost:3010`
- **Grafana**: `http://localhost:3100`
- **Tempo**: `http://localhost:3200`
- **PostgreSQL**: `localhost:5432`

## File Structure Notes
- All deployment scripts are in `devops/` directory
- Scripts auto-detect project root (parent of devops)
- Docker compose file is in project root (`docker-compose.yml`)
- Use `docker compose` (with space) for all Docker Compose commands
