# Backend Authentication Implementation Plan

**Status:** Ready for Implementation
**Estimated Time:** 4-6 hours
**Token Expiry:** 48 hours
**User Creation:** Via SQL only (no registration endpoint)

---

## Overview

This plan covers all backend changes needed to enable JWT-based authentication. The system uses:
- **JWT with RS256** signing (RSA keys)
- **OPA policies** for authorization
- **bcrypt** for password hashing
- **Basic Auth** for login (email:password)
- **Bearer Auth** for API requests (JWT token)

---

## Critical Bugs to Fix

### Bug 1: Migration SQL Syntax Error
**File:** `business/sdk/migrate/sql/migrate.sql`
**Line:** 29
**Issue:** Missing comma after PRIMARY KEY constraint

**Fix:**
```sql
PRIMARY KEY (think_id),  -- Add comma here
FOREIGN KEY (user_id) REFERENCES users(user_id) ON DELETE CASCADE
```

---

### Bug 2: Think Model Missing user_id Field
**File:** `business/domain/thinkbus/model.go`
**Lines:** 62-77

**Current:**
```go
type Think struct {
    ID          uuid.UUID
    Category    Category
    Content     content.Content
    DateCreated time.Time
    DateUpdated time.Time
}

type NewThink struct {
    Category Category
    Content  content.Content
}
```

**Fix:**
```go
type Think struct {
    ID          uuid.UUID
    UserID      uuid.UUID      // ADD THIS
    Category    Category
    Content     content.Content
    DateCreated time.Time
    DateUpdated time.Time
}

type NewThink struct {
    UserID   uuid.UUID      // ADD THIS
    Category Category
    Content  content.Content
}
```

---

### Bug 3: Think Database Model Missing user_id
**File:** `business/domain/thinkbus/stores/thinkdb/model.go`
**Lines:** 15-23

**Current:**
```go
type think struct {
    ID          uuid.UUID `db:"think_id"`
    Category    string    `db:"category"`
    Content     string    `db:"content"`
    DateCreated time.Time `db:"date_created"`
    DateUpdated time.Time `db:"date_updated"`
}
```

**Fix:**
```go
type think struct {
    ID          uuid.UUID `db:"think_id"`
    UserID      uuid.UUID `db:"user_id"`     // ADD THIS
    Category    string    `db:"category"`
    Content     string    `db:"content"`
    DateCreated time.Time `db:"date_created"`
    DateUpdated time.Time `db:"date_updated"`
}
```

---

### Bug 4: Think Database Queries Missing user_id
**File:** `business/domain/thinkbus/stores/thinkdb/thinkdb.go`

#### 4a. Create Method (Line 34)

**Current:**
```go
const q = `
INSERT INTO thinks
    (think_id, category, content, date_created, date_updated)
VALUES
    (:think_id, :category, :content, :date_created, :date_updated)`
```

**Fix:**
```go
const q = `
INSERT INTO thinks
    (think_id, user_id, category, content, date_created, date_updated)
VALUES
    (:think_id, :user_id, :category, :content, :date_created, :date_updated)`
```

#### 4b. Query Method (Line 57)

**Current:**
```go
const q = `
SELECT
    think_id, category, content, date_created, date_updated
FROM
    thinks`
```

**Fix:**
```go
const q = `
SELECT
    think_id, user_id, category, content, date_created, date_updated
FROM
    thinks
WHERE
    user_id = :user_id`
```

Add filter parameter:
```go
func (s *Store) Query(ctx context.Context, userID uuid.UUID, orderBy order.By, page page.Page) ([]thinkbus.Think, error) {
    data := map[string]any{
        "user_id":      userID,
        "offset":       (page.Number() - 1) * page.RowsPerPage(),
        "rows_per_page": page.RowsPerPage(),
    }
    // ... rest of method
}
```

#### 4c. Count Method (Line 96)

**Current:**
```go
const q = `SELECT COUNT(1) AS count FROM thinks`
```

**Fix:**
```go
const q = `SELECT COUNT(1) AS count FROM thinks WHERE user_id = :user_id`

func (s *Store) Count(ctx context.Context, userID uuid.UUID) (int, error) {
    data := map[string]any{
        "user_id": userID,
    }
    // ... rest of method
}
```

#### 4d. QueryByID Method (Line 115)

