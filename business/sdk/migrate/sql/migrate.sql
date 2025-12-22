-- Version: 1
-- Description: Create table users
CREATE TABLE IF NOT EXISTS users (
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


-- Version: 2
-- Description: Create table thinks
CREATE TABLE IF NOT EXISTS thinks (
    think_id      UUID      NOT NULL,
    user_id      UUID       NOT NULL,
    category      TEXT      NOT NULL,
    content       TEXT      NOT NULL,
    date_created  TIMESTAMP NOT NULL,
    date_updated  TIMESTAMP NOT NULL,

    PRIMARY KEY (think_id),
    FOREIGN KEY (user_id) REFERENCES users(user_id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS thinks_date_created_idx ON thinks(date_created DESC);
CREATE INDEX IF NOT EXISTS thinks_category_idx ON thinks(category);


-- Version: 3
-- Description: Create table moments for emotional tracking
CREATE TABLE IF NOT EXISTS moments (
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

CREATE INDEX IF NOT EXISTS moments_user_id_idx ON moments(user_id);
CREATE INDEX IF NOT EXISTS moments_user_date_idx ON moments(user_id, moment_date DESC);
CREATE INDEX IF NOT EXISTS moments_date_created_idx ON moments(date_created DESC);
CREATE INDEX IF NOT EXISTS moments_intensity_idx ON moments(intensity);

COMMENT ON TABLE moments IS 'Tracks emotional/difficult moments for psychological self-observation';
COMMENT ON COLUMN moments.moment_date IS 'When the observed moment actually occurred (user can backdate)';
COMMENT ON COLUMN moments.intensity IS 'Distress intensity on 0-10 scale';


-- Version: 4
-- Description: Create values table for personal values tracking
CREATE TABLE IF NOT EXISTS values (
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

CREATE INDEX IF NOT EXISTS values_user_id_idx ON values(user_id);
CREATE INDEX IF NOT EXISTS values_facet_idx ON values(facet);

COMMENT ON TABLE values IS 'Stores user core personal values with life facet categorization (max 10 per user)';
COMMENT ON COLUMN values.content IS 'Encrypted value statement (plaintext validated as 3-200 chars in business layer)';
COMMENT ON COLUMN values.facet IS 'Life domain categorization based on ACT therapy';
COMMENT ON COLUMN values.display_order IS 'User-controlled priority ranking (1=highest)';


-- Version: 5
-- Description: Remove unique constraint on display_order to allow atomic reordering
DROP INDEX IF EXISTS values_user_order_unique_idx;


-- Version: 6
-- Description: Create life_visions table for aspirational states of being
CREATE TABLE IF NOT EXISTS life_visions (
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

CREATE INDEX IF NOT EXISTS life_visions_user_id_idx ON life_visions(user_id);
CREATE INDEX IF NOT EXISTS life_visions_value_id_idx ON life_visions(value_id);
CREATE INDEX IF NOT EXISTS life_visions_date_created_idx ON life_visions(date_created DESC);

COMMENT ON TABLE life_visions IS 'Aspirational states of being linked to personal values';
COMMENT ON COLUMN life_visions.content IS 'Encrypted vision statement (plaintext validated as 10-500 chars in business layer)';


-- Version: 7
-- Description: Create view for combined moments and thinks export
CREATE OR REPLACE VIEW view_export_items AS
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


-- Version: 8
-- Description: Add River Queue tables for background job processing
DO $$ BEGIN
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
EXCEPTION WHEN duplicate_object THEN
    NULL;
END $$;

CREATE TABLE IF NOT EXISTS river_job(
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

CREATE INDEX IF NOT EXISTS river_job_kind ON river_job USING btree(kind);
CREATE INDEX IF NOT EXISTS river_job_state_and_finalized_at_index ON river_job USING btree(state, finalized_at) WHERE finalized_at IS NOT NULL;
CREATE INDEX IF NOT EXISTS river_job_prioritized_fetching_index ON river_job USING btree(state, queue, priority, scheduled_at, id);
CREATE INDEX IF NOT EXISTS river_job_args_index ON river_job USING GIN(args);
CREATE INDEX IF NOT EXISTS river_job_metadata_index ON river_job USING GIN(metadata);

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

DO $$ BEGIN
    CREATE UNIQUE INDEX river_job_unique_idx ON river_job (unique_key)
        WHERE unique_key IS NOT NULL
          AND unique_states IS NOT NULL
          AND river_job_state_in_bitmask(unique_states, state);
EXCEPTION WHEN duplicate_table THEN
    NULL;
END $$;

CREATE UNLOGGED TABLE IF NOT EXISTS river_leader(
    elected_at timestamptz NOT NULL,
    expires_at timestamptz NOT NULL,
    leader_id text NOT NULL,
    name text PRIMARY KEY DEFAULT 'default',
    CONSTRAINT name_length CHECK (name = 'default'),
    CONSTRAINT leader_id_length CHECK (char_length(leader_id) > 0 AND char_length(leader_id) < 128)
);

CREATE TABLE IF NOT EXISTS river_queue (
    name text PRIMARY KEY NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    metadata jsonb NOT NULL DEFAULT '{}' ::jsonb,
    paused_at timestamptz,
    updated_at timestamptz NOT NULL
);

CREATE UNLOGGED TABLE IF NOT EXISTS river_client (
    id text PRIMARY KEY NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    metadata jsonb NOT NULL DEFAULT '{}',
    paused_at timestamptz,
    updated_at timestamptz NOT NULL,
    CONSTRAINT name_length CHECK (char_length(id) > 0 AND char_length(id) < 128)
);

CREATE UNLOGGED TABLE IF NOT EXISTS river_client_queue (
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


-- Version: 9
-- Description: Add Telegram notifications support (users fields, messages table, content view)

-- Add Telegram connection fields to users table
ALTER TABLE users ADD COLUMN IF NOT EXISTS telegram_chat_id BIGINT;
ALTER TABLE users ADD COLUMN IF NOT EXISTS telegram_enabled BOOLEAN NOT NULL DEFAULT false;

CREATE INDEX IF NOT EXISTS users_telegram_enabled_idx
  ON users(telegram_enabled) WHERE telegram_enabled = true;

COMMENT ON COLUMN users.telegram_chat_id IS 'Telegram chat ID for sending notifications';
COMMENT ON COLUMN users.telegram_enabled IS 'Whether Telegram notifications are enabled';

-- Create notification_messages table (Support Domain)
CREATE TABLE IF NOT EXISTS notification_messages (
    message_id        UUID        NOT NULL,
    user_id           UUID        NOT NULL,
    message_type      TEXT        NOT NULL CHECK (message_type IN ('morning', 'evening', 'test', 'welcome')),
    content           TEXT        NOT NULL,
    telegram_msg_id   BIGINT,
    status            TEXT        NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'sent', 'failed')),
    error_message     TEXT,
    retry_count       INTEGER     NOT NULL DEFAULT 0,
    scheduled_at      TIMESTAMP   NOT NULL,
    sent_at           TIMESTAMP,
    date_created      TIMESTAMP   NOT NULL,

    PRIMARY KEY (message_id),
    FOREIGN KEY (user_id) REFERENCES users(user_id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS notification_messages_user_idx
  ON notification_messages(user_id, date_created DESC);
CREATE INDEX IF NOT EXISTS notification_messages_pending_idx
  ON notification_messages(scheduled_at) WHERE status = 'pending';

-- Unique constraint to prevent duplicate scheduled messages per user/type/date
-- Note: No WHERE clause so uniqueness applies regardless of status
CREATE UNIQUE INDEX IF NOT EXISTS notification_messages_schedule_unique_idx
  ON notification_messages(user_id, message_type, DATE(scheduled_at));

COMMENT ON TABLE notification_messages IS 'Telegram notification messages sent to users';
COMMENT ON COLUMN notification_messages.message_type IS 'morning, evening, test, or welcome message';
COMMENT ON COLUMN notification_messages.scheduled_at IS 'When message should be sent (UTC)';

-- Create view for notification content (values + life visions)
-- Note: ORDER BY intentionally omitted - consumers should order when querying
CREATE OR REPLACE VIEW view_notification_content AS
SELECT
    v.user_id,
    v.value_id,
    v.content AS value_content,
    v.facet AS value_facet,
    v.display_order AS value_order,
    lv.life_vision_id,
    lv.content AS life_vision_content
FROM values v
LEFT JOIN life_visions lv ON lv.value_id = v.value_id;

COMMENT ON VIEW view_notification_content IS 'Read-only view for notification message content';


-- Version: 10
-- Description: Fix duplicate message prevention - remove status condition from unique index

-- Drop the old partial unique index (only prevented duplicates for pending messages)
DROP INDEX IF EXISTS notification_messages_schedule_unique_idx;

-- Create new unique index without WHERE clause (prevents duplicates regardless of status)
CREATE UNIQUE INDEX IF NOT EXISTS notification_messages_user_type_date_unique_idx
  ON notification_messages(user_id, message_type, DATE(scheduled_at));

COMMENT ON INDEX notification_messages_user_type_date_unique_idx IS 'Prevents duplicate messages per user/type/day regardless of status';


-- Version: 11
-- Description: Add 'sending' status to prevent duplicate message delivery

-- Drop the old check constraint and add new one with 'sending' status
ALTER TABLE notification_messages
  DROP CONSTRAINT IF EXISTS notification_messages_status_check;

ALTER TABLE notification_messages
  ADD CONSTRAINT notification_messages_status_check
  CHECK (status IN ('pending', 'sending', 'sent', 'failed'));

COMMENT ON COLUMN notification_messages.status IS 'Message status: pending, sending, sent, or failed';


-- Version: 12
-- Description: Add soft delete (archive) support to values and life_visions

-- =============================================================================
-- Add status and archived_at columns
-- =============================================================================

-- Values table
ALTER TABLE values
  ADD COLUMN IF NOT EXISTS status TEXT NOT NULL DEFAULT 'active',
  ADD COLUMN IF NOT EXISTS archived_at TIMESTAMP NULL;

ALTER TABLE values
  DROP CONSTRAINT IF EXISTS values_status_check;

ALTER TABLE values
  ADD CONSTRAINT values_status_check
  CHECK (status IN ('active', 'archived'));

-- Life Visions table
ALTER TABLE life_visions
  ADD COLUMN IF NOT EXISTS status TEXT NOT NULL DEFAULT 'active',
  ADD COLUMN IF NOT EXISTS archived_at TIMESTAMP NULL;

ALTER TABLE life_visions
  DROP CONSTRAINT IF EXISTS life_visions_status_check;

ALTER TABLE life_visions
  ADD CONSTRAINT life_visions_status_check
  CHECK (status IN ('active', 'archived'));

-- =============================================================================
-- Add performance indexes
-- =============================================================================

-- Values indexes
CREATE INDEX IF NOT EXISTS values_user_active_idx
  ON values(user_id)
  WHERE status = 'active';

CREATE INDEX IF NOT EXISTS values_status_idx
  ON values(status);

-- Life Visions indexes
CREATE INDEX IF NOT EXISTS life_visions_user_active_idx
  ON life_visions(user_id)
  WHERE status = 'active';

CREATE INDEX IF NOT EXISTS life_visions_value_active_idx
  ON life_visions(value_id)
  WHERE status = 'active';

CREATE INDEX IF NOT EXISTS life_visions_status_idx
  ON life_visions(status);

-- =============================================================================
-- Change CASCADE to RESTRICT for life_visions -> values
-- =============================================================================

ALTER TABLE life_visions
  DROP CONSTRAINT IF EXISTS life_visions_value_id_fkey;

ALTER TABLE life_visions
  ADD CONSTRAINT life_visions_value_id_fkey
  FOREIGN KEY (value_id) REFERENCES values(value_id) ON DELETE RESTRICT;

-- =============================================================================
-- Database trigger: Prevent archiving value with active life visions
-- =============================================================================

CREATE OR REPLACE FUNCTION prevent_archive_with_active_life_visions()
RETURNS TRIGGER AS $$
BEGIN
    IF NEW.status = 'archived' AND (OLD.status IS NULL OR OLD.status != 'archived') THEN
        IF EXISTS (
            SELECT 1 FROM life_visions
            WHERE value_id = NEW.value_id
            AND status = 'active'
        ) THEN
            RAISE EXCEPTION 'Cannot archive value: has active life visions'
                USING ERRCODE = '23503';
        END IF;
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS check_archive_life_visions ON values;
CREATE TRIGGER check_archive_life_visions
    BEFORE UPDATE ON values
    FOR EACH ROW
    EXECUTE FUNCTION prevent_archive_with_active_life_visions();

-- =============================================================================
-- Comments
-- =============================================================================

COMMENT ON COLUMN values.status IS 'Entity status: active or archived (soft delete)';
COMMENT ON COLUMN values.archived_at IS 'Timestamp when entity was archived (NULL if active)';
COMMENT ON COLUMN life_visions.status IS 'Entity status: active or archived (soft delete)';
COMMENT ON COLUMN life_visions.archived_at IS 'Timestamp when entity was archived (NULL if active)';


-- Version: 13
-- Description: Create objetivos table for goal tracking with resultado/frecuencia types
CREATE TABLE IF NOT EXISTS objetivos (
    objetivo_id              UUID        NOT NULL,
    user_id                  UUID        NOT NULL,
    life_vision_id           UUID        NOT NULL,
    titulo                   TEXT        NOT NULL,
    tipo_tracking            TEXT        NOT NULL CHECK (tipo_tracking IN ('resultado', 'frecuencia')),
    status                   TEXT        NOT NULL DEFAULT 'activo' CHECK (status IN ('activo', 'completado', 'abandonado', 'pausado')),
    metrica_objetivo         INTEGER     NULL,
    metrica_actual           INTEGER     NULL,
    frecuencia_tipo          TEXT        NULL CHECK (frecuencia_tipo IS NULL OR frecuencia_tipo IN ('daily', 'n_por_semana', 'n_por_mes')),
    frecuencia_n             INTEGER     NULL,
    cumplimiento_target_pct  INTEGER     NULL CHECK (cumplimiento_target_pct IS NULL OR (cumplimiento_target_pct >= 1 AND cumplimiento_target_pct <= 100)),
    archived_at              TIMESTAMP   NULL,
    date_created             TIMESTAMP   NOT NULL,
    date_updated             TIMESTAMP   NOT NULL,

    PRIMARY KEY (objetivo_id),
    FOREIGN KEY (user_id) REFERENCES users(user_id) ON DELETE CASCADE,
    FOREIGN KEY (life_vision_id) REFERENCES life_visions(life_vision_id) ON DELETE CASCADE,

    -- Resultado type requires metrica_objetivo > 0
    CONSTRAINT objetivos_resultado_config CHECK (
        tipo_tracking != 'resultado' OR (metrica_objetivo IS NOT NULL AND metrica_objetivo > 0)
    ),

    -- Frecuencia type requires frecuencia_tipo and cumplimiento_target_pct
    CONSTRAINT objetivos_frecuencia_config CHECK (
        tipo_tracking != 'frecuencia' OR (
            frecuencia_tipo IS NOT NULL AND
            cumplimiento_target_pct IS NOT NULL
        )
    ),

    -- For n_por_semana and n_por_mes, frecuencia_n must have sensible bounds
    CONSTRAINT objetivos_frecuencia_n_config CHECK (
        frecuencia_tipo IS NULL OR
        frecuencia_tipo = 'daily' OR
        (frecuencia_tipo = 'n_por_semana' AND frecuencia_n IS NOT NULL AND frecuencia_n BETWEEN 1 AND 7) OR
        (frecuencia_tipo = 'n_por_mes' AND frecuencia_n IS NOT NULL AND frecuencia_n BETWEEN 1 AND 31)
    ),

    -- Metrica_actual cannot exceed metrica_objetivo and must be non-negative
    CONSTRAINT objetivos_progress_bounds CHECK (
        metrica_actual IS NULL OR (metrica_actual >= 0 AND (metrica_objetivo IS NULL OR metrica_actual <= metrica_objetivo))
    ),

    -- Mutual exclusivity: resultado type should not have frecuencia fields
    CONSTRAINT objetivos_resultado_exclusivity CHECK (
        tipo_tracking != 'resultado' OR (
            frecuencia_tipo IS NULL AND
            frecuencia_n IS NULL AND
            cumplimiento_target_pct IS NULL
        )
    ),

    -- Mutual exclusivity: frecuencia type should not have resultado fields
    CONSTRAINT objetivos_frecuencia_exclusivity CHECK (
        tipo_tracking != 'frecuencia' OR (
            metrica_objetivo IS NULL AND
            metrica_actual IS NULL
        )
    )
);

CREATE INDEX IF NOT EXISTS objetivos_user_id_idx ON objetivos(user_id);
CREATE INDEX IF NOT EXISTS objetivos_life_vision_id_idx ON objetivos(life_vision_id);
CREATE INDEX IF NOT EXISTS objetivos_status_idx ON objetivos(status);
CREATE INDEX IF NOT EXISTS objetivos_user_active_idx ON objetivos(user_id) WHERE archived_at IS NULL;
CREATE INDEX IF NOT EXISTS objetivos_tipo_tracking_idx ON objetivos(tipo_tracking);

COMMENT ON TABLE objetivos IS 'Measurable objectives linked to life visions with resultado or frecuencia tracking';
COMMENT ON COLUMN objetivos.titulo IS 'Encrypted objective title (plaintext validated as 5-200 chars in business layer)';
COMMENT ON COLUMN objetivos.tipo_tracking IS 'Tracking type: resultado (metric-based) or frecuencia (calendar compliance)';
COMMENT ON COLUMN objetivos.status IS 'Objective status: activo, completado (terminal), abandonado (terminal), pausado';
COMMENT ON COLUMN objetivos.metrica_objetivo IS 'Target metric for resultado tracking (e.g., 35 books)';
COMMENT ON COLUMN objetivos.metrica_actual IS 'Current progress for resultado tracking';
COMMENT ON COLUMN objetivos.frecuencia_tipo IS 'Frequency type: daily, n_por_semana, n_por_mes';
COMMENT ON COLUMN objetivos.frecuencia_n IS 'Target occurrences per period for n_por_semana/n_por_mes';
COMMENT ON COLUMN objetivos.cumplimiento_target_pct IS 'Target compliance percentage (1-100) for frecuencia tracking';


-- Version: 14
-- Description: Create objetivo_records table for frecuencia tracking records
CREATE TABLE IF NOT EXISTS objetivo_records (
    objetivo_record_id   UUID        NOT NULL,
    objetivo_id          UUID        NOT NULL,
    user_id              UUID        NOT NULL,
    fecha_registro       DATE        NOT NULL,
    status               TEXT        NOT NULL CHECK (status IN ('completado', 'omitido_intencional', 'omitido')),
    notes                TEXT        NULL,
    date_created         TIMESTAMP   NOT NULL,

    PRIMARY KEY (objetivo_record_id),
    FOREIGN KEY (objetivo_id) REFERENCES objetivos(objetivo_id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users(user_id) ON DELETE CASCADE
);

-- Unique constraint: one record per objetivo per day
CREATE UNIQUE INDEX IF NOT EXISTS objetivo_records_objetivo_date_unique_idx
    ON objetivo_records(objetivo_id, fecha_registro);

CREATE INDEX IF NOT EXISTS objetivo_records_objetivo_id_idx ON objetivo_records(objetivo_id);
CREATE INDEX IF NOT EXISTS objetivo_records_user_id_idx ON objetivo_records(user_id);
CREATE INDEX IF NOT EXISTS objetivo_records_fecha_idx ON objetivo_records(fecha_registro);
CREATE INDEX IF NOT EXISTS objetivo_records_status_idx ON objetivo_records(status);

COMMENT ON TABLE objetivo_records IS 'Daily completion records for frecuencia-type objectives';
COMMENT ON COLUMN objetivo_records.fecha_registro IS 'Date of the record (today or yesterday only, enforced in business layer)';
COMMENT ON COLUMN objetivo_records.status IS 'Record status: completado (done), omitido_intencional (rest day), omitido (missed)';
COMMENT ON COLUMN objetivo_records.notes IS 'Encrypted optional notes about the record';


-- Version: 15
-- Description: Rename objetivos tables to English (objectives/objective_records)
-- Note: Dropping and recreating since this is new functionality with no production data

-- Drop old tables (child first, then parent)
DROP TABLE IF EXISTS objetivo_records;
DROP TABLE IF EXISTS objetivos;

-- Create objectives table with English naming
CREATE TABLE IF NOT EXISTS objectives (
    objective_id             UUID        NOT NULL,
    user_id                  UUID        NOT NULL,
    life_vision_id           UUID        NOT NULL,
    title                    TEXT        NOT NULL,
    tracking_type            TEXT        NOT NULL CHECK (tracking_type IN ('result', 'frequency')),
    status                   TEXT        NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'completed', 'abandoned', 'paused')),
    target_metric            INTEGER     NULL,
    current_metric           INTEGER     NULL,
    frequency_type           TEXT        NULL CHECK (frequency_type IS NULL OR frequency_type IN ('daily', 'n_per_week', 'n_per_month')),
    frequency_count          INTEGER     NULL,
    compliance_target_pct    INTEGER     NULL CHECK (compliance_target_pct IS NULL OR (compliance_target_pct >= 1 AND compliance_target_pct <= 100)),
    archived_at              TIMESTAMP   NULL,
    date_created             TIMESTAMP   NOT NULL,
    date_updated             TIMESTAMP   NOT NULL,

    PRIMARY KEY (objective_id),
    FOREIGN KEY (user_id) REFERENCES users(user_id) ON DELETE CASCADE,
    FOREIGN KEY (life_vision_id) REFERENCES life_visions(life_vision_id) ON DELETE CASCADE,

    -- Result type requires target_metric > 0
    CONSTRAINT objectives_result_config CHECK (
        tracking_type != 'result' OR (target_metric IS NOT NULL AND target_metric > 0)
    ),

    -- Frequency type requires frequency_type and compliance_target_pct
    CONSTRAINT objectives_frequency_config CHECK (
        tracking_type != 'frequency' OR (
            frequency_type IS NOT NULL AND
            compliance_target_pct IS NOT NULL
        )
    ),

    -- For n_per_week and n_per_month, frequency_count must have sensible bounds
    CONSTRAINT objectives_frequency_count_config CHECK (
        frequency_type IS NULL OR
        frequency_type = 'daily' OR
        (frequency_type = 'n_per_week' AND frequency_count IS NOT NULL AND frequency_count BETWEEN 1 AND 7) OR
        (frequency_type = 'n_per_month' AND frequency_count IS NOT NULL AND frequency_count BETWEEN 1 AND 31)
    ),

    -- current_metric cannot exceed target_metric and must be non-negative
    CONSTRAINT objectives_progress_bounds CHECK (
        current_metric IS NULL OR (current_metric >= 0 AND (target_metric IS NULL OR current_metric <= target_metric))
    ),

    -- Mutual exclusivity: result type should not have frequency fields
    CONSTRAINT objectives_result_exclusivity CHECK (
        tracking_type != 'result' OR (
            frequency_type IS NULL AND
            frequency_count IS NULL AND
            compliance_target_pct IS NULL
        )
    ),

    -- Mutual exclusivity: frequency type should not have result fields
    CONSTRAINT objectives_frequency_exclusivity CHECK (
        tracking_type != 'frequency' OR (
            target_metric IS NULL AND
            current_metric IS NULL
        )
    )
);

CREATE INDEX IF NOT EXISTS objectives_user_id_idx ON objectives(user_id);
CREATE INDEX IF NOT EXISTS objectives_life_vision_id_idx ON objectives(life_vision_id);
CREATE INDEX IF NOT EXISTS objectives_status_idx ON objectives(status);
CREATE INDEX IF NOT EXISTS objectives_user_active_idx ON objectives(user_id) WHERE archived_at IS NULL;
CREATE INDEX IF NOT EXISTS objectives_tracking_type_idx ON objectives(tracking_type);

COMMENT ON TABLE objectives IS 'Measurable objectives linked to life visions with result or frequency tracking';
COMMENT ON COLUMN objectives.title IS 'Encrypted objective title (plaintext validated as 5-200 chars in business layer)';
COMMENT ON COLUMN objectives.tracking_type IS 'Tracking type: result (metric-based) or frequency (calendar compliance)';
COMMENT ON COLUMN objectives.status IS 'Objective status: active, completed (terminal), abandoned (terminal), paused';
COMMENT ON COLUMN objectives.target_metric IS 'Target metric for result tracking (e.g., 35 books)';
COMMENT ON COLUMN objectives.current_metric IS 'Current progress for result tracking';
COMMENT ON COLUMN objectives.frequency_type IS 'Frequency type: daily, n_per_week, n_per_month';
COMMENT ON COLUMN objectives.frequency_count IS 'Target occurrences per period for n_per_week/n_per_month';
COMMENT ON COLUMN objectives.compliance_target_pct IS 'Target compliance percentage (1-100) for frequency tracking';


-- Create objective_records table with English naming
CREATE TABLE IF NOT EXISTS objective_records (
    objective_record_id  UUID        NOT NULL,
    objective_id         UUID        NOT NULL,
    user_id              UUID        NOT NULL,
    record_date          DATE        NOT NULL,
    status               TEXT        NOT NULL CHECK (status IN ('completed', 'intentionally_skipped', 'skipped')),
    notes                TEXT        NULL,
    date_created         TIMESTAMP   NOT NULL,

    PRIMARY KEY (objective_record_id),
    FOREIGN KEY (objective_id) REFERENCES objectives(objective_id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users(user_id) ON DELETE CASCADE
);

-- Unique constraint: one record per objective per day
CREATE UNIQUE INDEX IF NOT EXISTS objective_records_objective_date_unique_idx
    ON objective_records(objective_id, record_date);

CREATE INDEX IF NOT EXISTS objective_records_objective_id_idx ON objective_records(objective_id);
CREATE INDEX IF NOT EXISTS objective_records_user_id_idx ON objective_records(user_id);
CREATE INDEX IF NOT EXISTS objective_records_date_idx ON objective_records(record_date);
CREATE INDEX IF NOT EXISTS objective_records_status_idx ON objective_records(status);

COMMENT ON TABLE objective_records IS 'Daily completion records for frequency-type objectives';
COMMENT ON COLUMN objective_records.record_date IS 'Date of the record (today or yesterday only, enforced in business layer)';
COMMENT ON COLUMN objective_records.status IS 'Record status: completed (done), intentionally_skipped (rest day), skipped (missed)';
COMMENT ON COLUMN objective_records.notes IS 'Encrypted optional notes about the record';


-- Version: 16
-- Description: Create tasks table for task management with inbox and objective-linked support
CREATE TABLE IF NOT EXISTS tasks (
    task_id          UUID        NOT NULL,
    user_id          UUID        NOT NULL,
    objective_id     UUID        NULL,
    title            TEXT        NOT NULL,
    description      TEXT        NULL,
    contribution     INTEGER     NULL CHECK (contribution IS NULL OR (contribution >= 1 AND contribution <= 10)),
    status           TEXT        NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'completed', 'cancelled')),
    completed_at     TIMESTAMP   NULL,
    date_created     TIMESTAMP   NOT NULL,
    date_updated     TIMESTAMP   NOT NULL,

    PRIMARY KEY (task_id),
    FOREIGN KEY (user_id) REFERENCES users(user_id) ON DELETE CASCADE,
    FOREIGN KEY (objective_id) REFERENCES objectives(objective_id) ON DELETE CASCADE,

    -- Contribution required when linked to objective
    CONSTRAINT tasks_contribution_required CHECK (
        objective_id IS NULL OR contribution IS NOT NULL
    ),

    -- Completed tasks must have completed_at timestamp
    CONSTRAINT tasks_completed_timestamp CHECK (
        (status = 'completed' AND completed_at IS NOT NULL) OR
        (status != 'completed' AND completed_at IS NULL)
    )
);

CREATE INDEX IF NOT EXISTS tasks_user_id_idx ON tasks(user_id);
CREATE INDEX IF NOT EXISTS tasks_objective_id_idx ON tasks(objective_id);
CREATE INDEX IF NOT EXISTS tasks_status_idx ON tasks(status);
CREATE INDEX IF NOT EXISTS tasks_user_pending_idx ON tasks(user_id) WHERE status = 'pending';
CREATE INDEX IF NOT EXISTS tasks_user_completed_idx ON tasks(user_id, completed_at DESC) WHERE status = 'completed';
CREATE INDEX IF NOT EXISTS tasks_date_created_idx ON tasks(date_created DESC);

COMMENT ON TABLE tasks IS 'Tasks for inbox or linked to objectives with contribution tracking';
COMMENT ON COLUMN tasks.title IS 'Encrypted task title (plaintext validated as 3-200 chars in business layer)';
COMMENT ON COLUMN tasks.description IS 'Encrypted optional task description (plaintext validated as max 2000 chars in business layer)';
COMMENT ON COLUMN tasks.contribution IS 'Contribution to objective progress (1-10 scale, NULL for inbox tasks)';
COMMENT ON COLUMN tasks.objective_id IS 'NULL for inbox tasks, UUID for objective-linked tasks';
COMMENT ON COLUMN tasks.status IS 'Task status: pending, completed, cancelled';
COMMENT ON COLUMN tasks.completed_at IS 'Timestamp when task was completed (NULL if not completed)';


-- Version: 17
-- Description: Allow tasks for frequency objectives with NULL contribution
-- Business layer now enforces:
-- - RESULT objectives: contribution REQUIRED (1-10)
-- - FREQUENCY objectives: contribution must be NULL
-- - Inbox tasks: contribution must be NULL
ALTER TABLE tasks DROP CONSTRAINT IF EXISTS tasks_contribution_required;


-- Version: 18
-- Description: Defense-in-depth trigger for tasks contribution validation
-- This trigger enforces contribution rules at the database level as a safety net
-- for the business layer validation. Rules:
-- - Inbox tasks (objective_id IS NULL): contribution MUST be NULL
-- - Result objectives: contribution REQUIRED and BETWEEN 1 AND 10
-- - Frequency objectives: contribution MUST be NULL

-- Create or replace the trigger function
CREATE OR REPLACE FUNCTION validate_task_contribution()
RETURNS TRIGGER AS $$
DECLARE
    objective_tracking_type TEXT;
BEGIN
    -- Rule 1: Inbox tasks (no objective) must have NULL contribution
    IF NEW.objective_id IS NULL THEN
        IF NEW.contribution IS NOT NULL THEN
            RAISE EXCEPTION 'Inbox tasks cannot have contribution (SQLSTATE 23514)'
                USING ERRCODE = '23514'; -- check_violation
        END IF;
        RETURN NEW;
    END IF;

    -- Get the tracking type for the objective
    SELECT tracking_type INTO objective_tracking_type
    FROM objectives
    WHERE objective_id = NEW.objective_id;

    -- If objective not found, let the FK constraint handle it
    IF objective_tracking_type IS NULL THEN
        RETURN NEW;
    END IF;

    -- Rule 2: Result objectives require contribution 1-10
    IF objective_tracking_type = 'result' THEN
        IF NEW.contribution IS NULL THEN
            RAISE EXCEPTION 'Result objective tasks require contribution (SQLSTATE 23514)'
                USING ERRCODE = '23514';
        END IF;
        IF NEW.contribution < 1 OR NEW.contribution > 10 THEN
            RAISE EXCEPTION 'Contribution must be between 1 and 10 (SQLSTATE 23514)'
                USING ERRCODE = '23514';
        END IF;
    END IF;

    -- Rule 3: Frequency objectives must have NULL contribution
    IF objective_tracking_type = 'frequency' THEN
        IF NEW.contribution IS NOT NULL THEN
            RAISE EXCEPTION 'Frequency objective tasks cannot have contribution (SQLSTATE 23514)'
                USING ERRCODE = '23514';
        END IF;
    END IF;

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Drop existing trigger if any
DROP TRIGGER IF EXISTS validate_task_contribution_trigger ON tasks;

-- Create the trigger for both INSERT and UPDATE
CREATE TRIGGER validate_task_contribution_trigger
    BEFORE INSERT OR UPDATE ON tasks
    FOR EACH ROW
    EXECUTE FUNCTION validate_task_contribution();

COMMENT ON FUNCTION validate_task_contribution() IS 'Defense-in-depth validation for task contribution rules';
