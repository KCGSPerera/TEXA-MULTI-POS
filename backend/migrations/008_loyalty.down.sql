DROP INDEX IF EXISTS idx_loyalty_tx_created_at;
DROP INDEX IF EXISTS idx_loyalty_tx_branch_id;
DROP INDEX IF EXISTS idx_loyalty_tx_customer_id;
DROP TABLE IF EXISTS loyalty_transactions;

ALTER TABLE sales DROP CONSTRAINT IF EXISTS fk_sales_customer;
ALTER TABLE sales DROP COLUMN IF EXISTS customer_id;
