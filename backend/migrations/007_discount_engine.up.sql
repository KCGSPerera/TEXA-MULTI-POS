CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE IF NOT EXISTS discount_rules (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    branch_id UUID NOT NULL,
    name VARCHAR(150) NOT NULL,
    type VARCHAR(30) NOT NULL,
    value NUMERIC(14,2) NOT NULL,
    min_order_amount NUMERIC(14,2) NOT NULL DEFAULT 0,
    max_discount_amount NUMERIC(14,2),
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT fk_discount_rules_branch FOREIGN KEY (branch_id) REFERENCES branches(id) ON DELETE CASCADE,
    CONSTRAINT ck_discount_rules_type CHECK (type IN ('PERCENTAGE', 'FLAT', 'LOYALTY', 'PROMO')),
    CONSTRAINT ck_discount_rules_value_non_negative CHECK (value >= 0),
    CONSTRAINT ck_discount_rules_min_order_non_negative CHECK (min_order_amount >= 0),
    CONSTRAINT ck_discount_rules_max_non_negative CHECK (max_discount_amount IS NULL OR max_discount_amount >= 0)
);

CREATE INDEX IF NOT EXISTS idx_discount_rules_branch_id ON discount_rules(branch_id);
CREATE INDEX IF NOT EXISTS idx_discount_rules_type ON discount_rules(type);
CREATE INDEX IF NOT EXISTS idx_discount_rules_active ON discount_rules(is_active);

CREATE TABLE IF NOT EXISTS sale_discounts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    sale_id UUID NOT NULL,
    discount_rule_id UUID,
    discount_amount NUMERIC(14,2) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT fk_sale_discounts_sale FOREIGN KEY (sale_id) REFERENCES sales(id) ON DELETE CASCADE,
    CONSTRAINT fk_sale_discounts_rule FOREIGN KEY (discount_rule_id) REFERENCES discount_rules(id),
    CONSTRAINT ck_sale_discounts_amount_non_negative CHECK (discount_amount >= 0)
);

CREATE INDEX IF NOT EXISTS idx_sale_discounts_sale_id ON sale_discounts(sale_id);
CREATE INDEX IF NOT EXISTS idx_sale_discounts_rule_id ON sale_discounts(discount_rule_id);
