ALTER TABLE products DROP COLUMN IF EXISTS sku;

ALTER TABLE product_attributes DROP COLUMN IF EXISTS value;

CREATE TABLE IF NOT EXISTS skus (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    product_id UUID NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    code VARCHAR(100) NOT NULL UNIQUE,
    attributes JSONB NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_skus_product_id ON skus(product_id);
