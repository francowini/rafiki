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


-- Version: 1.06
-- Description: Create life_visions table for aspirational states of being

CREATE TABLE life_visions (
    life_vision_id  UUID        NOT NULL,
    user_id         UUID        NOT NULL,
    value_id        UUID        NOT NULL,
    content         TEXT        NOT NULL,
    date_created    TIMESTAMP   NOT NULL,
    date_updated    TIMESTAMP   NOT NULL,

    PRIMARY KEY (life_vision_id),
    FOREIGN KEY (user_id) REFERENCES users(user_id) ON DELETE CASCADE,
    FOREIGN KEY (value_id) REFERENCES values(value_id) ON DELETE CASCADE
);

-- Performance indexes
CREATE INDEX life_visions_user_id_idx ON life_visions(user_id);
CREATE INDEX life_visions_value_id_idx ON life_visions(value_id);
CREATE INDEX life_visions_date_created_idx ON life_visions(date_created DESC);

COMMENT ON TABLE life_visions IS 'Aspirational states of being linked to personal values';
COMMENT ON COLUMN life_visions.content IS 'Encrypted vision statement (plaintext validated as 10-500 chars in business layer)';


-- Version: 1.07
-- Description: Create view for combined moments and thinks export

CREATE VIEW view_export_items AS
-- Select moments with type identifier
SELECT
    moment_id AS item_id,
    user_id,
    'moment' AS item_type,
    moment_date AS item_date,
    situation,
    thoughts,
    physical_symptoms,
    behavior,
    consequences,
    values_reflection,
    intensity,
    NULL AS category,
    NULL AS content,
    date_created
FROM moments

UNION ALL

-- Select thinks with type identifier
SELECT
    think_id AS item_id,
    user_id,
    'think' AS item_type,
    date_created AS item_date,
    NULL AS situation,
    NULL AS thoughts,
    NULL AS physical_symptoms,
    NULL AS behavior,
    NULL AS consequences,
    NULL AS values_reflection,
    NULL AS intensity,
    category,
    content,
    date_created
FROM thinks;


-- Version: 1.08
-- Description: Add River Queue tables for background job processing
--
-- Rollback Strategy:
--   Before rollback: Ensure all River jobs are drained or accept backlog loss.
--   To rollback, drop tables in this order (respects foreign keys):
--     DROP TABLE IF EXISTS river_client_queue;
--     DROP TABLE IF EXISTS river_client;
--     DROP TABLE IF EXISTS river_leader;
--     DROP TABLE IF EXISTS river_queue;
--     DROP TABLE IF EXISTS river_job;
--     DROP TYPE IF EXISTS river_job_state;
--
-- Version Compatibility:
--   Requires River Go client v0.15.0 or later (schema version 006).
--   See: https://riverqueue.com/docs/migrations

-- River job state enum
CREATE TYPE river_job_state AS ENUM(
  'available',
  'cancelled',
  'completed',
  'discarded',
  'pending',
  'retryable',
  'running',
  'scheduled'
);

-- Main jobs table
CREATE TABLE river_job(
  id bigserial PRIMARY KEY,
  state river_job_state NOT NULL DEFAULT 'available',
  attempt smallint NOT NULL DEFAULT 0,
  max_attempts smallint NOT NULL,
  attempted_at timestamptz,
  created_at timestamptz NOT NULL DEFAULT NOW(),
  finalized_at timestamptz,
  scheduled_at timestamptz NOT NULL DEFAULT NOW(),
  priority smallint NOT NULL DEFAULT 1,
  args jsonb NOT NULL,
  attempted_by text[],
  errors jsonb[],
  kind text NOT NULL,
  metadata jsonb NOT NULL DEFAULT '{}',
  queue text NOT NULL DEFAULT 'default',
  tags varchar(255)[] NOT NULL DEFAULT '{}',
  unique_key bytea,
  unique_states BIT(8),

  CONSTRAINT finalized_or_finalized_at_null CHECK (
    (finalized_at IS NULL AND state NOT IN ('cancelled', 'completed', 'discarded')) OR
    (finalized_at IS NOT NULL AND state IN ('cancelled', 'completed', 'discarded'))
  ),
  CONSTRAINT max_attempts_is_positive CHECK (max_attempts > 0),
  CONSTRAINT priority_in_range CHECK (priority >= 1 AND priority <= 4),
  CONSTRAINT queue_length CHECK (char_length(queue) > 0 AND char_length(queue) < 128),
  CONSTRAINT kind_length CHECK (char_length(kind) > 0 AND char_length(kind) < 128)
);

