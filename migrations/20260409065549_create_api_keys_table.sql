-- +goose Up
CREATE TABLE api_keys
(
    id           UUID PRIMARY KEY     DEFAULT gen_random_uuid(),
    project_id   UUID        NOT NULL REFERENCES projects (id),
    name         TEXT        NOT NULL,
    key_hash     TEXT        NOT NULL UNIQUE,
    key_prefix   TEXT        NOT NULL,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    last_used_at TIMESTAMPTZ,
    revoked_at   TIMESTAMPTZ
);

-- +goose Down
DROP TABLE api_keys;
