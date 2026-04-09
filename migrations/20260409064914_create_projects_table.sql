-- +goose Up
CREATE TABLE projects
(
    id         UUID PRIMARY KEY     DEFAULT gen_random_uuid(),
    account_id UUID        NOT NULL REFERENCES accounts (id),
    name       TEXT        NOT NULL,
    slug       TEXT        NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at TIMESTAMPTZ,
    UNIQUE (account_id, slug)
);

-- +goose Down
DROP TABLE projects;
