# CLAUDE.md - OpenTelemetry Package

This file provides guidance to Claude Code when working with the OpenTelemetry (otel) package in this repository.

## Package Overview

The `foundation/otel` package provides distributed tracing infrastructure using OpenTelemetry. It integrates with the Topifier application to enable observability through distributed traces, allowing monitoring and debugging of requests across services.

## Package Purpose

- **Distributed Tracing**: Track requests across service boundaries using OpenTelemetry standards
- **Trace Management**: Generate, propagate, and manage trace IDs throughout the application lifecycle
- **Custom Sampling**: Filter traces based on endpoint exclusion rules and probability-based sampling
- **Context Integration**: Store and retrieve tracing information in Go context objects

## Architecture

### Core Files

1. **[otel.go](otel.go)**: Main initialization and tracing setup
2. **[context.go](context.go)**: Context value management for tracers and trace IDs
3. **[sampler.go](sampler.go)**: Custom sampling logic to exclude endpoints and control sampling rates

### Key Components

#### Configuration (`Config` struct)
```go
type Config struct {
    ServiceName    string                 // Name of the service for trace identification
    Host           string                 // OTLP endpoint (empty = NOOP tracer)
    ExcludedRoutes map[string]struct{}    // Routes to exclude from tracing
    Probability    float64                // Sampling probability (0.0-1.0)
}
```

#### Tracer Modes
- **NOOP Mode**: When `Host` is empty, uses a no-op tracer (no traces exported)
- **Production Mode**: When `Host` is configured, exports traces via OTLP/gRPC

## Key Functions

### Initialization

#### `InitTracing(log *logger.Logger, cfg Config) (trace.TracerProvider, func(context.Context), error)`
Initializes OpenTelemetry tracing with the service.

**Usage:**
```go
traceProvider, teardown, err := otel.InitTracing(log, otel.Config{
    ServiceName:    "partner-service",
    Host:           "localhost:4317",
    ExcludedRoutes: map[string]struct{}{"/debug/readiness": {}},
    Probability:    0.5, // Sample 50% of traces
})
defer teardown(ctx)
```

**Behavior:**
- Sets up OTLP gRPC exporter
- Configures custom sampler with endpoint exclusion
- Sets global tracer provider and text map propagator
- Returns teardown function for graceful shutdown

**WARNING**: Current implementation uses insecure gRPC connection (`WithInsecure()`) - should be configurable for production

### Request Context Setup

#### `InjectTracing(ctx context.Context, tracer trace.Tracer) context.Context`
Initializes tracing for a request by storing the tracer and generating/extracting a trace ID.

**Usage:**
```go
ctx = otel.InjectTracing(r.Context(), tracer)
```

**Behavior:**
- Stores tracer in context for later span creation
- Extracts trace ID from existing span context, or generates new UUID if not present
- Stores trace ID in context for logging and correlation

### Span Management

#### `AddSpan(ctx context.Context, spanName string, keyValues ...attribute.KeyValue) (context.Context, trace.Span)`
Creates a new span within the current trace.

**Usage:**
```go
ctx, span := otel.AddSpan(ctx, "database.query",
    attribute.String("query", "SELECT * FROM users"),
    attribute.Int("limit", 100),
)
defer span.End()
```

**Behavior:**
- Retrieves tracer from context
- Creates child span with provided name
- Adds custom attributes to span
- Returns new context with span and the span itself

### Trace Propagation

#### `AddTraceToRequest(ctx context.Context, r *http.Request)`
Injects trace context into outgoing HTTP request headers for distributed tracing.

**Usage:**
```go
req, _ := http.NewRequest("GET", "http://api.example.com/users", nil)
otel.AddTraceToRequest(ctx, req)
client.Do(req)
```

**Behavior:**
- Uses W3C TraceContext propagation format
- Adds `traceparent` and `tracestate` headers to request

### Trace ID Retrieval

#### `GetTraceID(ctx context.Context) string`
Extracts the trace ID from context for logging or correlation.

**Usage:**
```go
traceID := otel.GetTraceID(ctx)
log.Info(ctx, "processing request", "trace_id", traceID)
```

**Returns:**
- Trace ID string if present
- Default trace ID (`00000000000000000000000000000000`) if not found

## Custom Sampling

The package implements a custom sampler (`endpointExcluder`) that:

1. **Endpoint Exclusion**: Drops traces for specified endpoints (e.g., health checks)
2. **Probability-Based Sampling**: Samples remaining traces based on configured probability

### Sampler Logic

```go
func (ee endpointExcluder) ShouldSample(parameters trace.SamplingParameters) trace.SamplingResult {
    // Extract endpoint from span attributes
    if ep := endpoint(parameters); ep != "" {
        // Drop if endpoint is in exclusion list
        if _, exists := ee.endpoints[ep]; exists {
            return trace.SamplingResult{Decision: trace.Drop}
        }
    }

    // Apply probability-based sampling for remaining traces
    return trace.TraceIDRatioBased(ee.probability).ShouldSample(parameters)
}
```

### Endpoint Detection

The sampler extracts endpoints from span attributes:
- `url.path`: Request path (e.g., `/api/users`)
- `url.query`: Query string (e.g., `limit=10`)

## Context Keys

The package uses typed context keys to prevent collisions:

