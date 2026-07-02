-- Deduplicate guest customers before unique indexes (keep most recent last_order_at).
DELETE FROM customers c
USING customers dup
WHERE c.type = 'guest'
  AND dup.type = 'guest'
  AND c.id <> dup.id
  AND LOWER(c.email) = LOWER(dup.email)
  AND c.email IS NOT NULL
  AND (
    COALESCE(c.last_order_at, c.created_at) < COALESCE(dup.last_order_at, dup.created_at)
    OR (
      COALESCE(c.last_order_at, c.created_at) = COALESCE(dup.last_order_at, dup.created_at)
      AND c.created_at < dup.created_at
    )
  );

DELETE FROM customers c
USING customers dup
WHERE c.type = 'guest'
  AND dup.type = 'guest'
  AND c.id <> dup.id
  AND c.phone IS NOT NULL
  AND c.phone <> ''
  AND c.phone = dup.phone
  AND (
    COALESCE(c.last_order_at, c.created_at) < COALESCE(dup.last_order_at, dup.created_at)
    OR (
      COALESCE(c.last_order_at, c.created_at) = COALESCE(dup.last_order_at, dup.created_at)
      AND c.created_at < dup.created_at
    )
  );

ALTER TABLE orders ADD COLUMN IF NOT EXISTS payment_expires_at TIMESTAMPTZ;

CREATE INDEX IF NOT EXISTS idx_orders_unpaid_expiry
  ON orders (payment_expires_at)
  WHERE payment_status = 'unpaid' AND status = 'pending';

CREATE UNIQUE INDEX IF NOT EXISTS idx_customers_guest_email_unique
  ON customers (LOWER(email))
  WHERE type = 'guest' AND email IS NOT NULL;

CREATE UNIQUE INDEX IF NOT EXISTS idx_customers_guest_phone_unique
  ON customers (phone)
  WHERE type = 'guest' AND phone IS NOT NULL AND phone <> '';
