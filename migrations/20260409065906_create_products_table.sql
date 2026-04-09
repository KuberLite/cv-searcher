-- +goose Up
CREATE TABLE products
(
    id          UUID PRIMARY KEY     DEFAULT gen_random_uuid(),
    project_id  UUID        NOT NULL REFERENCES projects (id),
    external_id TEXT        NOT NULL,
    name        TEXT        NOT NULL,
    brand       TEXT,
    description TEXT,
    attrs       JSONB,
    indexed_at  TIMESTAMPTZ,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at  TIMESTAMPTZ,
    UNIQUE (project_id, external_id)
);

CREATE INDEX idx_products_project_id ON products(project_id);
CREATE INDEX idx_products_updated_at ON products(project_id, updated_at)
    WHERE deleted_at IS NULL;

-- +goose Down
DROP TABLE products;