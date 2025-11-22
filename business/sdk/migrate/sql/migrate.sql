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
-- Description: Create facet_type ENUM and values table for personal values tracking

-- Create ENUM type for life facets
CREATE TYPE facet_type AS ENUM (
    'health',
    'relationships',
    'career',
    'personal_growth',
    'family',
    'creativity',
    'community',
    'spirituality'
);

-- Create values table
CREATE TABLE values (
    value_id       UUID        NOT NULL,
    user_id        UUID        NOT NULL,
    content        TEXT        NOT NULL,
    facet          facet_type  NOT NULL,
    display_order  INTEGER     NOT NULL,
    date_created   TIMESTAMP   NOT NULL,
    date_updated   TIMESTAMP   NOT NULL,

    PRIMARY KEY (value_id),
    FOREIGN KEY (user_id) REFERENCES users(user_id) ON DELETE CASCADE
);

-- Performance indexes
CREATE INDEX values_user_id_idx ON values(user_id);
CREATE INDEX values_user_order_idx ON values(user_id, display_order ASC);
CREATE INDEX values_facet_idx ON values(facet);

-- Unique constraint to prevent duplicate display_order per user
CREATE UNIQUE INDEX values_user_order_unique_idx ON values(user_id, display_order);

-- Documentation
COMMENT ON TABLE values IS 'Stores user core personal values with life facet categorization (max 10 per user)';
COMMENT ON COLUMN values.content IS 'Encrypted value statement (3-200 characters)';
COMMENT ON COLUMN values.facet IS 'Life domain categorization based on ACT therapy';
COMMENT ON COLUMN values.display_order IS 'User-controlled priority ranking (1=highest)';
