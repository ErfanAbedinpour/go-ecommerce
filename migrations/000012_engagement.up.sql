-- 000012 — Engagement Features: contact_messages, wishlist_items, product_reviews, product_questions

-- Contact Messages
CREATE TYPE contact_message_source AS ENUM ('homepage', 'about', 'contact_page');
CREATE TYPE contact_message_status AS ENUM ('unread', 'read', 'archived');

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

CREATE INDEX idx_contact_messages_status     ON contact_messages (status);
CREATE INDEX idx_contact_messages_created_at ON contact_messages (created_at DESC);

-- Wishlist Items
CREATE TABLE wishlist_items (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    customer_id UUID NOT NULL REFERENCES customers(id) ON DELETE CASCADE,
    product_id  UUID NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (customer_id, product_id)
);

CREATE INDEX idx_wishlist_items_customer ON wishlist_items (customer_id);

-- Product Reviews
CREATE TYPE product_review_status AS ENUM ('pending', 'approved', 'rejected');

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
CREATE INDEX idx_product_reviews_status  ON product_reviews (status);

-- Product Questions
CREATE TYPE product_question_status AS ENUM ('open', 'answered');

CREATE TABLE product_questions (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    product_id  UUID NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    asker_name  VARCHAR(255) NOT NULL,
    asker_email VARCHAR(255),
    question    TEXT NOT NULL,
    answer      TEXT,
    answered_at TIMESTAMPTZ,
    answered_by UUID REFERENCES admin_users(id) ON DELETE SET NULL,
    status      product_question_status NOT NULL DEFAULT 'open',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_product_questions_product ON product_questions (product_id);
CREATE INDEX idx_product_questions_status  ON product_questions (status);
