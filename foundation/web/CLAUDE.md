# CLAUDE.md - Web Framework Package

This file provides guidance to Claude Code when working with the web framework package in this repository.

## Package Overview

The `foundation/web` package is a lightweight HTTP web framework built on top of Go's standard library. It provides a structured approach to building HTTP APIs with built-in support for OpenTelemetry tracing, middleware, CORS, and encoding/decoding abstractions.

## Package Purpose

- **HTTP Handler Abstraction**: Simplify HTTP handler signatures by using context-based handlers
- **OpenTelemetry Integration**: Automatic trace propagation and span creation for HTTP requests
- **Response Encoding**: Flexible response encoding with content-type negotiation
- **Request Decoding**: Structured request body decoding with validation support
- **Middleware Support**: Composable middleware pattern for cross-cutting concerns
- **CORS Handling**: Built-in CORS preflight and header management
- **Static File Serving**: Support for embedded file systems and React SPAs

## Architecture

### Core Files

1. **[web.go](web.go)**: Main App type, HTTP routing, and handler registration
2. **[response.go](response.go)**: Response encoding and status code handling
3. **[request.go](request.go)**: Request parameter extraction and body decoding
4. **[context.go](context.go)**: Context value management for tracers and writers
5. **[middleware.go](middleware.go)**: Middleware composition utilities

### Key Types

#### App Struct
```go
type App struct {
    log     Logger               // Logging function
    tracer  trace.Tracer        // OpenTelemetry tracer
    mux     *http.ServeMux      // Standard library router
    otmux   http.Handler        // OpenTelemetry-wrapped handler
    mw      []MidFunc           // Application-level middleware
    origins []string            // CORS allowed origins
}
```

#### Handler Function Signature
```go
type HandlerFunc func(ctx context.Context, r *http.Request) Encoder
```

This signature differs from standard `http.HandlerFunc`:
- Takes `context.Context` as first parameter (not embedded in request)
- Returns an `Encoder` instead of writing directly to `ResponseWriter`
- Simplifies testing and encourages clean separation of concerns

## Core Concepts

### Handler Types

The framework provides three handler registration methods:

1. **`HandlerFunc`**: Full framework support (middleware + tracing)
2. **`HandlerFuncNoMid`**: No middleware or tracing (lightweight)
3. **`RawHandlerFunc`**: Bridge to standard `http.HandlerFunc` with middleware support

### Encoder Interface

Handlers return an `Encoder` instead of writing directly to the response:

```go
type Encoder interface {
    Encode() (data []byte, contentType string, err error)
}
```

**Benefits:**
- Decouples response generation from HTTP writing
- Enables easier testing (test encoder output, not HTTP response)
- Supports automatic content-type handling
- Centralizes error handling

### Decoder Interface

Request bodies are decoded using the `Decoder` interface:

```go
type Decoder interface {
    Decode(data []byte) error
}
```

If the decoder also implements `Validate()`, validation runs automatically after decoding.

## Key Functions

### Application Setup

#### `NewApp(log Logger, tracer trace.Tracer, mw ...MidFunc) *App`
Creates a new application instance with routing and middleware.

**Usage:**
```go
app := web.NewApp(
    log.Info,
    tracer,
    middleware.Logger(log),
    middleware.Errors(log),
    middleware.Metrics(),
    middleware.Panics(),
)
```

**Behavior:**
- Initializes `http.ServeMux` for routing
- Wraps mux with OpenTelemetry handler for automatic span creation
- Stores application-level middleware (applied to all routes)
- Configures W3C TraceContext propagation

### Handler Registration

#### `HandlerFunc(method, group, path string, handlerFunc HandlerFunc, mw ...MidFunc)`
Registers a handler with full framework support.

**Usage:**
```go
app.HandlerFunc("GET", "v1", "/users/{id}", getUserHandler)
app.HandlerFunc("POST", "v1", "/users", createUserHandler, authMiddleware)
```

**Parameters:**
- `method`: HTTP method (GET, POST, PUT, PATCH, DELETE)
- `group`: Route group prefix (e.g., "v1", "admin") - empty string for no group
- `path`: Route path with Go 1.22+ pattern support (e.g., "/users/{id}")
- `handlerFunc`: Your handler function
- `mw`: Route-specific middleware (optional)

**Behavior:**
- Applies route-specific middleware, then application middleware
- Injects tracer into context
- Stores `ResponseWriter` in context for access
- Injects OpenTelemetry headers into response
- Calls handler and encodes response
- Final path: `{method} /{group}{path}` (e.g., `GET /v1/users/{id}`)

