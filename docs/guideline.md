# Key Learnings from Think Saver Refactor

## Important Note on Notation
**`*` is used as a wildcard placeholder** in this document to represent patterns across different domains (product, user, home, think). In the actual codebase, replace `*` with the specific entity name.

Examples:
- `toBusNew*` means → `toBusNewProduct`, `toBusNewUser`, `toBusNewHome`, `toBusNewThink`
- `New*` means → `NewProduct`, `NewUser`, `NewHome`, `NewThink`
- `*bus` means → `productbus`, `userbus`, `homebus`, `thinkbus`

---

## 1. Validation Belongs in the App Layer

**Pattern**: All validation happens in app layer conversion functions via `Parse()` methods

```go
// App Layer - Actual example: toBusNewProduct
func toBusNewProduct(ctx context.Context, app NewProduct) (productbus.NewProduct, error) {
    userID, err := mid.GetUserID(ctx)
    if err != nil {
        return productbus.NewProduct{}, fmt.Errorf("getuserid: %w", err)
    }

    name, err := name.Parse(app.Name)          // ← Validate here
    if err != nil {
        return productbus.NewProduct{}, fmt.Errorf("parse name: %w", err)
    }

    cost, err := money.Parse(app.Cost)         // ← Validate here
    if err != nil {
        return productbus.NewProduct{}, fmt.Errorf("parse cost: %w", err)
    }

    return productbus.NewProduct{
        UserID: userID,
        Name:   name,    // Already validated types
        Cost:   cost,
    }, nil
}

// Business Layer - Zero validation
func (b *Business) Create(ctx context.Context, np NewProduct) (Product, error) {
    now := time.Now()

    prd := Product{
        ID:          uuid.New(),
        UserID:      np.UserID,
        Name:        np.Name,    // Already validated
        Cost:        np.Cost,    // Already validated
        DateCreated: now,
        DateUpdated: now,
    }

    if err := b.storer.Create(ctx, prd); err != nil {
        return Product{}, fmt.Errorf("create: %w", err)
    }

    return prd, nil
}
```

**Pattern applies to**:
- `toBusNewProduct` → validates Name, Money, Quantity
- `toBusNewUser` → validates Name, Email, Roles, Password
- `toBusNewHome` → validates Address, HomeType
- `toBusNewThink` → validates Category, Content

**Never**: Put validation in business layer `Create()` methods

---

## 2. Value Objects Pattern

All domain primitives are **typed value objects** with Parse validation:

```
business/types/
├── name/      → Name.Parse(string)       → validates regex ^[a-zA-Z0-9][a-zA-Z0-9' -]{2,19}$
├── money/     → Money.Parse(float64)     → validates range 0-1,000,000
├── quantity/  → Quantity.Parse(int)      → validates range 0-1,000,000
├── role/      → Role.Parse(string)       → validates enum (Admin, User)
├── hometype/  → HomeType.Parse(string)   → validates enum (Single, Condo)
└── content/   → Content.Parse(string)    → validates 2 paragraphs (new)
```

**Real examples from codebase**:

```go
// name/name.go
type Name struct {
    value string
}

func (n Name) String() string {
    return n.value
}

func Parse(value string) (Name, error) {
    if !nameRegEx.MatchString(value) {
        return Name{}, fmt.Errorf("invalid name %q", value)
    }
    return Name{value}, nil
}
```

**Benefits**:
- Type safety at compile time
- Validation centralized in one place
- Business layer works with types, never primitives
- Impossible to create invalid values

---

## 3. Three-Layer Architecture

```
App Layer (HTTP)      → Validation + Conversion (productapp, userapp, homeapp)
    ↓
Business Layer        → Business logic only (productbus, userbus, homebus)
    ↓
Store Layer (DB)      → Data persistence (productdb, userdb, homedb)
```

**Strict rules**:
- App never calls Store directly
- Business doesn't know about HTTP
- Store doesn't know about business logic

**Actual directory structure**:
```
app/domain/productapp/
    ├── productapp.go    # HTTP handlers
    ├── model.go         # Conversion + Validation
    └── route.go         # Route registration

business/domain/productbus/
    ├── productbus.go    # Business logic
    ├── model.go         # Business types
    └── stores/productdb/
        ├── productdb.go # SQL operations
        └── model.go     # DB models

business/types/
    ├── name/           # Shared value objects
    ├── money/
    └── quantity/
```

---

## 4. Import Direction Rules (Clean Architecture)

**The Golden Rule**: Imports flow toward abstractions, preventing circular dependencies.

### The Dependency Inversion Pattern (Store Layer)

The store layer uses **Dependency Inversion Principle** - one of the most critical patterns in this codebase:

```
Physical Structure:          Import Direction:           Runtime Flow:

business/domain/            productbus                  productBus.Create()
├── productbus/             (no import of productdb)          ↓
│   ├── productbus.go      ┌────────────┐              calls through interface
│   │   → Storer interface │            │                     ↓
│   └── stores/            │            ↓              productdb.Create()
│       └── productdb/     │      productdb
│           └── productdb  └──── (imports productbus)
                                 to implement interface
```

**Critical Insight**: Store is nested inside business folder BUT imports go INWARD to abstraction!

#### How It Works

**Step 1: Business defines interface** ([productbus/productbus.go](business/domain/productbus/productbus.go:29-38))
```go
// Package productbus - NO import of productdb anywhere!
package productbus

// Storer interface declares behavior for data persistence
type Storer interface {
    Create(ctx context.Context, prd Product) error
    Update(ctx context.Context, prd Product) error
    Delete(ctx context.Context, prd Product) error
    Query(ctx context.Context, filter QueryFilter, orderBy order.By, page page.Page) ([]Product, error)
    QueryByID(ctx context.Context, productID uuid.UUID) (Product, error)
}

// Business depends ONLY on interface (abstraction)
type Business struct {
    log      *logger.Logger
    userBus  userbus.ExtBusiness
    delegate *delegate.Delegate
    storer   Storer  // ← Interface, not concrete type!
}

func NewBusiness(log *logger.Logger, userBus userbus.ExtBusiness,
                 delegate *delegate.Delegate, storer Storer) *Business {
    return &Business{
        log:      log,
        userBus:  userBus,
        delegate: delegate,
        storer:   storer,  // ← Injected from outside
    }
}
```

**Step 2: Store implements interface** ([productdb/productdb.go](business/domain/productbus/stores/productdb/productdb.go:1-47))
```go
// Package productdb - DOES import productbus
package productdb

import (
    "github.com/ardanlabs/service/business/domain/productbus"  // ✅ Imports business to implement interface
    "github.com/ardanlabs/service/business/sdk/sqldb"
)

// Store implements productbus.Storer interface
type Store struct {
    log *logger.Logger
    db  sqlx.ExtContext
}

func NewStore(log *logger.Logger, db *sqlx.DB) *Store {
    return &Store{
        log: log,
        db:  db,
    }
}

// Implements productbus.Storer.Create
func (s *Store) Create(ctx context.Context, prd productbus.Product) error {
    // DB operations...
}
```

