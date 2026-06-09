DROP TABLE IF EXISTS skus;

ALTER TABLE product_attributes ADD COLUMN IF NOT EXISTS value VARCHAR(200) NOT NULL DEFAULT '';

ALTER TABLE products ADD COLUMN IF NOT EXISTS sku VARCHAR(100) NOT NULL DEFAULT '';
-- unique constraint may need manual cleanup on rollback if products.sku_key was dropped
