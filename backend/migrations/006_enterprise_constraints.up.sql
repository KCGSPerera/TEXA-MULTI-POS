ALTER TABLE inventory DROP CONSTRAINT IF EXISTS inventory_pkey;
ALTER TABLE inventory ADD CONSTRAINT inventory_pkey PRIMARY KEY (branch_id, product_id);

ALTER TABLE sales ADD COLUMN IF NOT EXISTS idempotency_key VARCHAR(100) NULL;

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint WHERE conname = 'uq_sales_branch_idempotency_key'
    ) THEN
        ALTER TABLE sales ADD CONSTRAINT uq_sales_branch_idempotency_key UNIQUE (branch_id, idempotency_key);
    END IF;
END $$;

ALTER TABLE sales DROP CONSTRAINT IF EXISTS ck_sales_net_amount_non_negative;
ALTER TABLE sales DROP CONSTRAINT IF EXISTS ck_sales_total_amount_non_negative;
ALTER TABLE sales ADD CONSTRAINT ck_sales_net_amount_non_negative CHECK (net_amount >= 0);
ALTER TABLE sales ADD CONSTRAINT ck_sales_total_amount_non_negative CHECK (total_amount >= 0);

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint WHERE conname = 'uq_products_branch_sku'
    ) THEN
        ALTER TABLE products ADD CONSTRAINT uq_products_branch_sku UNIQUE (branch_id, sku);
    END IF;
END $$;

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint WHERE conname = 'uq_categories_branch_name'
    ) THEN
        ALTER TABLE categories ADD CONSTRAINT uq_categories_branch_name UNIQUE (branch_id, name);
    END IF;
END $$;

CREATE INDEX IF NOT EXISTS idx_sales_branch_created_at ON sales(branch_id, created_at);
CREATE INDEX IF NOT EXISTS idx_sale_items_product_id ON sale_items(product_id);
CREATE INDEX IF NOT EXISTS idx_purchase_orders_branch_created_at ON purchase_orders(branch_id, created_at);