**Step 3: Main wires them together** ([main.go](api/services/sales/main.go:181))
```go
// Composition root - where dependency injection happens
func run(ctx context.Context, log *logger.Logger) error {
    // Create concrete store implementation
    productStore := productdb.NewStore(log, db)

    // Inject into business layer (satisfies Storer interface)
    productBus := productbus.NewBusiness(log, userBus, delegate, productStore)
    //                                                           ↑
    //                                    Concrete type injected, business only sees interface
}
```

### Layer Hierarchy and Import Rules

```
┌─────────────────────────────────────────────────────┐
│   App Layer (app/domain/*)                          │  ← Can import everything below
├─────────────────────────────────────────────────────┤
│   App SDK (app/sdk/*)                               │  ← Can import business layer
├─────────────────────────────────────────────────────┤
│   Business Layer (business/domain/*/[!stores])      │  ← Defines interfaces
│   - Depends on interfaces ONLY                      │
├─────────────────────────────────────────────────────┤
│   Store Layer (business/domain/*/stores/*db/)       │  ← Implements interfaces
│   - Imports parent business to implement interface  │
├─────────────────────────────────────────────────────┤
│   Business SDK (business/sdk/*)                     │  ← Pure utilities
├─────────────────────────────────────────────────────┤
│   Business Types (business/types/*)                 │  ← Value objects
├─────────────────────────────────────────────────────┤
│   Foundation (foundation/*)                         │  ← Base layer
└─────────────────────────────────────────────────────┘
```

### Import Rules by Layer

#### **Foundation Layer** (`foundation/`)
- ✅ Can import: Standard library, third-party packages
- ❌ Cannot import: `app/*`, `business/*`
- **Purpose**: Base infrastructure (logger, web, otel, docker)

#### **Business Types** (`business/types/`)
- ✅ Can import: Standard library only
- ❌ Cannot import: Any project packages
- **Purpose**: Primitive value objects (Name, Money, Quantity)

#### **Business SDK** (`business/sdk/`)
- ✅ Can import: `foundation/*`, `business/types/*`
- ❌ Cannot import: `business/domain/*` (except testing utilities)
- **Purpose**: Shared business utilities (order, page, delegate, sqldb)

#### **Store Layer** (`business/domain/*/stores/*db/`)
- ✅ Can import: Parent business domain ONLY, `business/sdk/*`, `business/types/*`, `foundation/*`
- ❌ Cannot import: Other business domains, sibling stores, `app/*`
- **Purpose**: Data persistence implementing parent's Storer interface
- **Critical**: Implements interface defined by parent business layer

**Example Import - Store Layer**:
```go
// productdb/productdb.go
import (
    "github.com/ardanlabs/service/business/domain/productbus"  // ✅ Parent domain
    "github.com/ardanlabs/service/business/sdk/sqldb"          // ✅ Business SDK
    "github.com/ardanlabs/service/business/types/money"        // ✅ Value objects
    "github.com/ardanlabs/service/foundation/logger"           // ✅ Foundation

    // ❌ NEVER import these:
    // "github.com/ardanlabs/service/business/domain/userbus"      // Other domains
    // "github.com/ardanlabs/service/business/domain/homebus"      // Other domains
    // "github.com/ardanlabs/service/app/domain/productapp"        // App layer
)
```

#### **Business Domains** (`business/domain/*/`)
- ✅ Can import: Other `business/domain/*` (one-way), `business/sdk/*`, `business/types/*`, `foundation/*`
- ❌ Cannot import: Own store (`stores/*db`), `app/*`
- **Purpose**: Business logic, defines Storer interfaces
- **Critical**: NEVER imports own store - only defines interface

**Example Import - Business Layer**:
```go
// productbus/productbus.go
import (
    "github.com/ardanlabs/service/business/domain/userbus"  // ✅ Other domain (one-way)
    "github.com/ardanlabs/service/business/sdk/delegate"    // ✅ Business SDK
    "github.com/ardanlabs/service/business/types/money"     // ✅ Value objects
    "github.com/ardanlabs/service/foundation/logger"        // ✅ Foundation

    // ❌ NEVER import these:
    // "github.com/ardanlabs/service/business/domain/productbus/stores/productdb"  // Own store!
    // "github.com/ardanlabs/service/app/domain/productapp"                         // App layer
)

// Define interface - don't import concrete implementation
type Storer interface {
    Create(ctx context.Context, prd Product) error
}
```

**Cross-Domain Imports (One-Way)**:
```go
// productbus → userbus ✅ OK (one direction)
productbus/productbus.go:
    import "github.com/ardanlabs/service/business/domain/userbus"

// userbus → productbus ❌ FORBIDDEN (would create cycle)
// Use delegate pattern instead!
userbus/userbus.go:
    import "github.com/ardanlabs/service/business/sdk/delegate"  // ✅ Use this
```

#### **App SDK** (`app/sdk/`)
- ✅ Can import: `business/domain/*`, `business/sdk/*`, `business/types/*`, `foundation/*`
- ❌ Cannot import: `app/domain/*`
- **Purpose**: HTTP middleware and utilities (domain-aware)
- **Note**: SDK packages are intermediate - can import business for domain awareness

#### **App Domains** (`app/domain/*/`)
- ✅ Can import: `business/domain/*`, `app/sdk/*`, `business/sdk/*`, `business/types/*`, `foundation/*`
- ❌ Cannot import: Other `app/domain/*`, `business/domain/*/stores/*`
- **Purpose**: HTTP handlers and API layer

### Visual Import Flow

```
main.go (Composition Root)
    │
    ├─→ Creates: productdb.NewStore(log, db)
    │           ↓
    │           implements productbus.Storer
    │
    └─→ Injects: productbus.NewBusiness(..., productStore)
                         ↓
                 Business only sees interface

Import Dependencies:

App Domain (productapp)
    │
    ├─→ App SDK (mid, errs)
    │       └─→ Business Domain (userbus)
    │
    ├─→ Business Domain (productbus)  ← Defines Storer interface
    │       │                         ← NEVER imports productdb
    │       │
    │       ├─→ Business Domain (userbus)  ← One-way only!
    │       ├─→ Business SDK (delegate)
    │       └─→ Business Types (money)
    │
    └─→ (Store wired at runtime via interface)

Store Layer (productdb)
    │
    ├─→ Business Domain (productbus)  ← To implement interface!
    ├─→ Business SDK (sqldb)
    ├─→ Business Types (money)
    └─→ Foundation (logger)
```

### Decision Matrix

| Scenario | Solution | Import Direction |
|----------|----------|------------------|
| App needs business logic | ✅ Direct import | app → business |
| Business needs data access | ✅ Define interface, inject store | business defines, store implements |
| Business needs own store | ❌ NEVER! Use interface | business ← store (store imports business) |
| Business A needs Business B | ✅ One-way import only | A → B (never B → A) |
| Business B reacts to A | ✅ Delegate pattern | Both → delegate |
| Store needs domain types | ✅ Import parent domain | store → parent business |
| Store needs sibling domain | ❌ Only through parent | store → parent → sibling |