**Current:**
```go
const q = `
SELECT
    think_id, category, content, date_created, date_updated
FROM
    thinks
WHERE
    think_id = :think_id`
```

**Fix:**
```go
const q = `
SELECT
    think_id, user_id, category, content, date_created, date_updated
FROM
    thinks
WHERE
    think_id = :think_id AND user_id = :user_id`

func (s *Store) QueryByID(ctx context.Context, thinkID uuid.UUID, userID uuid.UUID) (thinkbus.Think, error) {
    data := struct {
        ThinkID string `db:"think_id"`
        UserID  string `db:"user_id"`
    }{
        ThinkID: thinkID.String(),
        UserID:  userID.String(),
    }
    // ... rest of method
}
```

---

### Bug 5: Think Business Layer Signatures
**File:** `business/domain/thinkbus/thinkbus.go`

Update Storer interface (Line 35):
```go
type Storer interface {
    Create(ctx context.Context, think Think) error
    Query(ctx context.Context, userID uuid.UUID, orderBy order.By, page page.Page) ([]Think, error)
    Count(ctx context.Context, userID uuid.UUID) (int, error)
    QueryByID(ctx context.Context, thinkID uuid.UUID, userID uuid.UUID) (Think, error)
}
```

Update Business methods:
```go
func (b *Business) Query(ctx context.Context, userID uuid.UUID, orderBy order.By, page page.Page) ([]Think, error) {
    return b.storer.Query(ctx, userID, orderBy, page)
}

func (b *Business) Count(ctx context.Context, userID uuid.UUID) (int, error) {
    return b.storer.Count(ctx, userID)
}

func (b *Business) QueryByID(ctx context.Context, thinkID uuid.UUID, userID uuid.UUID) (Think, error) {
    return b.storer.QueryByID(ctx, thinkID, userID)
}
```

---

### Bug 6: Think App Handlers Missing user_id
**File:** `app/domain/thinkapp/thinkapp.go`

#### 6a. Create Handler (Line 39)

**Add import:**
```go
import (
    "github.com/francowini/rafiki/app/sdk/mid"  // ADD THIS
    // ... other imports
)
```

**Update create method:**
```go
func (a *app) create(ctx context.Context, r *http.Request) web.Encoder {
    var app NewThink
    if err := web.Decode(r, &app); err != nil {
        return errs.New(errs.InvalidArgument, err)
    }

    // ADD: Get user_id from JWT claims
    userID := mid.GetSubjectID(ctx)
    if userID == uuid.Nil {
        return errs.New(errs.Unauthenticated, errors.New("user not authenticated"))
    }

    nt, err := toBusNewThink(app)
    if err != nil {
        return errs.New(errs.InvalidArgument, err)
    }

    // ADD: Set user_id
    nt.UserID = userID

    think, err := a.thinkBus.Create(ctx, nt)
    if err != nil {
        return errs.Newf(errs.Internal, "create: think[%+v]: %s", app, err)
    }

    return toAppThink(think)
}
```

#### 6b. Query Handler (Line 56)

**Update query method:**
```go
func (a *app) query(ctx context.Context, r *http.Request) web.Encoder {
    qp := parseQueryParams(r)

    // ADD: Get user_id from JWT claims
    userID := mid.GetSubjectID(ctx)
    if userID == uuid.Nil {
        return errs.New(errs.Unauthenticated, errors.New("user not authenticated"))
    }

    page, err := page.Parse(qp.Page, qp.Rows)
    if err != nil {
        return errs.NewFieldErrors("page", err)
    }

    orderBy, err := order.Parse(orderByFields, qp.OrderBy, thinkbus.DefaultOrderBy)
    if err != nil {
        return errs.NewFieldErrors("order", err)
    }

    // MODIFY: Pass userID to Query
    thinks, err := a.thinkBus.Query(ctx, userID, orderBy, page)
    if err != nil {
        return errs.Newf(errs.Internal, "query: %s", err)
    }

    // MODIFY: Pass userID to Count
    total, err := a.thinkBus.Count(ctx, userID)
    if err != nil {
        return errs.Newf(errs.Internal, "count: %s", err)
    }

    return page.NewResponse(thinks, total, page.Number(), page.RowsPerPage())
}
```

#### 6c. QueryByID Handler (Line 91)

