CREATE EXTENSION IF NOT EXISTS pgcrypto;

ALTER TABLE categories ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMPTZ NULL;
ALTER TABLE products ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMPTZ NULL;
ALTER TABLE suppliers ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMPTZ NULL;
ALTER TABLE customers ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMPTZ NULL;

CREATE INDEX IF NOT EXISTS idx_categories_deleted_at ON categories(deleted_at);
CREATE INDEX IF NOT EXISTS idx_products_deleted_at ON products(deleted_at);
CREATE INDEX IF NOT EXISTS idx_suppliers_deleted_at ON suppliers(deleted_at);
CREATE INDEX IF NOT EXISTS idx_customers_deleted_at ON customers(deleted_at);

CREATE TABLE IF NOT EXISTS inventory (
    product_id UUID PRIMARY KEY,
    branch_id UUID NOT NULL,
    quantity NUMERIC(14,3) NOT NULL DEFAULT 0,
    reserved_quantity NUMERIC(14,3) NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT fk_inventory_product FOREIGN KEY (product_id) REFERENCES products(id) ON DELETE CASCADE,
    CONSTRAINT fk_inventory_branch FOREIGN KEY (branch_id) REFERENCES branches(id) ON DELETE CASCADE,
    CONSTRAINT ck_inventory_quantity_non_negative CHECK (quantity >= 0),
    CONSTRAINT ck_inventory_reserved_quantity_non_negative CHECK (reserved_quantity >= 0)
);

CREATE INDEX IF NOT EXISTS idx_inventory_branch_id ON inventory(branch_id);
CREATE INDEX IF NOT EXISTS idx_inventory_branch_product ON inventory(branch_id, product_id);

CREATE TABLE IF NOT EXISTS stock_movements (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    product_id UUID NOT NULL,
    branch_id UUID NOT NULL,
    movement_type VARCHAR(20) NOT NULL,
    quantity NUMERIC(14,3) NOT NULL,
    reference_type VARCHAR(30),
    reference_id UUID,
    created_by UUID,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT fk_stock_movements_product FOREIGN KEY (product_id) REFERENCES products(id),
    CONSTRAINT fk_stock_movements_branch FOREIGN KEY (branch_id) REFERENCES branches(id),
    CONSTRAINT fk_stock_movements_created_by FOREIGN KEY (created_by) REFERENCES users(id),
    CONSTRAINT ck_stock_movements_type CHECK (movement_type IN ('IN', 'OUT', 'ADJUSTMENT')),
    CONSTRAINT ck_stock_movements_qty_positive CHECK (quantity > 0)
);

CREATE INDEX IF NOT EXISTS idx_stock_movements_product_id ON stock_movements(product_id);
CREATE INDEX IF NOT EXISTS idx_stock_movements_branch_id ON stock_movements(branch_id);
CREATE INDEX IF NOT EXISTS idx_stock_movements_movement_type ON stock_movements(movement_type);
CREATE INDEX IF NOT EXISTS idx_stock_movements_created_at ON stock_movements(created_at);
CREATE INDEX IF NOT EXISTS idx_stock_movements_branch_created_at ON stock_movements(branch_id, created_at);

CREATE TABLE IF NOT EXISTS purchase_orders (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    branch_id UUID NOT NULL,
    supplier_id UUID NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'OPEN',
    total_amount NUMERIC(14,2) NOT NULL DEFAULT 0,
    created_by UUID,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT fk_purchase_orders_branch FOREIGN KEY (branch_id) REFERENCES branches(id) ON DELETE CASCADE,
    CONSTRAINT fk_purchase_orders_supplier FOREIGN KEY (supplier_id) REFERENCES suppliers(id),
    CONSTRAINT fk_purchase_orders_created_by FOREIGN KEY (created_by) REFERENCES users(id),
    CONSTRAINT ck_purchase_orders_status CHECK (status IN ('OPEN', 'PARTIAL', 'COMPLETED', 'CANCELLED')),
    CONSTRAINT ck_purchase_orders_total_amount_non_negative CHECK (total_amount >= 0)
);

CREATE INDEX IF NOT EXISTS idx_purchase_orders_branch_id ON purchase_orders(branch_id);
CREATE INDEX IF NOT EXISTS idx_purchase_orders_supplier_id ON purchase_orders(supplier_id);
CREATE INDEX IF NOT EXISTS idx_purchase_orders_status ON purchase_orders(status);
CREATE INDEX IF NOT EXISTS idx_purchase_orders_branch_status ON purchase_orders(branch_id, status);

CREATE TABLE IF NOT EXISTS purchase_order_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    purchase_order_id UUID NOT NULL,
    product_id UUID NOT NULL,
    quantity NUMERIC(14,3) NOT NULL,
    received_quantity NUMERIC(14,3) NOT NULL DEFAULT 0,
    unit_cost NUMERIC(12,2) NOT NULL,
    line_total NUMERIC(14,2) NOT NULL,
    CONSTRAINT fk_purchase_order_items_order FOREIGN KEY (purchase_order_id) REFERENCES purchase_orders(id) ON DELETE CASCADE,
    CONSTRAINT fk_purchase_order_items_product FOREIGN KEY (product_id) REFERENCES products(id),
    CONSTRAINT uq_purchase_order_item UNIQUE (purchase_order_id, product_id),
    CONSTRAINT ck_purchase_order_items_quantity_positive CHECK (quantity > 0),
    CONSTRAINT ck_purchase_order_items_received_qty_non_negative CHECK (received_quantity >= 0),
    CONSTRAINT ck_purchase_order_items_unit_cost_non_negative CHECK (unit_cost >= 0),
    CONSTRAINT ck_purchase_order_items_line_total_non_negative CHECK (line_total >= 0)
);