### Key Principles of Dependency Inversion

1. **Business defines contract (interface)** - What it needs, not how
2. **Store implements contract** - Concrete implementation details
3. **Business never imports store** - Depends on abstraction only
4. **Store imports business** - To implement interface and use types
5. **Main wires dependencies** - Composition root injects concrete into abstract
6. **Physical nesting ≠ import direction** - Store nested but imports upward

**The Brilliant Part**:
- Store is physically inside `business/domain/productbus/stores/`
- But logically, it's a separate layer that depends on business
- Business knows "I need a Storer" (what)
- Store provides "Here's how to store" (how)
- Main says "Use productdb for Storer" (wiring)

This is Clean Architecture's Dependency Rule in perfect form!

---

## 5. Order Field Mapping (3 Layers)

Brilliant indirection pattern to decouple layers:

**App Layer** (`productapp/order.go`):
```go
var orderByFields = map[string]string{
    "product_id": productbus.OrderByProductID,  // API → Business
    "name":       productbus.OrderByName,
    "cost":       productbus.OrderByCost,
}
```

**Business Layer** (`productbus/order.go`):
```go
const (
    OrderByProductID = "a"  // Short constants for indirection
    OrderByName      = "c"
    OrderByCost      = "d"
)
```

**Store Layer** (`productdb/order.go`):
```go
var orderByFields = map[string]string{
    productbus.OrderByProductID: "product_id",  // Business → DB
    productbus.OrderByName:      "name",
    productbus.OrderByCost:      "cost",
}
```

**Flow**: `"name"` (API) → `"c"` (Business) → `"name"` (DB column)

**Why**: API can rename fields without touching business/store layers

---

## 5. Parse Functions Everywhere

**Consistent signature**: `Parse(primitive) → (Type, error)`

**Domain Types**:
```go
name.Parse(string) → (Name, error)
money.Parse(float64) → (Money, error)
quantity.Parse(int) → (Quantity, error)
role.Parse(string) → (Role, error)
```

**SDK Types**:
```go
page.Parse(string, string) → (Page, error)                           // Defaults: page=1, rows=10
order.Parse(map, string, default) → (By, error)                      // With field mapping
```

**Business Enums**:
```go
thinkbus.ParseCategory(string) → (Category, error)                   // Enum validation
hometype.Parse(string) → (HomeType, error)                           // Enum validation
```

**Pattern**:
- Accepts primitives (string, int, float64)
- Validates business rules
- Returns typed values or error
- Provides `MustParse()` for testing (panics on error)

---

## 6. Error Flow Through Layers

**Actual example from productdb**:

```go
// Store Layer (productdb/productdb.go)
if err := sqldb.NamedQueryStruct(ctx, s.log, s.db, q, data, &dbPrd); err != nil {
    if errors.Is(err, sqldb.ErrDBNotFound) {                          // DB error
        return productbus.Product{}, fmt.Errorf("db: %w", productbus.ErrNotFound)  // → Domain error
    }
    return productbus.Product{}, fmt.Errorf("db: %w", err)
}
```

```go
// App Layer (productapp/productapp.go)
prd, err := mid.GetProduct(ctx)
if err != nil {
    return errs.Newf(errs.Internal, "product missing in context: %s", err)  // → Error code
}
```

**Flow**:
```
sqldb.ErrDBNotFound (sql.ErrNoRows)
    ↓ (store converts)
productbus.ErrNotFound (sentinel error)
    ↓ (app wraps)
errs.NotFound (error code)
    ↓ (middleware converts)
404 Not Found (HTTP status)
```

**Each layer converts errors to its abstraction level**

---

## 7. Sentinel Errors in Business Layer Only

**Business layer defines domain errors**:
```go
// productbus/productbus.go
var (
    ErrNotFound     = errors.New("product not found")
    ErrUserDisabled = errors.New("user disabled")
    ErrInvalidCost  = errors.New("cost not valid")
)

// userbus/userbus.go
var (
    ErrNotFound              = errors.New("user not found")
    ErrUniqueEmail           = errors.New("email is not unique")
    ErrAuthenticationFailure = errors.New("authentication failed")
)
```

**Store layer converts DB errors**:
```go
if errors.Is(err, sqldb.ErrDBNotFound) {
    return fmt.Errorf("db: %w", productbus.ErrNotFound)
}
```

**App layer uses error codes** (NOT sentinel errors):
```go
return errs.New(errs.NotFound, productbus.ErrNotFound)
return errs.New(errs.InvalidArgument, err)
return errs.Newf(errs.Internal, "create: %s", err)
```

**Never**: Return sentinel errors from app layer (use errs.ErrCode)

---

## 8. Delegate Pattern for Decoupling

Prevents circular dependencies between domains:

**Registration** (`productbus/event.go`):
```go
func (b *Business) registerDelegateFunctions() {
    b.delegate.Register(userbus.DomainName, userbus.ActionDeleted, b.actionUserDeleted)
}

func (b *Business) actionUserDeleted(ctx context.Context, data delegate.Data) error {
    // Clean up products when user is deleted
}
```

**Triggering** (`userbus/userbus.go`):
```go
func (b *Business) Delete(ctx context.Context, usr User) error {
    if err := b.storer.Delete(ctx, usr); err != nil {
        return fmt.Errorf("delete: %w", err)
    }

    // Notify other domains via delegate
    if err := b.delegate.Call(ctx, ActionDeletedData(usr.ID)); err != nil {
        return fmt.Errorf("delegate call: %w", err)
    }

    return nil
}
```

**Benefits**: In-process event bus without direct imports between domains

---

## 9. Conversion Functions at Every Layer

**Naming convention**:

**App Layer**:
```go
toAppProduct(productbus.Product) → Product              // Business → App
toAppProducts([]productbus.Product) → []Product         // Slice conversion

toBusNewProduct(NewProduct) → (productbus.NewProduct, error)    // App → Business (validates)
toBusUpdateProduct(UpdateProduct) → (productbus.UpdateProduct, error)
```

**Store Layer**:
```go
toDBProduct(productbus.Product) → product               // Business → DB
toBusProduct(product) → (productbus.Product, error)     // DB → Business (validates)
toBusProducts([]product) → ([]productbus.Product, error)
```

**Pattern**:
- `toApp*` - Business → App (JSON serialization)
- `toBus*` - App/DB → Business (validation via Parse)
- `toDB*` - Business → DB (for persistence)

**Always**: Keep layers isolated via conversion functions

---

## 10. SDK Packages - Separation of Concerns

**Critical Question**: Why do we have TWO SDK folders (`business/sdk/` and `app/sdk/`)?

**Answer**: They serve different layers with fundamentally different concerns and import rules.

### business/sdk/ - Domain-Agnostic Business Infrastructure

**Concern**: "How do we handle **data and business operations** that are common across all domains?"

**Key Principle**: **Domain-agnostic utilities** - knows NOTHING about specific domains (user, product, home, think)

