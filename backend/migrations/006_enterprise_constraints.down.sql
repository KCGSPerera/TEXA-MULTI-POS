DROP INDEX IF EXISTS idx_purchase_orders_branch_created_at;
DROP INDEX IF EXISTS idx_sale_items_product_id;
DROP INDEX IF EXISTS idx_sales_branch_created_at;

ALTER TABLE sales DROP CONSTRAINT IF EXISTS uq_sales_branch_idempotency_key;
ALTER TABLE sales DROP COLUMN IF EXISTS idempotency_key;

ALTER TABLE sales DROP CONSTRAINT IF EXISTS ck_sales_net_amount_non_negative;
ALTER TABLE sales DROP CONSTRAINT IF EXISTS ck_sales_total_amount_non_negative;
ALTER TABLE sales ADD CONSTRAINT ck_sales_net_amount_non_negative CHECK (net_amount >= 0);
ALTER TABLE sales ADD CONSTRAINT ck_sales_total_amount_non_negative CHECK (total_amount >= 0);

ALTER TABLE inventory DROP CONSTRAINT IF EXISTS inventory_pkey;
ALTER TABLE inventory ADD CONSTRAINT inventory_pkey PRIMARY KEY (product_id);
