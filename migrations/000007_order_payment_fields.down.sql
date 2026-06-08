DROP INDEX IF EXISTS idx_orders_transaction;

ALTER TABLE orders
    DROP COLUMN IF EXISTS payment_method,
    DROP COLUMN IF EXISTS transaction_id;