**Purpose**: Infrastructure that business domains use for common operations

```
business/sdk/
├── delegate/   → Event delegation between domains (prevents circular imports)
├── order/      → Ordering logic (Parse, NewBy) - works for any domain
├── page/       → Pagination (Parse, Number, RowsPerPage) - generic pagination
├── sqldb/      → Database utilities (NamedExecContext, transactions)
└── migrate/    → Database migrations
```

**Import Rules**:
- ✅ Can import: `foundation/*`, `business/types/*`
- ❌ Cannot import: `business/domain/*` (except testing utilities)
- **Must be reusable** by ALL business domains without coupling

**Examples**:

```go
// business/sdk/page/ - Domain-agnostic pagination
package page

type Page struct {
    number int
    rows   int
}

func Parse(page string, rowsPerPage string) (Page, error) {
    // Pure pagination logic - works for users, products, homes, thinks, etc.
    // NO knowledge of any specific domain
}
```

```go
// business/sdk/order/ - Domain-agnostic ordering
package order

type By struct {
    Field     string
    Direction string
}

func Parse(fieldMappings map[string]string, orderBy string, defaultOrder By) (By, error) {
    // Generic ordering - works for any domain
    // Caller provides field mappings
}
```

```go
// business/sdk/delegate/ - Domain-agnostic event system
package delegate

type Delegate struct {
    funcs map[domain]map[action][]Func
}

func (d *Delegate) Register(domainType string, actionType string, fn Func) {
    // Event bus - NO knowledge of specific domains
}
```

```go
// business/sdk/sqldb/ - Domain-agnostic database utilities
package sqldb

func NamedExecContext(ctx context.Context, log *logger.Logger,
                      db sqlx.ExtContext, query string, data any) error {
    // SQL execution wrapper - works for any domain
    // NO knowledge of products, users, homes, thinks
}
```

**Used by**: All business domains (userbus, productbus, homebus, thinkbus) and all stores

---

### app/sdk/ - Domain-Aware HTTP Infrastructure

**Concern**: "How do we handle **HTTP/Web operations** that need to understand our domains?"

**Key Principle**: **Domain-aware HTTP utilities** - knows about business domains for middleware/auth purposes

**Purpose**: Infrastructure that app layer uses for HTTP concerns

```
app/sdk/
├── errs/       → HTTP error codes and status mapping (domain errors → HTTP status)
├── mid/        → HTTP middleware (auth, logging, context) - stores domain types
├── query/      → JSON response formatting (wraps results with pagination)
├── auth/       → Authentication/authorization (works with userbus)
└── mux/        → Route configuration and setup
```

**Import Rules**:
- ✅ Can import: `business/domain/*`, `business/sdk/*`, `business/types/*`, `foundation/*`
- ❌ Cannot import: `app/domain/*`
- **Domain-aware** - knows about specific domains for middleware/auth

**Examples**:

```go
// app/sdk/errs/ - HTTP error codes (domain-aware for mapping)
package errs

type ErrCode struct {
    value int
}

var httpStatus = map[ErrCode]int{
    NotFound:        404,  // Maps business errors to HTTP status
    InvalidArgument: 400,
    Internal:        500,
}

// HTTP-specific - maps domain errors to status codes
```

```go
// app/sdk/mid/ - HTTP middleware (DOMAIN-AWARE!)
package mid

import (
    "github.com/ardanlabs/service/business/domain/userbus"     // ✅ Imports domain!
    "github.com/ardanlabs/service/business/domain/productbus"  // ✅ Imports domain!
)

// Stores domain objects in HTTP context
func GetUser(ctx context.Context) (userbus.User, error) {
    v, ok := ctx.Value(userKey).(userbus.User)  // ← Knows about userbus.User!
    return v, nil
}

func GetProduct(ctx context.Context) (productbus.Product, error) {
    v, ok := ctx.Value(productKey).(productbus.Product)  // ← Knows about productbus.Product!
    return v, nil
}
```

```go
// app/sdk/query/ - HTTP response formatting
package query

// Wraps query results with pagination metadata for JSON response
func NewResult(items any, total int, page page.Page) Result {
    return Result{
        Items:       items,        // Domain objects (products, users, etc.)
        Total:       total,
        Page:        page.Number(),
        RowsPerPage: page.RowsPerPage(),
    }
}

// HTTP-specific JSON response structure
```

```go
// app/sdk/auth/ - Authentication (DOMAIN-AWARE!)
package auth

import (
    "github.com/ardanlabs/service/business/domain/userbus"  // ✅ Imports domain!
)

// Authenticates and returns domain user
func Authenticate(ctx context.Context, userBus *userbus.Business,
                  email, password string) (Claims, error) {
    // Works directly with userbus domain!
}
```

---

### Visual Comparison

```
┌────────────────────────────────────────────────────────────────┐
│                    app/sdk/ (HTTP Layer)                       │
│  ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━  │
│  Concern: "How do we handle HTTP/Web operations?"             │
│                                                                │
│  errs/       → HTTP status codes (404, 500, etc.)             │
│  mid/        → HTTP middleware (auth, logging, context)       │
│  query/      → JSON response formatting                       │
│  auth/       → Authentication/authorization                   │
│  mux/        → Route configuration                            │
│                                                                │
│  Domain-Aware: ✅ CAN import business/domain/*                │
│  Why? Middleware needs to understand domain types             │
│  (storing userbus.User in context, checking permissions)      │
│                                                                │
│  Example: mid.GetUser() returns userbus.User                  │
└────────────────────────────────┬───────────────────────────────┘
                                 ↓
                         Uses for HTTP concerns
                                 ↓
┌────────────────────────────────────────────────────────────────┐
│              business/sdk/ (Business Layer)                    │
│  ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━  │
│  Concern: "How do we handle data/business operations?"        │
│                                                                │
│  page/       → Pagination (page number, rows per page)        │
│  order/      → Ordering (field, direction ASC/DESC)           │
│  delegate/   → Event system (domain decoupling)               │
│  sqldb/      → Database utilities (queries, transactions)     │
│  migrate/    → Database migrations                            │
│                                                                │
│  Domain-Agnostic: ❌ CANNOT import business/domain/*          │
│  Why? Must be reusable by ALL domains without coupling        │
│                                                                │
│  Example: page.Parse() works for users, products, homes, etc. │
└────────────────────────────────────────────────────────────────┘
```

---

### Decision Matrix: Which SDK?

| Need | Use | Why |
|------|-----|-----|
| HTTP error codes | `app/sdk/errs` | Map domain errors to HTTP status codes |
| HTTP middleware | `app/sdk/mid` | Store domain types in context, auth checks |
| JSON response format | `app/sdk/query` | Wrap results with pagination metadata |
| Pagination logic | `business/sdk/page` | Generic - works for all domains |
| Ordering logic | `business/sdk/order` | Generic - works for all domains |
| Database queries | `business/sdk/sqldb` | Generic - works for all domains |
| Event delegation | `business/sdk/delegate` | Notify domains without direct imports |

---