-- Performance indexes
CREATE INDEX river_job_kind ON river_job USING btree(kind);
CREATE INDEX river_job_state_and_finalized_at_index ON river_job USING btree(state, finalized_at) WHERE finalized_at IS NOT NULL;
CREATE INDEX river_job_prioritized_fetching_index ON river_job USING btree(state, queue, priority, scheduled_at, id);
CREATE INDEX river_job_args_index ON river_job USING GIN(args);
CREATE INDEX river_job_metadata_index ON river_job USING GIN(metadata);

-- Unique job deduplication function and index
CREATE OR REPLACE FUNCTION river_job_state_in_bitmask(bitmask BIT(8), state river_job_state)
RETURNS boolean
LANGUAGE SQL
IMMUTABLE
AS $$
    SELECT CASE state
        WHEN 'available' THEN get_bit(bitmask, 7)
        WHEN 'cancelled' THEN get_bit(bitmask, 6)
        WHEN 'completed' THEN get_bit(bitmask, 5)
        WHEN 'discarded' THEN get_bit(bitmask, 4)
        WHEN 'pending'   THEN get_bit(bitmask, 3)
        WHEN 'retryable' THEN get_bit(bitmask, 2)
        WHEN 'running'   THEN get_bit(bitmask, 1)
        WHEN 'scheduled' THEN get_bit(bitmask, 0)
        ELSE 0
    END = 1;
$$;

CREATE UNIQUE INDEX river_job_unique_idx ON river_job (unique_key)
    WHERE unique_key IS NOT NULL
      AND unique_states IS NOT NULL
      AND river_job_state_in_bitmask(unique_states, state);

-- Leader election table (unlogged for performance)
CREATE UNLOGGED TABLE river_leader(
    elected_at timestamptz NOT NULL,
    expires_at timestamptz NOT NULL,
    leader_id text NOT NULL,
    name text PRIMARY KEY DEFAULT 'default',
    CONSTRAINT name_length CHECK (name = 'default'),
    CONSTRAINT leader_id_length CHECK (char_length(leader_id) > 0 AND char_length(leader_id) < 128)
);

-- Queue metadata table
CREATE TABLE river_queue (
    name text PRIMARY KEY NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    metadata jsonb NOT NULL DEFAULT '{}' ::jsonb,
    paused_at timestamptz,
    updated_at timestamptz NOT NULL
);

-- Client tracking tables (unlogged for performance)
CREATE UNLOGGED TABLE river_client (
    id text PRIMARY KEY NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    metadata jsonb NOT NULL DEFAULT '{}',
    paused_at timestamptz,
    updated_at timestamptz NOT NULL,
    CONSTRAINT name_length CHECK (char_length(id) > 0 AND char_length(id) < 128)
);

CREATE UNLOGGED TABLE river_client_queue (
    river_client_id text NOT NULL REFERENCES river_client (id) ON DELETE CASCADE,
    name text NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    max_workers bigint NOT NULL DEFAULT 0,
    metadata jsonb NOT NULL DEFAULT '{}',
    num_jobs_completed bigint NOT NULL DEFAULT 0,
    num_jobs_running bigint NOT NULL DEFAULT 0,
    updated_at timestamptz NOT NULL,
    PRIMARY KEY (river_client_id, name),
    CONSTRAINT name_length CHECK (char_length(name) > 0 AND char_length(name) < 128),
    CONSTRAINT num_jobs_completed_zero_or_positive CHECK (num_jobs_completed >= 0),
    CONSTRAINT num_jobs_running_zero_or_positive CHECK (num_jobs_running >= 0)
);

COMMENT ON TABLE river_job IS 'River Queue: Background job storage';
COMMENT ON TABLE river_leader IS 'River Queue: Leader election for periodic jobs (single coordinator)';
COMMENT ON TABLE river_queue IS 'River Queue: Queue metadata and pause state';
COMMENT ON TABLE river_client IS 'River Queue: Active client tracking';
COMMENT ON TABLE river_client_queue IS 'River Queue: Per-client queue state';
