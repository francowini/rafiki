# Backend Development Guide

This guide covers backend development workflows, code quality tools, and best practices for the Rafiki Go API service.

## Tech Stack

- **Language**: Go 1.23+
- **Framework**: Standard library (`net/http`, Go 1.22+ router)
- **Database**: PostgreSQL 16
- **Configuration**: ardanlabs/conf/v3
- **Logging**: Custom structured logger (slog)
- **Deployment**: Docker on Hetzner VPS

## Quick Start

### Prerequisites

- Go 1.23+ installed
- PostgreSQL 16+ running
- Docker (optional, for containerized development)

### Setup

```bash
# Navigate to project root
cd /path/to/rafiki

# Install dependencies
go mod download

# Run the service
make run

# Or run directly
go run api/services/partners/main.go

# Service runs on:
# - API Server: http://localhost:3000
# - Debug Server: http://localhost:3010
```

## Code Quality Tools

The backend uses automated code quality tools as part of the CodeRabbit automation setup (Phase 1).

### golangci-lint (Linting & Formatting)

**Configuration**: [.golangci.yml](../.golangci.yml)

**Formatters (Auto-fixable - Tier 1)**:
- `gofmt`: Standard Go formatting
- `gofumpt`: Stricter gofmt variant
- `goimports`: Import organization
- `gci`: Custom import order (stdlib → external → local)

**Linters (Manual review - Tier 2/3)**:
- `errcheck`: Unchecked errors (critical for resource leaks)
- `govet`: Go vet checks
- `staticcheck`: Static analysis
- `gocritic`: Style and performance
- `revive`: Configurable linting
- `misspell`: Spelling errors
- `gosec`: Security issues (disabled - too strict)

**Commands**:
```bash
# Lint code
golangci-lint run

# Auto-fix safe issues
golangci-lint run --fix

# Lint specific path
golangci-lint run api/services/partners/...

# Runs automatically in CI
```

### Manual Formatting

```bash
# Format code (standard)
gofmt -w -s .

# Format code (strict)
gofumpt -w .

# Organize imports
goimports -w -local github.com/francowini/rafiki .
```

## Development Workflow

### Branch Strategy

**IMPORTANT**: The `main` branch is protected. All changes must go through pull requests.

1. **Create feature branch**:
   ```bash
   git checkout -b feature/my-feature
   # or: fix/bug-name, chore/task-name
   ```

2. **Make changes and commit**:
   ```bash
   git add .
   git commit -m "feat(auth): add JWT refresh token logic"
   # Pre-commit hooks auto-format code (Phase 3)
   ```

3. **Push and create PR**:
   ```bash
   git push origin feature/my-feature
   gh pr create --title "Add JWT refresh token logic" --body "Description"
   ```

4. **CodeRabbit review**:
   - CodeRabbit AI reviews PR automatically
   - Provides feedback categorized by tier
   - Security-sensitive paths get extra scrutiny

5. **Address feedback and merge**:
   - Fix any issues found
   - No approval required (but CodeRabbit feedback is helpful)
   - All CI checks must pass when enabled (Phase 2)

### Commit Message Format

Use conventional commits format:

```
<type>(<scope>): <subject>

<body>

<footer>
```

**Examples**:
```bash
git commit -m "feat(moments): add intensity validation"
git commit -m "fix(auth): prevent token refresh race condition"
git commit -m "chore(deps): update go.mod dependencies"
```

## Code Patterns

### Project Structure

```
rafiki/
├── api/
│   ├── services/
│   │   └── partners/           # Main service
│   │       ├── main.go         # Entry point
│   │       └── mux/            # HTTP handlers
│   ├── sdk/                    # SDK for external use (TODO)
│   └── tooling/
│       └── logfmt/             # Log formatting tool
├── business/
│   ├── domain/                 # Domain logic
│   │   ├── momentbus/          # Moment business logic
│   │   ├── thinkbus/           # Think business logic
│   │   └── userbus/            # User business logic
│   ├── types/                  # Business value types
│   │   ├── content/            # Text content validation
│   │   ├── intensity/          # 0-10 scale validation
│   │   ├── money/              # Monetary values
│   │   └── name/               # Name validation
│   └── sdk/
│       └── sqldb/              # Database utilities
├── foundation/
│   ├── logger/                 # Structured logging
│   └── keystore/               # Cryptographic keys
└── devops/                     # Deployment scripts
```