### Real-World Example: Query Products with Pagination

**APP LAYER** - uses both SDKs appropriately:

```go
func (a *app) query(ctx context.Context, r *http.Request) web.Encoder {
    qp := parseQueryParams(r)  // Parse HTTP query params

    // Use BUSINESS SDK for pagination (domain-agnostic)
    page, err := page.Parse(qp.Page, qp.Rows)  // ← business/sdk/page
    if err != nil {
        // Use APP SDK for HTTP errors
        return errs.NewFieldErrors("page", err)  // ← app/sdk/errs (HTTP 400)
    }

    // Use BUSINESS SDK for ordering (domain-agnostic)
    orderBy, err := order.Parse(orderByFields, qp.OrderBy,
                                productbus.DefaultOrderBy)  // ← business/sdk/order
    if err != nil {
        return errs.NewFieldErrors("order", err)  // ← app/sdk/errs
    }

    // Call business layer
    prds, err := a.productBus.Query(ctx, filter, orderBy, page)
    if err != nil {
        // Use APP SDK for HTTP errors
        return errs.Newf(errs.Internal, "query: %s", err)  // ← app/sdk/errs (HTTP 500)
    }

    total, err := a.productBus.Count(ctx, filter)

    // Use APP SDK for HTTP response formatting
    return query.NewResult(toAppProducts(prds), total, page)  // ← app/sdk/query (JSON)
}
```

**STORE LAYER** - uses business SDK only:

```go
func (s *Store) Query(ctx context.Context, filter productbus.QueryFilter,
                      orderBy order.By, page page.Page) ([]productbus.Product, error) {

    // Use BUSINESS SDK for pagination calculations (domain-agnostic)
    data := map[string]any{
        "offset":        (page.Number() - 1) * page.RowsPerPage(),  // ← business/sdk/page
        "rows_per_page": page.RowsPerPage(),                        // ← business/sdk/page
    }

    // Use BUSINESS SDK for database queries (domain-agnostic)
    if err := sqldb.NamedQuerySlice(ctx, s.log, s.db, q, data, &dbPrds); err != nil {
        // ← business/sdk/sqldb
        return nil, fmt.Errorf("namedqueryslice: %w", err)
    }

    return toBusProducts(dbPrds)
}
```

---

### Summary Table

| Aspect | app/sdk/ | business/sdk/ |
|--------|----------|---------------|
| **Concern** | HTTP/Web operations | Data/Business operations |
| **Knowledge** | Domain-aware (knows userbus, productbus) | Domain-agnostic (generic utilities) |
| **Import Domains?** | ✅ YES - needs domain types for middleware | ❌ NO - must be reusable by all |
| **Purpose** | HTTP infrastructure for app layer | Business infrastructure for all domains |
| **Examples** | HTTP errors, middleware, auth, JSON format | Pagination, ordering, DB utils, events |
| **Used By** | App layer (HTTP handlers) | Business layer (all domains & stores) |
| **Typical Imports** | `business/domain/*`, `business/sdk/*` | `business/types/*`, `foundation/*` only |

**Key Insight**:
- **app/sdk** = "HTTP helpers that understand our business domains"
- **business/sdk** = "Business helpers that know nothing about specific domains"

**Not in SDK**: Domain-specific logic (lives in `business/domain/*`)

---

## 11. Query Pattern with Pagination

**App layer** (`productapp/productapp.go`):
```go
func (a *app) query(ctx context.Context, r *http.Request) web.Encoder {
    qp := parseQueryParams(r)

    page, err := page.Parse(qp.Page, qp.Rows)  // Parse pagination
    if err != nil {
        return errs.NewFieldErrors("page", err)
    }

    orderBy, err := order.Parse(orderByFields, qp.OrderBy, productbus.DefaultOrderBy)
    if err != nil {
        return errs.NewFieldErrors("order", err)
    }

    prds, err := a.productBus.Query(ctx, filter, orderBy, page)
    if err != nil {
        return errs.Newf(errs.Internal, "query: %s", err)
    }

    total, err := a.productBus.Count(ctx, filter)
    if err != nil {
        return errs.Newf(errs.Internal, "count: %s", err)
    }

    return query.NewResult(toAppProducts(prds), total, page)  // Wraps with metadata
}
```

**Response structure**:
```json
{
  "items": [...],
  "total": 100,
  "page": 1,
  "rowsPerPage": 10
}
```

---

## 12. Logging Architecture (Cross-Cutting Concern)

Logging is a **cross-cutting concern** that flows through all layers via dependency injection, built on Go's `log/slog` with structured JSON output.

### Logger Location and Foundation

**Foundation Package** ([foundation/logger/](foundation/logger/)):
```
foundation/logger/
├── logger.go     # Core wrapper around log/slog
├── model.go      # Log levels and record types
├── handler.go    # Custom handlers for events
└── debug.go      # Build info logging
```

**Key Characteristics**:
- Wraps Go's standard `log/slog` package
- JSON-structured output format
- Supports four levels: Debug, Info, Warn, Error
- Automatic trace ID injection from context
- Event handlers for log level callbacks
- Source location tracking (file, line, function)

---

### Logger Initialization (Composition Root)

**Created once in main.go** ([api/services/sales/main.go](api/services/sales/main.go)):

```go
func run(ctx context.Context) error {
    // Event handlers - react to specific log levels
    events := logger.Events{
        Error: func(ctx context.Context, r logger.Record) {
            log.Info(ctx, "******* SEND ALERT *******")
        },
    }

    // Trace ID extractor - pulls from context
    traceIDFn := func(ctx context.Context) string {
        return otel.GetTraceID(ctx)
    }

    // Single logger instance for entire application
    log := logger.NewWithEvents(
        os.Stdout,           // Output destination
        logger.LevelInfo,    // Minimum log level
        "SALES",             // Service name (added to every log)
        traceIDFn,           // Trace ID function
        events,              // Event handlers
    )

    // ...
}
```

**Every log entry automatically includes**:
- Timestamp (ISO 8601)
- Log level
- Service name (`"SALES"`)
- Trace ID (from context)
- Source file and line number
- Custom key-value pairs

---

### Logger Propagation via Dependency Injection

Logger flows through **constructor injection** at every layer:

**Main → Business → Store**:
```go
// main.go - Composition root
func run(ctx context.Context, log *logger.Logger) error {
    // Business layer receives logger
    userBus := userbus.NewBusiness(log, delegate, userStorage, ...)
    productBus := productbus.NewBusiness(log, userBus, delegate, productStore)
    homeBus := homebus.NewBusiness(log, userBus, delegate, homeStore)

    // Store layer receives logger
    userStore := userdb.NewStore(log, db)
    productStore := productdb.NewStore(log, db)

    // Cache layer receives logger (wraps store)
    userStorage := usercache.NewStore(log, userStore, time.Minute)

    // Mux receives logger for middleware
    mux.WebAPI(mux.Config{
        Log:    log,
        Tracer: tracer,
        DB:     db,
        // ...
    }, routeAdder)
}
```

