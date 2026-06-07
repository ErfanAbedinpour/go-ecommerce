CREATE TABLE brands (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name        VARCHAR(100) NOT NULL,
    slug        VARCHAR(100) NOT NULL,
    description TEXT,
    is_active   BOOLEAN NOT NULL DEFAULT TRUE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at  TIMESTAMPTZ
);

CREATE UNIQUE INDEX idx_brands_name ON brands(name) WHERE deleted_at IS NULL;
CREATE UNIQUE INDEX idx_brands_slug ON brands(slug) WHERE deleted_at IS NULL;

CREATE TABLE product_attribute_definitions (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name        VARCHAR(100) NOT NULL,
    slug        VARCHAR(100) NOT NULL,
    sort_order  INT NOT NULL DEFAULT 0,
    is_active   BOOLEAN NOT NULL DEFAULT TRUE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at  TIMESTAMPTZ
);

CREATE UNIQUE INDEX idx_product_attr_defs_name ON product_attribute_definitions(name) WHERE deleted_at IS NULL;
CREATE UNIQUE INDEX idx_product_attr_defs_slug ON product_attribute_definitions(slug) WHERE deleted_at IS NULL;

CREATE TABLE product_attribute_values (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    attribute_id UUID NOT NULL REFERENCES product_attribute_definitions(id) ON DELETE CASCADE,
    value        VARCHAR(200) NOT NULL,
    sort_order   INT NOT NULL DEFAULT 0,
    is_active    BOOLEAN NOT NULL DEFAULT TRUE,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at   TIMESTAMPTZ
);

CREATE UNIQUE INDEX idx_product_attr_values_unique ON product_attribute_values(attribute_id, value) WHERE deleted_at IS NULL;
CREATE INDEX idx_product_attr_values_attribute ON product_attribute_values(attribute_id);