### Service Architecture

**Main Service**: `api/services/partners/main.go`

The service follows this lifecycle:

1. **Initialize logger** with events and trace ID function
2. **Parse configuration** from environment variables (prefix: `PARTNER_`)
3. **Start debug server** on separate goroutine (`:3010`)
   - pprof endpoints: `/debug/pprof/*`
   - expvar: `/debug/vars`
   - statsviz: `/debug/statsviz/*`
4. **Start API server** with graceful shutdown (`:3000`)
5. **Handle SIGINT/SIGTERM** for clean shutdown

**Example**:
```go
// api/services/partners/main.go
func main() {
    log := logger.New(os.Stdout, logger.LevelInfo, "PARTNER", events)

    cfg := struct {
        Web struct {
            APIHost string `conf:"default:0.0.0.0:3000"`
            DebugHost string `conf:"default:0.0.0.0:3010"`
        }
        DB struct {
            Host string `conf:"default:localhost"`
        }
    }{}

    // ... start servers
}
```

### HTTP Routing (Go 1.22+)

**Pattern**: Use standard library `http.ServeMux` with pattern matching

**File**: `api/services/partners/mux/mux.go`

```go
func Routes(log *logger.Logger) http.Handler {
    mux := http.NewServeMux()

    // Go 1.22+ pattern matching
    mux.HandleFunc("GET /v1/moments", getMoments)
    mux.HandleFunc("POST /v1/moments", createMoment)
    mux.HandleFunc("GET /v1/moments/{id}", getMoment)
    mux.HandleFunc("PUT /v1/moments/{id}", updateMoment)
    mux.HandleFunc("DELETE /v1/moments/{id}", deleteMoment)

    // Path variables available in request context
    mux.HandleFunc("GET /v1/users/{userId}/moments", getUserMoments)

    return mux
}
```

### Business Types Pattern

**CRITICAL**: Always use strong types from `business/types/` for domain values. Never use primitives.

**Why**:
- Validation happens once at parse time
- Type system enforces valid data
- Impossible to construct invalid business objects

**Example**: Intensity (0-10 scale)

```go
// business/types/intensity/intensity.go
package intensity

type Intensity struct {
    value int
}

func Parse(value int) (Intensity, error) {
    if value < 0 || value > 10 {
        return Intensity{}, fmt.Errorf("intensity must be 0-10, got %d", value)
    }
    return Intensity{value}, nil
}

func (i Intensity) Value() int {
    return i.value
}
```

**Usage in Domain Models**:

```go
// business/domain/momentbus/model.go
package momentbus

import (
    "github.com/francowini/rafiki/business/types/content"
    "github.com/francowini/rafiki/business/types/intensity"
    "github.com/google/uuid"
    "time"
)

type Moment struct {
    ID          uuid.UUID
    UserID      uuid.UUID
    Situation   content.Content      // Strong type, not string
    Intensity   intensity.Intensity  // Strong type, not int
    DateCreated time.Time
}
```

**App Layer Parsing**:

```go
// api/services/partners/mux/handlers.go
func createMoment(w http.ResponseWriter, r *http.Request) {
    var req struct {
        Situation string `json:"situation"`
        Intensity int    `json:"intensity"`
    }
    json.NewDecoder(r.Body).Decode(&req)

    // Parse primitives → business types
    situation, err := content.Parse(req.Situation)
    if err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    intensity, err := intensity.Parse(req.Intensity)
    if err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    // Business layer receives validated types
    moment := momentbus.NewMoment{
        Situation: situation,
        Intensity: intensity,
    }
}
```

### Existing Business Types

- **`business/types/content`**: Text content with validation
- **`business/types/name`**: Names with validation
- **`business/types/intensity`**: 0-10 scale (TODO: create this)
- **`business/types/money`**: Monetary values (TODO: create this)