**Update queryByID method:**
```go
func (a *app) queryByID(ctx context.Context, r *http.Request) web.Encoder {
    thinkID, err := uuid.Parse(web.Param(r, "think_id"))
    if err != nil {
        return errs.New(errs.InvalidArgument, err)
    }

    // ADD: Get user_id from JWT claims
    userID := mid.GetSubjectID(ctx)
    if userID == uuid.Nil {
        return errs.New(errs.Unauthenticated, errors.New("user not authenticated"))
    }

    // MODIFY: Pass userID to QueryByID
    think, err := a.thinkBus.QueryByID(ctx, thinkID, userID)
    if err != nil {
        return errs.Newf(errs.Internal, "querybyid: thinkID[%s]: %s", thinkID, err)
    }

    return toAppThink(think)
}
```

---

### Bug 7: Auth Not Initialized in main.go
**File:** `api/services/partners/main.go`
**Line:** 168

**Add imports:**
```go
import (
    "github.com/francowini/rafiki/app/sdk/auth"
    "github.com/francowini/rafiki/business/domain/userbus"
    "github.com/francowini/rafiki/business/domain/userbus/stores/userdb"
    "github.com/francowini/rafiki/foundation/keystore"
    // ... other imports
)
```

**Replace placeholder comment with:**
```go
// -------------------------------------------------------------------------
// Initialize authentication support

log.Info(ctx, "startup", "status", "initializing authentication support")

// Load RSA keys for JWT signing/verification
ks := keystore.New()

// Load from filesystem (development and production)
keysLoaded := 0
var err error

// Try loading from zarf/keys directory
keysFS := os.DirFS("./zarf/keys")
keysLoaded, err = ks.LoadByFileSystem(keysFS)
if err != nil {
    return fmt.Errorf("loading auth keys from filesystem: %w", err)
}

if keysLoaded == 0 {
    return fmt.Errorf("no authentication keys loaded - cannot start service")
}

log.Info(ctx, "startup", "keys_loaded", keysLoaded)

// Initialize UserBus (required for authentication)
userStore := userdb.NewStore(log, db)
userBus := userbus.NewBusiness(log, userStore)

// Initialize Auth
authInstance := auth.New(auth.Config{
    Log:       log,
    UserBus:   userBus,
    KeyLookup: ks,
    Issuer:    "rafiki-service",
})

log.Info(ctx, "startup", "status", "authentication support enabled")
```

**Update mux configuration (around line 215):**
```go
cfgMux := mux.Config{
    Build:  build,
    Log:    log,
    DB:     db,
    Tracer: tracer,
    BusConfig: mux.BusConfig{
        ThinkBus: thinkBus,
        UserBus:  userBus,      // ADD THIS
        Auth:     authInstance,  // ADD THIS
    },
}
```

---

### Bug 8: BusConfig Missing Fields
**File:** `app/sdk/mux/mux.go`
**Line:** 51

**Add imports:**
```go
import (
    "github.com/francowini/rafiki/app/sdk/auth"
    "github.com/francowini/rafiki/business/domain/userbus"
    // ... other imports
)
```

**Update BusConfig:**
```go
type BusConfig struct {
    ThinkBus *thinkbus.Business
    UserBus  userbus.ExtBusiness  // ADD THIS
    Auth     *auth.Auth            // ADD THIS
}
```

---

### Bug 9: Auth Routes Not Registered
**File:** `api/services/partners/all/all.go`

**Add import:**
```go
import (
    "github.com/francowini/rafiki/app/domain/authapp"  // ADD THIS
    "github.com/francowini/rafiki/app/domain/checkapp"
    "github.com/francowini/rafiki/app/domain/thinkapp"
    "github.com/francowini/rafiki/app/sdk/mux"
    "github.com/francowini/rafiki/foundation/web"
)
```

**Update Add function:**
```go
func (add) Add(app *web.App, cfg mux.Config) {
    checkapp.Routes(app, checkapp.Config{
        Build: cfg.Build,
        Log:   cfg.Log,
        DB:    cfg.DB,
    })

    // ADD: Register auth routes
    authapp.Routes(app, authapp.Config{
        Auth:    cfg.BusConfig.Auth,
        UserBus: cfg.BusConfig.UserBus,
    })

    thinkapp.Routes(app, thinkapp.Config{
        Log:      cfg.Log,
        ThinkBus: cfg.BusConfig.ThinkBus,
        Auth:     cfg.BusConfig.Auth,  // ADD THIS
    })
}
```

