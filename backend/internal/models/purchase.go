package models

import "time"

type PurchaseOrder struct {
	ID          string              `json:"id"`
	BranchID    string              `json:"branch_id"`
	SupplierID  string              `json:"supplier_id"`
	Status      string              `json:"status"`
	TotalAmount float64             `json:"total_amount"`
	CreatedBy   *string             `json:"created_by,omitempty"`
	CreatedAt   time.Time           `json:"created_at"`
	UpdatedAt   time.Time           `json:"updated_at"`
	Items       []PurchaseOrderItem `json:"items"`
}

type PurchaseOrderItem struct {
	ID               string  `json:"id"`
	PurchaseOrderID  string  `json:"purchase_order_id"`
	ProductID        string  `json:"product_id"`
	Quantity         float64 `json:"quantity"`
	ReceivedQuantity float64 `json:"received_quantity"`
	UnitCost         float64 `json:"unit_cost"`
	LineTotal        float64 `json:"line_total"`
}

type CreatePurchaseOrderItemRequest struct {
	ProductID string  `json:"product_id" binding:"required,uuid"`
	Quantity  float64 `json:"quantity" binding:"required,gt=0"`
	UnitCost  float64 `json:"unit_cost" binding:"required,gte=0"`
	LineTotal float64 `json:"line_total" binding:"required,gte=0"`
}

type CreatePurchaseOrderRequest struct {
	SupplierID string                           `json:"supplier_id" binding:"required,uuid"`
	Items      []CreatePurchaseOrderItemRequest `json:"items" binding:"required,min=1,dive"`
}

type GRN struct {
	ID              string     `json:"id"`
	PurchaseOrderID string     `json:"purchase_order_id"`
	BranchID        string     `json:"branch_id"`
	SupplierID      string     `json:"supplier_id"`
	Status          string     `json:"status"`
	TotalAmount     float64    `json:"total_amount"`
	CreatedBy       *string    `json:"created_by,omitempty"`
	ApprovedBy      *string    `json:"approved_by,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
	ApprovedAt      *time.Time `json:"approved_at,omitempty"`
	Items           []GRNItem  `json:"items"`
}

type GRNItem struct {
	ID        string  `json:"id"`
	GRNID     string  `json:"grn_id"`
	ProductID string  `json:"product_id"`
	Quantity  float64 `json:"quantity"`
	UnitCost  float64 `json:"unit_cost"`
	LineTotal float64 `json:"line_total"`
}

type CreateGRNItemRequest struct {
	ProductID string  `json:"product_id" binding:"required,uuid"`
	Quantity  float64 `json:"quantity" binding:"required,gt=0"`
	UnitCost  float64 `json:"unit_cost" binding:"required,gte=0"`
	LineTotal float64 `json:"line_total" binding:"required,gte=0"`
}

type CreateGRNRequest struct {
	Items []CreateGRNItemRequest `json:"items" binding:"required,min=1,dive"`
}