#### `HandlerFuncNoMid(method, group, path string, handlerFunc HandlerFunc)`
Registers a handler without middleware or tracing overhead.

**Use Cases:**
- Health check endpoints that need minimal latency
- High-frequency endpoints where tracing overhead matters
- Debug endpoints

**Behavior:**
- Bypasses all middleware
- No OpenTelemetry tracing
- Direct handler execution
- Still uses Encoder pattern

#### `RawHandlerFunc(method, group, path string, rawHandlerFunc http.HandlerFunc, mw ...MidFunc)`
Bridges standard `http.HandlerFunc` to the framework with middleware support.

**Usage:**
```go
// Wrap third-party handler
app.RawHandlerFunc("GET", "", "/metrics", promhttp.Handler().ServeHTTP)
```

**Use Cases:**
- Integrating third-party handlers (Prometheus, pprof)
- Migrating existing handlers incrementally
- Handlers that need direct `ResponseWriter` access

### CORS Configuration

#### `EnableCORS(origins []string)`
Enables CORS with specified allowed origins.

**Usage:**
```go
app.EnableCORS([]string{"https://app.example.com", "http://localhost:3000"})
// Allow all origins (not recommended for production)
app.EnableCORS([]string{"*"})
```

**Behavior:**
- Sets `Access-Control-Allow-Origin` header based on request's `Origin` header
- Only sets allowed origin if it matches one in the list
- Automatically adds CORS headers: `Access-Control-Allow-Methods`, `Access-Control-Allow-Headers`, `Access-Control-Max-Age`
- Enables preflight OPTIONS requests

**Headers Set:**
```
Access-Control-Allow-Methods: POST, PATCH, GET, OPTIONS, PUT, DELETE
Access-Control-Allow-Headers: Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization
Access-Control-Max-Age: 86400 (24 hours)
```

### Static File Serving

#### `FileServer(static embed.FS, dir, path string) error`
Serves static files from an embedded filesystem.

**Usage:**
```go
//go:embed static/*
var staticFiles embed.FS

app.FileServer(staticFiles, "static", "/static/")
```

**Use Cases:**
- Serving CSS, JavaScript, images
- Documentation files
- API schemas

#### `FileServerReact(static embed.FS, dir, path string) error`
Serves a React SPA with client-side routing support.

**Usage:**
```go
//go:embed build/*
var reactApp embed.FS

app.FileServerReact(reactApp, "build", "/app/")
```

**Behavior:**
- Serves actual files for requests with file extensions (`.js`, `.css`, `.png`)
- Returns `index.html` for all other routes (enables client-side routing)
- Properly sets `Content-Type: text/html` for index.html

**Difference from `FileServer`:**
- `FileServer`: Traditional file server (404 for missing files)
- `FileServerReact`: SPA-aware (index.html fallback for routes)

### Request Handling

#### `Param(r *http.Request, key string) string`
Extracts path parameters from the request.

**Usage:**
```go
// Route: GET /v1/users/{id}
func getUserHandler(ctx context.Context, r *http.Request) web.Encoder {
    userID := web.Param(r, "id")
    // ...
}
```

**Note:** Uses Go 1.22+ `PathValue` method.

#### `Decode(r *http.Request, v Decoder) error`
Decodes and validates request body.

**Usage:**
```go
type CreateUserRequest struct {
    Name  string `json:"name"`
    Email string `json:"email"`
}

func (c *CreateUserRequest) Decode(data []byte) error {
    return json.Unmarshal(data, c)
}

func (c *CreateUserRequest) Validate() error {
    if c.Name == "" {
        return errors.New("name is required")
    }
    return nil
}

func createUserHandler(ctx context.Context, r *http.Request) web.Encoder {
    var req CreateUserRequest
    if err := web.Decode(r, &req); err != nil {
        return ErrorResponse(err)
    }
    // ...
}
```

**Behavior:**
- Reads entire request body
- Calls `Decode` method
- If type implements `Validate()`, runs validation
- Returns descriptive errors

### Response Handling

#### `Respond(ctx context.Context, w http.ResponseWriter, resp Encoder) error`
Encodes and sends the response to the client.

**Usage:**
Typically called by the framework, but can be used manually:

```go
if err := web.Respond(ctx, w, response); err != nil {
    log.Error(ctx, "response failed", "error", err)
}
```

**Behavior:**
1. Checks for `NoResponse` type (skip sending)
2. Checks if context is canceled (client disconnected)
3. Determines HTTP status code:
   - From `HTTPStatus()` interface if implemented
   - `500` for error types
   - `204` for nil responses
   - `200` otherwise