CREATE INDEX IF NOT EXISTS idx_purchase_order_items_order_id ON purchase_order_items(purchase_order_id);
CREATE INDEX IF NOT EXISTS idx_purchase_order_items_product_id ON purchase_order_items(product_id);

CREATE TABLE IF NOT EXISTS grns (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    purchase_order_id UUID NOT NULL,
    branch_id UUID NOT NULL,
    supplier_id UUID NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'OPEN',
    total_amount NUMERIC(14,2) NOT NULL DEFAULT 0,
    created_by UUID,
    approved_by UUID,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    approved_at TIMESTAMPTZ NULL,
    CONSTRAINT fk_grns_purchase_order FOREIGN KEY (purchase_order_id) REFERENCES purchase_orders(id) ON DELETE CASCADE,
    CONSTRAINT fk_grns_branch FOREIGN KEY (branch_id) REFERENCES branches(id) ON DELETE CASCADE,
    CONSTRAINT fk_grns_supplier FOREIGN KEY (supplier_id) REFERENCES suppliers(id),
    CONSTRAINT fk_grns_created_by FOREIGN KEY (created_by) REFERENCES users(id),
    CONSTRAINT fk_grns_approved_by FOREIGN KEY (approved_by) REFERENCES users(id),
    CONSTRAINT ck_grns_status CHECK (status IN ('OPEN', 'PARTIAL', 'COMPLETED', 'CANCELLED')),
    CONSTRAINT ck_grns_total_amount_non_negative CHECK (total_amount >= 0)
);

CREATE INDEX IF NOT EXISTS idx_grns_purchase_order_id ON grns(purchase_order_id);
CREATE INDEX IF NOT EXISTS idx_grns_branch_id ON grns(branch_id);
CREATE INDEX IF NOT EXISTS idx_grns_status ON grns(status);

CREATE TABLE IF NOT EXISTS grn_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    grn_id UUID NOT NULL,
    product_id UUID NOT NULL,
    quantity NUMERIC(14,3) NOT NULL,
    unit_cost NUMERIC(12,2) NOT NULL,
    line_total NUMERIC(14,2) NOT NULL,
    CONSTRAINT fk_grn_items_grn FOREIGN KEY (grn_id) REFERENCES grns(id) ON DELETE CASCADE,
    CONSTRAINT fk_grn_items_product FOREIGN KEY (product_id) REFERENCES products(id),
    CONSTRAINT uq_grn_item UNIQUE (grn_id, product_id),
    CONSTRAINT ck_grn_items_quantity_positive CHECK (quantity > 0),
    CONSTRAINT ck_grn_items_unit_cost_non_negative CHECK (unit_cost >= 0),
    CONSTRAINT ck_grn_items_line_total_non_negative CHECK (line_total >= 0)
);

CREATE INDEX IF NOT EXISTS idx_grn_items_grn_id ON grn_items(grn_id);
CREATE INDEX IF NOT EXISTS idx_grn_items_product_id ON grn_items(product_id);

CREATE TABLE IF NOT EXISTS sale_payments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    sale_id UUID NOT NULL,
    payment_method VARCHAR(30) NOT NULL,
    amount NUMERIC(14,2) NOT NULL,
    paid_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT fk_sale_payments_sale FOREIGN KEY (sale_id) REFERENCES sales(id) ON DELETE CASCADE,
    CONSTRAINT ck_sale_payments_method CHECK (payment_method IN ('CASH', 'CARD', 'QR')),
    CONSTRAINT ck_sale_payments_amount_positive CHECK (amount > 0)
);

CREATE INDEX IF NOT EXISTS idx_sale_payments_sale_id ON sale_payments(sale_id);
CREATE INDEX IF NOT EXISTS idx_sale_payments_method ON sale_payments(payment_method);
CREATE INDEX IF NOT EXISTS idx_sale_payments_paid_at ON sale_payments(paid_at);

CREATE TABLE IF NOT EXISTS audit_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    entity_name VARCHAR(100) NOT NULL,
    entity_id UUID NOT NULL,
    action VARCHAR(20) NOT NULL,
    performed_by UUID,
    old_data JSONB,
    new_data JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT fk_audit_logs_performed_by FOREIGN KEY (performed_by) REFERENCES users(id),
    CONSTRAINT ck_audit_logs_action CHECK (action IN ('CREATE', 'UPDATE', 'DELETE'))
);

CREATE INDEX IF NOT EXISTS idx_audit_logs_entity ON audit_logs(entity_name, entity_id);
CREATE INDEX IF NOT EXISTS idx_audit_logs_created_at ON audit_logs(created_at);