---

### Bug 10: Think Routes Missing Authentication Middleware
**File:** `app/domain/thinkapp/route.go`

**Add imports:**
```go
import (
    "net/http"

    "github.com/francowini/rafiki/app/sdk/auth"    // ADD THIS
    "github.com/francowini/rafiki/app/sdk/mid"      // ADD THIS
    "github.com/francowini/rafiki/business/domain/thinkbus"
    "github.com/francowini/rafiki/foundation/logger"
    "github.com/francowini/rafiki/foundation/web"
)
```

**Update Config struct:**
```go
type Config struct {
    Log      *logger.Logger
    ThinkBus *thinkbus.Business
    Auth     *auth.Auth         // ADD THIS
}
```

**Update Routes function:**
```go
func Routes(app *web.App, cfg Config) {
    const version = "v1"

    // ADD: Create Bearer auth middleware
    bearer := mid.Bearer(cfg.Auth)

    api := newApp(cfg.ThinkBus)

    // ADD: Apply bearer middleware to all routes
    app.HandlerFunc(http.MethodGet, version, "/thinks", api.query, bearer)
    app.HandlerFunc(http.MethodGet, version, "/thinks/{think_id}", api.queryByID, bearer)
    app.HandlerFunc(http.MethodPost, version, "/thinks", api.create, bearer)
}
```

---

### Bug 11: Update Token Expiry to 48 Hours
**File:** `app/sdk/mid/authen.go`
**Line:** 73

**Current:**
```go
ExpiresAt: jwt.NewNumericDate(time.Now().UTC().Add(8760 * time.Hour)), // 1 year
```

**Fix:**
```go
ExpiresAt: jwt.NewNumericDate(time.Now().UTC().Add(48 * time.Hour)), // 48 hours
```

---

## RSA Key Setup

### Development Key (Commit to Repo)

```bash
# Create keys directory
mkdir -p zarf/keys

# Generate RSA private key (4096-bit)
openssl genrsa -out zarf/keys/54bb2165-71e1-41a6-af3e-7da4a0e1e2c1.pem 4096

# Verify key is valid
openssl rsa -in zarf/keys/54bb2165-71e1-41a6-af3e-7da4a0e1e2c1.pem -check -noout

# Set permissions
chmod 600 zarf/keys/54bb2165-71e1-41a6-af3e-7da4a0e1e2c1.pem

# Add to git (exception to .gitignore)
git add -f zarf/keys/54bb2165-71e1-41a6-af3e-7da4a0e1e2c1.pem
```

**Key ID (kid):** `54bb2165-71e1-41a6-af3e-7da4a0e1e2c1`
This is the filename without `.pem` extension

---

## Manual Testing Guide

### Step 1: Start Local Services

```bash
# Clean start (wipes database)
docker compose down -v

# Start all services
docker compose up -d --build

# Watch logs
docker compose logs -f partner-service

# Wait for "api router started" log message
```

### Step 2: Verify Database Schema

```bash
# Connect to PostgreSQL
docker exec -it rafiki-postgres psql -U rafiki -d rafiki

# Check tables exist
\dt

# Expected output:
#              List of relations
#  Schema |       Name        | Type  | Owner
# --------+-------------------+-------+--------
#  public | darwin_migrations | table | rafiki
#  public | thinks            | table | rafiki
#  public | users             | table | rafiki

# Check users table structure
\d users

# Check thinks table structure
\d thinks

# Verify foreign key exists
SELECT conname, conrelid::regclass, confrelid::regclass
FROM pg_constraint
WHERE contype = 'f' AND conrelid = 'thinks'::regclass;

# Expected: thinks_user_id_fkey | thinks | users

# Exit
\q
```

### Step 3: Create Test User via SQL

```bash
# Connect to database
docker exec -it rafiki-postgres psql -U rafiki -d rafiki
```

