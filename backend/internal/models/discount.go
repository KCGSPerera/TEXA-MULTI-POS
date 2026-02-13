package models

import "time"

type DiscountRule struct {
	ID                string    `json:"id"`
	BranchID          string    `json:"branch_id"`
	Name              string    `json:"name"`
	Type              string    `json:"type"`
	Value             float64   `json:"value"`
	MinOrderAmount    float64   `json:"min_order_amount"`
	MaxDiscountAmount *float64  `json:"max_discount_amount,omitempty"`
	IsActive          bool      `json:"is_active"`
	CreatedAt         time.Time `json:"created_at"`
}

type SaleDiscount struct {
	ID             string  `json:"id"`
	SaleID         string  `json:"sale_id"`
	DiscountRuleID *string `json:"discount_rule_id,omitempty"`
	DiscountAmount float64 `json:"discount_amount"`
}