4. Creates span: `web.send.response`
5. Encodes response data
6. Sets `Content-Type` header
7. Writes status code and body

#### `NoResponse` Type
Special response type that prevents automatic response sending.

**Usage:**
```go
func streamHandler(ctx context.Context, r *http.Request) web.Encoder {
    w := web.GetWriter(ctx)

    // Manually stream response
    w.Header().Set("Content-Type", "text/event-stream")
    w.WriteHeader(http.StatusOK)

    flusher := w.(http.Flusher)
    for event := range events {
        fmt.Fprintf(w, "data: %s\n\n", event)
        flusher.Flush()
    }

    return web.NewNoResponse()
}
```

**Use Cases:**
- Streaming responses (SSE, chunked encoding)
- WebSocket upgrades
- Custom response handling

#### `GetWriter(ctx context.Context) http.ResponseWriter`
Retrieves the response writer from context.

**Usage:**
```go
w := web.GetWriter(ctx)
w.Header().Set("X-Custom-Header", "value")
```

**Use Cases:**
- Setting custom headers before framework sends response
- Accessing writer in middleware
- Manual response writing (with `NoResponse`)

### HTTP Status Interface

Types can implement `HTTPStatus()` to control response status code:

```go
type httpStatus interface {
    HTTPStatus() int
}
```

**Usage:**
```go
type NotFoundError struct {
    Message string
}

func (e NotFoundError) Error() string {
    return e.Message
}

func (e NotFoundError) HTTPStatus() int {
    return http.StatusNotFound
}

func (e NotFoundError) Encode() ([]byte, string, error) {
    data, _ := json.Marshal(map[string]string{"error": e.Message})
    return data, "application/json", nil
}

// In handler
return NotFoundError{Message: "user not found"}
```

## Middleware System

### Middleware Function Type

```go
type MidFunc func(handler HandlerFunc) HandlerFunc
```

Middleware wraps handlers to add cross-cutting functionality.

### Middleware Composition

```go
func wrapMiddleware(mw []MidFunc, handler HandlerFunc) HandlerFunc
```

**Execution Order:**
- Application middleware runs first (outer layer)
- Route-specific middleware runs second (inner layer)
- Handler runs last (core)

**Example Flow:**
```go
app := web.NewApp(log, tracer, mw1, mw2)
app.HandlerFunc("GET", "v1", "/users", handler, mw3)

// Execution order: mw1 -> mw2 -> mw3 -> handler
```

### Example Middleware

#### Logger Middleware
```go
func Logger(log *logger.Logger) web.MidFunc {
    return func(handler web.HandlerFunc) web.HandlerFunc {
        return func(ctx context.Context, r *http.Request) web.Encoder {
            start := time.Now()

            log.Info(ctx, "request started",
                "method", r.Method,
                "path", r.URL.Path,
            )

            resp := handler(ctx, r)

            log.Info(ctx, "request completed",
                "method", r.Method,
                "path", r.URL.Path,
                "duration", time.Since(start),
            )

            return resp
        }
    }
}
```

#### Error Handler Middleware
```go
func Errors(log *logger.Logger) web.MidFunc {
    return func(handler web.HandlerFunc) web.HandlerFunc {
        return func(ctx context.Context, r *http.Request) web.Encoder {
            resp := handler(ctx, r)

            if err, ok := resp.(error); ok {
                log.Error(ctx, "handler error", "error", err)
            }

            return resp
        }
    }
}
```

#### Authentication Middleware
```go
func Authenticate(validator TokenValidator) web.MidFunc {
    return func(handler web.HandlerFunc) web.HandlerFunc {
        return func(ctx context.Context, r *http.Request) web.Encoder {
            token := r.Header.Get("Authorization")
            if token == "" {
                return UnauthorizedError{Message: "missing token"}
            }

            userID, err := validator.Validate(token)
            if err != nil {
                return UnauthorizedError{Message: "invalid token"}
            }

            ctx = context.WithValue(ctx, userKey, userID)
            return handler(ctx, r)
        }
    }
}
```

#### Panic Recovery Middleware
```go
func Panics(log *logger.Logger) web.MidFunc {
    return func(handler web.HandlerFunc) web.HandlerFunc {
        return func(ctx context.Context, r *http.Request) (resp web.Encoder) {
            defer func() {
                if rec := recover(); rec != nil {
                    log.Error(ctx, "panic recovered",
                        "panic", rec,
                        "stack", string(debug.Stack()),
                    )
                    resp = InternalServerError{Message: "internal server error"}
                }
            }()

            return handler(ctx, r)
        }
    }
}
```

