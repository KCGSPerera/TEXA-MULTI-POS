package models

import "time"

type Refund struct {
	ID          string       `json:"id"`
	SaleID      string       `json:"sale_id"`
	BranchID    string       `json:"branch_id"`
	TotalAmount float64      `json:"total_amount"`
	Reason      *string      `json:"reason,omitempty"`
	CreatedBy   *string      `json:"created_by,omitempty"`
	CreatedAt   time.Time    `json:"created_at"`
	Items       []RefundItem `json:"items"`
}

type RefundItem struct {
	ID        string  `json:"id"`
	RefundID  string  `json:"refund_id"`
	ProductID string  `json:"product_id"`
	Quantity  float64 `json:"quantity"`
	LineTotal float64 `json:"line_total"`
}

type CreateRefundItemRequest struct {
	ProductID string  `json:"product_id" binding:"required,uuid"`
	Quantity  float64 `json:"quantity" binding:"required,gt=0"`
	LineTotal float64 `json:"line_total" binding:"required,gte=0"`
}

type CreateRefundRequest struct {
	SaleID string                    `json:"sale_id" binding:"required,uuid"`
	Reason *string                   `json:"reason" binding:"omitempty,max=500"`
	Items  []CreateRefundItemRequest `json:"items" binding:"required,min=1,dive"`
}
