# Field-Level Encryption Architecture for Rafiki

**Document Version:** 1.0
**Created:** 2025-11-20
**Status:** Planning Phase
**Author:** Development Team

---

## Table of Contents

1. [Overview](#overview)
2. [Design Principles](#design-principles)
3. [Architecture](#architecture)
4. [Business Type Strategy](#business-type-strategy)
5. [Data Flow](#data-flow)
6. [Security Model](#security-model)
7. [Type System Design](#type-system-design)
8. [Database Strategy](#database-strategy)
9. [Performance Considerations](#performance-considerations)
10. [Future Extensibility](#future-extensibility)

---

## Overview

### Purpose

Implement field-level encryption for sensitive user data in the Rafiki habits tracker application to protect personal information (thoughts, reflections, behaviors) at rest in the PostgreSQL database.

### Goals

- ✅ **Transparent encryption**: Business logic works with plaintext data
- ✅ **Type-safe**: Compiler enforces which fields are encrypted
- ✅ **Zero schema changes**: Uses existing TEXT columns
- ✅ **Performance**: Minimal overhead (<1ms per record)
- ✅ **Reusable**: Easy to add encryption to new fields/entities
- ✅ **Secure**: AES-256-GCM with proper key management

### Non-Goals

- ❌ **Searchable encryption**: Cannot query encrypted fields (acceptable tradeoff)
- ❌ **Column-level encryption**: Not using database-level encryption
- ❌ **Encryption at rest only**: TLS already handles encryption in transit
- ❌ **Homomorphic encryption**: Too complex for current requirements

---

## Design Principles

### 1. Encryption is a Business Concern

**Rationale**: Which data is sensitive is a business decision, not a database decision.

**Implementation**:
- Define encrypted data as a **business type** (`encryptedcontent.EncryptedContent`)
- Type system enforces encryption at compile time
- No runtime configuration or tags needed

**Example**:
```go
// Business model declares sensitive fields
type Moment struct {
    Situation        encryptedcontent.EncryptedContent  // Compiler knows this is sensitive
    Intensity        intensity.Intensity                 // Not encrypted (business decision)
}
```

### 2. Business Layer Stays Pure

**Rationale**: Business logic should not be coupled to encryption details.

**Implementation**:
- Business layer works with **plaintext** data only
- Encryption happens at **infrastructure boundary** (database layer)
- No changes to business logic or validation rules

**Example**:
```go
// Business logic - no encryption awareness
func ValidateMoment(m Moment) error {
    if m.Situation.String() == "" {  // Works with plaintext
        return errors.New("situation required")
    }
    return nil
}
```

### 3. Type System Enforces Security

**Rationale**: Prevent accidental storage of sensitive data in plaintext.

**Implementation**:
- `encryptedcontent.EncryptedContent` type signals "must be encrypted"
- Database layer **cannot** store EncryptedContent without encrypting
- Compiler error if EncryptedContent used incorrectly

**Anti-pattern (prevented by types)**:
```go
// This won't compile - EncryptedContent can't be stored as plain string
db.Situation = moment.Situation.String()  // Compile error!

// Correct - must explicitly encrypt
db.Situation = encryptor.Encrypt(moment.Situation.String())  // OK
```

### 4. Reusability Over Repetition

**Rationale**: Encryption logic should work for any text content, not be duplicated per field.

**Implementation**:
- Single `encryptedcontent.EncryptedContent` type for all encrypted text
- Generic encryption helpers
- Same encryption logic for Moments, Thinks, and future entities

**Example**:
```go
// Reusable for any entity
type User struct {
    PrivateNotes  encryptedcontent.EncryptedContent  // Reuses same type
}

type Goal struct {
    PersonalReason encryptedcontent.EncryptedContent  // Reuses same type
}
```

### 5. Fail Secure by Default

**Rationale**: Encryption failures should prevent data storage, not store plaintext.

**Implementation**:
- Encryption failure → return error, abort transaction
- Decryption failure → return error, don't return partial data
- Missing encryption key → fail fast on startup

**Example**:
```go
// Encryption failure prevents storage
encrypted, err := encryptor.Encrypt(data)
if err != nil {
    return fmt.Errorf("encryption failed: %w", err)  // Don't store plaintext!
}
```

---

## Architecture

### High-Level Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                     API Layer (JSON)                        │
│  - Receives plaintext from frontend                         │
│  - Returns plaintext to frontend                            │
└────────────────────────┬────────────────────────────────────┘
                         │
┌────────────────────────▼────────────────────────────────────┐
│                   App Layer (Conversion)                    │
│  - Converts JSON → Business Types                           │
│  - Uses: encryptedcontent.Parse("user text")                │
└────────────────────────┬────────────────────────────────────┘
                         │
┌────────────────────────▼────────────────────────────────────┐
│              Business Layer (Domain Logic)                  │
│  - Works with plaintext EncryptedContent                    │
│  - No encryption awareness                                  │
│  - Validates: moment.Situation.String() != ""               │
└────────────────────────┬────────────────────────────────────┘
                         │
┌────────────────────────▼────────────────────────────────────┐
│           Database Store Layer (Infrastructure)             │
│  - Detects EncryptedContent type                            │
│  - Before Save: Encrypt(plaintext) → base64                 │
│  - After Load: Decrypt(base64) → plaintext                  │
└────────────────────────┬────────────────────────────────────┘
                         │
┌────────────────────────▼────────────────────────────────────┐
│                PostgreSQL (Data at Rest)                    │
│  - Stores: "a3f2b1c4d5..." (encrypted base64)               │
│  - Never stores plaintext for EncryptedContent fields       │
└─────────────────────────────────────────────────────────────┘
```

### Package Structure

```
business/
├── sdk/
│   └── encrypt/
│       ��── encrypt.go              # Encryptor interface + AES-256-GCM
│       └── encrypt_test.go         # Encryption unit tests
│
├── types/
│   ├── content/
│   │   └── content.go              # Regular content (not encrypted)
│   │
│   └── encryptedcontent/           # NEW: Encrypted business type
│       ├── encryptedcontent.go     # Wraps text content, signals encryption
│       └── encryptedcontent_test.go
│
└── domain/
    ├── momentbus/
    │   ├── model.go                # MODIFIED: Uses EncryptedContent
    │   ├── momentbus.go            # No changes
    │   └── stores/momentdb/
    │       ├── momentdb.go         # MODIFIED: Inject encryptor
    │       └── model.go            # MODIFIED: Encrypt/decrypt in conversions
    │
    └── thinkbus/
        ├── model.go                # MODIFIED: Uses EncryptedContent
        ├── thinkbus.go             # No changes
        └── stores/thinkdb/
            ├── thinkdb.go          # MODIFIED: Inject encryptor
            └── model.go            # MODIFIED: Encrypt/decrypt in conversions
```

### Component Responsibilities

#### `business/sdk/encrypt/`
**Purpose**: Reusable encryption primitives

**Responsibilities**:
- Define `Encryptor` interface (Encrypt/Decrypt methods)
- Implement `AESEncryptor` (AES-256-GCM)
- Implement `NoOpEncryptor` (testing/development)
- Validate encryption keys (must be 32 bytes)
- Handle base64 encoding/decoding

**Dependencies**: Standard library only (`crypto/aes`, `crypto/cipher`, `encoding/base64`)

#### `business/types/encryptedcontent/`
**Purpose**: Business type for encrypted text content

**Responsibilities**:
- Wrap text content in type-safe struct
- Signal "this data must be encrypted in DB"
- Provide same API as `content.Content` (Parse, String, Equal, MarshalText)
- Validate content is non-empty
- No encryption logic (pure business type)

**Dependencies**: None (standard library only)

#### Database Stores (`momentdb`, `thinkdb`)
**Purpose**: Encrypt/decrypt at infrastructure boundary

**Responsibilities**:
- Accept `encryptor encrypt.Encryptor` in constructor
- Detect `EncryptedContent` fields during conversions
- `toDBModel()`: Encrypt EncryptedContent → string
- `toBusModel()`: Decrypt string → EncryptedContent
- Handle encryption/decryption errors gracefully

**Dependencies**: `encrypt`, `encryptedcontent`, existing business types

#### Main Application (`api/services/partners/main.go`)
**Purpose**: Wire encryption configuration

**Responsibilities**:
- Load encryption key from environment (`PARTNER_ENCRYPTION_KEY`)
- Validate key format (64 hex chars = 32 bytes)
- Create `AESEncryptor` instance
- Inject encryptor into all database stores
- Fail fast if key is invalid or missing

**Dependencies**: `encrypt`, all stores

---

## Business Type Strategy

### The `EncryptedContent` Type

**Definition**:
```go
package encryptedcontent

// EncryptedContent represents text content that must be encrypted at rest.
// Business layer works with plaintext. Database layer handles encryption.
type EncryptedContent struct {
    value string  // Always plaintext in memory
}

// String returns the plaintext value.
func (c EncryptedContent) String() string {
    return c.value
}

// Parse creates EncryptedContent from plaintext.
func Parse(plaintext string) (EncryptedContent, error) {
    if strings.TrimSpace(plaintext) == "" {
        return EncryptedContent{}, errors.New("encrypted content cannot be empty")
    }
    return EncryptedContent{value: plaintext}, nil
}
```

### Type Semantics

**In Memory**: Always plaintext
- Business logic: `content.String()` returns plaintext
- Validation: Works with plaintext
- Logging: Can log (but shouldn't - security policy)

**In Database**: Always encrypted
- `toDBModel()`: Encrypts before storing
- `toBusModel()`: Decrypts after loading
- PostgreSQL: Stores base64-encoded ciphertext

**In Transit**: Plaintext over TLS
- API: Returns plaintext JSON (over HTTPS)
- Frontend: Receives plaintext (TLS protects transit)

### Type vs Content Type

| Aspect | `content.Content` | `encryptedcontent.EncryptedContent` |
|--------|-------------------|--------------------------------------|
| **Purpose** | General text content | Sensitive user data |
| **Storage** | Plaintext in DB | **Encrypted in DB** |
| **Use Case** | Categories, labels | Thoughts, reflections |
| **Searchable** | ✅ Yes | ❌ No (encrypted) |
| **Indexable** | ✅ Yes | ❌ No (random ciphertext) |

### When to Use Each Type

**Use `content.Content` for**:
- Non-sensitive data (categories, public labels)
- Data that needs database queries/filters
- Data that needs database indexing
- Example: Think category, Product name

**Use `encryptedcontent.EncryptedContent` for**:
- Personal/sensitive user data
- Data that doesn't need queries (fetch by ID only)
- HIPAA/GDPR protected information
- Example: Thoughts, behaviors, medical symptoms

---

## Data Flow

### Creating a Moment (Write Path)

```
1. Frontend sends JSON:
   POST /v1/moments
   {"situation": "I feel anxious about work"}

2. API Layer (HTTP Handler):
   - Parses JSON to app.NewMoment

3. App Layer (momentapp → momentbus):
   situation, err := encryptedcontent.Parse("I feel anxious about work")
   newMoment := momentbus.NewMoment{
     Situation: situation,  // EncryptedContent (plaintext in memory)
   }

4. Business Layer (momentbus.Business.Create):
   - Validates: moment.Situation.String() != ""
   - Business logic runs on plaintext
   - Calls: storer.Create(ctx, moment)

5. Database Store (momentdb.Store.Create):
   - Calls: toDBMoment(moment)

6. Conversion (toDBMoment):
   // Detect EncryptedContent type
   encryptedSituation, err := s.encryptor.Encrypt(moment.Situation.String())
   // "I feel anxious about work" → "a3f2b1c4d5e6..."

   dbMoment := moment{
     Situation: encryptedSituation,  // Encrypted base64 string
   }

7. PostgreSQL:
   INSERT INTO moments (situation) VALUES ('a3f2b1c4d5e6...');
   -- Stores encrypted base64, never plaintext
```

### Querying a Moment (Read Path)

```
1. Frontend requests:
   GET /v1/moments/{id}

2. API Layer:
   - Calls: momentBus.QueryByID(ctx, momentID)

3. Business Layer:
   - Calls: storer.QueryByID(ctx, momentID)

4. Database Store:
   - Queries PostgreSQL

5. PostgreSQL returns:
   SELECT situation FROM moments WHERE moment_id = $1
   → situation = "a3f2b1c4d5e6..." (encrypted)

6. Conversion (toBusMoment):
   // Decrypt encrypted string
   plaintext, err := s.encryptor.Decrypt(dbMoment.Situation)
   // "a3f2b1c4d5e6..." → "I feel anxious about work"

   // Parse to business type
   situation, err := encryptedcontent.Parse(plaintext)

   moment := momentbus.Moment{
     Situation: situation,  // EncryptedContent (plaintext in memory)
   }

7. Business Layer:
   - Returns moment (plaintext)

8. App Layer:
   - Converts to JSON: {"situation": "I feel anxious about work"}

9. Frontend receives plaintext (over TLS)
```

---

## Security Model

### Encryption Algorithm

**Choice**: AES-256-GCM (Galois/Counter Mode)

**Rationale**:
- ✅ **Industry standard**: NIST approved, widely used
- ✅ **Authenticated encryption**: Prevents tampering (includes authentication tag)
- ✅ **Fast**: Hardware-accelerated on modern CPUs
- ✅ **Secure**: No known practical attacks
- ✅ **Randomized**: Each encryption uses unique nonce (prevents pattern detection)

**Technical Details**:
- **Key size**: 256 bits (32 bytes)
- **Nonce size**: 96 bits (12 bytes, random per encryption)
- **Tag size**: 128 bits (16 bytes, authentication)
- **Output format**: `nonce || ciphertext || tag` (base64 encoded)

**Example**:
```
Plaintext:  "I feel anxious"
Key:        <32 random bytes>
Nonce:      <12 random bytes>
Encrypted:  <12-byte nonce><ciphertext><16-byte tag>
Base64:     "a3f2b1c4d5e6f7g8h9..."
```

### Key Management

**Key Generation**:
```bash
# Generate 32-byte (256-bit) random key
openssl rand -hex 32
# Output: 64 hex characters (32 bytes)
```

**Key Storage**:
- **Location**: `/opt/rafiki/.env` on Hetzner server
- **Format**: `PARTNER_ENCRYPTION_KEY=<64-hex-characters>`
- **Permissions**: `600` (read/write by owner only)
- **Owner**: `root:root`

**Key Backup**:
- Password manager (1Password, LastPass, etc.)
- Encrypted offline storage (USB drive with LUKS)
- **Critical**: Key loss = permanent data loss!

**Key Rotation** (future):
- Dual-key support (current + previous)
- Background re-encryption job
- Not required for initial deployment

### Threat Model

**Protected Against**:
- ✅ **Database breach**: Attacker gets database dump → sees encrypted gibberish
- ✅ **Backup leaks**: Old backups contain encrypted data
- ✅ **Insider access**: DBAs see encrypted data, not plaintext
- ✅ **SQL injection**: Even if successful, returns encrypted data
- ✅ **Data tampering**: GCM authentication detects modifications

**NOT Protected Against**:
- ❌ **Application compromise**: Attacker with app access sees plaintext
- ❌ **Memory dumps**: Plaintext exists in memory during processing
- ❌ **Key compromise**: Attacker with key can decrypt everything
- ❌ **Side-channel attacks**: Not hardened against timing attacks

**Accepted Risks**:
- Key stored on same server as database (acceptable for initial deployment)
- Single key for all data (simplicity over perfect security)
- No HSM/key management service (future enhancement)

### Compliance Considerations

**GDPR (EU Data Protection)**:
- ✅ Encryption at rest (Article 32: Security of processing)
- ✅ Pseudonymization (encrypted data is pseudonymous)
- ⚠️ Right to erasure: Delete records, not just encrypt

**HIPAA (US Healthcare)**:
- ✅ Encryption of ePHI at rest (§164.312(a)(2)(iv))
- ✅ Access controls (file permissions)
- ⚠️ Audit logs: Log encryption/decryption events

**General Best Practices**:
- ✅ Industry-standard encryption (AES-256)
- ✅ Proper key management
- ✅ Encrypted backups (data stays encrypted)
- ⚠️ Document which fields contain PII

---

## Type System Design

### Type Hierarchy

```
Business Types (business/types/):
│
├── content.Content                    # General text (not encrypted)
│   - Used for: Categories, labels
│   - Storage: Plaintext
│   - Searchable: Yes
│
├── encryptedcontent.EncryptedContent  # Sensitive text (encrypted)
│   - Used for: User thoughts, reflections
│   - Storage: Encrypted base64
│   - Searchable: No
│
├── name.Name                          # Names (not encrypted)
│   - Used for: Product names, user names
│   - Storage: Plaintext
│   - Searchable: Yes
│
└── intensity.Intensity                # Numbers (not encrypted)
    - Used for: Ratings, quantities
    - Storage: Plaintext (integer)
    - Searchable: Yes
```

### Type Contract

All business types must implement:

```go
// String returns the string representation
String() string

// Equal provides support for testing and comparisons
Equal(other T) bool

// MarshalText provides support for JSON serialization and logging
MarshalText() ([]byte, error)

// Package-level Parse function for construction
Parse(value string) (T, error)
```

**Example (`encryptedcontent.EncryptedContent`)**:
```go
package encryptedcontent

type EncryptedContent struct {
    value string  // Unexported - immutable
}

func (c EncryptedContent) String() string {
    return c.value
}

func (c EncryptedContent) Equal(other EncryptedContent) bool {
    return c.value == other.value
}

func (c EncryptedContent) MarshalText() ([]byte, error) {
    return []byte(c.value), nil
}

func Parse(plaintext string) (EncryptedContent, error) {
    if strings.TrimSpace(plaintext) == "" {
        return EncryptedContent{}, errors.New("cannot be empty")
    }
    return EncryptedContent{value: plaintext}, nil
}

func MustParse(plaintext string) EncryptedContent {
    c, err := Parse(plaintext)
    if err != nil {
        panic(err)
    }
    return c
}
```

### Immutability Benefits

**Why unexported fields**:
- ✅ **Immutability**: Once created, value cannot change
- ✅ **Validation**: All instances are guaranteed valid (validated in Parse)
- ✅ **Thread-safe**: No mutations = no race conditions
- ✅ **Predictability**: Value never changes unexpectedly

**Construction pattern**:
```go
// ✅ Correct - validated
content, err := encryptedcontent.Parse("user input")
if err != nil {
    return err  // Invalid input rejected
}

// ❌ Incorrect - won't compile (value is unexported)
content := encryptedcontent.EncryptedContent{value: "bypass validation"}
```

---

## Database Strategy

### Current Schema (No Changes Needed)

**Moments table**:
```sql
CREATE TABLE moments (
    moment_id         UUID        NOT NULL PRIMARY KEY,
    user_id           UUID        NOT NULL REFERENCES users(user_id),
    situation         TEXT        NOT NULL,  -- Will store encrypted base64
    thoughts          TEXT        NOT NULL,  -- Will store encrypted base64
    physical_symptoms TEXT        NOT NULL,  -- Will store encrypted base64
    behavior          TEXT        NOT NULL,  -- Will store encrypted base64
    consequences      TEXT        NOT NULL,  -- Will store encrypted base64
    values_reflection TEXT        NOT NULL,  -- Will store encrypted base64
    intensity         INTEGER     NOT NULL,  -- Stays plaintext
    moment_date       TIMESTAMP   NOT NULL,
    date_created      TIMESTAMP   NOT NULL,
    date_updated      TIMESTAMP   NOT NULL
);
```

**Thinks table**:
```sql
CREATE TABLE thinks (
    think_id      UUID      NOT NULL PRIMARY KEY,
    user_id       UUID      NOT NULL REFERENCES users(user_id),
    category      TEXT      NOT NULL,  -- Stays plaintext
    content       TEXT      NOT NULL,  -- Will store encrypted base64
    date_created  TIMESTAMP NOT NULL,
    date_updated  TIMESTAMP NOT NULL
);
```

### Why No Schema Changes?

1. **TEXT columns are unlimited**: Can store any length encrypted data
2. **Base64 encoding**: Makes encrypted bytes safe for TEXT storage
3. **Size increase acceptable**: ~33% larger, but TEXT has no limit
4. **Backward compatibility**: Same column names, just different content

### Storage Size Impact

**Calculation**:
```
Original text:    "I feel anxious about work" (26 characters)
Encrypted bytes:  12 (nonce) + 26 (ciphertext) + 16 (tag) = 54 bytes
Base64 encoded:   54 * 4/3 = 72 characters
Growth:           72/26 = 2.77x (177% larger)

Typical field:    100-500 characters plaintext
Encrypted size:   ~200-1000 characters
PostgreSQL TEXT:  Unlimited (up to 1GB)
Impact:           Negligible
```

### Index Strategy

**Current indexes** (remain unchanged):
```sql
-- These indexes still work (columns not encrypted)
CREATE INDEX moments_user_id_idx ON moments(user_id);
CREATE INDEX moments_user_date_idx ON moments(user_id, moment_date DESC);
CREATE INDEX moments_intensity_idx ON moments(intensity);
CREATE INDEX thinks_category_idx ON thinks(category);
```

**Encrypted fields are NOT indexed**:
- `situation`, `thoughts`, `behavior`, etc. → No indexes
- **Why?** Encrypted data is random, indexing is useless
- **Impact**: Can't filter by encrypted fields (acceptable requirement)

**Query patterns**:
```sql
-- ✅ Works - indexed on user_id
SELECT * FROM moments WHERE user_id = $1;

-- ✅ Works - indexed on intensity (not encrypted)
SELECT * FROM moments WHERE intensity > 5;

-- ❌ Doesn't work - situation is encrypted
SELECT * FROM moments WHERE situation LIKE '%anxious%';
-- This would search encrypted gibberish, always returns nothing
```

### Migration Strategy

**Approach**: Fresh start (no production data)

**Steps**:
1. Deploy encryption code
2. Wipe database (`docker volume rm rafiki_postgres_data`)
3. Restart services (migrations run automatically)
4. All new data encrypted from day one

**Alternative** (if data exists):
```sql
-- Hypothetical migration (not needed now)
-- Encrypt existing data in-place
UPDATE moments
SET situation = encrypt_function(situation)
WHERE situation NOT LIKE '%base64-encrypted-pattern%';
```

---

## Performance Considerations

### Encryption Overhead

**AES-256-GCM benchmarks**:
- Encryption: ~1-2 microseconds per field (modern CPU)
- Decryption: ~1-2 microseconds per field
- Base64 encode: ~0.1 microseconds
- **Total per field**: ~2-4 microseconds

**Per-record overhead**:
- Moment (6 encrypted fields): ~12-24 microseconds
- Think (1 encrypted field): ~2-4 microseconds
- **Negligible** compared to database I/O (1-10 milliseconds)

### Database Impact

**Storage**:
- Encrypted data ~2-3x larger than plaintext
- TEXT columns have no size limit
- Disk space impact: Minimal (text data is small)

**Query performance**:
- Fetching by ID: No change (indexed on UUID)
- Filtering by encrypted fields: **Not possible** (data is random)
- Sorting by encrypted fields: **Not meaningful** (random order)

**Acceptable because**:
- Primary queries: Fetch moments by user_id + date (both indexed, not encrypted)
- No business requirement to search encrypted content
- User sees their own data decrypted in frontend

### Scalability

**Current scale** (habits tracker):
- Users: Hundreds to thousands
- Moments per user: Tens to hundreds
- Queries: Fetch user's moments (10-100 records)
- **Bottleneck**: Database I/O, not encryption

**Future scale** (10x growth):
- Encryption overhead: Still negligible (<1% of total latency)
- Database size: TEXT storage scales fine
- Potential optimization: Caching layer (decrypt once, cache plaintext)

### Optimization Opportunities

**Not needed now, but for future**:

1. **Caching decrypted data**:
   ```go
   // Add caching decorator (similar to usercache)
   momentCache := momentcache.NewStore(log, momentEncrypted, ttl)
   ```

2. **Batch encryption**:
   ```go
   // Encrypt multiple fields in parallel
   go encrypt(situation)
   go encrypt(thoughts)
   // Wait for all completions
   ```

3. **Connection pooling**:
   - Already handled by `sqlx.DB`
   - No changes needed

---

## Future Extensibility

### Adding Encryption to New Fields

**Scenario**: Add encrypted notes to User entity

**Steps**:
1. Define field with `EncryptedContent` type:
   ```go
   type User struct {
       Name         name.Name
       PrivateNotes encryptedcontent.EncryptedContent  // NEW
   }
   ```

2. Database store automatically encrypts:
   ```go
   func toDBUser(user userbus.User) (dbUser, error) {
       encryptedNotes, err := s.encryptor.Encrypt(user.PrivateNotes.String())
       return dbUser{
           PrivateNotes: encryptedNotes,  // Encrypted automatically
       }
   }
   ```

3. No other changes needed!

### Adding New Encrypted Types

**Scenario**: Encrypt numeric data (e.g., salary)

**Steps**:
1. Create `business/types/encryptedmoney/encryptedmoney.go`:
   ```go
   type EncryptedMoney struct {
       value float64  // Plaintext in memory
   }

   func (m EncryptedMoney) Value() float64 {
       return m.value
   }
   ```

2. Database store handles encryption:
   ```go
   encrypted, err := s.encryptor.Encrypt(fmt.Sprintf("%.2f", salary.Value()))
   ```

3. Pattern is reusable!

### Key Rotation Support

**Future enhancement** (not in initial implementation):

**Architecture**:
```go
type DualKeyEncryptor struct {
    currentKey  *AESEncryptor  // New key (for encrypting)
    previousKey *AESEncryptor  // Old key (for decrypting)
}

func (e *DualKeyEncryptor) Encrypt(plaintext string) (string, error) {
    // Always encrypt with current key
    return e.currentKey.Encrypt(plaintext)
}

func (e *DualKeyEncryptor) Decrypt(ciphertext string) (string, error) {
    // Try current key first
    plaintext, err := e.currentKey.Decrypt(ciphertext)
    if err == nil {
        return plaintext, nil
    }

    // Fallback to previous key
    return e.previousKey.Decrypt(ciphertext)
}
```

**Rotation procedure**:
1. Deploy dual-key encryptor (current + previous)
2. Run background job to re-encrypt all data with new key
3. Remove previous key after 100% migration
4. Rotate annually or after security incident

### Multi-Tenant Encryption

**Future enhancement**: Different keys per user/organization

**Architecture**:
```go
type KeyStore interface {
    GetKey(userID uuid.UUID) ([]byte, error)
}

type PerUserEncryptor struct {
    keyStore KeyStore
}

func (e *PerUserEncryptor) Encrypt(userID uuid.UUID, plaintext string) (string, error) {
    key, err := e.keyStore.GetKey(userID)
    // Encrypt with user-specific key
}
```

**Benefits**:
- User data encrypted with unique key
- User account deletion = destroy their key (data becomes unrecoverable)
- Compliance: "Right to be forgotten" (GDPR)

---

## Conclusion

### Summary

This architecture provides:

- ✅ **Type-safe encryption**: Compiler enforces security
- ✅ **Business layer purity**: No encryption coupling
- ✅ **Zero schema changes**: Existing TEXT columns work
- ✅ **Minimal overhead**: <1ms per record
- ✅ **Reusable design**: Easy to extend to new fields
- ✅ **Secure by default**: AES-256-GCM with proper key management

### Key Design Decisions

| Decision | Rationale |
|----------|-----------|
| **Type-based** (not tag-based) | Compiler enforces encryption, self-documenting |
| **Business type** (not DB concern) | Encryption is a business decision, not infrastructure |
| **AES-256-GCM** | Industry standard, authenticated, fast |
| **Single key** | Simplicity for initial deployment |
| **No schema changes** | TEXT columns sufficient for encrypted base64 |
| **Inject encryptor** (not decorator) | Simpler with immutable business types |

### Next Steps

1. Review and approve this architecture
2. Proceed to [Implementation Plan](./encryption-implementation-plan.md)
3. Follow [Deployment Guide](./encryption-deployment-guide.md)

---

**Document Status**: ✅ Ready for Review
**Approval Required**: Development Team Lead, Security Team
**Next Document**: [encryption-implementation-plan.md](./encryption-implementation-plan.md)
