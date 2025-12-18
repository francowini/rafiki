Rafiki - Data Model

**Version 2.0 | December 2025**

---

## 1. Conceptual Context

This data model implements a personal life management system based on the principles of Acceptance and Commitment Therapy (ACT). The system organizes the user's life in hierarchical layers ranging from the most abstract (values) to the most concrete (life visions tied to those values).

### 1.1 Psychological Foundation

The model is based on ACT's premise that a meaningful life requires clarity in two dimensions:

- **Values:** Life directions that define who you want to be (not destinations to reach).
- **Committed Action:** Concrete behaviors aligned with those values.

The Values → Life Visions structure translates ACT philosophy into an operating system for daily life.

### 1.2 System Hierarchy

The model currently implements two levels of abstraction:

1. **Values:** Permanent life directions. Answer "Who do I want to be?"
2. **Life Visions:** Aspirational states of being linked to values. Answer "What does living this value look like?"

---

## 2. Relationship Diagram

```
USERS
    │
    ├──── 1:N ────► VALUES (max 10 per user)
    │                   │
    │                   │ 1:N
    │                   ▼
    │               LIFE_VISIONS
    │
    └──── (other entities: moments, thinks, notifications)
```

### Relationship Summary

| Source | Target | Type | Description |
|--------|--------|------|-------------|
| USERS | VALUES | 1:N | A user has up to 10 values |
| VALUES | LIFE_VISIONS | 1:N | A value can have multiple life visions |

### Key Constraints

- A user can have a maximum of 10 values
- A value cannot be archived if it has active life visions
- Life visions must be linked to a value owned by the same user
- Life visions can only be restored/reassigned to active values

---

## 3. Entity Definitions

### 3.1 USERS

Root entity of the system. Each user has their own complete set of values and life visions.

| Field | Type | Description |
|-------|------|-------------|
| user_id | UUID | Unique user identifier (PK) |
| name | TEXT | User display name |
| email | TEXT | User email (unique) |
| roles | TEXT[] | System roles (admin, user) |
| password_hash | TEXT | Hashed password |
| department | TEXT | Optional department |
| enabled | BOOLEAN | Whether user account is active |
| telegram_chat_id | BIGINT | Telegram chat ID for notifications |
| telegram_enabled | BOOLEAN | Whether Telegram notifications are enabled |
| date_created | TIMESTAMP | Account creation date |
| date_updated | TIMESTAMP | Last record update |

---

### 3.2 VALUES

Represent the user's life directions: who they want to be and what impact they want to have. They are the "north star" that guides all decisions. In ACT, values are not "achieved" but "lived".

**Key characteristics:**
- Maximum 10 values per user
- Organized by facet (life domain based on ACT therapy)
- User-controlled display order (1-10)
- Support soft delete (archive) with restore capability

| Field | Type | Description |
|-------|------|-------------|
| value_id | UUID | Unique value identifier (PK) |
| user_id | UUID | FK → USERS |
| content | TEXT | Value statement (encrypted, 3-200 chars) |
| facet | TEXT | Life domain categorization (see Facet Types below) |
| display_order | INTEGER | User-controlled priority ranking (1-10, 1=highest) |
| status | TEXT | 'active' \| 'archived' |
| archived_at | TIMESTAMP | Archive date (NULL if active) |
| date_created | TIMESTAMP | Creation date |
| date_updated | TIMESTAMP | Last update |

**Facet Types (based on ACT therapy):**
- `family_relationships` - Family and close relationships
- `friendships_social` - Friendships and social connections
- `romantic_relationships` - Romantic partnerships
- `work_career` - Professional life and career
- `education_growth` - Learning and personal development
- `recreation_leisure` - Hobbies and leisure activities
- `spirituality_meaning` - Spirituality and life meaning
- `health_wellbeing` - Physical and mental health
- `community_citizenship` - Community involvement

**Example:**
```
value_id: "a1b2c3d4-..."
user_id: "user-uuid-..."
content: "Being present and supportive for my family"
facet: "family_relationships"
display_order: 1
status: "active"
archived_at: NULL
```

**Business Rules:**
- `ErrMaxValues`: Cannot create more than 10 values per user
- `ErrDuplicateOrder`: Display order must be unique within user's active values
- `ErrHasActiveLifeVisions`: Cannot archive value with active life visions
- `ErrAlreadyArchived`: Cannot archive an already archived value
- `ErrNotArchived`: Cannot restore a value that is not archived

---

### 3.3 LIFE_VISIONS

Detailed visualization of what living a specific value looks like. Unlike abstract values, life visions are descriptive and aspirational. They help concretize what a value means in practice.

**Key characteristics:**
- Each life vision is linked to exactly one value
- Content describes an aspirational state of being
- Can be reassigned to a different value (same user only)
- Support soft delete (archive) with restore capability

| Field | Type | Description |
|-------|------|-------------|
| life_vision_id | UUID | Unique life vision identifier (PK) |
| user_id | UUID | FK → USERS |
| value_id | UUID | FK → VALUES (must be active for restore/reassign) |
| content | TEXT | Vision statement (encrypted, 10-500 chars) |
| status | TEXT | 'active' \| 'archived' |
| archived_at | TIMESTAMP | Archive date (NULL if active) |
| date_created | TIMESTAMP | Creation date |
| date_updated | TIMESTAMP | Last update |

**Example:**
```
life_vision_id: "e5f6g7h8-..."
user_id: "user-uuid-..."
value_id: "a1b2c3d4-..."
content: "Every evening, I sit down with my family for dinner without
          distractions. We share stories about our day, laugh together,
          and I feel deeply connected to each person at the table."
status: "active"
archived_at: NULL
```

