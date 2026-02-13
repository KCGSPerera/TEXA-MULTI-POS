DROP INDEX IF EXISTS idx_sale_discounts_rule_id;
DROP INDEX IF EXISTS idx_sale_discounts_sale_id;
DROP TABLE IF EXISTS sale_discounts;

DROP INDEX IF EXISTS idx_discount_rules_active;
DROP INDEX IF EXISTS idx_discount_rules_type;
DROP INDEX IF EXISTS idx_discount_rules_branch_id;
DROP TABLE IF EXISTS discount_rules;