**Generate bcrypt hash first:**
```bash
# In a separate terminal, create hash generator
cat > /tmp/gen-hash.go <<'EOF'
package main
import (
    "fmt"
    "os"
    "golang.org/x/crypto/bcrypt"
)
func main() {
    if len(os.Args) < 2 {
        fmt.Println("Usage: go run gen-hash.go <password>")
        os.Exit(1)
    }
    hash, err := bcrypt.GenerateFromPassword([]byte(os.Args[1]), bcrypt.DefaultCost)
    if err != nil {
        fmt.Printf("Error: %v\n", err)
        os.Exit(1)
    }
    fmt.Println(string(hash))
}
EOF

# Generate hash (use a test password like "password123")
go run /tmp/gen-hash.go "password123"

# Copy the output hash (starts with $2a$10$...)
```

**Insert user in psql:**
```sql
INSERT INTO users (
    user_id,
    name,
    email,
    roles,
    password_hash,
    department,
    enabled,
    date_created,
    date_updated
) VALUES (
    gen_random_uuid(),
    'Test User',
    'test@example.com',
    ARRAY['USER']::TEXT[],
    '$2a$10$YOUR_BCRYPT_HASH_HERE',  -- Replace with actual hash
    NULL,
    true,
    NOW(),
    NOW()
);

-- Verify user was created
SELECT user_id, name, email, roles, enabled FROM users;

-- Exit
\q
```

### Step 4: Test Health Endpoints

```bash
# Test readiness
curl http://localhost:3000/v1/readiness

# Expected: {"status":"ok"}

# Test liveness
curl http://localhost:3000/v1/liveness

# Expected: {"status":"ok"}
```

### Step 5: Test Authentication (Login)

```bash
# Test Basic Auth to get JWT token
curl -i -X GET \
  -H "Authorization: Basic $(echo -n 'test@example.com:password123' | base64)" \
  http://localhost:3000/v1/auth/token/54bb2165-71e1-41a6-af3e-7da4a0e1e2c1

# Expected response:
# HTTP/1.1 200 OK
# Content-Type: application/json
#
# {"token":"eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCIsImtpZCI6IjU0YmIyMTY1LTcxZTEtNDFhNi1hZjNlLTdkYTRhMGUxZTJjMSJ9..."}

# Copy the token value for next steps
```

**Decode token to verify claims (optional):**
```bash
# Use jwt.io or this bash command
TOKEN="your-token-here"
echo $TOKEN | cut -d. -f2 | base64 -d | jq .

# Expected output:
# {
#   "sub": "user-uuid-here",
#   "iss": "rafiki-service",
#   "exp": 1234567890,
#   "iat": 1234567890,
#   "roles": ["USER"]
# }
```

### Step 6: Test Invalid Authentication

```bash
# Test without Authorization header
curl -i -X GET http://localhost:3000/v1/auth/token/54bb2165-71e1-41a6-af3e-7da4a0e1e2c1

# Expected: HTTP/1.1 401 Unauthorized

# Test with wrong password
curl -i -X GET \
  -H "Authorization: Basic $(echo -n 'test@example.com:wrongpassword' | base64)" \
  http://localhost:3000/v1/auth/token/54bb2165-71e1-41a6-af3e-7da4a0e1e2c1

# Expected: HTTP/1.1 401 Unauthorized

# Test with non-existent user
curl -i -X GET \
  -H "Authorization: Basic $(echo -n 'fake@example.com:password123' | base64)" \
  http://localhost:3000/v1/auth/token/54bb2165-71e1-41a6-af3e-7da4a0e1e2c1

# Expected: HTTP/1.1 401 Unauthorized
```

### Step 7: Test Thinks with Bearer Token

**Set token variable:**
```bash
# Use token from Step 5
TOKEN="eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9..."
```

**Test creating a think:**
```bash
curl -i -X POST \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "category": "personal",
    "content": "This is my first authenticated think"
  }' \
  http://localhost:3000/v1/thinks

# Expected: HTTP/1.1 201 Created
# Response body: {"id":"think-uuid","category":"personal","content":"This is my first authenticated think",...}
```

**Test querying thinks:**
```bash
curl -i -X GET \
  -H "Authorization: Bearer $TOKEN" \
  http://localhost:3000/v1/thinks

# Expected: HTTP/1.1 200 OK
# Response: {"items":[{...}],"total":1,"page":1,"rowsPerPage":10}
```