**Business Rules:**
- `ErrNotValueOwner`: Cannot create/update life vision for a value you don't own
- `ErrTargetValueNotActive`: Cannot restore/reassign to an archived value
- `ErrAlreadyArchived`: Cannot archive an already archived life vision
- `ErrNotArchived`: Cannot restore a life vision that is not archived

---

## 4. Example Queries

### 4.1 Get user's active values ordered by priority

```sql
SELECT
  value_id,
  content,
  facet,
  display_order
FROM values
WHERE user_id = $1
  AND status = 'active'
ORDER BY display_order ASC;
```

### 4.2 Get life visions grouped by value

```sql
SELECT
  v.value_id,
  v.content AS value_content,
  v.facet,
  lv.life_vision_id,
  lv.content AS vision_content
FROM values v
LEFT JOIN life_visions lv ON lv.value_id = v.value_id AND lv.status = 'active'
WHERE v.user_id = $1
  AND v.status = 'active'
ORDER BY v.display_order, lv.date_created;
```

### 4.3 Check if value can be archived

```sql
SELECT EXISTS (
  SELECT 1 FROM life_visions
  WHERE value_id = $1
    AND status = 'active'
) AS has_active_life_visions;
```

### 4.4 Get notification content (values + visions for daily messages)

```sql
SELECT
  user_id,
  value_id,
  value_content,
  value_facet,
  value_order,
  life_vision_id,
  life_vision_content
FROM view_notification_content
WHERE user_id = $1
ORDER BY value_order;
```

---

## 5. Soft Delete Strategy

### 5.1 Archive vs Hard Delete

Both `values` and `life_visions` use soft delete (`archived_at`) to preserve historical integrity:

- **Archive**: Sets `status = 'archived'` and `archived_at = NOW()`
- **Restore**: Sets `status = 'active'` and `archived_at = NULL`
- **Hard Delete**: Only available for cleanup; removes record permanently

### 5.2 Archive Constraints

**Values:**
- Cannot archive if has active life visions (enforced by database trigger)
- User must first reassign or archive all related life visions

**Life Visions:**
- Can be archived independently
- Cannot be restored to an archived value

### 5.3 Cascade Behavior

- `values.value_id` → `life_visions.value_id`: ON DELETE RESTRICT
- This prevents accidental deletion of values that have life visions
- User must explicitly handle life visions before deleting a value

---

## 6. API Operations

### 6.1 Values API

| Operation | Endpoint | Description |
|-----------|----------|-------------|
| Create | POST /v1/values | Create new value (respects max 10 limit) |
| List | GET /v1/values | List values with optional filters |
| Get | GET /v1/values/{id} | Get single value |
| Update | PUT /v1/values/{id} | Update value content/facet/order |
| Delete | DELETE /v1/values/{id} | Hard delete value |
| Archive | PUT /v1/values/{id}/archive | Soft delete value |
| Restore | PUT /v1/values/{id}/restore | Restore archived value |
| Reorder | POST /v1/values/reorder | Atomically reorder multiple values |

### 6.2 Life Visions API

| Operation | Endpoint | Description |
|-----------|----------|-------------|
| Create | POST /v1/lifevisions | Create new life vision |
| List | GET /v1/lifevisions | List life visions with optional filters |
| Get | GET /v1/lifevisions/{id} | Get single life vision |
| Update | PUT /v1/lifevisions/{id} | Update life vision content |
| Delete | DELETE /v1/lifevisions/{id} | Hard delete life vision |
| Archive | PUT /v1/lifevisions/{id}/archive | Soft delete life vision |
| Restore | PUT /v1/lifevisions/{id}/restore | Restore archived life vision |
| Reassign | PUT /v1/lifevisions/{id}/reassign | Move to different value |
| By Value | GET /v1/values/{id}/lifevisions | List life visions for a value |

---

## 7. Future Implementation (Planned)

The following entities are planned for future implementation to complete the ACT-based hierarchy:

### 7.1 ROLES (Planned)

Life facets or "hats" from which the user lives their life. Each role groups related objectives and connects with one or more values.

**Planned fields:**
- role_id, user_id, name, type, facet, description, order, archived_at

**Planned types:**
- Personal (physical, mental, spiritual)
- Relational (family, partner, friend)
- Professional (employee, leader, mentor)
- Hobby/Cause (volunteer, athlete, artist)

### 7.2 ROLES_VALUES (Planned)

N:M junction table connecting roles with values. A role can manifest multiple values, and a value can be expressed through multiple roles.

### 7.3 OBJECTIVES (Planned)

Concrete, measurable, and time-bound actions. Each role has between 1 and 3 annual objectives.

**Planned tracking types:**
- Result: Measures a final achievement (e.g., "Read 35 books")
- Frequency: Measures habit consistency (e.g., "Meditate 80% of days")

### 7.4 OBJECTIVE_RECORDS (Planned)

Daily records for frequency-type objectives. Allows tracking day by day whether habits were completed.

---

## 8. Day-to-Day Data Flow

```
MORNING:
  User opens app
    → Load VALUES (reading, inspiration)
    → Load LIFE_VISIONS grouped by value
    → Display daily motivation message via Telegram

DURING THE DAY:
  User reflects on values
    → Review life visions for guidance
    → (Future: Track objectives and habits)

EVENING:
  User receives evening notification
    → Values + life visions in Telegram message
    → Opportunity for reflection
```

---

*Document updated to reflect Rafiki v2.0 implementation (December 2025)*
