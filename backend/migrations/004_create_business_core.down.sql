DROP INDEX IF EXISTS idx_sale_items_product_id;
DROP TABLE IF EXISTS sale_items;

DROP INDEX IF EXISTS idx_sales_created_at;
DROP INDEX IF EXISTS idx_sales_user_id;
DROP INDEX IF EXISTS idx_sales_branch_id;
DROP TABLE IF EXISTS sales;

DROP INDEX IF EXISTS idx_customers_mobile_number;
DROP INDEX IF EXISTS idx_customers_branch_id;
DROP TABLE IF EXISTS customers;

DROP INDEX IF EXISTS idx_suppliers_name;
DROP INDEX IF EXISTS idx_suppliers_branch_id;
DROP TABLE IF EXISTS suppliers;

DROP INDEX IF EXISTS idx_products_sku;
DROP INDEX IF EXISTS idx_products_category_id;
DROP INDEX IF EXISTS idx_products_branch_id;
DROP TABLE IF EXISTS products;

DROP INDEX IF EXISTS idx_categories_name;
DROP INDEX IF EXISTS idx_categories_branch_id;
DROP TABLE IF EXISTS categories;