## OpenTelemetry Integration

### Automatic Tracing

The framework automatically:
1. Creates root span for each request (via `otelhttp.NewHandler`)
2. Injects tracer into context
3. Propagates trace context via W3C headers
4. Creates span for response encoding

### Adding Custom Spans

```go
func handler(ctx context.Context, r *http.Request) web.Encoder {
    // Framework provides tracer in context
    ctx, span := addSpan(ctx, "custom.operation")
    defer span.End()

    // Your logic here...

    return response
}
```

**Note:** The `addSpan` function is internal. Use the tracer from context directly:

```go
import "go.opentelemetry.io/otel/trace"

func handler(ctx context.Context, r *http.Request) web.Encoder {
    tracer := trace.SpanFromContext(ctx).TracerProvider().Tracer("service-name")
    ctx, span := tracer.Start(ctx, "custom.operation")
    defer span.End()

    return response
}
```

### Trace Context Storage

The framework stores the tracer in context using typed keys:

```go
const (
    tracerKey ctxKey = iota + 1
    writerKey
)
```

## Security Headers

### HSTS (HTTP Strict Transport Security)

Automatically set on all responses:
```
Strict-Transport-Security: max-age=63072000; includeSubDomains; preload
```

**Configuration:**
- Max age: 2 years (63072000 seconds)
- Includes subdomains
- Preload-ready for browser HSTS lists

**Important:** Only use this if you serve your application exclusively over HTTPS.

## Common Patterns

### RESTful API Structure

```go
func setupRoutes(app *web.App) {
    // User routes
    app.HandlerFunc("GET",    "v1", "/users",     listUsers)
    app.HandlerFunc("GET",    "v1", "/users/{id}", getUser)
    app.HandlerFunc("POST",   "v1", "/users",     createUser, validateAdmin)
    app.HandlerFunc("PATCH",  "v1", "/users/{id}", updateUser, validateOwner)
    app.HandlerFunc("DELETE", "v1", "/users/{id}", deleteUser, validateAdmin)

    // Health checks (no middleware)
    app.HandlerFuncNoMid("GET", "", "/health", healthCheck)

    // Static files
    app.FileServer(staticFiles, "static", "/static/")
    app.FileServerReact(webapp, "build", "/app/")
}
```

### Request/Response Types

```go
// Request type with decoding and validation
type CreateUserRequest struct {
    Name  string `json:"name"`
    Email string `json:"email"`
}

func (r *CreateUserRequest) Decode(data []byte) error {
    return json.Unmarshal(data, r)
}

func (r *CreateUserRequest) Validate() error {
    if r.Name == "" {
        return errors.New("name required")
    }
    if !strings.Contains(r.Email, "@") {
        return errors.New("invalid email")
    }
    return nil
}

// Response type with encoding and status
type UserResponse struct {
    ID    string `json:"id"`
    Name  string `json:"name"`
    Email string `json:"email"`
}

func (r UserResponse) Encode() ([]byte, string, error) {
    data, err := json.Marshal(r)
    return data, "application/json", err
}

func (r UserResponse) HTTPStatus() int {
    return http.StatusCreated
}
```

### Error Responses

```go
type ErrorResponse struct {
    Error      string `json:"error"`
    StatusCode int    `json:"-"`
}

func (e ErrorResponse) Error() string {
    return e.Error
}

func (e ErrorResponse) HTTPStatus() int {
    return e.StatusCode
}

func (e ErrorResponse) Encode() ([]byte, string, error) {
    data, err := json.Marshal(e)
    return data, "application/json", err
}

func NewBadRequestError(msg string) ErrorResponse {
    return ErrorResponse{Error: msg, StatusCode: http.StatusBadRequest}
}

func NewNotFoundError(msg string) ErrorResponse {
    return ErrorResponse{Error: msg, StatusCode: http.StatusNotFound}
}
```

### Complete Handler Example

```go
func createUserHandler(ctx context.Context, r *http.Request) web.Encoder {
    // Decode and validate request
    var req CreateUserRequest
    if err := web.Decode(r, &req); err != nil {
        return NewBadRequestError(err.Error())
    }

    // Business logic with tracing
    user, err := userService.Create(ctx, req.Name, req.Email)
    if err != nil {
        return ErrorResponse{
            Error:      "failed to create user",
            StatusCode: http.StatusInternalServerError,
        }
    }

    // Return response
    return UserResponse{
        ID:    user.ID,
        Name:  user.Name,
        Email: user.Email,
    }
}
```

## Testing

### Handler Testing

