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

-- Assign all permissions to super_admin
INSERT INTO role_permissions (role_id, permission_id)
SELECT 'b0000000-0000-0000-0000-000000000001', id FROM permissions;

-- Assign permissions to admin (all except users:delete and users:manage_roles)
INSERT INTO role_permissions (role_id, permission_id)
SELECT 'b0000000-0000-0000-0000-000000000002', id FROM permissions
WHERE name NOT IN ('users:delete', 'users:manage_roles');

-- Assign permissions to manager
INSERT INTO role_permissions (role_id, permission_id)
SELECT 'b0000000-0000-0000-0000-000000000003', id FROM permissions
WHERE name IN (
    'dashboard:read', 'products:read', 'products:write', 'inventory:write',
    'categories:read', 'categories:write', 'orders:read', 'orders:write',
    'orders:cancel', 'coupons:read', 'customers:read'
);

-- Assign permissions to support
INSERT INTO role_permissions (role_id, permission_id)
SELECT 'b0000000-0000-0000-0000-000000000004', id FROM permissions
WHERE name IN (
    'dashboard:read', 'products:read', 'categories:read',
    'orders:read', 'coupons:read', 'customers:read'
);

-- Seed default super admin (password: Admin@123456)
-- bcrypt hash of "Admin@123456" with cost 12
INSERT INTO admin_users (id, email, password_hash, first_name, last_name, is_active)
VALUES (
    'c0000000-0000-0000-0000-000000000001',
    'admin@shop.com',
    '$2a$12$RZiAsIUKM1MczJ23.j0pX.gtg5uBUE3EZMZVUk.KoLGG4tANpxju2',
    'Admin',
    'User',
    TRUE
);

INSERT INTO admin_user_roles (admin_user_id, role_id)
VALUES ('c0000000-0000-0000-0000-000000000001', 'b0000000-0000-0000-0000-000000000001');
