CREATE TABLE store_themes (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name              VARCHAR(255) NOT NULL,
    slug              VARCHAR(255) NOT NULL UNIQUE,
    description       TEXT,
    preview_image_url VARCHAR(500),
    price             DECIMAL(10,2) NOT NULL DEFAULT 0,
    is_active         BOOLEAN NOT NULL DEFAULT true,
    default_colors    JSONB NOT NULL DEFAULT '{}',
    default_font      VARCHAR(100),
    created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE theme_purchases (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    theme_id     UUID NOT NULL REFERENCES store_themes(id) ON DELETE CASCADE,
    purchased_by UUID NOT NULL REFERENCES admin_users(id) ON DELETE CASCADE,
    purchased_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (theme_id, purchased_by)
);

CREATE TABLE store_style (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    active_theme_id UUID REFERENCES store_themes(id) ON DELETE SET NULL,
    colors          JSONB NOT NULL DEFAULT '{}',
    font_family     VARCHAR(100),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

INSERT INTO store_themes (id, name, slug, description, price, default_colors, default_font) VALUES
(
    'd1000000-0000-0000-0000-000000000001',
    'Modern Blue',
    'modern-blue',
    'Clean modern storefront with blue accents',
    0,
    '{"primary":"#062fe2","primary_foreground":"#ffffff","secondary":"#277ee4","secondary_foreground":"#ffffff","accent":"#277ee4","accent_foreground":"#ffffff","background":"#fcfdfc","foreground":"#1a1a1a","muted":"#f4f4f5","muted_foreground":"#71717a","border":"#e4e4e7","destructive":"#ef4444"}',
    'Inter'
),
(
    'd1000000-0000-0000-0000-000000000002',
    'Minimal Light',
    'minimal-light',
    'Minimal light theme for product-focused stores',
    0,
    '{"primary":"#18181b","primary_foreground":"#fafafa","secondary":"#f4f4f5","secondary_foreground":"#18181b","accent":"#3f3f46","accent_foreground":"#fafafa","background":"#ffffff","foreground":"#09090b","muted":"#f4f4f5","muted_foreground":"#71717a","border":"#e4e4e7","destructive":"#dc2626"}',
    'Inter'
),
(
    'd1000000-0000-0000-0000-000000000003',
    'Bold Dark',
    'bold-dark',
    'Bold dark theme with high contrast',
    0,
    '{"primary":"#3b82f6","primary_foreground":"#ffffff","secondary":"#1e293b","secondary_foreground":"#f8fafc","accent":"#60a5fa","accent_foreground":"#0f172a","background":"#0f172a","foreground":"#f8fafc","muted":"#1e293b","muted_foreground":"#94a3b8","border":"#334155","destructive":"#f87171"}',
    'Inter'
);

INSERT INTO store_style (id, active_theme_id, colors, font_family)
VALUES (
    'e1000000-0000-0000-0000-000000000001',
    'd1000000-0000-0000-0000-000000000001',
    '{"primary":"#062fe2","primary_foreground":"#ffffff","secondary":"#277ee4","secondary_foreground":"#ffffff","accent":"#277ee4","accent_foreground":"#ffffff","background":"#fcfdfc","foreground":"#1a1a1a","muted":"#f4f4f5","muted_foreground":"#71717a","border":"#e4e4e7","destructive":"#ef4444"}',
    'Inter'
);
