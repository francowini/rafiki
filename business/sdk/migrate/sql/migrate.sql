-- Version: 1.01
-- Description: Create table thinks

CREATE TABLE thinks (
    think_id      UUID      NOT NULL,
    category      TEXT      NOT NULL,
    content       TEXT      NOT NULL,
    date_created  TIMESTAMP NOT NULL,
    date_updated  TIMESTAMP NOT NULL,

    PRIMARY KEY (think_id)
);

-- Create index for faster retrieval by date
CREATE INDEX thinks_date_created_idx ON thinks(date_created DESC);
CREATE INDEX thinks_category_idx ON thinks(category);

