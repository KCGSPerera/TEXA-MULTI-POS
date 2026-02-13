package repositories

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/kcgsperera/texa-multi-pos/backend/internal/models"
)

type TransferRepository interface {
	CreateTransferTx(ctx context.Context, tx pgx.Tx, tr *models.BranchTransfer) (*models.BranchTransfer, error)
	CreateTransferItemsTx(ctx context.Context, tx pgx.Tx, transferID string, items []models.BranchTransferItem) ([]models.BranchTransferItem, error)
	GetTransferForUpdateTx(ctx context.Context, tx pgx.Tx, transferID, branchID string) (*models.BranchTransfer, error)
	GetTransferItemsTx(ctx context.Context, tx pgx.Tx, transferID string) ([]models.BranchTransferItem, error)
	UpdateTransferStatusTx(ctx context.Context, tx pgx.Tx, transferID, status string) error
}

type transferRepository struct{ db *pgxpool.Pool }

func NewTransferRepository(db *pgxpool.Pool) TransferRepository { return &transferRepository{db: db} }

func (r *transferRepository) CreateTransferTx(ctx context.Context, tx pgx.Tx, tr *models.BranchTransfer) (*models.BranchTransfer, error) {
	const query = `
		INSERT INTO branch_transfers (from_branch_id, to_branch_id, status, created_by)
		VALUES ($1,$2,$3,$4)
		RETURNING id, from_branch_id, to_branch_id, status, created_by, created_at
	`
	out := &models.BranchTransfer{}
	if err := tx.QueryRow(ctx, query, tr.FromBranchID, tr.ToBranchID, tr.Status, tr.CreatedBy).Scan(&out.ID, &out.FromBranchID, &out.ToBranchID, &out.Status, &out.CreatedBy, &out.CreatedAt); err != nil {
		return nil, err
	}
	return out, nil
}

func (r *transferRepository) CreateTransferItemsTx(ctx context.Context, tx pgx.Tx, transferID string, items []models.BranchTransferItem) ([]models.BranchTransferItem, error) {
	const query = `INSERT INTO branch_transfer_items (transfer_id, product_id, quantity) VALUES ($1,$2,$3) RETURNING id, transfer_id, product_id, quantity::float8`
	out := make([]models.BranchTransferItem, 0, len(items))
	for _, i := range items {
		var row models.BranchTransferItem
		if err := tx.QueryRow(ctx, query, transferID, i.ProductID, i.Quantity).Scan(&row.ID, &row.TransferID, &row.ProductID, &row.Quantity); err != nil {
			return nil, err
		}
		out = append(out, row)
	}
	return out, nil
}

func (r *transferRepository) GetTransferForUpdateTx(ctx context.Context, tx pgx.Tx, transferID, branchID string) (*models.BranchTransfer, error) {
	const query = `
		SELECT id, from_branch_id, to_branch_id, status, created_by, created_at
		FROM branch_transfers
		WHERE id = $1 AND to_branch_id = $2
		FOR UPDATE
	`
	out := &models.BranchTransfer{}
	if err := tx.QueryRow(ctx, query, transferID, branchID).Scan(&out.ID, &out.FromBranchID, &out.ToBranchID, &out.Status, &out.CreatedBy, &out.CreatedAt); err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return out, nil
}

func (r *transferRepository) GetTransferItemsTx(ctx context.Context, tx pgx.Tx, transferID string) ([]models.BranchTransferItem, error) {
	const query = `SELECT id, transfer_id, product_id, quantity::float8 FROM branch_transfer_items WHERE transfer_id = $1`
	rows, err := tx.Query(ctx, query, transferID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]models.BranchTransferItem, 0)
	for rows.Next() {
		var i models.BranchTransferItem
		if err := rows.Scan(&i.ID, &i.TransferID, &i.ProductID, &i.Quantity); err != nil {
			return nil, err
		}
		out = append(out, i)
	}
	return out, rows.Err()
}

func (r *transferRepository) UpdateTransferStatusTx(ctx context.Context, tx pgx.Tx, transferID, status string) error {
	_, err := tx.Exec(ctx, `UPDATE branch_transfers SET status=$2 WHERE id=$1`, transferID, status)
	return err
}
