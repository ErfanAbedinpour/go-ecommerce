ALTER TABLE orders
    ADD COLUMN payment_method  VARCHAR(50),
    ADD COLUMN transaction_id  VARCHAR(100);

CREATE INDEX idx_orders_transaction ON orders(transaction_id) WHERE transaction_id IS NOT NULL;
