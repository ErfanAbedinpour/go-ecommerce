-- customers.user_id links storefront auth users to order customers
ALTER TABLE customers ADD COLUMN user_id UUID REFERENCES admin_users(id) ON DELETE SET NULL;
CREATE UNIQUE INDEX idx_customers_user_id ON customers(user_id) WHERE user_id IS NOT NULL;

-- storefront hero (singleton)
CREATE TABLE storefront_hero (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    video_url           VARCHAR(500),
    title               VARCHAR(255),
    subtitle            TEXT,
    cta_primary_text    VARCHAR(100),
    cta_primary_url     VARCHAR(500),
    cta_secondary_text  VARCHAR(100),
    cta_secondary_url   VARCHAR(500),
    is_active           BOOLEAN NOT NULL DEFAULT true,
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

INSERT INTO storefront_hero (id) VALUES ('a1000000-0000-0000-0000-000000000001');

-- product slides (one per type)
CREATE TABLE product_slides (
    id                   UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    slide_type           VARCHAR(50) NOT NULL UNIQUE CHECK (slide_type IN ('featured', 'bestseller', 'discounted')),
    title                VARCHAR(255),
    autoplay_interval_ms INT NOT NULL DEFAULT 4500,
    is_active            BOOLEAN NOT NULL DEFAULT true,
    sort_order           INT NOT NULL DEFAULT 0,
    created_at           TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at           TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

INSERT INTO product_slides (id, slide_type, title) VALUES
    ('b1000000-0000-0000-0000-000000000001', 'featured', 'Featured Products'),
    ('b1000000-0000-0000-0000-000000000002', 'bestseller', 'Bestsellers'),
    ('b1000000-0000-0000-0000-000000000003', 'discounted', 'Discounted Products');

CREATE TABLE product_slide_items (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    slide_id   UUID NOT NULL REFERENCES product_slides(id) ON DELETE CASCADE,
    product_id UUID NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    sort_order INT NOT NULL DEFAULT 0,
    tab_label  VARCHAR(100)
);

CREATE INDEX idx_product_slide_items_slide ON product_slide_items(slide_id, sort_order);

CREATE TABLE pro_banners (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    desktop_image_url VARCHAR(500) NOT NULL,
    mobile_image_url  VARCHAR(500),
    link_url          VARCHAR(500),
    sort_order        INT NOT NULL DEFAULT 0,
    is_active         BOOLEAN NOT NULL DEFAULT true,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE partner_brands (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title       VARCHAR(255) NOT NULL,
    description TEXT,
    logo_url    VARCHAR(500) NOT NULL,
    link_url    VARCHAR(500),
    sort_order  INT NOT NULL DEFAULT 0,
    is_active   BOOLEAN NOT NULL DEFAULT true,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE homepage_reviews (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    customer_name VARCHAR(255) NOT NULL,
    photo_url     VARCHAR(500),
    review_text   TEXT NOT NULL,
    rating        SMALLINT CHECK (rating IS NULL OR (rating >= 1 AND rating <= 5)),
    sort_order    INT NOT NULL DEFAULT 0,
    is_active     BOOLEAN NOT NULL DEFAULT true,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE faq_sections (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    image_url  VARCHAR(500),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

INSERT INTO faq_sections (id) VALUES ('c1000000-0000-0000-0000-000000000001');

CREATE TABLE faq_items (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    question   TEXT NOT NULL,
    answer     TEXT NOT NULL,
    sort_order INT NOT NULL DEFAULT 0,
    is_active  BOOLEAN NOT NULL DEFAULT true
);

ALTER TABLE store_settings
    ADD COLUMN storefront_navigation JSONB NOT NULL DEFAULT '[]',
    ADD COLUMN contact_section_image_url VARCHAR(500);
