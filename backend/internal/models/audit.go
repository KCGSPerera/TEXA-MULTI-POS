package models

import "time"

type AuditLog struct {
	ID          string    `json:"id"`
	EntityName  string    `json:"entity_name"`
	EntityID    string    `json:"entity_id"`
	Action      string    `json:"action"`
	PerformedBy *string   `json:"performed_by,omitempty"`
	OldData     []byte    `json:"old_data,omitempty"`
	NewData     []byte    `json:"new_data,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}
