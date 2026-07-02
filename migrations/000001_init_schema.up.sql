-- 000001 — Core schema (normalized identity + commerce profile)
CREATE EXTENSION IF NOT EXISTS pgcrypto;

-- ============================================================
-- Enumerations
-- ============================================================

CREATE TYPE user_role AS ENUM ('admin', 'customer');
CREATE TYPE customer_type AS ENUM ('registered', 'guest');
CREATE TYPE address_type AS ENUM ('home', 'work', 'billing', 'shipping');
CREATE TYPE product_status AS ENUM ('draft', 'active', 'archived');
CREATE TYPE discount_type AS ENUM ('percentage', 'fixed_amount');
CREATE TYPE order_status AS ENUM ('pending', 'processing', 'shipped', 'delivered', 'cancelled', 'refunded');
CREATE TYPE payment_status AS ENUM ('unpaid', 'paid', 'refunded');
CREATE TYPE contact_message_source AS ENUM ('homepage', 'about', 'contact_page');
CREATE TYPE contact_message_status AS ENUM ('unread', 'read', 'archived');
CREATE TYPE product_review_status AS ENUM ('pending', 'approved', 'rejected');
CREATE TYPE product_question_status AS ENUM ('open', 'answered');
CREATE TYPE slide_type AS ENUM ('featured', 'bestseller', 'discounted');

-- ============================================================
-- Identity & access (single users table for admin + customer)
-- ============================================================

CREATE TABLE users (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email         VARCHAR(255) NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    first_name    VARCHAR(100) NOT NULL,
    last_name     VARCHAR(100) NOT NULL,
    phone         VARCHAR(20),
    role          user_role NOT NULL DEFAULT 'customer',
    is_active     BOOLEAN NOT NULL DEFAULT TRUE,
    last_login_at TIMESTAMPTZ,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at    TIMESTAMPTZ
);

CREATE UNIQUE INDEX idx_users_email_active ON users (LOWER(email)) WHERE deleted_at IS NULL;
CREATE INDEX idx_users_role ON users (role) WHERE deleted_at IS NULL;

