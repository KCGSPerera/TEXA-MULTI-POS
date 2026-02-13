package repositories

import (
	"context"
	"encoding/json"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type AuditRepository interface {
	Create(ctx context.Context, tx pgx.Tx, entityName, entityID, action string, performedBy, requestID *string, oldData, newData interface{}) error
}

type auditRepository struct {
	db *pgxpool.Pool
}

func NewAuditRepository(db *pgxpool.Pool) AuditRepository {
	return &auditRepository{db: db}
}

func (r *auditRepository) Create(ctx context.Context, tx pgx.Tx, entityName, entityID, action string, performedBy, requestID *string, oldData, newData interface{}) error {
	const query = `
		INSERT INTO audit_logs (entity_name, entity_id, action, performed_by, old_data, new_data, request_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	var oldJSON []byte
	var newJSON []byte
	var err error

	if oldData != nil {
		oldJSON, err = json.Marshal(oldData)
		if err != nil {
			return err
		}
	}
	if newData != nil {
		newJSON, err = json.Marshal(newData)
		if err != nil {
			return err
		}
	}

	if tx != nil {
		_, err = tx.Exec(ctx, query, entityName, entityID, action, performedBy, oldJSON, newJSON, requestID)
	} else {
		_, err = r.db.Exec(ctx, query, entityName, entityID, action, performedBy, oldJSON, newJSON, requestID)
	}

	return err
}
