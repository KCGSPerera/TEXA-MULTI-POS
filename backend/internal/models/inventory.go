package models

import "time"

type Inventory struct {
	ProductID         string    `json:"product_id"`
	BranchID          string    `json:"branch_id"`
	Quantity          float64   `json:"quantity"`
	ReservedQuantity  float64   `json:"reserved_quantity"`
	AvailableQuantity float64   `json:"available_quantity"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

type StockMovement struct {
	ID            string    `json:"id"`
	ProductID     string    `json:"product_id"`
	BranchID      string    `json:"branch_id"`
	MovementType  string    `json:"movement_type"`
	Quantity      float64   `json:"quantity"`
	ReferenceType *string   `json:"reference_type,omitempty"`
	ReferenceID   *string   `json:"reference_id,omitempty"`
	CreatedBy     *string   `json:"created_by,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
}

type AdjustInventoryRequest struct {
	ProductID          string  `json:"product_id" binding:"required,uuid"`
	AdjustmentQuantity float64 `json:"adjustment_quantity" binding:"required,ne=0"`
	ReferenceType      string  `json:"reference_type" binding:"omitempty,max=30"`
	ReferenceID        *string `json:"reference_id" binding:"omitempty,uuid"`
}
