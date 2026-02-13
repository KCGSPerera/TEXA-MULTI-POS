CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE IF NOT EXISTS refunds (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    sale_id UUID NOT NULL,
    branch_id UUID NOT NULL,
    total_amount NUMERIC(14,2) NOT NULL,
    reason TEXT,
    created_by UUID,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT fk_refunds_sale FOREIGN KEY (sale_id) REFERENCES sales(id),
    CONSTRAINT fk_refunds_branch FOREIGN KEY (branch_id) REFERENCES branches(id),
    CONSTRAINT fk_refunds_created_by FOREIGN KEY (created_by) REFERENCES users(id),
    CONSTRAINT ck_refunds_total_non_negative CHECK (total_amount >= 0)
);

CREATE INDEX IF NOT EXISTS idx_refunds_sale_id ON refunds(sale_id);
CREATE INDEX IF NOT EXISTS idx_refunds_branch_id ON refunds(branch_id);

CREATE TABLE IF NOT EXISTS refund_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    refund_id UUID NOT NULL,
    product_id UUID NOT NULL,
    quantity NUMERIC(14,3) NOT NULL,
    line_total NUMERIC(14,2) NOT NULL,
    CONSTRAINT fk_refund_items_refund FOREIGN KEY (refund_id) REFERENCES refunds(id) ON DELETE CASCADE,
    CONSTRAINT fk_refund_items_product FOREIGN KEY (product_id) REFERENCES products(id),
    CONSTRAINT ck_refund_items_qty_positive CHECK (quantity > 0),
    CONSTRAINT ck_refund_items_line_total_non_negative CHECK (line_total >= 0)
);

CREATE INDEX IF NOT EXISTS idx_refund_items_refund_id ON refund_items(refund_id);
CREATE INDEX IF NOT EXISTS idx_refund_items_product_id ON refund_items(product_id);
