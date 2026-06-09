-- Catalog attribute values live in product_attribute_values (FK -> product_attribute_definitions).
-- Product variant values need a separate table (FK -> product_attributes).

CREATE TABLE IF NOT EXISTS product_variant_attribute_values (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    attribute_id UUID NOT NULL REFERENCES product_attributes(id) ON DELETE CASCADE,
    value VARCHAR(200) NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_product_variant_attribute_values_attribute_id
    ON product_variant_attribute_values(attribute_id);