```go
func TestCreateUser(t *testing.T) {
    // Setup
    ctx := context.Background()
    body := strings.NewReader(`{"name":"Alice","email":"alice@example.com"}`)
    req := httptest.NewRequest("POST", "/v1/users", body)

    // Call handler directly
    resp := createUserHandler(ctx, req)

    // Test response
    data, contentType, err := resp.Encode()
    assert.NoError(t, err)
    assert.Equal(t, "application/json", contentType)

    var user UserResponse
    json.Unmarshal(data, &user)
    assert.Equal(t, "Alice", user.Name)
}
```

### Integration Testing

```go
func TestAPI(t *testing.T) {
    // Create app
    app := web.NewApp(testLogger, testTracer)
    setupRoutes(app)

    // Create test server
    server := httptest.NewServer(app)
    defer server.Close()

    // Make request
    resp, err := http.Get(server.URL + "/v1/users")
    assert.NoError(t, err)
    assert.Equal(t, http.StatusOK, resp.StatusCode)
}
```

## Best Practices

### Handler Design

1. **Keep handlers thin**: Move business logic to service layer
2. **Use typed requests/responses**: Create dedicated types for clarity
3. **Return errors as Encoders**: Don't panic, return error types
4. **Validate input**: Implement `Validate()` on request types
5. **Use context for request-scoped data**: Pass user ID, trace ID, etc. via context

### Middleware Guidelines

1. **Order matters**: Apply middleware in correct order (auth before authorization)
2. **Keep middleware focused**: Each middleware should do one thing
3. **Return early**: Return error responses in middleware when needed
4. **Preserve context**: Pass context through middleware chain
5. **Document side effects**: Clearly document what each middleware does

### Response Patterns

1. **Consistent error format**: Use standard error response structure
2. **Appropriate status codes**: Use correct HTTP status for each scenario
3. **Content-Type accuracy**: Set correct content type in Encode()
4. **Avoid nil returns**: Return explicit `NoResponse` or error types

### Performance Considerations

1. **Use `HandlerFuncNoMid` for hot paths**: Skip middleware overhead when safe
2. **Pool allocations**: Use sync.Pool for frequently allocated objects
3. **Stream large responses**: Use `NoResponse` with manual streaming
4. **Limit middleware**: Only apply necessary middleware per route

## Dependencies

- `net/http`: Standard library HTTP server
- `go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp`: OpenTelemetry HTTP instrumentation
- `go.opentelemetry.io/otel`: Core OpenTelemetry API
- `embed`: For static file embedding
- `regexp`: For file pattern matching in React SPA server

## Integration with Project

This framework integrates with:
- **[foundation/otel](../otel/CLAUDE.md)**: Distributed tracing
- **[foundation/logger](../logger/)**: Structured logging
- **[api/services/partner/mux](../../api/services/partner/mux/)**: Route definitions

### Typical Service Setup

```go
// In main.go
func main() {
    // Initialize tracing
    tracer := traceProvider.Tracer("partner-service")

    // Create app
    app := web.NewApp(log.Info, tracer,
        middleware.Logger(log),
        middleware.Errors(log),
        middleware.Panics(),
    )

    // Setup routes
    mux.RegisterRoutes(app)

    // Enable CORS
    app.EnableCORS(cfg.Web.CORSAllowedOrigins)

    // Start server
    http.ListenAndServe(":3000", app)
}
```

## Troubleshooting

### Handler Not Called

1. **Check route pattern**: Ensure method and path match exactly
2. **Verify group prefix**: Group adds `/{group}` prefix to path
3. **CORS preflight**: OPTIONS requests need CORS enabled

### Response Not Sent

1. **Check Encoder implementation**: Ensure `Encode()` returns valid data
2. **Context cancellation**: Client may have disconnected
3. **NoResponse return**: Handler may be returning `NoResponse`

### Middleware Not Executing

1. **Check registration order**: App-level middleware before route-specific
2. **Verify middleware wrapping**: Ensure middleware calls wrapped handler
3. **Early returns**: Middleware may be returning early

### Tracing Not Working

1. **Tracer initialization**: Verify tracer passed to `NewApp`
2. **Context flow**: Ensure context is passed through all calls
3. **Sampling**: Check if route is excluded from tracing

## Related Documentation

- Go 1.22 HTTP Routing: https://go.dev/blog/routing-enhancements
- OpenTelemetry Go: https://opentelemetry.io/docs/languages/go/
- CORS: https://developer.mozilla.org/en-US/docs/Web/HTTP/CORS
- HSTS: https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Strict-Transport-Security