```go
type ctxKey int

const (
    tracerKey  ctxKey = iota + 1  // Stores trace.Tracer
    traceIDKey                     // Stores trace ID string
)
```

## Integration with Application

### Typical Initialization Flow

1. **Service Startup**: Initialize tracing in `main.go`
```go
traceProvider, teardown, err := otel.InitTracing(log, cfg)
defer teardown(ctx)
```

2. **Request Handling**: Inject tracing into request context
```go
ctx = otel.InjectTracing(r.Context(), tracer)
```

3. **Operation Tracking**: Add spans for key operations
```go
ctx, span := otel.AddSpan(ctx, "operation.name")
defer span.End()
```

4. **Service Calls**: Propagate traces to downstream services
```go
otel.AddTraceToRequest(ctx, outgoingRequest)
```

## Dependencies

- `go.opentelemetry.io/otel`: Core OpenTelemetry API
- `go.opentelemetry.io/otel/sdk/trace`: OpenTelemetry SDK for trace implementation
- `go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc`: OTLP gRPC exporter
- `github.com/google/uuid`: UUID generation for trace IDs
- `github.com/ardanlabs/service/foundation/logger`: Project's logging infrastructure

## Common Patterns

### Adding Tracing to New Endpoints

```go
func Handler(ctx context.Context, r *http.Request) web.Encoder {
    // Add span for the handler
    ctx, span := otel.AddSpan(ctx, "handler.name")
    defer span.End()

    // Add custom attributes
    span.SetAttributes(
        attribute.String("user_id", userID),
        attribute.Int("count", itemCount),
    )

    // Business logic here...

    return response
}
```

### Tracing External API Calls

```go
func callExternalAPI(ctx context.Context, url string) error {
    ctx, span := otel.AddSpan(ctx, "external.api.call",
        attribute.String("url", url),
    )
    defer span.End()

    req, _ := http.NewRequest("GET", url, nil)
    otel.AddTraceToRequest(ctx, req)

    resp, err := client.Do(req)
    if err != nil {
        span.RecordError(err)
        return err
    }

    span.SetAttributes(attribute.Int("status_code", resp.StatusCode))
    return nil
}
```

### Excluding Health Check Endpoints

```go
excludedRoutes := map[string]struct{}{
    "/debug/readiness":  {},
    "/debug/liveness":   {},
    "/health":           {},
}

cfg := otel.Config{
    ServiceName:    "partner-service",
    Host:           otlpHost,
    ExcludedRoutes: excludedRoutes,
    Probability:    1.0, // Sample all non-excluded traces
}
```

## Configuration Best Practices

1. **Sampling Rate**:
   - Development: Set to `1.0` (sample everything)
   - Production: Start with `0.1` (10%) and adjust based on volume

2. **Excluded Routes**:
   - Always exclude health checks (`/debug/readiness`, `/health`)
   - Consider excluding high-frequency, low-value endpoints

3. **Service Name**:
   - Use consistent naming: `{service}-service` pattern
   - Matches the application's primary identifier

4. **OTLP Endpoint**:
   - Development: Use local collector (e.g., `localhost:4317`)
   - Production: Use environment variable for flexibility
   - Empty string disables tracing (NOOP mode)

## Observability Setup

### Local Development with Jaeger

```bash
# Run Jaeger all-in-one with OTLP support
docker run -d --name jaeger \
  -e COLLECTOR_OTLP_ENABLED=true \
  -p 4317:4317 \
  -p 16686:16686 \
  jaegertracing/all-in-one:latest

# Configure service to use Jaeger
export PARTNER_OTEL_HOST=localhost:4317
```

### Production Considerations

- **Security**: Replace `WithInsecure()` with proper TLS configuration
- **Batching**: Current settings use defaults - tune for your workload
- **Resource Attributes**: Service name is set - consider adding version, environment
- **Error Handling**: Exporter creation errors are handled, but consider fallback strategies

## Troubleshooting

### No Traces Appearing

1. **Check OTLP Endpoint**: Verify `Host` is reachable
2. **Sampling Rate**: Ensure probability > 0 and endpoint not excluded
3. **Propagation**: Verify `InjectTracing` is called for each request
4. **Teardown**: Ensure teardown function is called on shutdown to flush spans

### Missing Trace IDs in Logs

1. **Context Flow**: Ensure context with trace ID is passed to logger
2. **Injection**: Verify `InjectTracing` is called before logging
3. **Logger Integration**: Ensure logger is configured to extract trace ID from context

### High Cardinality Attributes

- Avoid adding user-generated content directly to spans
- Use consistent attribute keys (see OpenTelemetry semantic conventions)
- Limit attribute values to bounded sets where possible

## Future Enhancements

When extending this package, consider:

1. **TLS Support**: Make insecure mode configurable
2. **Multiple Exporters**: Support exporting to multiple backends
3. **Metrics Integration**: Add OpenTelemetry metrics alongside traces
4. **Baggage Propagation**: Utilize baggage for cross-cutting concerns
5. **Sampling Strategies**: Add more sophisticated sampling algorithms
6. **Resource Attributes**: Add deployment environment, version, instance ID

## Related Documentation

- OpenTelemetry Go: https://opentelemetry.io/docs/languages/go/
- W3C TraceContext: https://www.w3.org/TR/trace-context/
- OTLP Specification: https://github.com/open-telemetry/opentelemetry-specification/blob/main/specification/protocol/otlp.md
