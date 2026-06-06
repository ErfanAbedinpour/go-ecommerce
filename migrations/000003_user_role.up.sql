-- Simplify authorization: single role column on users (admin | customer).
-- Authorization is enforced at the application/router layer, not via DB permissions.

ALTER TABLE admin_users
    ADD COLUMN IF NOT EXISTS role VARCHAR(20) NOT NULL DEFAULT 'customer'
    CHECK (role IN ('admin', 'customer'));

UPDATE admin_users SET role = 'admin' WHERE email = 'admin@shop.com';
