CREATE EXTENSION IF NOT EXISTS pgcrypto;

ALTER TABLE sales ADD COLUMN IF NOT EXISTS customer_id UUID NULL;
ALTER TABLE sales DROP CONSTRAINT IF EXISTS fk_sales_customer;
ALTER TABLE sales ADD CONSTRAINT fk_sales_customer FOREIGN KEY (customer_id) REFERENCES customers(id);

CREATE TABLE IF NOT EXISTS loyalty_transactions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    customer_id UUID NOT NULL,
    branch_id UUID NOT NULL,
    type VARCHAR(20) NOT NULL,
    points NUMERIC(14,2) NOT NULL,
    reference_type VARCHAR(30),
    reference_id UUID,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT fk_loyalty_tx_customer FOREIGN KEY (customer_id) REFERENCES customers(id),
    CONSTRAINT fk_loyalty_tx_branch FOREIGN KEY (branch_id) REFERENCES branches(id),
    CONSTRAINT ck_loyalty_tx_type CHECK (type IN ('EARN', 'REDEEM', 'ADJUST')),
    CONSTRAINT ck_loyalty_tx_points_non_negative CHECK (points >= 0)
);

CREATE INDEX IF NOT EXISTS idx_loyalty_tx_customer_id ON loyalty_transactions(customer_id);
CREATE INDEX IF NOT EXISTS idx_loyalty_tx_branch_id ON loyalty_transactions(branch_id);
CREATE INDEX IF NOT EXISTS idx_loyalty_tx_created_at ON loyalty_transactions(created_at);
