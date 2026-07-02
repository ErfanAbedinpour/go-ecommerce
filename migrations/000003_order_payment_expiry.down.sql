DROP INDEX IF EXISTS idx_customers_guest_phone_unique;
DROP INDEX IF EXISTS idx_customers_guest_email_unique;
DROP INDEX IF EXISTS idx_orders_unpaid_expiry;
ALTER TABLE orders DROP COLUMN IF EXISTS payment_expires_at;
