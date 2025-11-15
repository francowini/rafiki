# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Topifier is a personal development tracking application built in Go, designed to help track ideales (ideals), valores (values), hábitos (habits), metas (goals), and objetivos (objectives). The service is containerized and deployed on Hetzner servers.

## Architecture

- **Service Pattern**: Main service is `partner-service` with dual HTTP servers:
  - API Server (port 3000): Main application endpoints
  - Debug Server (port 3010): Debug/profiling endpoints (pprof, expvar, statsviz)
- **Configuration**: Uses `ardanlabs/conf/v3` with environment variables prefixed by `PARTNER_`
- **Logging**: Custom structured logger (`foundation/logger`) using Go's `slog` package with JSON output
- **HTTP Routing**: Standard library `http.ServeMux` (Go 1.22+ pattern matching)
- **Project Structure**:
  - `api/services/partners/`: Main service entry point
  - `api/services/partners/mux/`: HTTP route handlers
  - `api/services/api/debug/`: Debug endpoint handlers
  - `foundation/logger/`: Shared logging infrastructure
  - `api/tooling/logfmt/`: Log formatting tool

## Common Commands

### Development
```bash
# Run the service locally (with log formatting)
make run

# Start all services with Docker Compose
make up

# View logs
make logs

# Check health
make health

# Stop services
make down
```

## Service Lifecycle

The main service ([api/services/partners/main.go](api/services/partners/main.go)) follows this pattern:

1. Initialize logger with events and trace ID function
2. Parse configuration from environment variables
3. Start debug server on separate goroutine (pprof, expvar, statsviz)
4. Start API server with graceful shutdown support
5. Handle SIGINT/SIGTERM for clean shutdown


## Deployment

Target platform: Hetzner CPX11 servers (178.156.170.37)

### Production Deployment (from local machine)
```bash
# Deploy with one command
make deploy

# View production logs
make deploy-logs

# Check production status
make deploy-status

# SSH to server
make ssh
```

### Production Deployment (on server)
```bash
# SSH to server
ssh root@178.156.170.37

# Pull latest changes and deploy
cd /opt/rafiki
git pull origin main
sudo ./devops/deploy.sh
```

### Important Notes
- **One-Time Setup**: JWT keys, ADMIN user, and .env file are created ONCE only
- **Regular Deployments**: Only run `make deploy` or `./devops/deploy.sh`
- **Database Migrations**: Run automatically on every deployment (idempotent)
- **Full Guide**: See [devops/DEPLOYMENT_GUIDE.md](devops/DEPLOYMENT_GUIDE.md)

### Service Endpoints
- API: `https://api.rafiki.lat` (production via nginx)
- Direct API: `http://localhost:3000` (backend only)
- Debug/Metrics: `http://localhost:3010`
- Frontend: `https://app.rafiki.lat` (Vercel)

## Code Patterns

### Adding New Routes
Routes are registered in [api/services/partners/mux/mux.go](api/services/partners/mux/mux.go) using Go 1.22+ patterns:
```go
mux.HandleFunc("GET /endpoint", handlerFunc)
```

### Logging
Use the structured logger with context:
```go
log.Info(ctx, "message", "key", value)
log.Error(ctx, "error message", "err", err)
```

### Service Configuration
Add new config fields to the `cfg` struct in `main.go` with `conf` tags for defaults and environment variable mapping.