**Pattern**: Single logger instance created at startup, injected through constructors into every layer.

---

### Logging by Layer

#### **App Layer - Heavy Logging** (HTTP Request Lifecycle)

**Request/Response Middleware** ([app/sdk/mid/logging.go](app/sdk/mid/logging.go)):

```go
func Logger(log *logger.Logger) web.MidFunc {
    m := func(next web.HandlerFunc) web.HandlerFunc {
        h := func(ctx context.Context, r *http.Request) web.Encoder {
            now := time.Now()

            path := r.URL.Path
            if r.URL.RawQuery != "" {
                path = fmt.Sprintf("%s?%s", path, r.URL.RawQuery)
            }

            // Log request start
            log.Info(ctx, "request started",
                "method", r.Method,
                "path", path,
                "remoteaddr", r.RemoteAddr)

            resp := next(ctx, r)

            // Determine status code
            var statusCode = errs.None
            if err := isError(resp); err != nil {
                statusCode = errs.Internal
                var v *errs.Error
                if errors.As(err, &v) {
                    statusCode = v.Code
                }
            }

            // Log request completion
            log.Info(ctx, "request completed",
                "method", r.Method,
                "path", path,
                "remoteaddr", r.RemoteAddr,
                "statuscode", statusCode,        // HTTP status
                "since", time.Since(now).String())  // Duration

            return resp
        }
        return h
    }
    return m
}
```

**Error Middleware** ([app/sdk/mid/errors.go](app/sdk/mid/errors.go)):

```go
func Errors(log *logger.Logger) web.MidFunc {
    m := func(next web.HandlerFunc) web.HandlerFunc {
        h := func(ctx context.Context, r *http.Request) web.Encoder {
            resp := next(ctx, r)
            err := isError(resp)
            if err == nil {
                return resp
            }

            _, span := otel.AddSpan(ctx, "app.sdk.mid.error")
            span.RecordError(err)
            defer span.End()

            var appErr *errs.Error
            if !errors.As(err, &appErr) {
                appErr = errs.Newf(errs.Internal, "Internal Server Error")
            }

            // Log error with source location
            log.Error(ctx, "handled error during request",
                "err", err,
                "source_err_file", path.Base(appErr.FileName),
                "source_err_func", path.Base(appErr.FuncName))

            if appErr.Code == errs.InternalOnlyLog {
                appErr = errs.Newf(errs.Internal, "Internal Server Error")
            }

            return appErr
        }
        return h
    }
    return m
}
```

**Middleware Stack Order** ([app/sdk/mux/mux.go](app/sdk/mux/mux.go)):
```go
app := web.NewApp(
    cfg.Log.Info,          // Log writer callback
    cfg.Tracer,
    mid.Otel(cfg.Tracer),  // 1. Inject trace ID into context
    mid.Logger(cfg.Log),   // 2. Log request start/completion
    mid.Errors(cfg.Log),   // 3. Log errors with source location
    mid.Metrics(),         // 4. Collect metrics
    mid.Panics(),          // 5. Recover panics
)
```

**App Layer Logs**:
- ✅ Every HTTP request (method, path, remote address)
- ✅ Every HTTP response (status code, duration)
- ✅ Every error (with source file and function)
- ✅ Panic recovery

---

#### **Business Layer - Minimal Logging** (Available but Rarely Used)

**Stores Logger, Focuses on Error Propagation** ([business/domain/userbus/userbus.go](business/domain/userbus/userbus.go)):

```go
type Business struct {
    log      *logger.Logger  // ← Logger available
    storer   Storer
    delegate *delegate.Delegate
}

func NewBusiness(log *logger.Logger, delegate *delegate.Delegate,
                 storer Storer, extensions ...Extension) ExtBusiness {
    b := ExtBusiness(&Business{
        log:      log,         // ← Stored for potential use
        delegate: delegate,
        storer:   storer,
    })
    // Extension wrapping...
    return b
}

// Business methods typically don't log - they propagate errors
func (b *Business) Create(ctx context.Context, nu NewUser) (User, error) {
    // Business logic - no logging
    // Errors propagate to app layer for logging
    if err := b.storer.Create(ctx, usr); err != nil {
        return User{}, fmt.Errorf("create: %w", err)
    }
    return usr, nil
}
```

**Business Layer Pattern**:
- ❌ Minimal direct logging
- ✅ Error propagation to app layer
- ✅ Logger stored for critical operations only
- **Why?** Business layer focuses on logic, not observability. App layer handles request/response logging.

---

#### **Store Layer - Database Operation Logging** (Via SDK Helpers)

**Store stores logger, passes to sqldb helpers** ([business/domain/userbus/stores/userdb/userdb.go](business/domain/userbus/stores/userdb/userdb.go)):

```go
type Store struct {
    log *logger.Logger  // ← Logger for DB operations
    db  sqlx.ExtContext
}

func NewStore(log *logger.Logger, db *sqlx.DB) *Store {
    return &Store{
        log: log,
        db:  db,
    }
}

func (s *Store) Create(ctx context.Context, usr userbus.User) error {
    const q = `
    INSERT INTO users
        (user_id, name, email, password_hash, roles, department, enabled, date_created, date_updated)
    VALUES
        (:user_id, :name, :email, :password_hash, :roles, :department, :enabled, :date_created, :date_updated)`

    // Pass logger to sqldb helper
    if err := sqldb.NamedExecContext(ctx, s.log, s.db, q, toDBUser(usr)); err != nil {
        if errors.Is(err, sqldb.ErrDBDuplicatedEntry) {
            return fmt.Errorf("namedexeccontext: %w", userbus.ErrUniqueEmail)
        }
        return fmt.Errorf("namedexeccontext: %w", err)
    }

    return nil
}
```

**Actual logging happens in business/sdk/sqldb** ([business/sdk/sqldb/sqldb.go](business/sdk/sqldb/sqldb.go)):

```go
func NamedExecContext(ctx context.Context, log *logger.Logger,
                      db sqlx.ExtContext, query string, data any) (err error) {

    // Build query string with interpolated parameters for debugging
    q := queryString(query, data)

    // Defer logging - only logs on error
    defer func() {
        if err != nil {
            switch data.(type) {
            case struct{}:
                log.Infoc(ctx, 6, "database.NamedExecContext",
                    "query", q, "ERROR", err)
            default:
                log.Infoc(ctx, 5, "database.NamedExecContext",
                    "query", q, "ERROR", err)
            }
        }
    }()

    ctx, span := otel.AddSpan(ctx, "business.sdk.sqldb.exec",
        attribute.String("query", q))
    defer span.End()

    if _, err := sqlx.NamedExecContext(ctx, db, query, data); err != nil {
        if errors.Is(err, sql.ErrNoRows) {
            return ErrDBNotFound
        }
        if pqerr, ok := err.(*pq.Error); ok {
            if pqerr.Code == undefinedTable {
                return ErrDBNotFound
            }
            if pqerr.Code == uniqueViolation {
                return ErrDBDuplicatedEntry
            }
        }
        return err
    }

    return nil
}
```

