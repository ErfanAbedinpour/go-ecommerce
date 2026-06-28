ALTER TABLE store_settings DROP COLUMN IF EXISTS contact_section_image_url;
ALTER TABLE store_settings DROP COLUMN IF EXISTS storefront_navigation;

DROP TABLE IF EXISTS faq_items;
DROP TABLE IF EXISTS faq_sections;
DROP TABLE IF EXISTS homepage_reviews;
DROP TABLE IF EXISTS partner_brands;
DROP TABLE IF EXISTS pro_banners;
DROP TABLE IF EXISTS product_slide_items;
DROP TABLE IF EXISTS product_slides;
DROP TABLE IF EXISTS storefront_hero;

DROP INDEX IF EXISTS idx_customers_user_id;
ALTER TABLE customers DROP COLUMN IF EXISTS user_id;
