# OpenTelemetry Observability with Docker Compose

Complete guide to setting up and using OpenTelemetry (OTEL), Tempo, and Grafana with Docker Compose.

---

## Table of Contents

1. [What is Tempo?](#what-is-tempo)
2. [What is Grafana?](#what-is-grafana)
3. [How OTEL Works in This Service](#how-otel-works-in-this-service)
4. [Complete Docker Compose Setup](#complete-docker-compose-setup)
5. [Step-by-Step Implementation](#step-by-step-implementation)
6. [Viewing Traces and Logs](#viewing-traces-and-logs)
7. [Troubleshooting](#troubleshooting)
8. [Advanced Configuration](#advanced-configuration)

---

## What is Tempo?

**Grafana Tempo** is a distributed tracing backend - essentially a "database for traces."

### Key Features:
- **High-scale, low-cost**: Uses object storage (S3, GCS, or local disk)
- **OTLP Compatible**: Accepts traces via OpenTelemetry Protocol (gRPC/HTTP)
- **Minimal indexing**: Stores complete trace data with minimal overhead
- **TraceQL**: Query traces using a powerful query language

### How It Works in Your Service:

```go
// api/services/sales/main.go:197-212
traceProvider, teardown, err := otel.InitTracing(log, otel.Config{
    ServiceName: cfg.Tempo.ServiceName,  // "sales"
    Host:        cfg.Tempo.Host,          // "tempo:4317" (OTLP gRPC endpoint)
    Probability: cfg.Tempo.Probability,  // 0.05 = 5% sampling rate
})
```

**Flow:**
1. Your Go service generates spans using OpenTelemetry SDK
2. Spans are sent to Tempo via OTLP gRPC (port 4317)
3. Tempo stores the traces
4. Grafana queries Tempo to display traces in the UI

---

## What is Grafana?

**Grafana** is an observability platform for visualizing metrics, logs, and traces.

### In Your Setup:
- **Version**: 12.2.0
- **Purpose**: Unified dashboard for all observability data
- **Datasource**: Tempo (traces)
- **Features**:
  - TraceQL editor for advanced trace queries
  - Anonymous admin access (dev mode)
  - Real-time trace visualization
  - Correlation between logs and traces

---

## How OTEL Works in This Service

### Architecture Overview

```
┌─────────────────────────────────────────────────────────────┐
│                    HTTP Request Flow                         │
└─────────────────────────────────────────────────────────────┘
                              ↓
┌─────────────────────────────────────────────────────────────┐
│  HTTP Middleware (mid.Otel)                                  │
│  - Injects tracer into context                               │
│  - Creates trace ID                                          │
│  Location: app/sdk/mid/otel.go                               │
└─────────────────────────────────────────────────────────────┘
                              ↓
┌─────────────────────────────────────────────────────────────┐
│  Business Layer (userotel, productbus, homebus)              │
│  - Each operation creates a span                             │
│  - Example: otel.AddSpan(ctx, "business.userbus.create")     │
│  Location: business/domain/*/                                │
└─────────────────────────────────────────────────────────────┘
                              ↓
┌─────────────────────────────────────────────────────────────┐
│  Database Layer (sqldb)                                      │
│  - Database queries are traced                               │
│  Location: business/sdk/sqldb/sqldb.go                       │
└─────────────────────────────────────────────────────────────┘
                              ↓
┌─────────────────────────────────────────────────────────────┐
│  OTLP Exporter                                               │
│  - Batches spans                                             │
│  - Sends to Tempo via gRPC (port 4317)                       │
│  Location: foundation/otel/otel.go                           │
└─────────────────────────────────────────────────────────────┘
                              ↓
                          Tempo Storage
```

### Key Components

#### 1. Foundation Layer (`foundation/otel/`)

**`otel.go`**: Core tracing initialization
```go
// Initializes trace provider with OTLP exporter
InitTracing(log, cfg) -> (TracerProvider, teardown, error)

// Injects tracer and trace ID into context
InjectTracing(ctx, tracer) -> context.Context

// Creates spans for operations
AddSpan(ctx, spanName, attributes...) -> (context.Context, Span)

// Extracts trace ID from context (for logging)
GetTraceID(ctx) -> string
```

**`sampler.go`**: Custom sampling logic
- Excludes health check endpoints (`/v1/liveness`, `/v1/readiness`)
- Probabilistic sampling (default 5%)
- Prevents trace spam from monitoring tools

**`context.go`**: Context management
- Stores tracer and trace ID in context
- Thread-safe context operations

#### 2. Middleware Layer (`app/sdk/mid/otel.go`)

Automatically applied to all HTTP requests:
```go
func Otel(tracer trace.Tracer) web.MidFunc {
    return func(next web.HandlerFunc) web.HandlerFunc {
        return func(ctx context.Context, r *http.Request) web.Encoder {
            ctx = otel.InjectTracing(ctx, tracer)
            return next(ctx, r)
        }
    }
}
```

Applied in: `app/sdk/mux/mux.go:96`

#### 3. Business Layer Extensions

**User Operations** (`business/domain/userbus/extensions/userotel/`)
```go
func (ext *Extension) Create(ctx context.Context, actorID uuid.UUID, nu userbus.NewUser) (userbus.User, error) {
    ctx, span := otel.AddSpan(ctx, "business.userbus.create")
    defer span.End()

    return ext.bus.Create(ctx, actorID, nu)
}
```

Similar patterns in:
- `business/domain/productbus/productbus.go:87` - Product operations
- `business/domain/homebus/homebus.go` - Home operations
- `business/domain/auditbus/auditbus.go` - Audit operations

#### 4. Trace Propagation

For downstream service calls:
```go
// app/sdk/authclient/authclient.go
otel.AddTraceToRequest(ctx, r)
```

This injects trace context into HTTP headers so traces span across services.

### Configuration

Environment variables:
```bash
SALES_TEMPO_HOST=tempo:4317           # Tempo endpoint
SALES_TEMPO_SERVICE_NAME=sales        # Service identifier in traces
SALES_TEMPO_PROBABILITY=0.05          # 5% sampling (use 1.0 for 100%)
```

Defined in: `api/services/sales/main.go:110-117`

### Trace ID in Logs

Every log entry includes the trace ID for correlation:
```go
// api/services/sales/main.go:61-65
traceIDFn := func(ctx context.Context) string {
    return otel.GetTraceID(ctx)
}
log = logger.NewWithEvents(os.Stdout, logger.LevelInfo, "SALES", traceIDFn, events)
```

**Log Output:**
```json
{"service":"SALES","ts":"2024-10-31T12:34:56Z","trace_id":"abc123...","msg":"user.create","user_id":"..."}
```

---

## Complete Docker Compose Setup

### Overview

The current `docker_compose.yaml` only includes the application services. We'll add Tempo and Grafana for complete observability.

### Architecture

```
┌──────────────────────────────────────────────────────────┐
│                    Docker Network                         │
│                  sales-system-network                     │
│                                                           │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐              │
│  │PostgreSQL│  │   Auth   │  │  Sales   │              │
│  │  :5432   │  │  :6000   │  │  :3000   │              │
│  └──────────┘  └──────────┘  └──────────┘              │
│                      │            │                      │
│                      └────────────┘                      │
│                            │ OTLP                        │
│                            ↓                             │
│                   ┌──────────────┐                       │
│                   │    Tempo     │                       │
│                   │    :4317     │                       │
│                   │    :3200     │                       │
│                   └──────────────┘                       │
│                            ↑                             │
│                            │ Query                       │
│                   ┌──────────────┐                       │
│                   │   Grafana    │                       │
│                   │    :3100     │                       │
│                   └──────────────┘                       │
└──────────────────────────────────────────────────────────┘
```

---

## Step-by-Step Implementation

### Step 1: Create Tempo Configuration File

Create file: `zarf/compose/tempo-config.yaml`

```yaml
usage_report:
  reporting_enabled: false

server:
  http_listen_port: 3200

distributor:
  receivers:
    otlp:
      protocols:
        http:
          endpoint: "0.0.0.0:4318"
        grpc:
          endpoint: "0.0.0.0:4317"

ingester:
  trace_idle_period: 10s
  max_block_bytes: 1_000_000
  max_block_duration: 5m

compactor:
  compaction:
    compaction_window: 1h
    max_block_bytes: 100_000_000
    block_retention: 1h
    compacted_block_retention: 10m

storage:
  trace:
    backend: local
    block:
      bloom_filter_false_positive: .05
      v2_index_downsample_bytes: 1000
      v2_encoding: zstd
    wal:
      path: /tmp/tempo/wal
      v2_encoding: snappy
    local:
      path: /tmp/tempo/blocks
    pool:
      max_workers: 100
      queue_depth: 10000
```

### Step 2: Create Grafana Datasource Configuration

Create file: `zarf/compose/grafana-datasources.yaml`

```yaml
apiVersion: 1

datasources:
  - name: Tempo
    type: tempo
    access: proxy
    url: http://tempo:3200
    isDefault: true
    jsonData:
      httpMethod: GET
      nodeGraph:
        enabled: true
      search:
        hide: false
    editable: true
```

### Step 3: Create Complete Docker Compose File

Create file: `zarf/compose/docker_compose_with_observability.yaml`

```yaml
services:
  # ==============================================================================
  # Database Service
  # ==============================================================================

  database:
    image: postgres:18.0
    container_name: database
    restart: unless-stopped
    ports:
      - "5432:5432"
    environment:
      - POSTGRES_PASSWORD=postgres
    volumes:
      - ./database-data:/var/lib/postgresql/data
      - ./pg_hba.conf:/etc/pg_hba.conf
    command: ["-c", "hba_file=/etc/pg_hba.conf"]
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -h localhost -U postgres"]
      interval: 10s
      timeout: 5s
      retries: 5
      start_period: 30s
    networks:
      sales-system-network:
        ipv4_address: 10.5.0.2

  # ==============================================================================
  # Database Initialization
  # ==============================================================================

  init-migrate-seed:
    image: localhost/ardanlabs/sales:0.0.1
    pull_policy: never
    container_name: init-migrate-seed
    restart: "no"
    entrypoint: ["./admin", "migrate-seed"]
    environment:
      - SALES_DB_USER=postgres
      - SALES_DB_PASSWORD=postgres
      - SALES_DB_HOST=database
      - SALES_DB_DISABLE_TLS=true
    networks:
      sales-system-network:
        ipv4_address: 10.5.0.10
    depends_on:
      database:
        condition: service_healthy

  # ==============================================================================
  # Observability Stack
  # ==============================================================================

  tempo:
    image: grafana/tempo:2.8.1
    container_name: tempo
    restart: unless-stopped
    command: ["-config.file=/etc/tempo.yaml"]
    volumes:
      - ./tempo-data:/tmp/tempo
      - ./tempo-config.yaml:/etc/tempo.yaml:ro
    ports:
      - "3200:3200"   # Tempo query port
      - "4317:4317"   # OTLP gRPC
      - "4318:4318"   # OTLP HTTP
    networks:
      sales-system-network:
        ipv4_address: 10.5.0.30
    healthcheck:
      test: ["CMD-SHELL", "wget -qO- http://localhost:3200/ready || exit 1"]
      interval: 10s
      timeout: 5s
      retries: 5
      start_period: 10s

  grafana:
    image: grafana/grafana:12.2.0
    container_name: grafana
    restart: unless-stopped
    environment:
      - GF_AUTH_ANONYMOUS_ENABLED=true
      - GF_AUTH_ANONYMOUS_ORG_ROLE=Admin
      - GF_AUTH_DISABLE_LOGIN_FORM=true
      - GF_FEATURE_TOGGLES_ENABLE=traceqlEditor
      - GF_SERVER_HTTP_PORT=3100
    volumes:
      - ./grafana-datasources.yaml:/etc/grafana/provisioning/datasources/datasources.yaml:ro
      - grafana-data:/var/lib/grafana
    ports:
      - "3100:3100"
    networks:
      sales-system-network:
        ipv4_address: 10.5.0.31
    depends_on:
      tempo:
        condition: service_healthy
    healthcheck:
      test: ["CMD-SHELL", "wget -qO- http://localhost:3100/api/health || exit 1"]
      interval: 10s
      timeout: 5s
      retries: 5
      start_period: 10s

  # ==============================================================================
  # Application Services
  # ==============================================================================

  auth:
    image: localhost/ardanlabs/auth:0.0.1
    pull_policy: never
    container_name: auth
    restart: unless-stopped
    ports:
      - "6000:6000"
      - "6010:6010"
    environment:
      - GOMAXPROCS=2
      - AUTH_DB_USER=postgres
      - AUTH_DB_PASSWORD=postgres
      - AUTH_DB_HOST=database
      - AUTH_DB_DISABLE_TLS=true
      - AUTH_TEMPO_HOST=tempo:4317
      - AUTH_TEMPO_SERVICE_NAME=auth
      - AUTH_TEMPO_PROBABILITY=1.0
    healthcheck:
      test: ["CMD-SHELL", "wget -qO- http://localhost:6000/v1/liveness || exit 1"]
      interval: 5s
      timeout: 5s
      retries: 2
      start_period: 5s
    networks:
      sales-system-network:
        ipv4_address: 10.5.0.5
    depends_on:
      database:
        condition: service_healthy
      tempo:
        condition: service_healthy

  sales:
    image: localhost/ardanlabs/sales:0.0.1
    pull_policy: never
    container_name: sales
    restart: unless-stopped
    ports:
      - "3000:3000"
      - "3010:3010"
    environment:
      - GOMAXPROCS=${GOMAXPROCS:-0}
      - GOGC=off
      - GOMEMLIMIT=${GOMEMLIMIT:-0}
      - SALES_DB_USER=postgres
      - SALES_DB_PASSWORD=postgres
      - SALES_DB_HOST=database
      - SALES_DB_DISABLE_TLS=true
      - SALES_AUTH_HOST=http://auth:6000
      - SALES_TEMPO_HOST=tempo:4317
      - SALES_TEMPO_SERVICE_NAME=sales
      - SALES_TEMPO_PROBABILITY=1.0
    healthcheck:
      test: ["CMD-SHELL", "wget -qO- http://localhost:3000/v1/liveness || exit 1"]
      interval: 5s
      timeout: 5s
      retries: 2
      start_period: 5s
    networks:
      sales-system-network:
        ipv4_address: 10.5.0.15
    depends_on:
      init-migrate-seed:
        condition: service_completed_successfully
      tempo:
        condition: service_healthy

  metrics:
    image: localhost/ardanlabs/metrics:0.0.1
    pull_policy: never
    container_name: metrics
    restart: unless-stopped
    ports:
      - "4000:4000"
      - "4010:4010"
      - "4020:4020"
    environment:
      - GOMAXPROCS=1
      - METRICS_COLLECT_FROM=http://sales:3010/debug/vars
    networks:
      sales-system-network:
        ipv4_address: 10.5.0.20
    depends_on:
      sales:
        condition: service_healthy

# ==============================================================================
# Volumes
# ==============================================================================

volumes:
  grafana-data:
    driver: local

# ==============================================================================
# Networks
# ==============================================================================

networks:
  sales-system-network:
    driver: bridge
    ipam:
      config:
        - subnet: 10.5.0.0/24
```

### Step 4: Alternative - Update Existing docker_compose.yaml

If you prefer to update your existing `docker_compose.yaml`, add these sections:

**Add to services section:**

```yaml
  # Add Tempo service
  tempo:
    image: grafana/tempo:2.8.1
    container_name: tempo
    restart: unless-stopped
    command: ["-config.file=/etc/tempo.yaml"]
    volumes:
      - ./tempo-data:/tmp/tempo
      - ./tempo-config.yaml:/etc/tempo.yaml:ro
    ports:
      - "3200:3200"
      - "4317:4317"
      - "4318:4318"
    networks:
      sales-system-network:
        ipv4_address: 10.5.0.30
    healthcheck:
      test: ["CMD-SHELL", "wget -qO- http://localhost:3200/ready || exit 1"]
      interval: 10s
      timeout: 5s
      retries: 5

  # Add Grafana service
  grafana:
    image: grafana/grafana:12.2.0
    container_name: grafana
    restart: unless-stopped
    environment:
      - GF_AUTH_ANONYMOUS_ENABLED=true
      - GF_AUTH_ANONYMOUS_ORG_ROLE=Admin
      - GF_AUTH_DISABLE_LOGIN_FORM=true
      - GF_FEATURE_TOGGLES_ENABLE=traceqlEditor
      - GF_SERVER_HTTP_PORT=3100
    volumes:
      - ./grafana-datasources.yaml:/etc/grafana/provisioning/datasources/datasources.yaml:ro
      - grafana-data:/var/lib/grafana
    ports:
      - "3100:3100"
    networks:
      sales-system-network:
        ipv4_address: 10.5.0.31
    depends_on:
      - tempo
    healthcheck:
      test: ["CMD-SHELL", "wget -qO- http://localhost:3100/api/health || exit 1"]
      interval: 10s
      timeout: 5s
      retries: 5
```

**Update sales service environment:**

```yaml
  sales:
    environment:
      # ... existing vars ...
      - SALES_TEMPO_HOST=tempo:4317
      - SALES_TEMPO_SERVICE_NAME=sales
      - SALES_TEMPO_PROBABILITY=1.0  # 100% sampling for dev
    depends_on:
      - tempo  # Add this
```

**Update auth service environment:**

```yaml
  auth:
    environment:
      # ... existing vars ...
      - AUTH_TEMPO_HOST=tempo:4317
      - AUTH_TEMPO_SERVICE_NAME=auth
      - AUTH_TEMPO_PROBABILITY=1.0
    depends_on:
      - tempo  # Add this
```

**Add to volumes section:**

```yaml
volumes:
  # ... existing volumes ...
  grafana-data: {}
```

### Step 5: Start the Stack

```bash
# Navigate to compose directory
cd zarf/compose

# Start all services
docker compose -f docker_compose_with_observability.yaml up -d

# Or if you updated existing file:
docker compose up -d

# Check all services are running
docker compose ps

# View logs
docker compose logs -f
```

### Step 6: Verify Everything is Running

```bash
# Check Tempo health
curl http://localhost:3200/ready

# Check Grafana health
curl http://localhost:3100/api/health

# Check Sales API
curl http://localhost:3000/v1/liveness

# View service logs
docker compose logs sales
docker compose logs tempo
docker compose logs grafana
```

---

## Viewing Traces and Logs

### Access Grafana

1. **Open browser**: http://localhost:3100
2. **Login**: Automatic (anonymous admin enabled)
3. **Navigate to Explore** (compass icon on left sidebar)
4. **Select Tempo** datasource

### Generate Test Traffic

```bash
# First, get authentication token
export TOKEN=$(curl -s -X POST http://localhost:6000/v1/auth \
  -H "Content-Type: application/json" \
  -d '{
    "email": "admin@example.com",
    "password": "gophers"
  }' | jq -r '.token')

# Generate traffic to create traces
for i in {1..20}; do
  curl -H "Authorization: Bearer $TOKEN" http://localhost:3000/v1/users
  sleep 1
done
```

### Query Traces in Grafana

#### By Service Name
```
{ service.name = "sales" }
```

#### By Status (Errors)
```
{ service.name = "sales" && status = error }
```

#### By HTTP Method
```
{ service.name = "sales" && span.http.method = "POST" }
```

#### By Duration (Slow Requests)
```
{ service.name = "sales" && duration > 100ms }
```

#### By Span Name
```
{ service.name = "sales" && name = "business.userbus.create" }
```

#### Complex Query
```
{ service.name = "sales" && span.http.method = "GET" && duration > 50ms }
```

### Understanding Trace Output

**Trace View Components:**
- **Trace ID**: Unique identifier for the entire request
- **Spans**: Individual operations (HTTP handler, business logic, DB queries)
- **Duration**: Time each operation took
- **Attributes**: Metadata (HTTP method, status code, user ID, etc.)
- **Trace Timeline**: Visual representation of span hierarchy

**Example Trace Hierarchy:**
```
sales [200ms total]
├── HTTP GET /v1/users [200ms]
│   ├── business.userbus.query [150ms]
│   │   └── postgres.query [140ms]
│   └── response.encode [10ms]
```

### Correlate Logs with Traces

Since logs include trace IDs:

1. Find trace ID in Grafana (e.g., `abc123def456`)
2. Search application logs:
   ```bash
   docker compose logs sales | grep abc123def456
   ```

3. Or filter in real-time:
   ```bash
   docker compose logs -f sales | grep --line-buffered abc123def456
   ```

---

## Troubleshooting

### Services Won't Start

**Check Docker logs:**
```bash
docker compose logs

# Or specific service
docker compose logs tempo
docker compose logs grafana
```

**Check container status:**
```bash
docker compose ps
```

**Restart specific service:**
```bash
docker compose restart tempo
docker compose restart sales
```

### Tempo Not Receiving Traces

**Check Tempo health:**
```bash
curl http://localhost:3200/ready
curl http://localhost:3200/status
```

**Check application logs:**
```bash
docker compose logs sales | grep -i "tempo\|trace\|otel"
```

**Verify OTEL configuration in sales logs:**
```bash
# Should show: OTEL tracer tempo:4317
docker compose logs sales | grep "OTEL"
```

**Common issues:**
- Tempo host misconfigured (should be `tempo:4317` not `localhost:4317`)
- Sampling probability too low (set to 1.0 for development)
- Network issues between containers
- Tempo container not healthy

### No Traces Appearing in Grafana

**Verify Tempo datasource in Grafana:**
1. Open Grafana: http://localhost:3100
2. Go to Configuration → Data Sources
3. Click "Tempo"
4. Click "Save & Test"
5. Should show: "Data source is working"

**Check if traces exist in Tempo:**
```bash
# Query Tempo API directly
curl "http://localhost:3200/api/search?tags=service.name=sales&limit=10"
```

**Verify traffic was generated:**
```bash
# Check sales logs for requests
docker compose logs sales | grep "GET /v1/users"
```

### Grafana Connection Issues

**Check Grafana health:**
```bash
curl http://localhost:3100/api/health
```

**Check datasource configuration:**
```bash
docker compose exec grafana cat /etc/grafana/provisioning/datasources/datasources.yaml
```

**Restart Grafana:**
```bash
docker compose restart grafana
```

**Check Grafana logs:**
```bash
docker compose logs grafana
```

### Sampling Rate Too Low

If you're not seeing traces, verify sampling rate:

```yaml
# In docker_compose.yaml
  sales:
    environment:
      - SALES_TEMPO_PROBABILITY=1.0  # 100% of requests traced
```

Then restart:
```bash
docker compose restart sales
```

### Port Conflicts

If ports are already in use:

```bash
# Check what's using a port
lsof -i :3100  # Grafana
lsof -i :4317  # Tempo OTLP

# Change ports in docker_compose.yaml
ports:
  - "3101:3100"  # Map to different host port
```

### Database Connection Issues

```bash
# Check database is ready
docker compose logs database

# Check database health
docker compose exec database pg_isready

# Restart database
docker compose restart database
```

---

## Advanced Configuration

### Production Configuration

For production deployments, update these settings:

**1. Sampling Rate** (`docker_compose.yaml`)

```yaml
  sales:
    environment:
      - SALES_TEMPO_PROBABILITY=0.05  # 5% sampling
```

**2. Tempo Storage** (`tempo-config.yaml`)

Use S3/GCS instead of local disk:

```yaml
storage:
  trace:
    backend: s3
    s3:
      bucket: your-tempo-traces
      endpoint: s3.amazonaws.com
      access_key: ${S3_ACCESS_KEY}
      secret_key: ${S3_SECRET_KEY}
```

**3. Grafana Security** (`docker_compose.yaml`)

Disable anonymous access:

```yaml
  grafana:
    environment:
      - GF_AUTH_ANONYMOUS_ENABLED=false
      - GF_SECURITY_ADMIN_USER=admin
      - GF_SECURITY_ADMIN_PASSWORD=${GRAFANA_PASSWORD}
```

### Custom Span Attributes

Add custom attributes to spans for better debugging:

```go
import "go.opentelemetry.io/otel/attribute"

ctx, span := otel.AddSpan(ctx, "business.userbus.create",
    attribute.String("user.email", email),
    attribute.String("user.id", userID.String()),
    attribute.Bool("user.enabled", true),
)
defer span.End()
```

### Error Tracking

Mark spans as errors:

```go
import (
    "go.opentelemetry.io/otel/codes"
    "go.opentelemetry.io/otel/trace"
)

ctx, span := otel.AddSpan(ctx, "business.userbus.create")
defer span.End()

if err != nil {
    span.RecordError(err)
    span.SetStatus(codes.Error, err.Error())
    return err
}
```

### Performance Monitoring

Use TraceQL to identify slow operations:

```
// In Grafana Explore
{ service.name = "sales" && duration > 500ms }
```

### Distributed Tracing

When calling downstream services, traces automatically propagate:

```go
// In authclient.go
req, err := http.NewRequestWithContext(ctx, method, url, body)
otel.AddTraceToRequest(ctx, req)  // Injects trace context headers

// The downstream service (auth) will continue the same trace
```

This creates a single trace across multiple services.

### Resource Limits

Add resource limits for production:

```yaml
  tempo:
    deploy:
      resources:
        limits:
          cpus: '2.0'
          memory: 2G
        reservations:
          cpus: '1.0'
          memory: 1G

  grafana:
    deploy:
      resources:
        limits:
          cpus: '1.0'
          memory: 512M
        reservations:
          cpus: '0.5'
          memory: 256M
```

### Backup and Persistence

Ensure data persistence:

```yaml
volumes:
  grafana-data:
    driver: local
    driver_opts:
      type: none
      device: /path/to/grafana/data
      o: bind
```

Backup script:

```bash
#!/bin/bash
# backup-grafana.sh

docker compose exec grafana grafana-cli admin export-dashboard > grafana-backup.json
tar -czf grafana-data-$(date +%Y%m%d).tar.gz ./grafana-data
```

---

## Docker Commands Reference

### Starting and Stopping

```bash
# Start all services
docker compose up -d

# Start specific service
docker compose up -d grafana

# Stop all services
docker compose down

# Stop but keep volumes
docker compose stop

# Restart service
docker compose restart sales
```

### Viewing Logs

```bash
# All logs
docker compose logs

# Follow logs
docker compose logs -f

# Specific service
docker compose logs sales

# Last 100 lines
docker compose logs --tail=100 sales

# Filter logs
docker compose logs sales | grep ERROR
```

### Checking Status

```bash
# List all containers
docker compose ps

# Show resource usage
docker stats

# Inspect container
docker compose exec sales env
```

### Cleaning Up

```bash
# Stop and remove containers
docker compose down

# Remove containers and volumes
docker compose down -v

# Remove everything including images
docker compose down -v --rmi all

# Clean up Docker system
docker system prune -a
```

---

## Port Reference

| Service  | Port | Purpose                    | URL                          |
|----------|------|----------------------------|------------------------------|
| Sales    | 3000 | API                        | http://localhost:3000        |
| Sales    | 3010 | Debug/metrics              | http://localhost:3010/debug  |
| Grafana  | 3100 | Web UI                     | http://localhost:3100        |
| Tempo    | 3200 | Query API                  | http://localhost:3200        |
| Tempo    | 4317 | OTLP gRPC (traces in)      | tempo:4317                   |
| Tempo    | 4318 | OTLP HTTP (traces in)      | http://tempo:4318            |
| Auth     | 6000 | API                        | http://localhost:6000        |
| Auth     | 6010 | Debug/metrics              | http://localhost:6010/debug  |
| Database | 5432 | PostgreSQL                 | postgres://localhost:5432    |

---

## Summary

### What You Get:

1. **Complete observability stack** running in Docker
2. **Tempo** for distributed tracing
3. **Grafana** for visualization
4. **Automatic tracing** of all HTTP requests and business operations
5. **Correlation** between logs and traces via trace IDs

### Quick Start:

```bash
# 1. Create config files
cd zarf/compose

# 2. Copy configurations from this guide:
#    - tempo-config.yaml
#    - grafana-datasources.yaml
#    - docker_compose_with_observability.yaml

# 3. Start everything
docker compose -f docker_compose_with_observability.yaml up -d

# 4. Open Grafana
open http://localhost:3100

# 5. Generate traffic and view traces!
```

### Key Benefits:

- No Kubernetes required
- Simple Docker Compose setup
- Works on any machine with Docker
- Perfect for development and small deployments
- Easy to maintain and debug
- Low resource usage

### Next Steps:

1. Customize sampling rates for your needs
2. Add custom span attributes to your business logic
3. Create Grafana dashboards for your metrics
4. Set up alerts for errors and slow requests
5. Deploy to production with appropriate security settings

---

## Resources

- **OpenTelemetry Go SDK**: https://opentelemetry.io/docs/languages/go/
- **Grafana Tempo**: https://grafana.com/docs/tempo/latest/
- **TraceQL**: https://grafana.com/docs/tempo/latest/traceql/
- **OTLP Specification**: https://opentelemetry.io/docs/specs/otlp/
- **Docker Compose**: https://docs.docker.com/compose/
