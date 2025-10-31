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
  - `api/services/partner/`: Main service entry point
  - `api/services/partner/mux/`: HTTP route handlers
  - `api/services/api/debug/`: Debug endpoint handlers
  - `foundation/logger/`: Shared logging infrastructure
  - `api/tooling/logfmt/`: Log formatting tool

## Common Commands

### Development
```bash
# Run the service locally (with log formatting)
make run

# View help and available flags
make help

# Check version
make version
```

## Service Lifecycle

The main service ([api/services/partner/main.go](api/services/partner/main.go)) follows this pattern:

1. Initialize logger with events and trace ID function
2. Parse configuration from environment variables
3. Start debug server on separate goroutine (pprof, expvar, statsviz)
4. Start API server with graceful shutdown support
5. Handle SIGINT/SIGTERM for clean shutdown


## Deployment

Target platform: Hetzner CPX11 servers

- Deployment script: `./deploy.sh` (must run as root on server)
- Credentials stored in Bitwarden with specific naming convention
- Full deployment documentation: [devops/DEPLOYMENT.md](devops/DEPLOYMENT.md)
- Service endpoints:
  - API: `http://SERVER_IP:3000`
  - Debug/Metrics: `http://SERVER_IP:3010`

## Code Patterns

### Adding New Routes
Routes are registered in [api/services/partner/mux/mux.go](api/services/partner/mux/mux.go) using Go 1.22+ patterns:
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
