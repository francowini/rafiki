# Rafiki DevOps Context

## Project Identity
- **Project Name**: Rafiki
- **Deployment Path**: `/opt/rafiki` (on server)
- **Main Branch**: `main`

## Server Infrastructure
- **Provider**: Hetzner CPX11
- **OS**: Debian/Ubuntu
- **Docker Compose**: Uses modern plugin syntax (`docker compose` with space, NOT `docker-compose` with hyphen)

## Important Conventions
- **Branch naming**: `feature/{{linear-issue}}-description`, `fix/{{linear-issue}}-description`, or `internal/{{linear-issue}}-description`
- **NEVER force push to main/master branch**

## Deployment Configuration
- **Production compose file**: `devops/docker-compose.prod.yml`
- **Deployment script**: `deploy.sh` (in project root)
- **Deploy to server script**: `devops/deploy-to-hetzner.sh`
- **Required env file**: `.env` (must be present on server at `/opt/rafiki/.env`)

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

### Deploy to production
```bash
cd /opt/rafiki
./deploy.sh
```

### View logs
```bash
docker compose logs -f
docker compose logs -f partner-service
```

### Check service health
```bash
curl http://localhost:3010/debug/readiness
```

## File Structure Notes
- Main deployment scripts are in project root (`deploy.sh`)
- DevOps documentation and configs are in `devops/` directory
- Use `docker compose` (with space) for all Docker Compose commands
