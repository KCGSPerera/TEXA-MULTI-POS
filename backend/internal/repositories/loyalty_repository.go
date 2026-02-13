package repositories

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/kcgsperera/texa-multi-pos/backend/internal/models"
)

type LoyaltyRepository interface {
	CreateTx(ctx context.Context, tx pgx.Tx, t *models.LoyaltyTransaction) error
	GetBalanceTx(ctx context.Context, tx pgx.Tx, customerID, branchID string) (float64, error)
	UpdateCustomerPointsTx(ctx context.Context, tx pgx.Tx, customerID, branchID string, points float64) error
}

type loyaltyRepository struct{ db *pgxpool.Pool }

func NewLoyaltyRepository(db *pgxpool.Pool) LoyaltyRepository { return &loyaltyRepository{db: db} }

func (r *loyaltyRepository) CreateTx(ctx context.Context, tx pgx.Tx, t *models.LoyaltyTransaction) error {
	const query = `INSERT INTO loyalty_transactions (customer_id, branch_id, type, points, reference_type, reference_id) VALUES ($1,$2,$3,$4,$5,$6)`
	_, err := tx.Exec(ctx, query, t.CustomerID, t.BranchID, t.Type, t.Points, t.ReferenceType, t.ReferenceID)
	return err
}

func (r *loyaltyRepository) GetBalanceTx(ctx context.Context, tx pgx.Tx, customerID, branchID string) (float64, error) {
	const query = `
		SELECT COALESCE(SUM(CASE WHEN type='REDEEM' THEN -points ELSE points END),0)::float8
		FROM loyalty_transactions
		WHERE customer_id = $1 AND branch_id = $2
	`
	var points float64
	if err := tx.QueryRow(ctx, query, customerID, branchID).Scan(&points); err != nil {
		return 0, err
	}
	return points, nil
}

func (r *loyaltyRepository) UpdateCustomerPointsTx(ctx context.Context, tx pgx.Tx, customerID, branchID string, points float64) error {
	const query = `UPDATE customers SET loyalty_points = $1, updated_at = now() WHERE id = $2 AND branch_id = $3 AND deleted_at IS NULL`
	_, err := tx.Exec(ctx, query, points, customerID, branchID)
	return err
}
