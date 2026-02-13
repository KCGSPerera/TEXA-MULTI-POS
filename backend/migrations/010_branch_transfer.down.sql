DROP INDEX IF EXISTS idx_branch_transfer_items_product;
DROP INDEX IF EXISTS idx_branch_transfer_items_transfer;
DROP TABLE IF EXISTS branch_transfer_items;

DROP INDEX IF EXISTS idx_branch_transfers_status;
DROP INDEX IF EXISTS idx_branch_transfers_to;
DROP INDEX IF EXISTS idx_branch_transfers_from;
DROP TABLE IF EXISTS branch_transfers;

ALTER TABLE stock_movements DROP CONSTRAINT IF EXISTS ck_stock_movements_type;
ALTER TABLE stock_movements ADD CONSTRAINT ck_stock_movements_type CHECK (movement_type IN ('IN', 'OUT', 'ADJUSTMENT'));