-- Legacy RBAC tables (seeded; authorization enforced in application layer)
CREATE TABLE roles (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name        VARCHAR(50) NOT NULL UNIQUE,
    description TEXT,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE permissions (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name        VARCHAR(100) NOT NULL UNIQUE,
    description TEXT,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE user_roles (
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role_id UUID NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    PRIMARY KEY (user_id, role_id)
);

CREATE TABLE role_permissions (
    role_id       UUID NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    permission_id UUID NOT NULL REFERENCES permissions(id) ON DELETE CASCADE,
    PRIMARY KEY (role_id, permission_id)
);

CREATE TABLE refresh_tokens (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id    UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash VARCHAR(255) NOT NULL,
    family_id  UUID NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    revoked_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_refresh_tokens_hash ON refresh_tokens (token_hash);
CREATE INDEX idx_refresh_tokens_user ON refresh_tokens (user_id);

CREATE TABLE password_reset_tokens (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id    UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash VARCHAR(255) NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    used_at    TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_password_reset_tokens_hash ON password_reset_tokens (token_hash);
CREATE INDEX idx_password_reset_tokens_user ON password_reset_tokens (user_id);

-- ============================================================
-- Catalog
-- ============================================================

CREATE TABLE categories (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    parent_id   UUID REFERENCES categories(id) ON DELETE SET NULL,
    name        VARCHAR(200) NOT NULL,
    slug        VARCHAR(200) NOT NULL,
    description TEXT,
    image_url   VARCHAR(500),
    sort_order  INT NOT NULL DEFAULT 0,
    is_active   BOOLEAN NOT NULL DEFAULT TRUE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at  TIMESTAMPTZ
);

CREATE UNIQUE INDEX idx_categories_slug_active ON categories (slug) WHERE deleted_at IS NULL;
CREATE INDEX idx_categories_parent ON categories (parent_id) WHERE deleted_at IS NULL;

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

CREATE UNIQUE INDEX idx_brands_name_active ON brands (name) WHERE deleted_at IS NULL;
CREATE UNIQUE INDEX idx_brands_slug_active ON brands (slug) WHERE deleted_at IS NULL;

CREATE TABLE products (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    category_id       UUID REFERENCES categories(id) ON DELETE SET NULL,
    name              VARCHAR(300) NOT NULL,
    slug              VARCHAR(300) NOT NULL,
    description       TEXT,
    short_description VARCHAR(500),
    price             DECIMAL(12, 2) NOT NULL CHECK (price >= 0),
    sale_price        DECIMAL(12, 2) CHECK (sale_price IS NULL OR sale_price >= 0),
    brand             VARCHAR(100),
    is_featured       BOOLEAN NOT NULL DEFAULT FALSE,
    status            product_status NOT NULL DEFAULT 'draft',
    created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at        TIMESTAMPTZ
);

CREATE UNIQUE INDEX idx_products_slug_active ON products (slug) WHERE deleted_at IS NULL;
CREATE INDEX idx_products_category ON products (category_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_products_status ON products (status) WHERE deleted_at IS NULL;

CREATE TABLE product_images (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    product_id UUID NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    url        VARCHAR(500) NOT NULL,
    alt_text   VARCHAR(200),
    sort_order INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_product_images_product ON product_images (product_id);

CREATE TABLE inventories (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    product_id          UUID NOT NULL UNIQUE REFERENCES products(id) ON DELETE CASCADE,
    quantity            INT NOT NULL DEFAULT 0 CHECK (quantity >= 0),
    low_stock_threshold INT NOT NULL DEFAULT 10,
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE product_attributes (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    product_id UUID NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    name       VARCHAR(100) NOT NULL
);

CREATE INDEX idx_product_attributes_product ON product_attributes (product_id);

CREATE TABLE product_variant_attribute_values (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    attribute_id UUID NOT NULL REFERENCES product_attributes(id) ON DELETE CASCADE,
    value        VARCHAR(200) NOT NULL
);

CREATE INDEX idx_product_variant_attribute_values_attribute ON product_variant_attribute_values (attribute_id);

CREATE TABLE skus (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    product_id UUID NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    code       VARCHAR(100) NOT NULL UNIQUE,
    attributes JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_skus_product ON skus (product_id);

CREATE TABLE product_attribute_definitions (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name       VARCHAR(100) NOT NULL,
    slug       VARCHAR(100) NOT NULL,
    sort_order INT NOT NULL DEFAULT 0,
    is_active  BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE UNIQUE INDEX idx_product_attr_defs_name_active ON product_attribute_definitions (name) WHERE deleted_at IS NULL;
CREATE UNIQUE INDEX idx_product_attr_defs_slug_active ON product_attribute_definitions (slug) WHERE deleted_at IS NULL;

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

CREATE UNIQUE INDEX idx_product_attr_values_unique_active
    ON product_attribute_values (attribute_id, value) WHERE deleted_at IS NULL;
CREATE INDEX idx_product_attr_values_attribute ON product_attribute_values (attribute_id);

-- ============================================================
-- Commerce profiles (customers = storefront buyer profile)
-- Registered: identity lives in users; guest: contact snapshot stored here
-- ============================================================

CREATE TABLE customers (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id       UUID UNIQUE REFERENCES users(id) ON DELETE CASCADE,
    email         VARCHAR(255),
    first_name    VARCHAR(100),
    last_name     VARCHAR(100),
    phone         VARCHAR(20),
    type          customer_type NOT NULL,
    total_orders  INT NOT NULL DEFAULT 0 CHECK (total_orders >= 0),
    total_spent   DECIMAL(12, 2) NOT NULL DEFAULT 0 CHECK (total_spent >= 0),
    last_order_at TIMESTAMPTZ,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_customers_identity CHECK (
        (
            type = 'registered'
            AND user_id IS NOT NULL
            AND email IS NULL
            AND first_name IS NULL
            AND last_name IS NULL
            AND phone IS NULL
        )
        OR (
            type = 'guest'
            AND user_id IS NULL
            AND email IS NOT NULL
            AND first_name IS NOT NULL
            AND last_name IS NOT NULL
        )
    )
);

CREATE INDEX idx_customers_user_id ON customers (user_id) WHERE user_id IS NOT NULL;
CREATE INDEX idx_customers_guest_email ON customers (LOWER(email)) WHERE type = 'guest';
CREATE INDEX idx_customers_type ON customers (type);
CREATE INDEX idx_customers_created_at ON customers (created_at DESC);

CREATE TABLE customer_addresses (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    customer_id UUID NOT NULL REFERENCES customers(id) ON DELETE CASCADE,
    type        address_type NOT NULL,
    street      VARCHAR(300) NOT NULL,
    city        VARCHAR(100) NOT NULL,
    state       VARCHAR(100),
    postal_code VARCHAR(20) NOT NULL,
    country     CHAR(2) NOT NULL,
    is_default  BOOLEAN NOT NULL DEFAULT FALSE
);

CREATE INDEX idx_customer_addresses_customer ON customer_addresses (customer_id);

CREATE TABLE coupons (
    id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    code             VARCHAR(50) NOT NULL,
    discount_type    discount_type NOT NULL,
    discount_value   DECIMAL(12, 2) NOT NULL CHECK (discount_value > 0),
    min_order_amount DECIMAL(12, 2) NOT NULL DEFAULT 0,
    max_usage        INT,
    usage_count      INT NOT NULL DEFAULT 0,
    expires_at       TIMESTAMPTZ,
    is_active        BOOLEAN NOT NULL DEFAULT TRUE,
    note             TEXT,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at       TIMESTAMPTZ
);

CREATE UNIQUE INDEX idx_coupons_code_active ON coupons (LOWER(code)) WHERE deleted_at IS NULL;

CREATE TABLE orders (
    id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_number     VARCHAR(20) NOT NULL UNIQUE,
    customer_id      UUID NOT NULL REFERENCES customers(id),
    coupon_id        UUID REFERENCES coupons(id) ON DELETE SET NULL,
    status           order_status NOT NULL DEFAULT 'pending',
    payment_status   payment_status NOT NULL DEFAULT 'unpaid',
    subtotal         DECIMAL(12, 2) NOT NULL,
    discount_amount  DECIMAL(12, 2) NOT NULL DEFAULT 0,
    shipping_amount  DECIMAL(12, 2) NOT NULL DEFAULT 0,
    tax_amount       DECIMAL(12, 2) NOT NULL DEFAULT 0,
    total            DECIMAL(12, 2) NOT NULL,
    notes            TEXT,
    payment_method   VARCHAR(50),
    transaction_id   VARCHAR(100),
    billing_address  JSONB NOT NULL,
    shipping_address JSONB NOT NULL,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_orders_customer ON orders (customer_id);
CREATE INDEX idx_orders_status ON orders (status);
CREATE INDEX idx_orders_created_at ON orders (created_at DESC);
CREATE INDEX idx_orders_status_created ON orders (status, created_at DESC);
CREATE INDEX idx_orders_transaction ON orders (transaction_id) WHERE transaction_id IS NOT NULL;

CREATE TABLE order_items (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id     UUID NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    product_id   UUID NOT NULL REFERENCES products(id),
    product_name VARCHAR(300) NOT NULL,
    product_sku  VARCHAR(100) NOT NULL,
    quantity     INT NOT NULL CHECK (quantity > 0),
    unit_price   DECIMAL(12, 2) NOT NULL,
    total_price  DECIMAL(12, 2) NOT NULL
);

CREATE INDEX idx_order_items_order ON order_items (order_id);
CREATE INDEX idx_order_items_product ON order_items (product_id);

CREATE TABLE order_status_history (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id    UUID NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    from_status order_status,
    to_status   order_status NOT NULL,
    note        TEXT,
    changed_by  UUID REFERENCES users(id) ON DELETE SET NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_order_status_history_order ON order_status_history (order_id);

CREATE TABLE audit_logs (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id       UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    action        VARCHAR(50) NOT NULL,
    resource_type VARCHAR(50) NOT NULL,
    resource_id   VARCHAR(100) NOT NULL,
    old_value     JSONB,
    new_value     JSONB,
    ip_address    VARCHAR(45),
    user_agent    TEXT,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_audit_logs_user ON audit_logs (user_id);
CREATE INDEX idx_audit_logs_resource ON audit_logs (resource_type, resource_id);
CREATE INDEX idx_audit_logs_created ON audit_logs (created_at DESC);

-- ============================================================
-- Store settings
-- ============================================================

CREATE TABLE store_settings (
    id                       UUID PRIMARY KEY DEFAULT 'f0000000-0000-0000-0000-000000000001',
    site                     JSONB NOT NULL DEFAULT '{}',
    contact                  JSONB NOT NULL DEFAULT '{}',
    social                   JSONB NOT NULL DEFAULT '{}',
    seo                      JSONB NOT NULL DEFAULT '{}',
    about                    JSONB NOT NULL DEFAULT '{}',
    checkout                 JSONB NOT NULL DEFAULT '{
        "min_order_toman": 100000,
        "payment_methods": ["online", "cod"],
        "cod_enabled": true,
        "cod_cities": ["تهران", "کرج", "Tehran", "Karaj"],
        "currency_label": "تومان"
    }',
    navigation               JSONB NOT NULL DEFAULT '[]',
    storefront_navigation    JSONB NOT NULL DEFAULT '[]',
    contact_section_image_url VARCHAR(500),
    created_at               TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at               TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- ============================================================
-- Storefront content
-- ============================================================

CREATE TABLE storefront_hero (
    id                 UUID PRIMARY KEY DEFAULT 'a1000000-0000-0000-0000-000000000001',
    video_url          VARCHAR(500),
    title              VARCHAR(255),
    subtitle           TEXT,
    cta_primary_text   VARCHAR(100),
    cta_primary_url    VARCHAR(500),
    cta_secondary_text VARCHAR(100),
    cta_secondary_url  VARCHAR(500),
    is_active          BOOLEAN NOT NULL DEFAULT TRUE,
    updated_at         TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE product_slides (
    id                   UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    slide_type           slide_type NOT NULL UNIQUE,
    title                VARCHAR(255),
    autoplay_interval_ms INT NOT NULL DEFAULT 4500,
    is_active            BOOLEAN NOT NULL DEFAULT TRUE,
    sort_order           INT NOT NULL DEFAULT 0,
    created_at           TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at           TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE product_slide_items (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    slide_id   UUID NOT NULL REFERENCES product_slides(id) ON DELETE CASCADE,
    product_id UUID NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    sort_order INT NOT NULL DEFAULT 0,
    tab_label  VARCHAR(100),
    UNIQUE (slide_id, product_id)
);

CREATE INDEX idx_product_slide_items_slide ON product_slide_items (slide_id, sort_order);

CREATE TABLE pro_banners (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    desktop_image_url VARCHAR(500) NOT NULL,
    mobile_image_url  VARCHAR(500),
    link_url          VARCHAR(500),
    sort_order        INT NOT NULL DEFAULT 0,
    is_active         BOOLEAN NOT NULL DEFAULT TRUE,
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
    is_active   BOOLEAN NOT NULL DEFAULT TRUE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE homepage_reviews (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    customer_name VARCHAR(255) NOT NULL,
    photo_url     VARCHAR(500),
    review_text   TEXT NOT NULL,
    rating        SMALLINT CHECK (rating IS NULL OR rating BETWEEN 1 AND 5),
    sort_order    INT NOT NULL DEFAULT 0,
    is_active     BOOLEAN NOT NULL DEFAULT TRUE,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE faq_sections (
    id         UUID PRIMARY KEY DEFAULT 'c1000000-0000-0000-0000-000000000001',
    image_url  VARCHAR(500),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE faq_items (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    question   TEXT NOT NULL,
    answer     TEXT NOT NULL,
    sort_order INT NOT NULL DEFAULT 0,
    is_active  BOOLEAN NOT NULL DEFAULT TRUE
);

-- ============================================================
-- Themes
-- ============================================================

CREATE TABLE store_themes (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name              VARCHAR(255) NOT NULL,
    slug              VARCHAR(255) NOT NULL UNIQUE,
    description       TEXT,
    preview_image_url VARCHAR(500),
    price             DECIMAL(10, 2) NOT NULL DEFAULT 0,
    is_active         BOOLEAN NOT NULL DEFAULT TRUE,
    default_colors    JSONB NOT NULL DEFAULT '{}',
    default_font      VARCHAR(100),
    created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE theme_purchases (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    theme_id     UUID NOT NULL REFERENCES store_themes(id) ON DELETE CASCADE,
    purchased_by UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    purchased_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (theme_id, purchased_by)
);

CREATE TABLE store_style (
    id              UUID PRIMARY KEY DEFAULT 'e1000000-0000-0000-0000-000000000001',
    active_theme_id UUID REFERENCES store_themes(id) ON DELETE SET NULL,
    colors          JSONB NOT NULL DEFAULT '{}',
    font_family     VARCHAR(100),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- ============================================================
-- Engagement
-- ============================================================

CREATE TABLE contact_messages (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name       VARCHAR(255) NOT NULL,
    email      VARCHAR(255) NOT NULL,
    phone      VARCHAR(50),
    subject    VARCHAR(500),
    message    TEXT NOT NULL,
    source     contact_message_source NOT NULL DEFAULT 'homepage',
    status     contact_message_status NOT NULL DEFAULT 'unread',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_contact_messages_status ON contact_messages (status);
CREATE INDEX idx_contact_messages_created_at ON contact_messages (created_at DESC);

CREATE TABLE wishlist_items (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    customer_id UUID NOT NULL REFERENCES customers(id) ON DELETE CASCADE,
    product_id  UUID NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (customer_id, product_id)
);

CREATE INDEX idx_wishlist_items_customer ON wishlist_items (customer_id);

CREATE TABLE product_reviews (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    product_id  UUID NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    customer_id UUID REFERENCES customers(id) ON DELETE SET NULL,
    author_name VARCHAR(255) NOT NULL,
    rating      SMALLINT NOT NULL CHECK (rating BETWEEN 1 AND 5),
    title       VARCHAR(255),
    content     TEXT NOT NULL,
    status      product_review_status NOT NULL DEFAULT 'pending',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_product_reviews_product ON product_reviews (product_id);
CREATE INDEX idx_product_reviews_status ON product_reviews (status);

CREATE TABLE product_questions (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    product_id  UUID NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    asker_name  VARCHAR(255) NOT NULL,
    asker_email VARCHAR(255),
    question    TEXT NOT NULL,
    answer      TEXT,
    answered_at TIMESTAMPTZ,
    answered_by UUID REFERENCES users(id) ON DELETE SET NULL,
    status      product_question_status NOT NULL DEFAULT 'open',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_product_questions_product ON product_questions (product_id);
CREATE INDEX idx_product_questions_status ON product_questions (status);

-- ============================================================
-- Blog
-- ============================================================

CREATE TABLE blog_categories (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name        VARCHAR(255) NOT NULL,
    slug        VARCHAR(255) NOT NULL UNIQUE,
    description TEXT,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE blog_posts (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title          VARCHAR(255) NOT NULL,
    slug           VARCHAR(255) NOT NULL UNIQUE,
    content        TEXT NOT NULL,
    summary        TEXT,
    featured_image VARCHAR(500),
    category_id    UUID REFERENCES blog_categories(id) ON DELETE SET NULL,
    author_id      UUID REFERENCES users(id) ON DELETE SET NULL,
    status         VARCHAR(50) NOT NULL DEFAULT 'draft' CHECK (status IN ('draft', 'published')),
    published_at   TIMESTAMPTZ,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_blog_posts_category ON blog_posts (category_id);
CREATE INDEX idx_blog_posts_status ON blog_posts (status);
CREATE INDEX idx_blog_posts_published_at ON blog_posts (published_at DESC);

CREATE TABLE blog_comments (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    post_id      UUID NOT NULL REFERENCES blog_posts(id) ON DELETE CASCADE,
    author_name  VARCHAR(255) NOT NULL,
    author_email VARCHAR(255) NOT NULL,
    content      TEXT NOT NULL,
    status       VARCHAR(50) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'approved', 'rejected')),
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_blog_comments_post ON blog_comments (post_id);
CREATE INDEX idx_blog_comments_status ON blog_comments (status);

-- ============================================================
-- Carts (PostgreSQL-backed; replaces Redis)
-- ============================================================

CREATE TABLE carts (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     UUID REFERENCES users(id) ON DELETE CASCADE,
    guest_token VARCHAR(255),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_carts_owner CHECK (
        (user_id IS NOT NULL AND guest_token IS NULL)
        OR (user_id IS NULL AND guest_token IS NOT NULL)
    )
);

CREATE UNIQUE INDEX idx_carts_user_id ON carts (user_id) WHERE user_id IS NOT NULL;
CREATE UNIQUE INDEX idx_carts_guest_token ON carts (guest_token) WHERE guest_token IS NOT NULL;

CREATE TABLE cart_items (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    cart_id    UUID NOT NULL REFERENCES carts(id) ON DELETE CASCADE,
    product_id UUID NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    sku_id     UUID REFERENCES skus(id) ON DELETE SET NULL,
    quantity   INT NOT NULL CHECK (quantity > 0),
    added_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX idx_cart_items_unique_line
    ON cart_items (cart_id, product_id, COALESCE(sku_id, '00000000-0000-0000-0000-000000000000'::uuid));
CREATE INDEX idx_cart_items_cart ON cart_items (cart_id);
