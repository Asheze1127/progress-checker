-- sqlc schema: references idempotency_keys from database/postgres/schema.sql
-- This minimal definition is required because sqlc cannot use the full schema
-- due to a name collision between the slack_channel_purpose enum and the
-- slack_channel_purposes table (sqlc singularizes table names for Go structs).
CREATE TABLE idempotency_keys (
    key        VARCHAR     PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    expires_at TIMESTAMPTZ NOT NULL
);
