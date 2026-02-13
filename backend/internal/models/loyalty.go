package models

import "time"

type LoyaltyTransaction struct {
	ID            string    `json:"id"`
	CustomerID    string    `json:"customer_id"`
	BranchID      string    `json:"branch_id"`
	Type          string    `json:"type"`
	Points        float64   `json:"points"`
	ReferenceType *string   `json:"reference_type,omitempty"`
	ReferenceID   *string   `json:"reference_id,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
}
