-- 000013 — Blog CMS: blog_categories, blog_posts, blog_comments

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
    author_id      UUID REFERENCES admin_users(id) ON DELETE SET NULL,
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