**Test querying by ID:**
```bash
# Use the think_id from create response
THINK_ID="uuid-from-create-response"

curl -i -X GET \
  -H "Authorization: Bearer $TOKEN" \
  http://localhost:3000/v1/thinks/$THINK_ID

# Expected: HTTP/1.1 200 OK
# Response: {"id":"...","category":"personal",...}
```

### Step 8: Test Unauthenticated Access (Should Fail)

```bash
# Try to create think without token
curl -i -X POST \
  -H "Content-Type: application/json" \
  -d '{"category":"personal","content":"test"}' \
  http://localhost:3000/v1/thinks

# Expected: HTTP/1.1 401 Unauthorized

# Try to query thinks without token
curl -i -X GET http://localhost:3000/v1/thinks

# Expected: HTTP/1.1 401 Unauthorized
```

### Step 9: Test User Isolation

**Create second user:**
```bash
docker exec -it rafiki-postgres psql -U rafiki -d rafiki

INSERT INTO users (user_id, name, email, roles, password_hash, department, enabled, date_created, date_updated)
VALUES (gen_random_uuid(), 'User Two', 'user2@example.com', ARRAY['USER']::TEXT[], '$2a$10$YOUR_BCRYPT_HASH', NULL, true, NOW(), NOW());

\q
```

**Login as second user:**
```bash
curl -X GET \
  -H "Authorization: Basic $(echo -n 'user2@example.com:password123' | base64)" \
  http://localhost:3000/v1/auth/token/54bb2165-71e1-41a6-af3e-7da4a0e1e2c1

# Copy the token
TOKEN2="new-token-here"
```

**Verify user2 has no thinks:**
```bash
curl -X GET \
  -H "Authorization: Bearer $TOKEN2" \
  http://localhost:3000/v1/thinks

# Expected: {"items":[],"total":0,"page":1,"rowsPerPage":10}
```

**Verify user2 cannot access user1's think:**
```bash
# Use the THINK_ID from Step 7 (created by user1)
curl -i -X GET \
  -H "Authorization: Bearer $TOKEN2" \
  http://localhost:3000/v1/thinks/$THINK_ID

# Expected: HTTP/1.1 404 Not Found or 500 Internal Server Error
# (User2 cannot see User1's think)
```

### Step 10: Test Token Expiry

**Check token expiration time:**
```bash
# Decode token and check 'exp' claim
echo $TOKEN | cut -d. -f2 | base64 -d | jq .exp

# Expected: Unix timestamp 48 hours from 'iat' (issued at)
# Calculate: exp - iat should equal 172800 seconds (48 hours)
```

**Test with expired token (optional - requires waiting or manipulation):**
```bash
# Create token with past expiry (requires code change for testing)
# Or wait 48 hours and test that token is rejected
```

### Step 11: Verify Database Relationships

```bash
docker exec -it rafiki-postgres psql -U rafiki -d rafiki

-- Check all thinks have user_id set
SELECT think_id, user_id, category, content FROM thinks;

-- Verify foreign key constraint
-- Try to insert think with invalid user_id (should fail)
INSERT INTO thinks (think_id, user_id, category, content, date_created, date_updated)
VALUES (gen_random_uuid(), '00000000-0000-0000-0000-000000000000', 'test', 'test', NOW(), NOW());

-- Expected error: ERROR:  insert or update on table "thinks" violates foreign key constraint "thinks_user_id_fkey"

-- Test CASCADE delete
-- Delete a user and verify their thinks are deleted
SELECT user_id FROM users WHERE email = 'test@example.com';
-- Copy the user_id

SELECT COUNT(*) FROM thinks WHERE user_id = 'user-id-here';
-- Note the count

DELETE FROM users WHERE email = 'test@example.com';

SELECT COUNT(*) FROM thinks WHERE user_id = 'user-id-here';
-- Expected: 0 (thinks were cascade deleted)

\q
```

---

## Testing Checklist

- [ ] Database migrations run successfully
- [ ] users table exists with correct schema
- [ ] thinks table has user_id foreign key
- [ ] Foreign key constraint enforces referential integrity
- [ ] Test user can be created via SQL
- [ ] Health endpoints return 200 OK
- [ ] Login with valid credentials returns JWT token
- [ ] Token contains correct claims (sub, iss, exp, roles)
- [ ] Token expiry is 48 hours (172800 seconds)
- [ ] Login with invalid credentials returns 401
- [ ] Create think with token succeeds
- [ ] Query thinks with token returns only user's thinks
- [ ] Query think by ID with token succeeds
- [ ] Create think without token returns 401
- [ ] Query thinks without token returns 401
- [ ] User2 cannot see User1's thinks
- [ ] Cascade delete removes user's thinks
- [ ] Service logs show "keys_loaded: 1"
- [ ] Service logs show "authentication support enabled"

