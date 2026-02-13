DROP INDEX IF EXISTS idx_audit_logs_created_at;
DROP INDEX IF EXISTS idx_audit_logs_entity;
DROP TABLE IF EXISTS audit_logs;

DROP INDEX IF EXISTS idx_sale_payments_paid_at;
DROP INDEX IF EXISTS idx_sale_payments_method;
DROP INDEX IF EXISTS idx_sale_payments_sale_id;
DROP TABLE IF EXISTS sale_payments;

DROP INDEX IF EXISTS idx_grn_items_product_id;
DROP INDEX IF EXISTS idx_grn_items_grn_id;
DROP TABLE IF EXISTS grn_items;

DROP INDEX IF EXISTS idx_grns_status;
DROP INDEX IF EXISTS idx_grns_branch_id;
DROP INDEX IF EXISTS idx_grns_purchase_order_id;
DROP TABLE IF EXISTS grns;

DROP INDEX IF EXISTS idx_purchase_order_items_product_id;
DROP INDEX IF EXISTS idx_purchase_order_items_order_id;
DROP TABLE IF EXISTS purchase_order_items;

DROP INDEX IF EXISTS idx_purchase_orders_branch_status;
DROP INDEX IF EXISTS idx_purchase_orders_status;
DROP INDEX IF EXISTS idx_purchase_orders_supplier_id;
DROP INDEX IF EXISTS idx_purchase_orders_branch_id;
DROP TABLE IF EXISTS purchase_orders;

DROP INDEX IF EXISTS idx_stock_movements_branch_created_at;
DROP INDEX IF EXISTS idx_stock_movements_created_at;
DROP INDEX IF EXISTS idx_stock_movements_movement_type;
DROP INDEX IF EXISTS idx_stock_movements_branch_id;
DROP INDEX IF EXISTS idx_stock_movements_product_id;
DROP TABLE IF EXISTS stock_movements;

DROP INDEX IF EXISTS idx_inventory_branch_product;
DROP INDEX IF EXISTS idx_inventory_branch_id;
DROP TABLE IF EXISTS inventory;

DROP INDEX IF EXISTS idx_customers_deleted_at;
DROP INDEX IF EXISTS idx_suppliers_deleted_at;
DROP INDEX IF EXISTS idx_products_deleted_at;
DROP INDEX IF EXISTS idx_categories_deleted_at;

ALTER TABLE customers DROP COLUMN IF EXISTS deleted_at;
ALTER TABLE suppliers DROP COLUMN IF EXISTS deleted_at;
ALTER TABLE products DROP COLUMN IF EXISTS deleted_at;
ALTER TABLE categories DROP COLUMN IF EXISTS deleted_at;