**Pattern to follow**: See `content` and `name` implementations.

### Logging Pattern

**Configuration**: `foundation/logger`

```go
import "github.com/francowini/rafiki/foundation/logger"

// Initialize logger
log := logger.New(os.Stdout, logger.LevelInfo, "PARTNER", events)

// Log with context
log.Info(ctx, "moment created", "moment_id", momentID, "user_id", userID)

// Log errors
log.Error(ctx, "failed to create moment", "err", err)

// Log with trace ID (automatically added from context)
ctx = context.WithValue(ctx, "trace_id", uuid.New().String())
log.Info(ctx, "processing request") // trace_id auto-added
```

### Configuration Pattern

**Library**: `ardanlabs/conf/v3`

```go
cfg := struct {
    Web struct {
        APIHost string `conf:"default:0.0.0.0:3000"`
    }
    DB struct {
        Host     string `conf:"default:localhost"`
        User     string `conf:"default:postgres"`
        Password string `conf:"default:postgres,mask"`
    }
}{}

// Parse from environment: PARTNER_WEB_APIHOST, PARTNER_DB_HOST, etc.
if err := conf.Parse("PARTNER", &cfg); err != nil {
    log.Error(ctx, "parsing config", "err", err)
    return err
}
```

### Database Patterns

**Utilities**: `business/sdk/sqldb`

```go
// Query with proper error handling
rows, err := db.Query(ctx, "SELECT * FROM moments WHERE user_id = $1", userID)
if err != nil {
    return nil, fmt.Errorf("query moments: %w", err)
}
defer func() {
    if err := rows.Close(); err != nil {
        log.Error(ctx, "close rows", "err", err)
    }
}()

// Always handle Close() errors to prevent resource leaks
```

## Testing

### Unit Tests

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run tests for specific package
go test ./business/types/content

# Run with race detector
go test -race ./...
```

### Test Patterns

```go
func TestIntensityParse(t *testing.T) {
    tests := []struct {
        name    string
        value   int
        wantErr bool
    }{
        {"valid min", 0, false},
        {"valid max", 10, false},
        {"invalid low", -1, true},
        {"invalid high", 11, true},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            _, err := intensity.Parse(tt.value)
            if (err != nil) != tt.wantErr {
                t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
            }
        })
    }
}
```

## Building and Running

### Local Development

```bash
# Run with live-reload (using make)
make run

# Run directly
go run api/services/partners/main.go

# Build binary
go build -o bin/partners api/services/partners/main.go

# Run binary
./bin/partners
```

### Docker

```bash
# Build image
docker build -t rafiki-api -f api/services/partners/Dockerfile .

# Run container
docker run -p 3000:3000 -p 3010:3010 \
  -e PARTNER_DB_HOST=postgres \
  rafiki-api
```

### Environment Variables

Create `.env` file in project root:

```bash
# Database
PARTNER_DB_HOST=localhost
PARTNER_DB_USER=postgres
PARTNER_DB_PASSWORD=postgres
PARTNER_DB_NAME=rafiki
PARTNER_DB_DISABLE_TLS=true

# Web
PARTNER_WEB_APIHOST=0.0.0.0:3000
PARTNER_WEB_DEBUGHOST=0.0.0.0:3010

