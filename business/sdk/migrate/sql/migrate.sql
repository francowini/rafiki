-- Version: 1.01
-- Description: Create table users
CREATE TABLE users (
	user_id       UUID        NOT NULL,
	name          TEXT        NOT NULL,
	email         TEXT UNIQUE NOT NULL,
	roles         TEXT[]      NOT NULL,
	password_hash TEXT        NOT NULL,
    department    TEXT        NULL,
    enabled       BOOLEAN     NOT NULL,
	date_created  TIMESTAMP   NOT NULL,
	date_updated  TIMESTAMP   NOT NULL,

	PRIMARY KEY (user_id)
);


-- Version: 1.02
-- Description: Create table thinks

CREATE TABLE thinks (
    think_id      UUID      NOT NULL,
    user_id      UUID       NOT NULL,
    category      TEXT      NOT NULL,
    content       TEXT      NOT NULL,
    date_created  TIMESTAMP NOT NULL,
    date_updated  TIMESTAMP NOT NULL,

    PRIMARY KEY (think_id),
    FOREIGN KEY (user_id) REFERENCES users(user_id) ON DELETE CASCADE  

);

-- Create index for faster retrieval by date
CREATE INDEX thinks_date_created_idx ON thinks(date_created DESC);
CREATE INDEX thinks_category_idx ON thinks(category);


-- Version: 1.03
-- Description: Create table moments for emotional tracking

CREATE TABLE moments (
    moment_id         UUID        NOT NULL,
    user_id           UUID        NOT NULL,
    moment_date       TIMESTAMP   NOT NULL,
    situation         TEXT        NOT NULL,
    thoughts          TEXT        NOT NULL,
    physical_symptoms TEXT        NOT NULL,
    behavior          TEXT        NOT NULL,
    consequences      TEXT        NOT NULL,
    values_reflection TEXT        NOT NULL,
    intensity         INTEGER     NOT NULL CHECK (intensity >= 0 AND intensity <= 10),
    date_created      TIMESTAMP   NOT NULL,
    date_updated      TIMESTAMP   NOT NULL,

    PRIMARY KEY (moment_id),
    FOREIGN KEY (user_id) REFERENCES users(user_id) ON DELETE CASCADE
);

CREATE INDEX moments_user_id_idx ON moments(user_id);
CREATE INDEX moments_user_date_idx ON moments(user_id, moment_date DESC);
CREATE INDEX moments_date_created_idx ON moments(date_created DESC);
CREATE INDEX moments_intensity_idx ON moments(intensity);

COMMENT ON TABLE moments IS 'Tracks emotional/difficult moments for psychological self-observation';
COMMENT ON COLUMN moments.moment_date IS 'When the observed moment actually occurred (user can backdate)';
COMMENT ON COLUMN moments.intensity IS 'Distress intensity on 0-10 scale';


-- Version: 1.04
-- Description: Create values table for personal values tracking

-- Create values table
CREATE TABLE values (
    value_id       UUID        NOT NULL,
    user_id        UUID        NOT NULL,
    content        TEXT        NOT NULL,
    facet          TEXT        NOT NULL,
    display_order  INTEGER     NOT NULL CHECK (display_order BETWEEN 1 AND 10),
    date_created   TIMESTAMP   NOT NULL,
    date_updated   TIMESTAMP   NOT NULL,

    PRIMARY KEY (value_id),
    FOREIGN KEY (user_id) REFERENCES users(user_id) ON DELETE CASCADE
);

-- Performance indexes
CREATE INDEX values_user_id_idx ON values(user_id);
CREATE INDEX values_facet_idx ON values(facet);

-- Unique constraint to prevent duplicate display_order per user
CREATE UNIQUE INDEX values_user_order_unique_idx ON values(user_id, display_order);

-- Documentation
COMMENT ON TABLE values IS 'Stores user core personal values with life facet categorization (max 10 per user)';
COMMENT ON COLUMN values.content IS 'Encrypted value statement (plaintext validated as 3-200 chars in business layer)';
COMMENT ON COLUMN values.facet IS 'Life domain categorization based on ACT therapy';
COMMENT ON COLUMN values.display_order IS 'User-controlled priority ranking (1=highest)';


-- Version: 1.05
-- Description: Remove unique constraint on display_order to allow atomic reordering

-- Drop the unique index that prevents atomic batch reordering within a transaction.
--
-- Trade-off: DB no longer enforces unique (user_id, display_order), so business layer
-- validates no-duplicates for ALL write operations:
--
-- 1. Create: valuebus.Create() queries existing values, rejects duplicate display_order
-- 2. Update: valuebus.Update() checks for conflicts when display_order changes
-- 3. Reorder: valuebus.Reorder() validates via request payload + transactional batch
--
-- Additional safeguards:
-- - display_order range (1-10) validated at parse time via displayorder.Parse
-- - Concurrent requests could theoretically race, but:
--   * Each operation queries current state before writing
--   * Reorder uses explicit transaction (all-or-nothing)
--   * Worst case: one request fails, UI refreshes from server state
--
-- This enables atomic batch reordering where swapping positions (e.g., 1↔2)
-- would otherwise violate the unique constraint during the intermediate state.
DROP INDEX IF EXISTS values_user_order_unique_idx;
