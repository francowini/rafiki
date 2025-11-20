# Backend Implementation - Database Schema

**Task Category**: Database
**Estimated Time**: 4-6 hours
**Prerequisites**: None (first task)
**Dependencies**: None

---

## Overview

This document describes the database schema changes for Telegram integration. The schema is **generic and reusable** - it supports moments now, and habits/goals in the future without schema changes.

---

## Migration Version: 1.04

**File**: `business/sdk/migrate/sql/migrate.sql`

### New Tables

#### 1. telegram_users

**Purpose**: Store Telegram user data separately from main users table (normalized design)

**Why Separate Table**:
- Clean separation of concerns (Telegram data isolated)
- Easy to add WhatsApp/Discord in future (similar pattern)
- Can drop table without affecting users table

```sql
-- Version: 1.04
-- Description: Telegram integration tables

-- Separate table for Telegram users (not part of main users table)
CREATE TABLE telegram_users (
    telegram_user_id   BIGINT       PRIMARY KEY,
    username           TEXT         NULL,
    first_name         TEXT         NULL,
    last_name          TEXT         NULL,
    language_code      TEXT         NULL,  -- e.g., "en", "es"
    is_bot             BOOLEAN      NOT NULL DEFAULT false,
    date_created       TIMESTAMP    NOT NULL,
    date_updated       TIMESTAMP    NOT NULL
);

CREATE INDEX telegram_users_username_idx ON telegram_users(username)
    WHERE username IS NOT NULL;

COMMENT ON TABLE telegram_users IS 'Telegram user data (separate from main users table)';
COMMENT ON COLUMN telegram_users.telegram_user_id IS 'Telegram''s unique user ID (from Telegram API)';
```

