CREATE EXTENSION IF NOT EXISTS pgcrypto;

ALTER TABLE stock_movements DROP CONSTRAINT IF EXISTS ck_stock_movements_type;
ALTER TABLE stock_movements ADD CONSTRAINT ck_stock_movements_type CHECK (movement_type IN ('IN', 'OUT', 'ADJUSTMENT', 'TRANSFER_OUT', 'TRANSFER_IN'));

CREATE TABLE IF NOT EXISTS branch_transfers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    from_branch_id UUID NOT NULL,
    to_branch_id UUID NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'PENDING',
    created_by UUID,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT fk_branch_transfers_from FOREIGN KEY (from_branch_id) REFERENCES branches(id),
    CONSTRAINT fk_branch_transfers_to FOREIGN KEY (to_branch_id) REFERENCES branches(id),
    CONSTRAINT fk_branch_transfers_created_by FOREIGN KEY (created_by) REFERENCES users(id),
    CONSTRAINT ck_branch_transfers_status CHECK (status IN ('PENDING', 'IN_TRANSIT', 'COMPLETED')),
    CONSTRAINT ck_branch_transfers_branch_diff CHECK (from_branch_id <> to_branch_id)
);

CREATE INDEX IF NOT EXISTS idx_branch_transfers_from ON branch_transfers(from_branch_id);
CREATE INDEX IF NOT EXISTS idx_branch_transfers_to ON branch_transfers(to_branch_id);
CREATE INDEX IF NOT EXISTS idx_branch_transfers_status ON branch_transfers(status);

CREATE TABLE IF NOT EXISTS branch_transfer_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    transfer_id UUID NOT NULL,
    product_id UUID NOT NULL,
    quantity NUMERIC(14,3) NOT NULL,
    CONSTRAINT fk_branch_transfer_items_transfer FOREIGN KEY (transfer_id) REFERENCES branch_transfers(id) ON DELETE CASCADE,
    CONSTRAINT fk_branch_transfer_items_product FOREIGN KEY (product_id) REFERENCES products(id),
    CONSTRAINT ck_branch_transfer_items_qty_positive CHECK (quantity > 0)
);

CREATE INDEX IF NOT EXISTS idx_branch_transfer_items_transfer ON branch_transfer_items(transfer_id);
CREATE INDEX IF NOT EXISTS idx_branch_transfer_items_product ON branch_transfer_items(product_id);