**Store Layer Pattern**:
- ✅ Logs database queries **on error only**
- ✅ Includes actual SQL with interpolated parameters
- ✅ Adds OpenTelemetry spans for tracing
- ✅ Uses `Infoc(ctx, caller, ...)` to control call stack depth for accurate source location
- **Note**: `queryString()` interpolates parameters for debugging (never sent to database)

---

### Context and Trace ID Propagation

**Trace ID Injection** ([foundation/otel/otel.go](foundation/otel/otel.go)):

```go
func InjectTracing(ctx context.Context, tracer trace.Tracer) context.Context {
    ctx = setTracer(ctx, tracer)

    // Extract or generate trace ID
    traceID := trace.SpanFromContext(ctx).SpanContext().TraceID().String()
    if traceID == defaultTraceID {
        traceID = uuid.NewString()  // Generate if not present
    }
    ctx = setTraceID(ctx, traceID)

    return ctx
}

func GetTraceID(ctx context.Context) string {
    v, ok := ctx.Value(traceIDKey).(string)
    if !ok {
        return defaultTraceID
    }
    return v
}
```

**Logger extracts trace ID automatically** ([foundation/logger/logger.go](foundation/logger/logger.go)):

```go
func (log *Logger) write(ctx context.Context, level Level, caller int,
                         msg string, args ...any) {
    slogLevel := slog.Level(level)

    if !log.handler.Enabled(ctx, slogLevel) {
        return
    }

    var pcs [1]uintptr
    runtime.Callers(caller, pcs[:])

    r := slog.NewRecord(time.Now(), slogLevel, msg, pcs[0])

    // Automatically add trace ID to all logs
    if log.traceIDFn != nil {
        args = append(args, "trace_id", log.traceIDFn(ctx))
    }
    r.Add(args...)

    log.handler.Handle(ctx, r)
}
```

**Pattern**:
- Context flows through all layers
- Trace ID stored in context (injected by Otel middleware)
- Logger automatically extracts trace ID via `traceIDFn`
- **Every log entry** for a request shares the same trace ID

---

### Structured Logging Convention

**Key-Value Pairs**:
```go
log.Info(ctx, "message", "key1", value1, "key2", value2)
```

**Example Output** (JSON):
```json
{
  "time": "2025-11-02T10:15:30.123Z",
  "level": "INFO",
  "source": "productapp/productapp.go:145",
  "msg": "request completed",
  "service": "SALES",
  "trace_id": "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
  "method": "GET",
  "path": "/v1/products",
  "remoteaddr": "192.168.1.100:54321",
  "statuscode": 200,
  "since": "45.2ms"
}
```

**Special Methods**:

```go
// Standard logging
log.Info(ctx, "message", "key", value)
log.Error(ctx, "error occurred", "err", err)

// Control call stack depth (for accurate source location in wrappers)
log.Infoc(ctx, 5, "message", "key", value)

// Log build metadata at startup
log.BuildInfo(ctx)
```

---

### Log Levels and Usage Patterns

| Level | Usage | Example |
|-------|-------|---------|
| **Debug** | Development debugging | Detailed state inspection |
| **Info** | Normal operations | Request/response logs, lifecycle events |
| **Warn** | Recoverable issues | Deprecated API usage, fallback triggered |
| **Error** | Failures requiring attention | Database errors, validation failures |

**Actual usage in codebase**:
- **Info**: HTTP requests, responses, successful operations
- **Error**: Handled errors during request processing
- **Debug/Warn**: Rarely used (most logging is Info or Error)

---

### Cache Layer Example

**Wraps store, delegates logging** ([business/domain/userbus/stores/usercache/usercache.go](business/domain/userbus/stores/usercache/usercache.go)):

```go
type Store struct {
    log    *logger.Logger    // ← Receives logger
    storer userbus.Storer    // ← Wraps store layer
    cache  *sturdyc.Client[userbus.User]
}

func NewStore(log *logger.Logger, storer userbus.Storer, ttl time.Duration) *Store {
    return &Store{
        log:    log,
        storer: storer,
        cache:  sturdyc.New[userbus.User](capacity, numShards, ttl, evictionPercentage),
    }
}

func (s *Store) QueryByID(ctx context.Context, userID uuid.UUID) (userbus.User, error) {
    // Cache hit/miss - no logging (could add if needed)
    usr, err := s.cache.GetOrFetch(ctx, userID.String(),
        func(ctx context.Context) (userbus.User, error) {
            return s.storer.QueryByID(ctx, userID)  // ← Delegates to store (which logs DB errors)
        },
    )
    if err != nil {
        return userbus.User{}, fmt.Errorf("getorfetch: %w", err)
    }
    return usr, nil
}
```

**Cache Layer Pattern**:
- ✅ Receives logger for potential use
- ✅ Delegates to underlying store (which handles DB logging)
- ✅ Could add cache-specific logging if needed

---

### Summary: Logging Architecture Patterns

**Foundation**:
- Custom wrapper around `log/slog`
- JSON-structured output
- Automatic trace ID injection
- Event handlers for log levels

**Propagation**:
- Single logger created at startup
- Dependency injection through constructors
- Available in all layers

**Usage by Layer**:
- **App Layer**: Heavy logging (requests, responses, errors)
- **Business Layer**: Minimal logging (stored but rarely used)
- **Store Layer**: Database logging on errors (via sqldb helpers)

**Context Flow**:
- Trace ID in context → automatically in all logs
- Correlate all logs for a single request

**Conventions**:
- Structured key-value pairs
- Every log includes: timestamp, level, service, trace ID, source location
- Log at boundaries (HTTP, database), not in business logic

**Key Insight**: Logging is a **cross-cutting concern** that provides **observability at boundaries** (HTTP requests, database operations) while keeping business logic clean and focused. The logger is **available everywhere** but **used sparingly** where it matters most.

---

## Summary: The Golden Rules

