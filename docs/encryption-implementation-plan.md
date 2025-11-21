# Field-Level Encryption Implementation Plan

**Document Version:** 1.0
**Created:** 2025-11-20
**Status:** Planning Phase
**Estimated Effort:** 2-3 days

---

## Table of Contents

1. [Overview](#overview)
2. [Prerequisites](#prerequisites)
3. [Phase 1: Core Encryption Package](#phase-1-core-encryption-package)
4. [Phase 2: Encrypted Business Type](#phase-2-encrypted-business-type)
5. [Phase 3: Update Business Models](#phase-3-update-business-models)
6. [Phase 4: Update Database Stores](#phase-4-update-database-stores)
7. [Phase 5: Update App Layer](#phase-5-update-app-layer)
8. [Phase 6: Configuration Integration](#phase-6-configuration-integration)
9. [Phase 7: Manual Testing](#phase-7-manual-testing)
10. [Phase 8: Documentation](#phase-8-documentation)
11. [Validation Checklist](#validation-checklist)
12. [Rollback Plan](#rollback-plan)

---

## Overview

### Goal

Implement field-level encryption for sensitive user data (thoughts, reflections, behaviors) in the Rafiki habits tracker.

### Scope

**Entities affected**:
- **Moment**: 6 fields encrypted (situation, thoughts, physicalSymptoms, behavior, consequences, valuesReflection)
- **Think**: 1 field encrypted (content)

**Changes required**:
- ✅ New package: `business/sdk/encrypt/`
- ✅ New package: `business/types/encryptedcontent/`
- ✅ Modified: `business/domain/momentbus/model.go`
- ✅ Modified: `business/domain/thinkbus/model.go`
- ✅ Modified: `business/domain/momentbus/stores/momentdb/*`
- ✅ Modified: `business/domain/thinkbus/stores/thinkdb/*`
- ✅ Modified: `app/domain/momentapp/*`
- ✅ Modified: `app/domain/thinkapp/*`
- ✅ Modified: `api/services/partners/main.go`
- ✅ Modified: `.env.example`

### Success Criteria

- [ ] Code compiles without errors
- [ ] Data encrypted in database (verify with direct query)
- [ ] Data decrypted in API responses (verify with curl)
- [ ] Business logic unaware of encryption
- [ ] Type-safe (compiler enforces encryption)
- [ ] Zero schema changes needed
- [ ] golangci-lint passing

---

## Prerequisites

### Development Environment

**Required tools**:
```bash
# Go 1.23+
go version

# PostgreSQL (via Docker)
docker --version

# golangci-lint (for code quality)
golangci-lint version

# Make (for build commands)
make --version
```

**Project setup**:
```bash
# Clone repository
cd /Users/francowini/Documents/rafiki

# Ensure dependencies are up to date
go mod tidy

# Verify project builds
make build
```

### Knowledge Requirements

**Developer should understand**:
- Go interfaces and struct embedding
- AES encryption fundamentals (basic level)
- Rafiki's business types pattern ([CLAUDE.md](../CLAUDE.md))
- Database conversion patterns (`toDBModel`, `toBusModel`)

---

## Phase 1: Core Encryption Package

**Estimated Time**: 1-2 hours

### 1.1 Create Package Structure

```bash
# Create directories
mkdir -p business/sdk/encrypt

# Create files
touch business/sdk/encrypt/encrypt.go
```

### 1.2 Implement Encryptor Interface

**File**: `business/sdk/encrypt/encrypt.go`

**Implementation**:

```go
package encrypt

import (
    "crypto/aes"
    "crypto/cipher"
    "crypto/rand"
    "encoding/base64"
    "fmt"
    "io"
)

// Encryptor provides field-level encryption/decryption.
type Encryptor interface {
    Encrypt(plaintext string) (string, error)
    Decrypt(ciphertext string) (string, error)
}

// =============================================================================
// AES-256-GCM Implementation
// =============================================================================

// AESEncryptor implements AES-256-GCM encryption.
type AESEncryptor struct {
    gcm cipher.AEAD
}

// NewAESEncryptor creates an AES-256-GCM encryptor.
// key must be exactly 32 bytes for AES-256.
func NewAESEncryptor(key []byte) (*AESEncryptor, error) {
    if len(key) != 32 {
        return nil, fmt.Errorf("encryption key must be exactly 32 bytes for AES-256, got %d bytes", len(key))
    }

    block, err := aes.NewCipher(key)
    if err != nil {
        return nil, fmt.Errorf("create cipher: %w", err)
    }

    gcm, err := cipher.NewGCM(block)
    if err != nil {
        return nil, fmt.Errorf("create gcm: %w", err)
    }

    return &AESEncryptor{gcm: gcm}, nil
}

// Encrypt encrypts plaintext and returns base64-encoded ciphertext.
// Format: base64(nonce + ciphertext + authentication_tag)
func (e *AESEncryptor) Encrypt(plaintext string) (string, error) {
    // Handle empty strings - return empty (no encryption needed)
    if plaintext == "" {
        return "", nil
    }

    // Generate random nonce (12 bytes for GCM)
    nonce := make([]byte, e.gcm.NonceSize())
    if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
        return "", fmt.Errorf("generate nonce: %w", err)
    }

    // Encrypt and authenticate
    // GCM Seal appends: nonce + (ciphertext + tag)
    ciphertext := e.gcm.Seal(nonce, nonce, []byte(plaintext), nil)

    // Base64 encode for safe storage in VARCHAR/TEXT columns
    return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// Decrypt decrypts base64-encoded ciphertext and returns plaintext.
func (e *AESEncryptor) Decrypt(ciphertext string) (string, error) {
    // Handle empty strings
    if ciphertext == "" {
        return "", nil
    }

    // Base64 decode
    data, err := base64.StdEncoding.DecodeString(ciphertext)
    if err != nil {
        return "", fmt.Errorf("decode base64: %w", err)
    }

    nonceSize := e.gcm.NonceSize()
    if len(data) < nonceSize {
        return "", fmt.Errorf("ciphertext too short: expected at least %d bytes, got %d", nonceSize, len(data))
    }

    // Extract nonce and encrypted data
    nonce, encryptedData := data[:nonceSize], data[nonceSize:]

    // Decrypt and verify authentication tag
    plaintext, err := e.gcm.Open(nil, nonce, encryptedData, nil)
    if err != nil {
        return "", fmt.Errorf("decrypt: %w", err)
    }

    return string(plaintext), nil
}

// =============================================================================
// NoOp Implementation (Development/Testing)
// =============================================================================

// NoOpEncryptor is a pass-through encryptor for development and testing.
type NoOpEncryptor struct{}

// NewNoOpEncryptor creates an encryptor that doesn't encrypt.
func NewNoOpEncryptor() *NoOpEncryptor {
    return &NoOpEncryptor{}
}

// Encrypt returns plaintext unchanged.
func (e *NoOpEncryptor) Encrypt(plaintext string) (string, error) {
    return plaintext, nil
}

// Decrypt returns ciphertext unchanged.
func (e *NoOpEncryptor) Decrypt(ciphertext string) (string, error) {
    return ciphertext, nil
}
```

**Reference**: See `docs/product-encryption.md` lines 244-360 for detailed explanation

### 1.3 Validation

**Checklist**:
- [ ] Code compiles: `go build ./business/sdk/encrypt`
- [ ] golangci-lint passing: `golangci-lint run business/sdk/encrypt/`
- [ ] No syntax errors

---

## Phase 2: Encrypted Business Type

**Estimated Time**: 30-60 minutes

### 2.1 Create Package Structure

```bash
# Create directories
mkdir -p business/types/encryptedcontent

# Create files
touch business/types/encryptedcontent/encryptedcontent.go
```

### 2.2 Implement EncryptedContent Type

**File**: `business/types/encryptedcontent/encryptedcontent.go`

```go
// Package encryptedcontent represents sensitive text content that must be encrypted at rest.
package encryptedcontent

import (
    "errors"
    "strings"
)

// EncryptedContent represents text content that must be encrypted at rest in the database.
// The business layer works with plaintext. The database layer handles encryption/decryption.
type EncryptedContent struct {
    value string  // Always plaintext in memory (unexported = immutable)
}

// String returns the plaintext value.
func (c EncryptedContent) String() string {
    return c.value
}

// Equal provides support for the go-cmp package and testing.
func (c EncryptedContent) Equal(c2 EncryptedContent) bool {
    return c.value == c2.value
}

// MarshalText provides support for logging and JSON marshaling.
func (c EncryptedContent) MarshalText() ([]byte, error) {
    return []byte(c.value), nil
}

// =============================================================================

// ErrContentEmpty is returned when content is empty.
var ErrContentEmpty = errors.New("encrypted content cannot be empty")

// Parse validates and creates an EncryptedContent from plaintext.
func Parse(plaintext string) (EncryptedContent, error) {
    plaintext = strings.TrimSpace(plaintext)

    if plaintext == "" {
        return EncryptedContent{}, ErrContentEmpty
    }

    return EncryptedContent{value: plaintext}, nil
}

// MustParse is like Parse but panics on error. Use in tests only.
func MustParse(plaintext string) EncryptedContent {
    content, err := Parse(plaintext)
    if err != nil {
        panic(err)
    }
    return content
}
```

**Key points**:
- Mirrors `content.Content` API for consistency
- Unexported field ensures immutability
- Type name signals "this must be encrypted in DB"
- No actual encryption logic (pure business type)

### 2.3 Validation

**Checklist**:
- [ ] Code compiles: `go build ./business/types/encryptedcontent`
- [ ] API consistent with `content.Content`
- [ ] golangci-lint passing
- [ ] No syntax errors

---

## Phase 3: Update Business Models

**Estimated Time**: 30 minutes

### 3.1 Update Moment Model

**File**: `business/domain/momentbus/model.go`

**Changes**:

```go
package momentbus

import (
    // ... existing imports
    "github.com/francowini/rafiki/business/types/encryptedcontent"  // ADD this import
)

// Moment represents a tracked emotional/difficult moment in the system.
type Moment struct {
    ID               uuid.UUID
    UserID           uuid.UUID
    MomentDate       time.Time
    Situation        encryptedcontent.EncryptedContent  // CHANGED from content.Content
    Thoughts         encryptedcontent.EncryptedContent  // CHANGED from content.Content
    PhysicalSymptoms encryptedcontent.EncryptedContent  // CHANGED from content.Content
    Behavior         encryptedcontent.EncryptedContent  // CHANGED from content.Content
    Consequences     encryptedcontent.EncryptedContent  // CHANGED from content.Content
    ValuesReflection encryptedcontent.EncryptedContent  // CHANGED from content.Content
    Intensity        intensity.Intensity                 // UNCHANGED (not encrypted)
    DateCreated      time.Time
    DateUpdated      time.Time
}

// NewMoment contains information needed to create a new moment.
type NewMoment struct {
    UserID           uuid.UUID
    MomentDate       time.Time
    Situation        encryptedcontent.EncryptedContent  // CHANGED
    Thoughts         encryptedcontent.EncryptedContent  // CHANGED
    PhysicalSymptoms encryptedcontent.EncryptedContent  // CHANGED
    Behavior         encryptedcontent.EncryptedContent  // CHANGED
    Consequences     encryptedcontent.EncryptedContent  // CHANGED
    ValuesReflection encryptedcontent.EncryptedContent  // CHANGED
    Intensity        intensity.Intensity
}

// UpdateMoment contains information needed to update a moment.
type UpdateMoment struct {
    MomentDate       *time.Time
    Situation        *encryptedcontent.EncryptedContent  // CHANGED
    Thoughts         *encryptedcontent.EncryptedContent  // CHANGED
    PhysicalSymptoms *encryptedcontent.EncryptedContent  // CHANGED
    Behavior         *encryptedcontent.EncryptedContent  // CHANGED
    Consequences     *encryptedcontent.EncryptedContent  // CHANGED
    ValuesReflection *encryptedcontent.EncryptedContent  // CHANGED
    Intensity        *intensity.Intensity
}
```

### 3.2 Update Think Model

**File**: `business/domain/thinkbus/model.go`

**Changes**:

```go
package thinkbus

import (
    // ... existing imports
    "github.com/francowini/rafiki/business/types/encryptedcontent"  // ADD this import
)

// Think represents a thought/note in the system
type Think struct {
    ID          uuid.UUID
    UserID      uuid.UUID
    Category    Category                              // UNCHANGED (not encrypted)
    Content     encryptedcontent.EncryptedContent     // CHANGED from content.Content
    DateCreated time.Time
    DateUpdated time.Time
}

// NewThink contains information needed to create a new think
type NewThink struct {
    UserID   uuid.UUID
    Category Category
    Content  encryptedcontent.EncryptedContent  // CHANGED
}
```

### 3.3 Validation

**Checklist**:
- [ ] Code compiles: `go build ./business/domain/momentbus`
- [ ] Code compiles: `go build ./business/domain/thinkbus`
- [ ] Import cycles resolved
- [ ] golangci-lint passing

---

## Phase 4: Update Database Stores

**Estimated Time**: 2-3 hours (most complex phase)

### 4.1 Update Moment Database Store

#### 4.1.1 Inject Encryptor

**File**: `business/domain/momentbus/stores/momentdb/momentdb.go`

**Add import**:
```go
import (
    // ... existing imports
    "github.com/francowini/rafiki/business/sdk/encrypt"  // ADD
)
```

**Update Store struct**:
```go
// Store manages the set of APIs for moment database access.
type Store struct {
    log       *logger.Logger
    db        sqlx.ExtContext
    encryptor encrypt.Encryptor  // ADD this field
}
```

**Update NewStore constructor**:
```go
// NewStore constructs the api for data access.
func NewStore(log *logger.Logger, db *sqlx.DB, encryptor encrypt.Encryptor) *Store {  // ADD parameter
    return &Store{
        log:       log,
        db:        db,
        encryptor: encryptor,  // ADD initialization
    }
}
```

#### 4.1.2 Update Conversion Functions

**File**: `business/domain/momentbus/stores/momentdb/model.go`

**Add import**:
```go
import (
    // ... existing imports
    "github.com/francowini/rafiki/business/types/encryptedcontent"  // ADD
)
```

**Update `toDBMoment` signature and implementation**:

```go
// OLD signature (function)
// func toDBMoment(bus momentbus.Moment) moment

// NEW signature (method with error handling)
func (s *Store) toDBMoment(bus momentbus.Moment) (moment, error) {
    // Encrypt Situation
    encryptedSituation, err := s.encryptor.Encrypt(bus.Situation.String())
    if err != nil {
        return moment{}, fmt.Errorf("encrypt situation: %w", err)
    }

    // Encrypt Thoughts
    encryptedThoughts, err := s.encryptor.Encrypt(bus.Thoughts.String())
    if err != nil {
        return moment{}, fmt.Errorf("encrypt thoughts: %w", err)
    }

    // Encrypt PhysicalSymptoms
    encryptedPhysicalSymptoms, err := s.encryptor.Encrypt(bus.PhysicalSymptoms.String())
    if err != nil {
        return moment{}, fmt.Errorf("encrypt physical_symptoms: %w", err)
    }

    // Encrypt Behavior
    encryptedBehavior, err := s.encryptor.Encrypt(bus.Behavior.String())
    if err != nil {
        return moment{}, fmt.Errorf("encrypt behavior: %w", err)
    }

    // Encrypt Consequences
    encryptedConsequences, err := s.encryptor.Encrypt(bus.Consequences.String())
    if err != nil {
        return moment{}, fmt.Errorf("encrypt consequences: %w", err)
    }

    // Encrypt ValuesReflection
    encryptedValuesReflection, err := s.encryptor.Encrypt(bus.ValuesReflection.String())
    if err != nil {
        return moment{}, fmt.Errorf("encrypt values_reflection: %w", err)
    }

    return moment{
        ID:               bus.ID,
        UserID:           bus.UserID,
        MomentDate:       bus.MomentDate.UTC(),
        Situation:        encryptedSituation,        // Encrypted base64 string
        Thoughts:         encryptedThoughts,         // Encrypted base64 string
        PhysicalSymptoms: encryptedPhysicalSymptoms, // Encrypted base64 string
        Behavior:         encryptedBehavior,         // Encrypted base64 string
        Consequences:     encryptedConsequences,     // Encrypted base64 string
        ValuesReflection: encryptedValuesReflection, // Encrypted base64 string
        Intensity:        bus.Intensity.Value(),     // Plaintext (not encrypted)
        DateCreated:      bus.DateCreated.UTC(),
        DateUpdated:      bus.DateUpdated.UTC(),
    }, nil
}
```

**Update `toBusMoment` signature and implementation**:

```go
// OLD signature (function)
// func toBusMoment(db moment) (momentbus.Moment, error)

// NEW signature (method with decryption)
func (s *Store) toBusMoment(db moment) (momentbus.Moment, error) {
    // Decrypt Situation
    situationPlain, err := s.encryptor.Decrypt(db.Situation)
    if err != nil {
        return momentbus.Moment{}, fmt.Errorf("decrypt situation: %w", err)
    }

    situation, err := encryptedcontent.Parse(situationPlain)  // Changed from content.Parse
    if err != nil {
        return momentbus.Moment{}, fmt.Errorf("parse situation: %w", err)
    }

    // Decrypt Thoughts
    thoughtsPlain, err := s.encryptor.Decrypt(db.Thoughts)
    if err != nil {
        return momentbus.Moment{}, fmt.Errorf("decrypt thoughts: %w", err)
    }

    thoughts, err := encryptedcontent.Parse(thoughtsPlain)
    if err != nil {
        return momentbus.Moment{}, fmt.Errorf("parse thoughts: %w", err)
    }

    // Decrypt PhysicalSymptoms
    physicalSymptomsPlain, err := s.encryptor.Decrypt(db.PhysicalSymptoms)
    if err != nil {
        return momentbus.Moment{}, fmt.Errorf("decrypt physical_symptoms: %w", err)
    }

    physicalSymptoms, err := encryptedcontent.Parse(physicalSymptomsPlain)
    if err != nil {
        return momentbus.Moment{}, fmt.Errorf("parse physical_symptoms: %w", err)
    }

    // Decrypt Behavior
    behaviorPlain, err := s.encryptor.Decrypt(db.Behavior)
    if err != nil {
        return momentbus.Moment{}, fmt.Errorf("decrypt behavior: %w", err)
    }

    behavior, err := encryptedcontent.Parse(behaviorPlain)
    if err != nil {
        return momentbus.Moment{}, fmt.Errorf("parse behavior: %w", err)
    }

    // Decrypt Consequences
    consequencesPlain, err := s.encryptor.Decrypt(db.Consequences)
    if err != nil {
        return momentbus.Moment{}, fmt.Errorf("decrypt consequences: %w", err)
    }

    consequences, err := encryptedcontent.Parse(consequencesPlain)
    if err != nil {
        return momentbus.Moment{}, fmt.Errorf("parse consequences: %w", err)
    }

    // Decrypt ValuesReflection
    valuesReflectionPlain, err := s.encryptor.Decrypt(db.ValuesReflection)
    if err != nil {
        return momentbus.Moment{}, fmt.Errorf("decrypt values_reflection: %w", err)
    }

    valuesReflection, err := encryptedcontent.Parse(valuesReflectionPlain)
    if err != nil {
        return momentbus.Moment{}, fmt.Errorf("parse values_reflection: %w", err)
    }

    // Parse intensity (unchanged)
    intensityVal, err := intensity.Parse(db.Intensity)
    if err != nil {
        return momentbus.Moment{}, fmt.Errorf("parse intensity: %w", err)
    }

    return momentbus.Moment{
        ID:               db.ID,
        UserID:           db.UserID,
        MomentDate:       db.MomentDate.In(time.Local),
        Situation:        situation,
        Thoughts:         thoughts,
        PhysicalSymptoms: physicalSymptoms,
        Behavior:         behavior,
        Consequences:     consequences,
        ValuesReflection: valuesReflection,
        Intensity:        intensityVal,
        DateCreated:      db.DateCreated.In(time.Local),
        DateUpdated:      db.DateUpdated.In(time.Local),
    }, nil
}
```

**Update `toBusMoments`**:
```go
func (s *Store) toBusMoments(dbs []moment) ([]momentbus.Moment, error) {
    moments := make([]momentbus.Moment, len(dbs))

    for i, db := range dbs {
        var err error
        moments[i], err = s.toBusMoment(db)  // Now a method call
        if err != nil {
            return nil, err
        }
    }

    return moments, nil
}
```

**Update all callers in `momentdb.go`**:

Find all calls to `toDBMoment` and `toBusMoment` and update them:

```go
// In Create method
func (s *Store) Create(ctx context.Context, moment momentbus.Moment) error {
    dbMoment, err := s.toDBMoment(moment)  // Now returns error
    if err != nil {
        return err
    }
    // ... rest of method
}

// In Update method
func (s *Store) Update(ctx context.Context, moment momentbus.Moment) error {
    dbMoment, err := s.toDBMoment(moment)  // Now returns error
    if err != nil {
        return err
    }
    // ... rest of method
}

// In Query method
func (s *Store) Query(ctx context.Context, filter momentbus.QueryFilter) ([]momentbus.Moment, error) {
    // ... query database
    return s.toBusMoments(dbMoments)  // Now a method call
}

// In QueryByID method
func (s *Store) QueryByID(ctx context.Context, momentID uuid.UUID) (momentbus.Moment, error) {
    // ... query database
    return s.toBusMoment(dbMoment)  // Now a method call
}
```

### 4.2 Update Think Database Store

**Apply same pattern to Think store**:

**File**: `business/domain/thinkbus/stores/thinkdb/thinkdb.go`

1. Add import: `"github.com/francowini/rafiki/business/sdk/encrypt"`
2. Add field to Store: `encryptor encrypt.Encryptor`
3. Update NewStore: `func NewStore(log *logger.Logger, db *sqlx.DB, encryptor encrypt.Encryptor) *Store`

**File**: `business/domain/thinkbus/stores/thinkdb/model.go`

1. Add import: `"github.com/francowini/rafiki/business/types/encryptedcontent"`
2. Update `toDBThink`:
   ```go
   func (s *Store) toDBThink(bus thinkbus.Think) (think, error) {
       encryptedContent, err := s.encryptor.Encrypt(bus.Content.String())
       if err != nil {
           return think{}, fmt.Errorf("encrypt content: %w", err)
       }

       return think{
           ID:          bus.ID,
           UserID:      bus.UserID,
           Category:    bus.Category.String(),
           Content:     encryptedContent,  // Encrypted
           DateCreated: bus.DateCreated.UTC(),
           DateUpdated: bus.DateUpdated.UTC(),
       }, nil
   }
   ```

3. Update `toBusThink`:
   ```go
   func (s *Store) toBusThink(db think) (thinkbus.Think, error) {
       category, err := thinkbus.ParseCategory(db.Category)
       if err != nil {
           return thinkbus.Think{}, fmt.Errorf("parse category: %w", err)
       }

       // Decrypt content
       contentPlain, err := s.encryptor.Decrypt(db.Content)
       if err != nil {
           return thinkbus.Think{}, fmt.Errorf("decrypt content: %w", err)
       }

       cnt, err := encryptedcontent.Parse(contentPlain)  // Changed
       if err != nil {
           return thinkbus.Think{}, fmt.Errorf("parse content: %w", err)
       }

       return thinkbus.Think{
           ID:          db.ID,
           UserID:      db.UserID,
           Category:    category,
           Content:     cnt,
           DateCreated: db.DateCreated.In(time.Local),
           DateUpdated: db.DateUpdated.In(time.Local),
       }, nil
   }
   ```

4. Update `toBusThinks` and all callers

### 4.3 Validation

**Checklist**:
- [ ] Code compiles: `go build ./business/domain/momentbus/stores/momentdb`
- [ ] Code compiles: `go build ./business/domain/thinkbus/stores/thinkdb`
- [ ] All method signatures updated
- [ ] Error handling complete
- [ ] golangci-lint passing

---

## Phase 5: Update App Layer

**Estimated Time**: 1 hour

### 5.1 Update Moment App Layer

**File**: `app/domain/momentapp/momentapp.go`

**Add import**:
```go
import (
    // ... existing imports
    "github.com/francowini/rafiki/business/types/encryptedcontent"  // ADD
)
```

**Update `toBusNewMoment`**:

```go
func toBusNewMoment(app NewMoment) (momentbus.NewMoment, error) {
    // ... existing code for parsing userID, momentDate, intensity ...

    // Parse encrypted content fields (CHANGED from content.Parse)
    situation, err := encryptedcontent.Parse(app.Situation)
    if err != nil {
        return momentbus.NewMoment{}, fmt.Errorf("parse situation: %w", err)
    }

    thoughts, err := encryptedcontent.Parse(app.Thoughts)
    if err != nil {
        return momentbus.NewMoment{}, fmt.Errorf("parse thoughts: %w", err)
    }

    physicalSymptoms, err := encryptedcontent.Parse(app.PhysicalSymptoms)
    if err != nil {
        return momentbus.NewMoment{}, fmt.Errorf("parse physical_symptoms: %w", err)
    }

    behavior, err := encryptedcontent.Parse(app.Behavior)
    if err != nil {
        return momentbus.NewMoment{}, fmt.Errorf("parse behavior: %w", err)
    }

    consequences, err := encryptedcontent.Parse(app.Consequences)
    if err != nil {
        return momentbus.NewMoment{}, fmt.Errorf("parse consequences: %w", err)
    }

    valuesReflection, err := encryptedcontent.Parse(app.ValuesReflection)
    if err != nil {
        return momentbus.NewMoment{}, fmt.Errorf("parse values_reflection: %w", err)
    }

    return momentbus.NewMoment{
        UserID:           userID,
        MomentDate:       momentDate,
        Situation:        situation,
        Thoughts:         thoughts,
        PhysicalSymptoms: physicalSymptoms,
        Behavior:         behavior,
        Consequences:     consequences,
        ValuesReflection: valuesReflection,
        Intensity:        intensityVal,
    }, nil
}
```

**Update `toBusUpdateMoment`**:

```go
func toBusUpdateMoment(app UpdateMoment) (momentbus.UpdateMoment, error) {
    bus := momentbus.UpdateMoment{}

    // ... existing code for MomentDate, Intensity ...

    if app.Situation != nil {
        situation, err := encryptedcontent.Parse(*app.Situation)  // CHANGED
        if err != nil {
            return momentbus.UpdateMoment{}, fmt.Errorf("parse situation: %w", err)
        }
        bus.Situation = &situation
    }

    if app.Thoughts != nil {
        thoughts, err := encryptedcontent.Parse(*app.Thoughts)  // CHANGED
        if err != nil {
            return momentbus.UpdateMoment{}, fmt.Errorf("parse thoughts: %w", err)
        }
        bus.Thoughts = &thoughts
    }

    // ... similar for PhysicalSymptoms, Behavior, Consequences, ValuesReflection

    return bus, nil
}
```

### 5.2 Update Think App Layer

**File**: `app/domain/thinkapp/thinkapp.go`

**Add import**:
```go
import (
    // ... existing imports
    "github.com/francowini/rafiki/business/types/encryptedcontent"  // ADD
)
```

**Update `toBusNewThink`**:

```go
func toBusNewThink(app NewThink) (thinkbus.NewThink, error) {
    // ... existing code for userID, category ...

    content, err := encryptedcontent.Parse(app.Content)  // CHANGED from content.Parse
    if err != nil {
        return thinkbus.NewThink{}, fmt.Errorf("parse content: %w", err)
    }

    return thinkbus.NewThink{
        UserID:   userID,
        Category: category,
        Content:  content,
    }, nil
}
```

### 5.3 Validation

**Checklist**:
- [ ] Code compiles: `go build ./app/domain/momentapp`
- [ ] Code compiles: `go build ./app/domain/thinkapp`
- [ ] All conversion functions updated
- [ ] golangci-lint passing

---

## Phase 6: Configuration Integration

**Estimated Time**: 1 hour

### 6.1 Update Main Configuration

**File**: `api/services/partners/main.go`

**Add imports**:
```go
import (
    "encoding/hex"
    "github.com/francowini/rafiki/business/sdk/encrypt"
)
```

**Update config struct** (find the existing `cfg` struct definition):

```go
cfg := struct {
    conf.Version
    Web struct {
        // ... existing fields
    }
    DB struct {
        // ... existing fields
    }
    Tempo struct {
        // ... existing fields
    }
    Encryption struct {                    // ADD this section
        Key string `conf:"required,mask"`  // PARTNER_ENCRYPTION_KEY
    }
}{
    Version: conf.Version{
        Build: build,
        Desc:  "Partner",
    },
}
```

**Add encryption initialization** (after config parsing, before database store creation):

```go
// -------------------------------------------------------------------------
// Encryption Support

log.Info(ctx, "startup", "status", "initializing encryption")

// Validate key format (must be 64 hex characters = 32 bytes)
if len(cfg.Encryption.Key) != 64 {
    return fmt.Errorf("encryption key must be 64 hex characters (32 bytes), got %d", len(cfg.Encryption.Key))
}

// Decode hex string to bytes
encryptionKey, err := hex.DecodeString(cfg.Encryption.Key)
if err != nil {
    return fmt.Errorf("decode encryption key: %w", err)
}

// Create AES-256-GCM encryptor
encryptor, err := encrypt.NewAESEncryptor(encryptionKey)
if err != nil {
    return fmt.Errorf("create encryptor: %w", err)
}

log.Info(ctx, "startup", "status", "encryption initialized", "algorithm", "AES-256-GCM")
```

### 6.2 Inject Encryptor into Stores

**Find existing store creation** (search for `momentdb.NewStore` and `thinkdb.NewStore`):

```go
// OLD: Without encryptor
momentStore := momentdb.NewStore(log, db)
thinkStore := thinkdb.NewStore(log, db)

// NEW: With encryptor
momentStore := momentdb.NewStore(log, db, encryptor)  // ADD encryptor parameter
thinkStore := thinkdb.NewStore(log, db, encryptor)    // ADD encryptor parameter

// Business layer creation remains unchanged
momentBus := momentbus.NewBusiness(log, momentStore)
thinkBus := thinkbus.NewBusiness(log, thinkStore)
```

### 6.3 Update Environment Configuration

**File**: `.env.example`

**Add encryption key section**:

```bash
# ==============================================================================
# Encryption Configuration
# ==============================================================================

# Field-level encryption key for sensitive user data (thoughts, reflections, behaviors)
# REQUIRED for production deployment
#
# Key requirements:
# - Must be exactly 32 bytes (64 hex characters)
# - Generate with: openssl rand -hex 32
# - Store securely in password manager (key loss = permanent data loss!)
# - Never commit this key to version control
#
# Encrypted fields:
# - Moments: situation, thoughts, physicalSymptoms, behavior, consequences, valuesReflection
# - Thinks: content
#
# Example key generation:
# $ openssl rand -hex 32
# a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4a5b6c7d8e9f0a1b2
#
PARTNER_ENCRYPTION_KEY=

```

### 6.4 Validation

**Checklist**:
- [ ] Application compiles: `go build ./api/services/partners`
- [ ] Config validation works (try with missing key - should fail)
- [ ] Config validation works (try with wrong key length - should fail)
- [ ] golangci-lint passing

---

## Phase 7: Manual Testing

**Estimated Time**: 1-2 hours

### 7.1 Local Environment Setup

**Generate test encryption key**:
```bash
# Generate 32-byte key (64 hex characters)
openssl rand -hex 32

# Example output:
# a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4a5b6c7d8e9f0a1b2
```

**Add to local `.env`**:
```bash
# Edit .env file
nano .env

# Add line:
PARTNER_ENCRYPTION_KEY=a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4a5b6c7d8e9f0a1b2
```

**Start services**:
```bash
# Start database and service
make up

# Watch logs
make logs
```

**Verify encryption initialized**:
```bash
# Look for log message
make logs | grep -i encryption

# Expected output:
# {"level":"INFO","msg":"startup","status":"initializing encryption"}
# {"level":"INFO","msg":"startup","status":"encryption initialized","algorithm":"AES-256-GCM"}
```

### 7.2 API Testing

**Create a moment with sensitive data**:

```bash
# Get auth token first (if needed)
TOKEN="your-auth-token"

# Create moment
curl -X POST http://localhost:3000/v1/moments \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "momentDate": "2025-11-20T10:00:00Z",
    "situation": "I feel very anxious about my work presentation",
    "thoughts": "I think everyone will judge me and notice my mistakes",
    "physicalSymptoms": "Racing heart, sweaty palms, tight chest",
    "behavior": "Avoiding preparation, procrastinating",
    "consequences": "More anxiety, less prepared, worse performance",
    "valuesReflection": "I value competence but fear vulnerability",
    "intensity": 8
  }' | jq

# Save the moment_id from response
MOMENT_ID="<uuid-from-response>"
```

### 7.3 Database Verification

**Check encrypted data in database**:

```bash
# Connect to database
docker exec -it rafiki-postgres psql -U rafiki -d rafiki

# Query moment (should see ENCRYPTED data)
SELECT
    moment_id,
    LEFT(situation, 50) as situation_preview,
    LEFT(thoughts, 50) as thoughts_preview,
    intensity
FROM moments
WHERE moment_id = '<paste-moment-id-here>';

# Expected output:
# situation_preview: "dGhpcyBpcyBlbmNyeXB0ZWQgZGF0YQ==" (base64 gibberish)
# thoughts_preview: "YW5vdGhlciBlbmNyeXB0ZWQ=" (base64 gibberish)
# intensity: 8 (plaintext number)

# Exit psql
\q
```

### 7.4 API Response Verification

**Query moment via API**:

```bash
# Query moment (should see DECRYPTED data)
curl http://localhost:3000/v1/moments/$MOMENT_ID \
  -H "Authorization: Bearer $TOKEN" | jq

# Expected response (plaintext):
# {
#   "id": "...",
#   "situation": "I feel very anxious about my work presentation",
#   "thoughts": "I think everyone will judge me and notice my mistakes",
#   "physicalSymptoms": "Racing heart, sweaty palms, tight chest",
#   ...
# }
```

### 7.5 Think Testing

**Test Think entity** (same pattern):

```bash
# Create think
curl -X POST http://localhost:3000/v1/thinks \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "category": "personal",
    "content": "Private thoughts about my personal struggles and goals"
  }' | jq

# Verify in database (content should be encrypted)
# Verify in API (content should be decrypted)
```

### 7.6 Validation

**Checklist**:
- [ ] Service starts successfully
- [ ] Encryption initialization logged
- [ ] Moments can be created via API
- [ ] Thinks can be created via API
- [ ] Database contains encrypted base64 (not plaintext)
- [ ] API returns decrypted plaintext
- [ ] Intensity/category fields NOT encrypted (plaintext in DB)

---

## Phase 8: Documentation

**Estimated Time**: 30 minutes

### 8.1 Update CLAUDE.md

**File**: `CLAUDE.md`

**Add security section**:

```markdown
## Security

### Field-Level Encryption

Rafiki uses field-level encryption to protect sensitive user data at rest in the database.

**Encrypted Fields**:
- **Moments**: situation, thoughts, physicalSymptoms, behavior, consequences, valuesReflection
- **Thinks**: content

**Architecture**:
- Business layer: Works with plaintext (`encryptedcontent.EncryptedContent` type)
- Database layer: Encrypts before save, decrypts after load
- Algorithm: AES-256-GCM (authenticated encryption)
- Key: 32 bytes, stored in `PARTNER_ENCRYPTION_KEY` environment variable

**Local Development**:
```bash
# Generate encryption key
openssl rand -hex 32

# Add to .env
PARTNER_ENCRYPTION_KEY=<generated-key>
```

**Key Management**:
- Back up key in password manager (key loss = permanent data loss!)
- File permissions: `chmod 600 .env`
- Never commit key to git (already in `.gitignore`)

**Documentation**:
- Architecture: [docs/encryption-architecture.md](docs/encryption-architecture.md)
- Implementation: [docs/encryption-implementation-plan.md](docs/encryption-implementation-plan.md)
- Deployment: [docs/encryption-deployment-guide.md](docs/encryption-deployment-guide.md)
```

### 8.2 Validation

**Checklist**:
- [ ] CLAUDE.md updated
- [ ] Documentation links working
- [ ] Clear instructions for developers

---

## Validation Checklist

### Pre-Deployment

- [ ] All code compiles without errors
- [ ] golangci-lint passing on all packages
- [ ] Manual API testing complete
- [ ] Database encryption verified (base64 in DB)
- [ ] API decryption verified (plaintext in response)
- [ ] Documentation complete

### Code Quality

- [ ] No hardcoded encryption keys
- [ ] Error handling complete (encrypt/decrypt failures)
- [ ] Type safety enforced (EncryptedContent type)
- [ ] Immutability maintained (unexported fields)
- [ ] Consistent naming conventions
- [ ] No sensitive data in logs

### Security

- [ ] Encryption key 32 bytes (64 hex characters)
- [ ] Key stored in .env (not in code)
- [ ] AES-256-GCM properly implemented
- [ ] Random nonces per encryption
- [ ] Authentication tag validated on decrypt
- [ ] No keys in git history

---

## Rollback Plan

### If Implementation Fails

**Revert changes**:

```bash
# Discard all changes
git checkout main
git branch -D docs/field-level-encryption-plan

# Or revert specific commit
git revert <commit-hash>
```

**Database**: No schema changes, so no migration rollback needed.

### If Deployment Fails

**Rollback procedure**:

1. SSH to server: `ssh root@178.156.170.37`
2. Stop services: `docker compose --profile production down`
3. Checkout previous commit: `git checkout <previous-commit>`
4. Remove encryption key: Edit `.env`, comment out `PARTNER_ENCRYPTION_KEY`
5. Redeploy: `./devops/deploy.sh`

---

## Next Steps

After completing this implementation:

1. ✅ Review code with team
2. ✅ Test in local environment
3. ✅ Proceed to [Deployment Guide](./encryption-deployment-guide.md)
4. ✅ Deploy to production
5. ✅ Monitor for 24 hours

---

**Document Status**: ✅ Ready for Implementation
**Next Document**: [encryption-deployment-guide.md](./encryption-deployment-guide.md)
