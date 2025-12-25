-- @cgraph-id sql-root
-- @cgraph-deps html-root

CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    email TEXT NOT NULL,
    created_at TIMESTAMPTZ DEFAULT now()
);

-- note: indexes handled separately
