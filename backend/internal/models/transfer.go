package models

import "time"

type BranchTransfer struct {
	ID           string               `json:"id"`
	FromBranchID string               `json:"from_branch_id"`
	ToBranchID   string               `json:"to_branch_id"`
	Status       string               `json:"status"`
	CreatedBy    *string              `json:"created_by,omitempty"`
	CreatedAt    time.Time            `json:"created_at"`
	Items        []BranchTransferItem `json:"items"`
}

type BranchTransferItem struct {
	ID         string  `json:"id"`
	TransferID string  `json:"transfer_id"`
	ProductID  string  `json:"product_id"`
	Quantity   float64 `json:"quantity"`
}

type CreateBranchTransferItemRequest struct {
	ProductID string  `json:"product_id" binding:"required,uuid"`
	Quantity  float64 `json:"quantity" binding:"required,gt=0"`
}

type CreateBranchTransferRequest struct {
	ToBranchID string                            `json:"to_branch_id" binding:"required,uuid"`
	Items      []CreateBranchTransferItemRequest `json:"items" binding:"required,min=1,dive"`
}