---

## Common Issues and Solutions

### Issue: "key not found" error

**Symptoms:** Service fails to start with "key not found" error

**Solutions:**
```bash
# Verify key exists
ls -la zarf/keys/

# Verify key is valid PEM format
openssl rsa -in zarf/keys/54bb2165-71e1-41a6-af3e-7da4a0e1e2c1.pem -check -noout

# Regenerate key if corrupted
rm zarf/keys/54bb2165-71e1-41a6-af3e-7da4a0e1e2c1.pem
openssl genrsa -out zarf/keys/54bb2165-71e1-41a6-af3e-7da4a0e1e2c1.pem 4096
```

### Issue: "user not found" or authentication failure

**Symptoms:** Valid credentials return 401

**Solutions:**
```bash
# Verify user exists
docker exec -it rafiki-postgres psql -U rafiki -d rafiki \
  -c "SELECT email, enabled FROM users WHERE email='test@example.com';"

# Verify user is enabled
# enabled column should be 't' (true)

# Verify password hash format
docker exec -it rafiki-postgres psql -U rafiki -d rafiki \
  -c "SELECT email, password_hash FROM users WHERE email='test@example.com';"

# Hash should start with $2a$ or $2b$ (bcrypt format)

# Regenerate password hash if incorrect
go run /tmp/gen-hash.go "password123"
# Update user with new hash
```

### Issue: "violates foreign key constraint"

**Symptoms:** Cannot insert think, foreign key error

**Solutions:**
```bash
# Verify user_id exists in users table
docker exec -it rafiki-postgres psql -U rafiki -d rafiki \
  -c "SELECT user_id FROM users WHERE email='test@example.com';"

# Check if Think model has UserID field set
# Verify mid.GetSubjectID(ctx) is being called in handlers
# Check service logs for "user not authenticated" errors
```

### Issue: Token validation fails

**Symptoms:** Valid token returns 401

**Solutions:**
```bash
# Verify token format (3 parts separated by dots)
echo $TOKEN | tr '.' '\n' | wc -l
# Expected: 3

# Decode token header
echo $TOKEN | cut -d. -f1 | base64 -d | jq .
# Should show: {"alg":"RS256","typ":"JWT","kid":"54bb2165-71e1-41a6-af3e-7da4a0e1e2c1"}

# Verify kid in token matches key filename
# Verify issuer matches "rafiki-service"

# Check service logs for OPA validation errors
docker compose logs partner-service | grep -i opa
```

### Issue: User can see other users' thinks

**Symptoms:** User sees thinks they didn't create

**Solutions:**
```bash
# Verify WHERE user_id = :user_id clause in queries
# Check thinkdb/thinkdb.go Query and Count methods

# Verify user_id is being passed from handlers
# Check thinkapp/thinkapp.go query and create methods

# Test with SQL query
docker exec -it rafiki-postgres psql -U rafiki -d rafiki \
  -c "SELECT user_id, COUNT(*) FROM thinks GROUP BY user_id;"

# Should show thinks grouped by different user_ids
```

---

## Deployment to Hetzner

See separate document: `docs/HETZNER_DEPLOYMENT_PLAN.md`

---

## Success Criteria

✅ All 11 bugs fixed
✅ Development RSA key generated and committed
✅ Service starts without errors
✅ Auth endpoints accessible
✅ Users can login and receive JWT
✅ Thinks require authentication
✅ Users only see their own thinks
✅ Token expiry set to 48 hours
✅ Cascade delete works correctly

---

## Next Steps

After backend implementation is complete:
1. Deploy to Hetzner (see deployment plan)
2. Frontend team can begin login UI implementation
3. Create additional admin/test users as needed
4. Monitor logs for authentication errors
5. Plan user registration endpoint (future feature)

---

**Questions?** Review the multi-mind analysis document for detailed architecture explanation.