1. ✅ **Validate in App Layer** - via Parse functions in `toBusNew*` conversions
2. ✅ **Value Objects** - typed primitives with validation (business/types/*)
3. ✅ **Zero Validation in Business** - Create() assumes valid inputs
4. ✅ **Three-Layer Isolation** - strict boundaries (app/business/store)
5. ✅ **Dependency Inversion** - business defines interface, store implements it, main wires
6. ✅ **Import Direction** - imports flow toward abstractions (store → business, never business → store)
7. ✅ **Parse Pattern** - consistent `Parse(primitive) → (Type, error)` signature
8. ✅ **Error Conversion** - each layer translates errors to its abstraction
9. ✅ **Sentinel Errors** - only in business layer (var Err* = errors.New)
10. ✅ **Conversion Functions** - at every layer boundary (toApp*, toBus*, toDB*)
11. ✅ **Order Mapping** - three-layer indirection (API → const → DB)
12. ✅ **Delegate Pattern** - event-based decoupling between domains (prevents circular imports)
13. ✅ **Logging at Boundaries** - app layer (HTTP), store layer (DB errors), minimal in business; trace ID in context automatically added to all logs

---

## The Core Insights

### 1. Validation is a Boundary Concern

**Validation is a boundary concern, not a business concern.**

The business layer should assume it receives **valid, typed values** and focus purely on business logic. All validation happens at the **app layer boundary** via Parse functions that convert primitives to typed value objects.

This creates:
- **Type safety** throughout the business layer
- **Single source of truth** for validation rules
- **Clear separation of concerns** across layers
- **Impossible states** at compile time (can't create invalid Name, Money, etc.)

### 2. Dependencies Point Inward (Clean Architecture)

**Source code dependencies must point inward toward abstractions.**

The business layer defines **what it needs** (interfaces), not **how it's implemented**. Infrastructure (stores) depends on business abstractions, not the other way around. The composition root (main.go) wires concrete implementations into abstract interfaces.

This creates:
- **Testability** - easily mock interfaces for unit tests
- **Flexibility** - swap implementations without changing business logic
- **No circular dependencies** - imports flow one direction toward abstractions
- **Physical nesting ≠ logical dependency** - store nested in folder but imports parent

### Complete Architecture Diagram

```
┌─────────────────────────────────────────────────────────────────────┐
│                         HTTP Request (JSON)                         │
└────────────────────────────────┬────────────────────────────────────┘
                                 ↓
┌─────────────────────────────────────────────────────────────────────┐
│  APP LAYER (app/domain/productapp/)                                 │
│  ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━  │
│  HTTP Handlers + Validation + Conversion                            │
│                                                                     │
│  1. Decode JSON → primitives (string, float64, int)                │
│  2. Validate via Parse() → typed values                            │
│  3. toBusNew*() → convert to business types                        │
│  4. Call business layer                                            │
│  5. toApp*() → convert back to JSON                                │
│                                                                     │
│  Imports: business/domain/*, app/sdk/*, business/sdk/*             │
└────────────────────────────────┬────────────────────────────────────┘
                                 ↓
┌─────────────────────────────────────────────────────────────────────┐
│  APP SDK (app/sdk/)                                                 │
│  ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━  │
│  HTTP Middleware: errs, mid, query                                 │
│  Domain-aware utilities (can import business domains)              │
│                                                                     │
│  Imports: business/domain/*, business/sdk/*, foundation/*          │
└────────────────────────────────┬────────────────────────────────────┘
                                 ↓
┌─────────────────────────────────────────────────────────────────────┐
│  BUSINESS LAYER (business/domain/productbus/)                       │
│  ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━  │
│  Business Logic ONLY (zero validation)                             │
│                                                                     │
│  - Defines Storer INTERFACE (abstraction)                          │
│  - Works with typed value objects (Name, Money, Quantity)          │
│  - Contains business rules and orchestration                       │
│  - NEVER imports own store implementation                          │
│                                                                     │
│  type Storer interface {                                           │
│      Create(ctx, Product) error  ← What we need                    │
│  }                                                                  │
│                                                                     │
│  Imports: business/domain/* (one-way), business/sdk/*,             │
│           business/types/*, foundation/*                            │
│  ❌ NEVER: business/domain/productbus/stores/productdb              │
└────────────────────────────────┬────────────────────────────────────┘
                                 ↓ (interface only)
                          ┌──────┴────────┐
                          │  Storer       │ ← Interface contract
                          │  interface    │
                          └──────┬────────┘
                                 ↑ (implements)
┌─────────────────────────────────────────────────────────────────────┐
│  STORE LAYER (business/domain/productbus/stores/productdb/)         │
│  ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━  │
│  Data Persistence (implements Storer interface)                    │
│                                                                     │
│  - Implements productbus.Storer interface                          │
│  - IMPORTS productbus (for interface & types)                      │
│  - SQL operations and database logic                               │
│  - Converts between DB models and business types                   │
│                                                                     │
│  func (s *Store) Create(ctx, productbus.Product) error {           │
│      // SQL operations... ← How we do it                           │
│  }                                                                  │
│                                                                     │
│  Imports: business/domain/productbus (parent ONLY),                │
│           business/sdk/sqldb, business/types/*, foundation/*       │
│  ❌ NEVER: other business domains, app/*                            │
└────────────────────────────────┬────────────────────────────────────┘
                                 ↓
┌─────────────────────────────────────────────────────────────────────┐
│  BUSINESS SDK (business/sdk/)                                      │
│  ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━  │
│  Shared Utilities: order, page, delegate, sqldb, migrate           │
│  Pure utilities (NO domain imports)                                │
│                                                                     │
│  Imports: business/types/*, foundation/* ONLY                      │
└────────────────────────────────┬────────────────────────────────────┘
                                 ↓
┌─────────────────────────────────────────────────────────────────────┐
│  BUSINESS TYPES (business/types/)                                  │
│  ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━  │
│  Value Objects: Name, Money, Quantity, Role, HomeType              │
│  Primitive wrappers with validation                                │
│                                                                     │
│  Imports: Standard library ONLY                                    │
└────────────────────────────────┬────────────────────────────────────┘
                                 ↓
┌─────────────────────────────────────────────────────────────────────┐
│  FOUNDATION (foundation/)                                           │
│  ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━  │
│  Base Infrastructure: logger, web, otel, docker, keystore          │
│                                                                     │
│  Imports: Standard library, third-party packages ONLY              │
│  ❌ NEVER: app/*, business/*                                        │
└─────────────────────────────────────────────────────────────────────┘

                         DEPENDENCY WIRING (main.go)

                    productStore := productdb.NewStore(log, db)
                              ↓ (concrete implementation)
                    productBus := productbus.NewBusiness(..., productStore)
                              ↓ (injected as Storer interface)
                    Business only sees interface, not concrete type
```

### Import Flow Rules (Quick Reference)

| From ↓ / To → | Foundation | Types | Bus SDK | Store | Business | App SDK | App |
|---------------|------------|-------|---------|-------|----------|---------|-----|
| **App**       | ✅         | ✅    | ✅      | ❌    | ✅       | ✅      | ❌  |
| **App SDK**   | ✅         | ✅    | ✅      | ❌    | ✅       | ❌      | ❌  |
| **Business**  | ✅         | ✅    | ✅      | ❌    | ✅ (1-way)| ❌     | ❌  |
| **Store**     | ✅         | ✅    | ✅      | ❌    | ✅ (parent)| ❌    | ❌  |
| **Bus SDK**   | ✅         | ✅    | ❌      | ❌    | ❌       | ❌      | ❌  |
| **Types**     | ❌         | ❌    | ❌      | ❌    | ❌       | ❌      | ❌  |
| **Foundation**| ❌         | ❌    | ❌      | ❌    | ❌       | ❌      | ❌  |

**Key**: ✅ Can import | ❌ Cannot import | 1-way = only one direction allowed

This architecture ensures:
- **Validation at boundaries** - Parse functions convert primitives to types
- **Type safety in business** - impossible to have invalid values
- **Testability** - interfaces enable mocking
- **No circular dependencies** - imports flow one direction
- **Flexibility** - swap implementations without changing logic