# Auth
PARTNER_AUTH_KEYS_FOLDER=keys/
```

**IMPORTANT**: Never commit `.env` files with secrets.

## CodeRabbit Integration

### Path-Specific Rules

CodeRabbit has special instructions for backend paths:

- **`business/types/**`**: Business validation types - **NEVER auto-fix**
- **`business/domain/**/model.go`**: Domain models (API contracts) - manual review
- **`app/sdk/auth/**`**: Authentication - security-sensitive
- **`foundation/keystore/**`**: Cryptographic keys - security-sensitive
- **`api/services/partners/main.go`**: Main entry point - careful review

### Review Tiers

**Tier 1 (Auto-fix)**: CodeRabbit auto-fixes and commits
- Formatting (gofmt, gofumpt, goimports)
- Import organization (gci)
- Spelling errors (misspell)

**Tier 2 (Interactive)**: CodeRabbit suggests, you approve
- Unchecked errors (errcheck)
- Go vet issues (govet)
- Code improvements (gocritic, revive)

**Tier 3 (Manual)**: Flagged for manual review
- Business type validation changes
- API contract changes
- Authentication/authorization logic
- Database schema changes

## Common Issues

### Resource Leaks

**Issue**: Unchecked errors on Close()

```go
// ❌ Bad: Ignores close error
defer rows.Close()

// ✅ Good: Handles close error
defer func() {
    if err := rows.Close(); err != nil {
        log.Error(ctx, "close rows", "err", err)
    }
}()
```

**Recent fixes** (Task 1.2):
- `api/services/partners/main.go:152,302` - DB/server close errors
- `foundation/keystore/keystore.go:100` - File close error
- `business/sdk/sqldb/sqldb.go:216,292` - SQL rows close errors

### Import Organization

**Issue**: Imports not organized correctly

```bash
# Fix: Run gci
goimports -w -local github.com/francowini/rafiki .
```

**Order**: stdlib → external → local packages

```go
import (
    // Standard library
    "context"
    "fmt"

    // External packages
    "github.com/google/uuid"

    // Local packages
    "github.com/francowini/rafiki/business/types/content"
)
```

### Business Type Misuse

**Issue**: Using primitives instead of business types

```go
// ❌ Bad: No validation
type Moment struct {
    Intensity int // Could be -100 or 1000!
}

// ✅ Good: Validated business type
import "github.com/francowini/rafiki/business/types/intensity"

type Moment struct {
    Intensity intensity.Intensity // Guaranteed 0-10
}
```

## Performance Best Practices

### Database Connection Pooling

```go
// Configure connection pool
db.SetMaxOpenConns(25)
db.SetMaxIdleConns(25)
db.SetConnMaxLifetime(5 * time.Minute)
```

### Context Timeouts

```go
// Set timeout for database operations
ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
defer cancel()

rows, err := db.QueryContext(ctx, query)
```

### Profiling

Debug server exposes profiling endpoints:

```bash
# CPU profile
curl http://localhost:3010/debug/pprof/profile?seconds=30 > cpu.prof
go tool pprof cpu.prof

# Heap profile
curl http://localhost:3010/debug/pprof/heap > heap.prof
go tool pprof heap.prof

# Live stats visualization
open http://localhost:3010/debug/statsviz/
```

## Deployment

Backend is deployed to **Hetzner VPS** via Docker.

### Production Server

- **Host**: 178.156.170.37
- **SSH**: `ssh root@178.156.170.37`
- **Project**: `/opt/rafiki`

### Deployment Commands

```bash
# From local machine
make deploy

# On server
cd /opt/rafiki
git pull origin main
sudo ./devops/deploy.sh
```

### Service Endpoints

- **API**: https://api.rafiki.lat (via nginx)
- **Direct**: http://localhost:3000 (backend only)
- **Debug**: http://localhost:3010 (not publicly exposed)

### Configuration

See:
- [Deployment Guide](./DEPLOYMENT_GUIDE.md)
- [Auth Deployment Guide](./AUTH_DEPLOYMENT_GUIDE.md)
- [Docker Cleanup](./DOCKER_CLEANUP.md)

## Getting Help

- **Linter errors**: Run `golangci-lint run --fix` first
- **Build errors**: Check `go.mod` dependencies with `go mod tidy`
- **Test failures**: Run `go test -v ./...` for verbose output
- **CodeRabbit issues**: Check `.coderabbit.yaml` path instructions

## Related Documentation

- [Frontend Development Guide](./FRONTEND_DEVELOPMENT.md)
- [Deployment Guide](./DEPLOYMENT_GUIDE.md)
- [Auth Deployment Guide](./AUTH_DEPLOYMENT_GUIDE.md)
- [CodeRabbit Configuration](../docs/coderabbit-automation-implementation-plan.md)
