# Business Model Import Dependencies Architecture

## Table of Contents
1. [Overview](#overview)
2. [Domain Model Structure](#domain-model-structure)
3. [Dependency Hierarchy](#dependency-hierarchy)
4. [One-Directional Import Enforcement](#one-directional-import-enforcement)
5. [Asynchronous Delete Pattern](#asynchronous-delete-pattern)
6. [Cross-Model Relationships](#cross-model-relationships)
7. [Database Views for Multi-Model Queries](#database-views-for-multi-model-queries)
8. [Architectural Patterns](#architectural-patterns)
9. [Best Practices](#best-practices)

---

## Overview

This architecture implements a **Domain-Driven Design (DDD)** approach with strict **one-directional dependency** rules between business models. The architecture prevents circular dependencies and maintains clean separation of concerns through:

- **Dependency Injection** at construction time
- **Interface-based contracts** between domains
- **Event-driven cascade operations** via the Delegate pattern
- **Database views** for cross-domain read operations
- **Extension system** for composable behaviors (audit, tracing, caching)

---

## Domain Model Structure

### Domain Types

| Domain Type | Purpose | Dependency Rules | Example Use Cases |
|-------------|---------|------------------|-------------------|
| **Root Domain** | Core entities that stand alone | No business domain dependencies | User, Account, Organization |
| **Child Domain** | Entities owned by root domains | Can import parent domain interfaces only | Products owned by users, Posts by authors |
| **Support Domain** | Cross-cutting concerns | No business domain dependencies | Audit logs, Notifications, Analytics |
| **Query Domain** | Read-only aggregated views | No business domain dependencies | Dashboard data, Reports, Joined views |

### Domain Hierarchy Example

```
userbus (root)
  ├── productbus (child - products owned by users)
  ├── orderbus (child - orders owned by users)
  └── [support domains - isolated]

[query domains - isolated views]
```

---

## Dependency Hierarchy

### Dependency Graph

```
                    [Support Domains]
                    (isolated, no imports)

           rootbus (e.g., userbus)
              ▲
              │
              │ import parent interface
              │
         childbus (e.g., productbus)
              ▲
              │
              │ import parent interface
              │
      grandchildbus (e.g., orderitembus)


    [Query Domains - isolated islands]
    (read from database views only)
```

### Import Rules

#### Rule 1: Child Imports Parent Only
**Child domains** can import **parent domain** interfaces, but **never the reverse**.

**Example:**
```go
package productbus

import (
    "yourapp/business/domain/userbus"
)

type Business struct {
    userBus userbus.ExtBusiness  // Child imports parent interface
    storer  Storer
}

func NewBusiness(userBus userbus.ExtBusiness, storer Storer) ExtBusiness {
    return &Business{
        userBus: userBus,
        storer:  storer,
    }
}
```

**Anti-pattern (DO NOT DO THIS):**
```go
package userbus

import (
    "yourapp/business/domain/productbus"  // ❌ Parent importing child!
)

type Business struct {
    productBus productbus.ExtBusiness  // ❌ Creates circular dependency
}
```

#### Rule 2: Support Domains Have No Dependencies
**Support domains** provide cross-cutting functionality and have **no dependencies** on other business domains, ensuring they can be used universally across the application.

**What are Support Domains?**

Support domains handle concerns that span multiple business domains without being part of any specific domain's core logic:

- **Audit Logging**: Track who did what, when, and why across all domains
- **Notifications**: Send emails, SMS, push notifications triggered by any domain
- **Analytics/Metrics**: Collect usage statistics and performance metrics
- **Search Indexing**: Maintain search indices for entities from multiple domains
- **File Storage**: Handle file uploads/downloads for any domain
- **Background Jobs**: Queue and process async tasks from any domain

**Example Structure:**
```go
package auditbus

import (
    // Only SDK types and standard library
    "context"
    "time"
    "yourapp/business/sdk/order"
    "yourapp/business/sdk/page"
    // ❌ NO imports from userbus, productbus, or any business domain
)

type Audit struct {
    ID        uuid.UUID
    Domain    string      // Which domain: "user", "product", etc.
    Action    string      // What happened: "created", "updated", "deleted"
    ActorID   uuid.UUID   // Who did it
    ObjectID  uuid.UUID   // What was affected
    Data      json.RawMessage
    Timestamp time.Time
}

type Business struct {
    storer Storer
    // No references to other business domains
}

func NewBusiness(storer Storer) ExtBusiness {
    return &Business{storer: storer}
}

// Generic method that works for any domain
func (b *Business) Create(ctx context.Context, audit Audit) (Audit, error) {
    // No domain-specific logic
    return b.storer.Create(ctx, audit)
}
```

**Why Keep Support Domains Isolated?**

1. **Universal Reusability**: Any domain can use support domains without creating dependencies
2. **Easy Testing**: Support domains can be tested independently
3. **Flexible Deployment**: Can be deployed as separate services if needed
4. **Clear Boundaries**: Support logic stays separate from business logic

**How to Connect Support Domains:**

Use the **Extension Pattern** (see below) to bridge support domains with business domains:

```go
// Extension bridges userbus with auditbus WITHOUT userbus importing auditbus
package useraudit

import (
    "yourapp/business/domain/userbus"
    "yourapp/business/domain/auditbus"
)

type Extension struct {
    bus      userbus.ExtBusiness
    auditBus auditbus.ExtBusiness  // Support domain injected here
}

func NewExtension(auditBus auditbus.ExtBusiness) userbus.Extension {
    return func(b userbus.ExtBusiness) userbus.ExtBusiness {
        return &Extension{bus: b, auditBus: auditBus}
    }
}

func (ext *Extension) Create(ctx context.Context, nu userbus.NewUser) (userbus.User, error) {
    // Call core business logic
    user, err := ext.bus.Create(ctx, nu)
    if err != nil {
        return userbus.User{}, err
    }

    // Add audit side-effect
    audit := auditbus.Audit{
        Domain:   "user",
        Action:   "created",
        ObjectID: user.ID,
        Data:     marshal(nu),
    }
    ext.auditBus.Create(ctx, audit)

    return user, nil
}
```

#### Rule 3: Query Domains Are Read-Only Islands
**Query domains** are specialized read-only domains that provide efficient access to data spanning multiple business domains through database views. They have **no dependencies** on other business domains.

**What are Query Domains?**

Query domains solve the problem of efficiently retrieving related data from multiple domains without creating import dependencies or making N+1 database queries.

**When to Use Query Domains:**

- **Dashboard/Reporting**: Displaying aggregated data from multiple sources
- **List Views**: Showing entities with related entity data (e.g., products with owner names)
- **Search Results**: Combining data from multiple tables for search functionality
- **API Responses**: Returning denormalized data optimized for client consumption

**Example Problem:**

You want to display a product list with owner information:
```
Product: Laptop ($999)
Owner: John Doe (john@example.com)
```

This requires data from two domains: `productbus` and `userbus`.

**Bad Solutions:**

```go
// ❌ BAD: Import userbus into productbus (creates coupling)
package productbus
import "userbus"

type Product struct {
    ID   uuid.UUID
    Name string
    User userbus.User  // ❌ Tight coupling
}

// ❌ BAD: N+1 queries
products, _ := productBus.Query(ctx, filter)
for _, product := range products {
    user, _ := userBus.QueryByID(ctx, product.UserID)  // N additional queries!
}
```

**Good Solution: Query Domain Pattern**

**Step 1: Create Database View**
```sql
-- Database handles the JOIN efficiently
CREATE OR REPLACE VIEW view_products AS
SELECT
    p.product_id,
    p.name,
    p.price,
    p.user_id,
    u.name AS user_name,
    u.email AS user_email
FROM
    products AS p
JOIN
    users AS u ON u.user_id = p.user_id;
```

**Step 2: Create Query Domain**
```go
package vproductbus  // 'v' prefix indicates view/query domain

import (
    // Only SDK types and primitives
    "context"
    "yourapp/business/sdk/order"
    "yourapp/business/sdk/page"
    // ❌ NO imports from userbus or productbus
)

// Model includes fields from multiple source domains
type Product struct {
    ID        uuid.UUID
    Name      string
    Price     float64
    UserID    uuid.UUID
    UserName  string   // From users table via view
    UserEmail string   // From users table via view
}

// Only read operations
type Storer interface {
    Query(ctx context.Context, filter QueryFilter, orderBy order.By, page page.Page) ([]Product, error)
    Count(ctx context.Context, filter QueryFilter) (int, error)
    // ❌ NO Create, Update, Delete methods
}

type Business struct {
    storer Storer  // Only needs view access
}

func NewBusiness(storer Storer) ExtBusiness {
    return &Business{storer: storer}
}

// Read-only operations
func (b *Business) Query(ctx context.Context, filter QueryFilter, orderBy order.By, page page.Page) ([]Product, error) {
    return b.storer.Query(ctx, filter, orderBy, page)
}
```

**Step 3: Use in API Layer**
```go
// For writes - use core domains
newProduct := productbus.NewProduct{
    UserID: userID,
    Name:   "Laptop",
    Price:  999.99,
}
product, _ := productBus.Create(ctx, newProduct)

// For reads with joins - use query domain
filter := vproductbus.QueryFilter{Name: "Laptop"}
products, _ := vproductBus.Query(ctx, filter, orderBy, page)

// products already contains user data from the view!
for _, p := range products {
    fmt.Printf("%s - Owner: %s\n", p.Name, p.UserName)
}
```

**Benefits of Query Domains:**

1. **No Import Dependencies**: Query domains don't import business domains
2. **Performance**: Single database query with JOIN instead of N+1 queries
3. **Separation of Concerns**: Writes through core domains, reads through query domains
4. **Optimized Models**: Query models shaped exactly for UI/API needs
5. **Database Optimization**: Views can be indexed, materialized, or cached

**Comparison: Core Domain vs Query Domain**

| Aspect | Core Domain | Query Domain |
|--------|-------------|--------------|
| **Operations** | Create, Update, Delete, Query | Query, Count only |
| **Dependencies** | May import parent domains | No domain imports |
| **Data Source** | Base tables | Database views |
| **Model** | Minimal fields | Extended/denormalized |
| **Purpose** | Business logic, writes | Efficient reads |
| **Validation** | Full business rules | None (read-only) |
| **Transactions** | Supports transactions | Not needed |

---

## One-Directional Import Enforcement

### Enforcement Mechanism 1: Dependency Injection

Dependencies are injected **only at construction time**, establishing a compile-time enforced hierarchy.

```go
// Step 1: Create root domain with NO business dependencies
userBus := userbus.NewBusiness(log, delegate, userStore)

// Step 2: Create child domain, injecting parent
productBus := productbus.NewBusiness(log, userBus, delegate, productStore)
//                                         ^^^^^^^ Unidirectional dependency

// Step 3: Create grandchild domain, injecting parent
orderBus := orderbus.NewBusiness(log, productBus, delegate, orderStore)
//                                     ^^^^^^^^^^ Unidirectional dependency

// Step 4: Create support domain with NO domain dependencies
auditBus := auditbus.NewBusiness(auditStore)

// Step 5: Create query domain with NO domain dependencies
vproductBus := vproductbus.NewBusiness(vproductStore)
```

**Key Points:**
- Root domains are created **first** with no business dependencies
- Child domains **receive** parent as constructor parameter
- Parents **never receive** children
- Compile-time error if you try to reverse the dependency

### Enforcement Mechanism 2: Interface Contracts

Child domains accept **interfaces**, not concrete types, preventing reverse casting or circular references.

```go
type Business struct {
    log     *logger.Logger
    userBus userbus.ExtBusiness    // ← Interface, not *userbus.Business
    storer  Storer
}

func NewBusiness(
    log *logger.Logger,
    userBus userbus.ExtBusiness,  // ← Interface enforces contract
    storer Storer,
) ExtBusiness {
    // Implementation can only use interface methods
    // Cannot cast back to concrete type or access internals
    return &Business{
        log:     log,
        userBus: userBus,
        storer:  storer,
    }
}
```

**Benefits:**
- Child domains cannot access internal implementation details
- Prevents reverse dependencies through casting
- Enables mocking and testing
- Clear contract boundaries

---

## Asynchronous Delete Pattern

### The Problem

When a parent entity is deleted, **all related child entities** must also be deleted. However, we cannot:
- Have parent domain import child domain (violates one-directional rule)
- Use database cascade delete alone (loses business logic, audit trails)

### The Solution: Delegate Pattern

The **Delegate** is an event bus that allows domains to:
1. **Publish events** when important actions occur
2. **Subscribe to events** from other domains
3. **Execute business logic** in response to events

```go
package delegate

type Delegate struct {
    log   *logger.Logger
    funcs map[domain]map[action][]Func
}

// Register a handler for a domain's action
func (d *Delegate) Register(domain string, action string, fn Func) {
    // Multiple handlers can register for same event
}

// Call all handlers for an event
func (d *Delegate) Call(ctx context.Context, data Data) error {
    if dMap, ok := d.funcs[domain(data.Domain)]; ok {
        if funcs, ok := dMap[action(data.Action)]; ok {
            for _, fn := range funcs {
                if err := fn(ctx, data); err != nil {
                    return fmt.Errorf("delegate: %w", err)
                }
            }
        }
    }
    return nil
}
```

### Implementation Steps

**Step 1: Define Event in Parent Domain**

```go
package userbus

const DomainName = "user"
const ActionDeleted = "deleted"

type ActionDeletedParms struct {
    UserID uuid.UUID
}

func ActionDeletedData(userID uuid.UUID) delegate.Data {
    params := ActionDeletedParms{UserID: userID}
    rawParams, _ := json.Marshal(params)

    return delegate.Data{
        Domain:    DomainName,
        Action:    ActionDeleted,
        RawParams: rawParams,
    }
}
```

**Step 2: Parent Publishes Event**

```go
package userbus

func (b *Business) Delete(ctx context.Context, user User) error {
    // 1. Delete from database
    if err := b.storer.Delete(ctx, user); err != nil {
        return err
    }

    // 2. Publish event for other domains
    if err := b.delegate.Call(ctx, ActionDeletedData(user.ID)); err != nil {
        return fmt.Errorf("failed to execute `%s` action: %w", ActionDeleted, err)
    }

    return nil
}
```

**Step 3: Child Domain Subscribes**

```go
package productbus

func (b *Business) registerDelegateFunctions() {
    if b.delegate != nil {
        b.delegate.Register(
            userbus.DomainName,
            userbus.ActionDeleted,
            b.actionUserDeleted,
        )
    }
}

func (b *Business) actionUserDeleted(ctx context.Context, data delegate.Data) error {
    var params userbus.ActionDeletedParms
    if err := json.Unmarshal(data.RawParams, &params); err != nil {
        return err
    }

    // Apply business logic
    return b.storer.DeleteByUserID(ctx, params.UserID)
}
```

### Flow Diagram

```
┌─────────────────────────────────────────────────────────┐
│                 Delete Cascade Flow                      │
└─────────────────────────────────────────────────────────┘

    API Request          userbus              Delegate
        │                   │                    │
        │  Delete(user)     │                    │
        ├──────────────────>│                    │
        │                   │  DELETE FROM users │
        │                   ├───────────────────>│
        │                   │                    │
        │                   │  Call("user.deleted")
        │                   ├───────────────────>│
        │                   │                    │
        │                   │                    ├─┐ Notify
        │                   │                    │<┘ subscribers
        │                   │                    │
        ▼                   ▼                    ▼

                      productbus          orderbus
                          │                  │
                 actionUserDeleted()  actionUserDeleted()
                          │                  │
                 DELETE products      DELETE orders
                 WHERE user_id=?      WHERE user_id=?
```

### Benefits

1. **No Circular Dependencies**: Parent doesn't import children
2. **Loose Coupling**: Domains communicate through events
3. **Extensible**: New domains can subscribe without modifying existing code
4. **Business Logic Preserved**: Can audit, validate before deletion
5. **Testable**: Can mock delegate in tests

---

## Cross-Model Relationships

### Pattern 1: Parent Validation During Child Creation

When creating a child entity, validate the parent exists and satisfies business rules.

```go
package productbus

func (b *Business) Create(ctx context.Context, np NewProduct) (Product, error) {
    // 1. Query parent domain for validation
    user, err := b.userBus.QueryByID(ctx, np.UserID)
    if err != nil {
        return Product{}, fmt.Errorf("user.querybyid: %w", err)
    }

    // 2. Enforce business rule
    if !user.Enabled {
        return Product{}, ErrUserDisabled
    }

    // 3. Create product if validation passes
    product := Product{
        ID:          uuid.New(),
        UserID:      np.UserID,
        Name:        np.Name,
        Price:       np.Price,
        DateCreated: time.Now(),
    }

    return b.storer.Create(ctx, product)
}
```

**Business Rules Enforced:**
- Parent must exist (referential integrity)
- Parent must satisfy conditions (e.g., enabled)
- Returns specific business errors

### Pattern 2: Query by Parent ID

```go
// Business layer
func (b *Business) QueryByUserID(ctx context.Context, userID uuid.UUID) ([]Product, error) {
    return b.storer.QueryByUserID(ctx, userID)
}

// Store interface
type Storer interface {
    QueryByUserID(ctx context.Context, userID uuid.UUID) ([]Product, error)
}

// Database implementation
// SELECT * FROM products WHERE user_id = $1
```

### Pattern 3: Transactional Cross-Domain Operations

When operations span multiple domains, use transactions.

```go
// Each domain supports transactions
func (b *Business) NewWithTx(tx sqldb.CommitRollbacker) (ExtBusiness, error) {
    // Create transactional store
    storer, err := b.storer.NewWithTx(tx)
    if err != nil {
        return nil, err
    }

    // Wrap parent domain with same transaction
    userBus, err := b.userBus.NewWithTx(tx)
    if err != nil {
        return nil, err
    }

    // Return new business layer with transactional dependencies
    return NewBusiness(b.log, userBus, storer), nil
}
```

**Usage:**
```go
// Begin transaction
tx, _ := db.Begin()

// Create transactional business layers
userBusTx, _ := userBus.NewWithTx(tx)
productBusTx, _ := productBus.NewWithTx(tx)

// Operations on both domains use same transaction
user, _ := userBusTx.Create(ctx, newUser)
product, _ := productBusTx.Create(ctx, newProduct)

// Commit or rollback together
tx.Commit()  // Both succeed or both fail
```

---

## Database Views for Multi-Model Queries

### Creating Database Views

```sql
-- Join products with user data
CREATE OR REPLACE VIEW view_products AS
SELECT
    p.product_id,
    p.user_id,
    p.name AS product_name,
    p.price,
    p.quantity,
    p.date_created,
    u.name AS user_name,
    u.email AS user_email
FROM
    products AS p
JOIN
    users AS u ON u.user_id = p.user_id;
```

### Query Domain Implementation

```go
package vproductbus

// Model with fields from multiple sources
type Product struct {
    ID          uuid.UUID
    UserID      uuid.UUID
    ProductName string
    Price       float64
    Quantity    int
    DateCreated time.Time
    UserName    string    // From users table
    UserEmail   string    // From users table
}

// Query the view
func (s *Store) Query(ctx context.Context, filter QueryFilter) ([]Product, error) {
    const q = `
    SELECT
        product_id,
        user_id,
        product_name,
        price,
        quantity,
        date_created,
        user_name,
        user_email
    FROM
        view_products`

    // Execute single query with all joined data
    rows, err := s.db.QueryContext(ctx, q)
    // ... scan results
}
```

### Benefits

1. **No Domain Dependencies**: Query domain doesn't import other domains
2. **Performance**: Single query instead of N+1
3. **Separation**: Writes through core domains, reads through query domains
4. **Database Optimization**: Views can be indexed or materialized

---

## Architectural Patterns

### Pattern 1: Extension/Middleware System

Extensions allow composable behaviors without modifying core logic.

```go
// Extension type
type Extension func(ExtBusiness) ExtBusiness

// Constructor applies extensions
func NewBusiness(log *logger.Logger, storer Storer, extensions ...Extension) ExtBusiness {
    b := ExtBusiness(&Business{
        log:    log,
        storer: storer,
    })

    // Apply extensions in reverse (last applied first)
    for i := len(extensions) - 1; i >= 0; i-- {
        if extensions[i] != nil {
            b = extensions[i](b)
        }
    }

    return b
}
```

**Example Extension:**
```go
// Audit extension
type AuditExtension struct {
    bus      userbus.ExtBusiness
    auditBus auditbus.ExtBusiness
}

func (ext *AuditExtension) Create(ctx context.Context, nu NewUser) (User, error) {
    // Call wrapped layer
    user, err := ext.bus.Create(ctx, nu)
    if err != nil {
        return User{}, err
    }

    // Add audit side-effect
    audit := auditbus.Audit{
        Domain:   "user",
        Action:   "created",
        ObjectID: user.ID,
    }
    ext.auditBus.Create(ctx, audit)

    return user, nil
}
```

**Usage:**
```go
tracingExt := NewTracingExtension()
auditExt := NewAuditExtension(auditBus)

userBus := userbus.NewBusiness(log, store, tracingExt, auditExt)
// Execution: tracing → audit → core business logic
```

### Pattern 2: Strong Type Safety

Use custom types instead of primitives to enforce business rules.

```go
// Instead of primitives
type User struct {
    ID    uuid.UUID      // Not string
    Name  name.Name      // Not string
    Email mail.Address   // Not string
    Age   age.Age        // Not int
}

// Custom types with validation
package name

type Name struct {
    value string
}

func Parse(value string) (Name, error) {
    if len(value) == 0 {
        return Name{}, errors.New("name cannot be empty")
    }
    if len(value) > 100 {
        return Name{}, errors.New("name too long")
    }
    return Name{value: value}, nil
}

func (n Name) String() string {
    return n.value
}
```

**Benefits:**
- Validation at creation time
- Cannot pass wrong type
- Self-documenting code
- Centralized business rules

---

## Best Practices

### ✅ DO: Follow Dependency Hierarchy

```go
// GOOD: Child imports parent interface
package productbus

import "yourapp/business/domain/userbus"

type Business struct {
    userBus userbus.ExtBusiness
}
```

### ❌ DON'T: Create Reverse Dependencies

```go
// BAD: Parent imports child
package userbus

import "yourapp/business/domain/productbus"  // Circular dependency!

type Business struct {
    productBus productbus.ExtBusiness
}
```

### ✅ DO: Use Delegate for Cross-Domain Events

```go
// GOOD: Publish event
func (b *Business) Delete(ctx context.Context, user User) error {
    if err := b.storer.Delete(ctx, user); err != nil {
        return err
    }
    return b.delegate.Call(ctx, ActionDeletedData(user.ID))
}
```

### ❌ DON'T: Directly Call Child Domains

```go
// BAD: Direct call creates dependency
func (b *Business) Delete(ctx context.Context, user User) error {
    if err := b.storer.Delete(ctx, user); err != nil {
        return err
    }
    return b.productBus.DeleteByUserID(ctx, user.ID)  // Circular dependency!
}
```

### ✅ DO: Use Database Views for Multi-Model Reads

```go
// GOOD: Query through view domain
products, _ := vproductBus.Query(ctx, filter)
// Single query with JOIN
```

### ❌ DON'T: Make N+1 Queries

```go
// BAD: N+1 queries
products, _ := productBus.Query(ctx, filter)
for _, product := range products {
    user, _ := userBus.QueryByID(ctx, product.UserID)  // N queries!
}
```

### ✅ DO: Validate Parent Before Creating Child

```go
// GOOD: Validate parent exists and satisfies rules
func (b *Business) Create(ctx context.Context, np NewProduct) (Product, error) {
    user, err := b.userBus.QueryByID(ctx, np.UserID)
    if err != nil {
        return Product{}, err
    }
    if !user.Enabled {
        return Product{}, ErrUserDisabled
    }
    return b.storer.Create(ctx, product)
}
```

### ✅ DO: Use Transactions for Multi-Domain Operations

```go
// GOOD: Atomic operation
tx, _ := db.Begin()
userBusTx, _ := userBus.NewWithTx(tx)
productBusTx, _ := productBus.NewWithTx(tx)

user, _ := userBusTx.Create(ctx, newUser)
product, _ := productBusTx.Create(ctx, newProduct)

tx.Commit()  // Both succeed or both fail
```

---

## Summary

This architecture enforces **strict dependency rules** through:

1. **One-Directional Imports**: Child → Parent only, enforced at compile time
2. **Delegate Pattern**: Event-driven operations without reverse dependencies
3. **Database Views**: Multi-model queries through query domains
4. **Extension System**: Composable behaviors without tight coupling
5. **Strong Types**: Business rules enforced in the type system
6. **Interface Contracts**: Dependencies as interfaces, not concrete types
7. **Transactional Support**: Atomic operations across domains

**Core Principle:** Parent domains know nothing about children. Children know about parents through interfaces. Cross-domain communication happens through events or database views.