**Rationale**:
- `telegram_user_id` is the primary key (Telegram's unique ID)
- `username` can be NULL (not all Telegram users have usernames)
- `language_code` stored for future multi-language support (unused for now)
- `is_bot` prevents bots from linking accounts

#### 2. user_telegram_links

**Purpose**: Link Rafiki users to Telegram users (many-to-one relationship)

**Why Separate Table**:
- Clean separation: user linking is different from user data
- Easy to query: "Which Rafiki user is linked to telegram_user_id X?"
- Easy to unlink: Just delete row

```sql
-- Linking table: Rafiki user ↔ Telegram user
CREATE TABLE user_telegram_links (
    user_id            UUID         NOT NULL,
    telegram_user_id   BIGINT       NOT NULL,
    linked_at          TIMESTAMP    NOT NULL,
    date_created       TIMESTAMP    NOT NULL,
    date_updated       TIMESTAMP    NOT NULL,

    PRIMARY KEY (user_id),
    UNIQUE (telegram_user_id),
    FOREIGN KEY (user_id) REFERENCES users(user_id) ON DELETE CASCADE,
    FOREIGN KEY (telegram_user_id) REFERENCES telegram_users(telegram_user_id) ON DELETE CASCADE
);

CREATE INDEX user_telegram_links_telegram_id_idx ON user_telegram_links(telegram_user_id);

COMMENT ON TABLE user_telegram_links IS 'Links Rafiki users to Telegram users (1:1)';
```

**Constraints**:
- One Rafiki user can link to ONE Telegram account (PRIMARY KEY on user_id)
- One Telegram account can link to ONE Rafiki user (UNIQUE on telegram_user_id)
- Cascade delete: If Rafiki user deleted → link deleted

#### 3. telegram_link_codes

**Purpose**: Temporary link codes for account linking (expires after 5 minutes)

**Why Separate Table**:
- Ephemeral data (deleted after use or expiry)
- Should not pollute users table
- Easy to cleanup (delete expired codes)

```sql
-- Temporary link codes for account linking
CREATE TABLE telegram_link_codes (
    code               TEXT         PRIMARY KEY,
    user_id            UUID         NOT NULL,
    expires_at         TIMESTAMP    NOT NULL,
    consumed_at        TIMESTAMP    NULL,
    consumed_by        BIGINT       NULL,  -- telegram_user_id that consumed it
    date_created       TIMESTAMP    NOT NULL,

    FOREIGN KEY (user_id) REFERENCES users(user_id) ON DELETE CASCADE
);

CREATE INDEX telegram_link_codes_expiry_idx ON telegram_link_codes(expires_at)
    WHERE consumed_at IS NULL;

CREATE INDEX telegram_link_codes_user_id_idx ON telegram_link_codes(user_id);

COMMENT ON TABLE telegram_link_codes IS 'Temporary codes for linking Telegram accounts (5-minute TTL)';
COMMENT ON COLUMN telegram_link_codes.code IS 'Random 16-character code (e.g., "RAFIKI-2X9K4P7M")';
```

**Lifecycle**:
1. User requests link code → INSERT with `expires_at = NOW() + 5 minutes`
2. User sends `/link CODE` in Telegram → UPDATE `consumed_at`, `consumed_by`
3. Cleanup job runs hourly → DELETE WHERE `expires_at < NOW()`

#### 4. conversation_states

**Purpose**: Store active Telegram conversations (generic state machine)

**Why Generic**: Same table supports moments, habits, goals - just different `conversation_type`

```sql
-- Generic conversation state machine
CREATE TABLE conversation_states (
    conversation_id    UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    telegram_user_id   BIGINT       NOT NULL,
    user_id            UUID         NOT NULL,
    conversation_type  TEXT         NOT NULL,  -- 'moment', 'habit', 'goal'
    current_step       TEXT         NOT NULL,  -- 'moment:awaiting_situation', 'habit:awaiting_name'
    data               JSONB        NOT NULL DEFAULT '{}',
    last_message_id    INTEGER      NULL,
    last_activity      TIMESTAMP    NOT NULL,
    date_created       TIMESTAMP    NOT NULL,
    date_updated       TIMESTAMP    NOT NULL,

    FOREIGN KEY (user_id) REFERENCES users(user_id) ON DELETE CASCADE,
    FOREIGN KEY (telegram_user_id) REFERENCES telegram_users(telegram_user_id) ON DELETE CASCADE
);

CREATE UNIQUE INDEX conversation_states_telegram_user_idx
    ON conversation_states(telegram_user_id)
    WHERE current_step != 'completed' AND current_step != 'cancelled';

CREATE INDEX conversation_states_last_activity_idx ON conversation_states(last_activity);
CREATE INDEX conversation_states_user_type_idx ON conversation_states(user_id, conversation_type);

COMMENT ON TABLE conversation_states IS 'Active Telegram conversations (generic state machine)';
COMMENT ON COLUMN conversation_states.current_step IS 'Namespaced state (e.g., "moment:awaiting_situation")';
COMMENT ON COLUMN conversation_states.data IS 'Accumulated conversation data (JSONB for flexibility)';
```

**State Namespacing**:
- Moments: `moment:awaiting_situation`, `moment:awaiting_thoughts`, etc.
- Habits (future): `habit:awaiting_name`, `habit:awaiting_frequency`, etc.
- Special states: `completed`, `cancelled`

**Unique Constraint**: One active conversation per Telegram user (prevents overlapping conversations)

**Cleanup**: Conversations with `last_activity > 5 minutes` are considered abandoned and deleted

### Updated Tables

#### 5. moments (Add source tracking)

**Purpose**: Track which moments came from Telegram vs. web

```sql
-- Extend moments table for source tracking
ALTER TABLE moments
  ADD COLUMN source TEXT NOT NULL DEFAULT 'web',
  ADD COLUMN source_metadata JSONB NULL;

CREATE INDEX moments_source_idx ON moments(user_id, source, date_created DESC);

COMMENT ON COLUMN moments.source IS 'Source of moment creation: "web" or "telegram"';
COMMENT ON COLUMN moments.source_metadata IS 'Source-specific data (e.g., {"message_id": 12345})';
```

**Examples**:
```json
// Web-created moment
{ "source": "web", "source_metadata": null }

// Telegram-created moment
{
  "source": "telegram",
  "source_metadata": {
    "message_id": 12345,
    "conversation_id": "uuid-here",
    "duration_seconds": 45
  }
}
```

---

## Complete Migration SQL

**File**: `business/sdk/migrate/sql/migrate.sql`

Add this to the end of the file:

```sql
-- ============================================================================
-- Version 1.04: Telegram Integration
-- ============================================================================

-- 1. Telegram users (separate from main users table)
CREATE TABLE IF NOT EXISTS telegram_users (
    telegram_user_id   BIGINT       PRIMARY KEY,
    username           TEXT         NULL,
    first_name         TEXT         NULL,
    last_name          TEXT         NULL,
    language_code      TEXT         NULL,
    is_bot             BOOLEAN      NOT NULL DEFAULT false,
    date_created       TIMESTAMP    NOT NULL,
    date_updated       TIMESTAMP    NOT NULL
);

CREATE INDEX IF NOT EXISTS telegram_users_username_idx
    ON telegram_users(username) WHERE username IS NOT NULL;

-- 2. User-Telegram linking table
CREATE TABLE IF NOT EXISTS user_telegram_links (
    user_id            UUID         NOT NULL,
    telegram_user_id   BIGINT       NOT NULL,
    linked_at          TIMESTAMP    NOT NULL,
    date_created       TIMESTAMP    NOT NULL,
    date_updated       TIMESTAMP    NOT NULL,

    PRIMARY KEY (user_id),
    UNIQUE (telegram_user_id),
    FOREIGN KEY (user_id) REFERENCES users(user_id) ON DELETE CASCADE,
    FOREIGN KEY (telegram_user_id) REFERENCES telegram_users(telegram_user_id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS user_telegram_links_telegram_id_idx
    ON user_telegram_links(telegram_user_id);

-- 3. Temporary link codes (5-minute TTL)
CREATE TABLE IF NOT EXISTS telegram_link_codes (
    code               TEXT         PRIMARY KEY,
    user_id            UUID         NOT NULL,
    expires_at         TIMESTAMP    NOT NULL,
    consumed_at        TIMESTAMP    NULL,
    consumed_by        BIGINT       NULL,
    date_created       TIMESTAMP    NOT NULL,

    FOREIGN KEY (user_id) REFERENCES users(user_id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS telegram_link_codes_expiry_idx
    ON telegram_link_codes(expires_at) WHERE consumed_at IS NULL;

CREATE INDEX IF NOT EXISTS telegram_link_codes_user_id_idx
    ON telegram_link_codes(user_id);

-- 4. Generic conversation state machine
CREATE TABLE IF NOT EXISTS conversation_states (
    conversation_id    UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    telegram_user_id   BIGINT       NOT NULL,
    user_id            UUID         NOT NULL,
    conversation_type  TEXT         NOT NULL,
    current_step       TEXT         NOT NULL,
    data               JSONB        NOT NULL DEFAULT '{}',
    last_message_id    INTEGER      NULL,
    last_activity      TIMESTAMP    NOT NULL,
    date_created       TIMESTAMP    NOT NULL,
    date_updated       TIMESTAMP    NOT NULL,

    FOREIGN KEY (user_id) REFERENCES users(user_id) ON DELETE CASCADE,
    FOREIGN KEY (telegram_user_id) REFERENCES telegram_users(telegram_user_id) ON DELETE CASCADE
);

CREATE UNIQUE INDEX IF NOT EXISTS conversation_states_telegram_user_idx
    ON conversation_states(telegram_user_id)
    WHERE current_step != 'completed' AND current_step != 'cancelled';

CREATE INDEX IF NOT EXISTS conversation_states_last_activity_idx
    ON conversation_states(last_activity);

CREATE INDEX IF NOT EXISTS conversation_states_user_type_idx
    ON conversation_states(user_id, conversation_type);

-- 5. Extend moments table for source tracking
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_name = 'moments' AND column_name = 'source'
    ) THEN
        ALTER TABLE moments ADD COLUMN source TEXT NOT NULL DEFAULT 'web';
    END IF;

    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_name = 'moments' AND column_name = 'source_metadata'
    ) THEN
        ALTER TABLE moments ADD COLUMN source_metadata JSONB NULL;
    END IF;
END $$;

CREATE INDEX IF NOT EXISTS moments_source_idx
    ON moments(user_id, source, date_created DESC);

-- Comments
COMMENT ON TABLE telegram_users IS 'Telegram user data (separate from main users table)';
COMMENT ON TABLE user_telegram_links IS 'Links Rafiki users to Telegram users (1:1)';
COMMENT ON TABLE telegram_link_codes IS 'Temporary codes for linking Telegram accounts (5-minute TTL)';
COMMENT ON TABLE conversation_states IS 'Active Telegram conversations (generic state machine)';
COMMENT ON COLUMN moments.source IS 'Source of moment creation: "web" or "telegram"';
COMMENT ON COLUMN moments.source_metadata IS 'Source-specific data (e.g., {"message_id": 12345})';
```

---

## Testing the Migration

### Test 1: Idempotency

Run migration twice - should not error:

```bash
# First run
make migrate

# Second run (should be safe)
make migrate
```

### Test 2: Verify Tables Created

```sql
-- Check all tables exist
SELECT table_name FROM information_schema.tables
WHERE table_name IN (
    'telegram_users',
    'user_telegram_links',
    'telegram_link_codes',
    'conversation_states'
)
ORDER BY table_name;

-- Should return 4 rows
```

### Test 3: Verify Indexes

```sql
-- Check indexes created
SELECT indexname FROM pg_indexes
WHERE tablename LIKE 'telegram%' OR tablename = 'conversation_states'
ORDER BY tablename, indexname;

-- Should return 8+ indexes
```

### Test 4: Verify Foreign Keys

```sql
-- Check foreign key constraints
SELECT
    tc.table_name,
    kcu.column_name,
    ccu.table_name AS foreign_table_name
FROM information_schema.table_constraints AS tc
JOIN information_schema.key_column_usage AS kcu
  ON tc.constraint_name = kcu.constraint_name
JOIN information_schema.constraint_column_usage AS ccu
  ON ccu.constraint_name = tc.constraint_name
WHERE tc.constraint_type = 'FOREIGN KEY'
  AND tc.table_name LIKE '%telegram%';

-- Should return foreign key relationships
```

### Test 5: Insert Test Data

```sql
-- Test data insertion (rollback after)
BEGIN;

-- Insert test Telegram user
INSERT INTO telegram_users (telegram_user_id, username, first_name, date_created, date_updated)
VALUES (123456789, 'test_user', 'Test', NOW(), NOW());

-- Insert test link code
INSERT INTO telegram_link_codes (code, user_id, expires_at, date_created)
VALUES ('TEST-CODE-1234', (SELECT user_id FROM users LIMIT 1), NOW() + INTERVAL '5 minutes', NOW());

-- Insert test conversation
INSERT INTO conversation_states (
    telegram_user_id,
    user_id,
    conversation_type,
    current_step,
    last_activity,
    date_created,
    date_updated
)
VALUES (
    123456789,
    (SELECT user_id FROM users LIMIT 1),
    'moment',
    'moment:awaiting_situation',
    NOW(),
    NOW(),
    NOW()
);

-- Verify inserts
SELECT * FROM telegram_users;
SELECT * FROM telegram_link_codes;
SELECT * FROM conversation_states;

ROLLBACK;  -- Don't commit test data
```

---

## Rollback Plan

If migration fails or needs to be rolled back:

**File**: `scripts/rollback-telegram-migration.sql`

```sql
-- Rollback script for Telegram migration (Version 1.04)
-- WARNING: This deletes all Telegram data

BEGIN;

-- Drop tables in reverse order (respecting foreign keys)
DROP TABLE IF EXISTS conversation_states CASCADE;
DROP TABLE IF EXISTS telegram_link_codes CASCADE;
DROP TABLE IF EXISTS user_telegram_links CASCADE;
DROP TABLE IF EXISTS telegram_users CASCADE;

-- Remove columns from moments table
ALTER TABLE moments DROP COLUMN IF EXISTS source;
ALTER TABLE moments DROP COLUMN IF EXISTS source_metadata;

COMMIT;
```

**Run rollback**:
```bash
psql -U rafiki -d rafiki < scripts/rollback-telegram-migration.sql
```

---

## Database Access Layer (Next Step)

After migration is complete, implement database access in:
- `business/sdk/sqldb/telegramdb/telegram_users.go`
- `business/sdk/sqldb/telegramdb/link_codes.go`
- `business/sdk/sqldb/telegramdb/conversation_states.go`

See [02-backend-telegramdb.md](./02-backend-telegramdb.md) for implementation details.

---

## Checklist

- [ ] Add migration SQL to `migrate.sql`
- [ ] Test migration on local database
- [ ] Test idempotency (run twice)
- [ ] Verify all tables created
- [ ] Verify all indexes created
- [ ] Verify all foreign keys work
- [ ] Test sample data insertion
- [ ] Create rollback script
- [ ] Test rollback script (on copy of database)
- [ ] Commit migration to git
- [ ] Deploy to staging environment
- [ ] Verify migration on staging
- [ ] Document any issues encountered

---

**Status**: ⏭️ Ready for Implementation
**Next Task**: [02-backend-telegramdb.md](./02-backend-telegramdb.md) - Database access layer
