-- Seed permissions
INSERT INTO permissions (id, name, description) VALUES
    ('a0000000-0000-0000-0000-000000000001', 'dashboard:read', 'View dashboard statistics'),
    ('a0000000-0000-0000-0000-000000000002', 'products:read', 'View products'),
    ('a0000000-0000-0000-0000-000000000003', 'products:write', 'Create and update products'),
    ('a0000000-0000-0000-0000-000000000004', 'products:delete', 'Delete products'),
    ('a0000000-0000-0000-0000-000000000005', 'inventory:write', 'Manage product inventory'),
    ('a0000000-0000-0000-0000-000000000006', 'categories:read', 'View categories'),
    ('a0000000-0000-0000-0000-000000000007', 'categories:write', 'Create and update categories'),
    ('a0000000-0000-0000-0000-000000000008', 'categories:delete', 'Delete categories'),
    ('a0000000-0000-0000-0000-000000000009', 'orders:read', 'View orders'),
    ('a0000000-0000-0000-0000-00000000000a', 'orders:write', 'Update order status'),
    ('a0000000-0000-0000-0000-00000000000b', 'orders:cancel', 'Cancel orders'),
    ('a0000000-0000-0000-0000-00000000000c', 'orders:refund', 'Refund orders'),
    ('a0000000-0000-0000-0000-00000000000d', 'coupons:read', 'View coupons'),
    ('a0000000-0000-0000-0000-00000000000e', 'coupons:write', 'Create and update coupons'),
    ('a0000000-0000-0000-0000-00000000000f', 'coupons:delete', 'Delete coupons'),
    ('a0000000-0000-0000-0000-000000000010', 'customers:read', 'View customers'),
    ('a0000000-0000-0000-0000-000000000011', 'users:read', 'View admin users'),
    ('a0000000-0000-0000-0000-000000000012', 'users:write', 'Create and update admin users'),
    ('a0000000-0000-0000-0000-000000000013', 'users:delete', 'Delete admin users'),
    ('a0000000-0000-0000-0000-000000000014', 'users:manage_roles', 'Assign roles to admin users'),
    ('a0000000-0000-0000-0000-000000000015', 'audit:read', 'View audit logs');

-- Seed roles
INSERT INTO roles (id, name, description) VALUES
    ('b0000000-0000-0000-0000-000000000001', 'super_admin', 'Full system access'),
    ('b0000000-0000-0000-0000-000000000002', 'admin', 'Standard admin access'),
    ('b0000000-0000-0000-0000-000000000003', 'manager', 'Product and order management'),
    ('b0000000-0000-0000-0000-000000000004', 'support', 'Read-only support access');

INSERT INTO role_permissions (role_id, permission_id)
SELECT 'b0000000-0000-0000-0000-000000000001', id FROM permissions;

INSERT INTO role_permissions (role_id, permission_id)
SELECT 'b0000000-0000-0000-0000-000000000002', id FROM permissions
WHERE name NOT IN ('users:delete', 'users:manage_roles');

INSERT INTO role_permissions (role_id, permission_id)
SELECT 'b0000000-0000-0000-0000-000000000003', id FROM permissions
WHERE name IN (
    'dashboard:read', 'products:read', 'products:write', 'inventory:write',
    'categories:read', 'categories:write', 'orders:read', 'orders:write',
    'orders:cancel', 'coupons:read', 'customers:read'
);

INSERT INTO role_permissions (role_id, permission_id)
SELECT 'b0000000-0000-0000-0000-000000000004', id FROM permissions
WHERE name IN (
    'dashboard:read', 'products:read', 'categories:read',
    'orders:read', 'coupons:read', 'customers:read'
);

-- Default super admin (password: Admin@123456)
INSERT INTO users (id, email, password_hash, first_name, last_name, role, is_active)
VALUES (
    'c0000000-0000-0000-0000-000000000001',
    'admin@shop.com',
    '$2a$12$RZiAsIUKM1MczJ23.j0pX.gtg5uBUE3EZMZVUk.KoLGG4tANpxju2',
    'Admin',
    'User',
    'admin',
    TRUE
);

INSERT INTO user_roles (user_id, role_id)
VALUES ('c0000000-0000-0000-0000-000000000001', 'b0000000-0000-0000-0000-000000000001');

INSERT INTO store_settings (id) VALUES ('f0000000-0000-0000-0000-000000000001');

INSERT INTO product_slides (id, slide_type, title) VALUES
    ('b1000000-0000-0000-0000-000000000001', 'featured', 'Featured Products'),
    ('b1000000-0000-0000-0000-000000000002', 'bestseller', 'Bestsellers'),
    ('b1000000-0000-0000-0000-000000000003', 'discounted', 'Discounted Products');

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
