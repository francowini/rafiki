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
- **External Database**: Optional - use PlanetScale/Neon/RDS instead of local PostgreSQL (saves ~256MB RAM)
- **Complete Deployment Guide**: See [devops/DEPLOYMENT_GUIDE.md](devops/DEPLOYMENT_GUIDE.md) (includes everything: setup, deployment, troubleshooting, security, rollback)

### Service Endpoints
- API: `https://api.rafiki.lat` (production via nginx)
- Direct API: `http://localhost:3000` (backend only)
- Debug/Metrics: `http://localhost:3010`
- Frontend: `https://app.rafiki.lat` (Vercel)

## Code Quality & Development Guides

### Automated Code Quality (CodeRabbit Phase 1)

This project uses automated code quality tools:

- **CodeRabbit Pro**: AI-powered code review on all PRs
- **golangci-lint**: Go linting and formatting (backend)
- **Prettier + ESLint**: TypeScript/React linting and formatting (frontend)
- **Branch Protection**: Main branch blocks direct pushes (must use PRs)
- **Pre-commit Hooks**: Auto-formatting before commits (Phase 3)
- **GitHub Actions**: Automated testing and linting on PRs (Phase 2)

### Development Workflow

**IMPORTANT**: The `main` branch is protected. All changes must go through pull requests.

1. Create feature branch: `git checkout -b feature/my-feature`
2. Make changes and commit (pre-commit hooks will auto-format)
3. Push and create PR: `gh pr create`
4. CodeRabbit reviews automatically
5. Approve your PR (click "Approve" in GitHub)
6. Merge when approved (you can approve your own PRs)

### Comprehensive Development Guides

For detailed development workflows, code patterns, and best practices:

- **Backend (Go)**: [devops/BACKEND_DEVELOPMENT.md](devops/BACKEND_DEVELOPMENT.md)
  - golangci-lint configuration and usage
  - Business types pattern (CRITICAL)
  - Service architecture and patterns
  - Testing, profiling, and deployment

- **Frontend (TypeScript/Next.js)**: [devops/FRONTEND_DEVELOPMENT.md](devops/FRONTEND_DEVELOPMENT.md)
  - Prettier and ESLint configuration
  - Next.js 16 App Router patterns
  - Component patterns and best practices
  - API client and authentication

- **Deployment**: [devops/DEPLOYMENT_GUIDE.md](devops/DEPLOYMENT_GUIDE.md)
- **Branch Protection**: [scripts/setup-branch-protection.sh](scripts/setup-branch-protection.sh)

### Quick Formatting Commands

**Backend (Go)**:
```bash
# Auto-fix linter issues
golangci-lint run --fix

# Format code
gofmt -w -s .
goimports -w -local github.com/francowini/rafiki .
```

**Frontend (TypeScript)**:
```bash
cd frontend

# Auto-fix linter issues
npm run lint:fix

# Format code
npm run format

# Run all checks
npm run check
```

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

### Business Types with Validation
IMPORTANT: The business layer ALWAYS uses strong types with validation for domain values. Never use primitive types (int, string, float64) directly in business domain models.

**Pattern**: Create dedicated types in `business/types/` for any value that has validation rules:

```go
// Example: business/types/intensity/intensity.go
package intensity

import "fmt"

// Intensity represents a validated intensity value (0-10 scale).
type Intensity struct {
	value int
}

// Value returns the int value of the intensity.
func (i Intensity) Value() int {
	return i.value
}

// String returns the string representation.
func (i Intensity) String() string {
	return fmt.Sprintf("%d", i.value)
}

// Equal provides support for the go-cmp package and testing.
func (i Intensity) Equal(i2 Intensity) bool {
	return i.value == i2.value
}

// MarshalText provides support for logging and any marshal needs.
func (i Intensity) MarshalText() ([]byte, error) {
	return []byte(i.String()), nil
}

// Parse validates and creates an Intensity. This is where validation happens.
func Parse(value int) (Intensity, error) {
	if value < 0 || value > 10 {
		return Intensity{}, fmt.Errorf("intensity must be between 0 and 10, got %d", value)
	}
	return Intensity{value}, nil
}

// MustParse is like Parse but panics on error. Use in tests only.
func MustParse(value int) Intensity {
	intensity, err := Parse(value)
	if err != nil {
		panic(err)
	}
	return intensity
}
```

**Usage in Business Domain Models**:
```go
// business/domain/momentbus/model.go
type Moment struct {
	ID               uuid.UUID
	UserID           uuid.UUID
	Situation        content.Content    // Strong type for text content
	Intensity        intensity.Intensity // Strong type for 0-10 scale
	Cost             money.Money         // Strong type for monetary values
	Quantity         quantity.Quantity   // Strong type for quantities
	DateCreated      time.Time
}
```

**Existing Business Types**:
- `business/types/content` - Text content with validation
- `business/types/name` - Names with validation
- Create new types as needed following this pattern

**Why This Pattern**:
- Validation happens once at parse time
- Type system enforces valid data throughout the application
- Impossible to construct invalid business objects
- Clear separation of concerns: app layer parses primitives → business types
