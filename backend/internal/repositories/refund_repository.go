package repositories

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/kcgsperera/texa-multi-pos/backend/internal/models"
)

type RefundRepository interface {
	CreateRefundTx(ctx context.Context, tx pgx.Tx, refund *models.Refund) (*models.Refund, error)
	CreateRefundItemsTx(ctx context.Context, tx pgx.Tx, refundID string, items []models.RefundItem) ([]models.RefundItem, error)
	GetSaleItemsTx(ctx context.Context, tx pgx.Tx, saleID string) ([]models.SaleItem, error)
	GetRefundedQtyBySaleProductTx(ctx context.Context, tx pgx.Tx, saleID, productID string) (float64, error)
}

type refundRepository struct{ db *pgxpool.Pool }

func NewRefundRepository(db *pgxpool.Pool) RefundRepository { return &refundRepository{db: db} }

func (r *refundRepository) CreateRefundTx(ctx context.Context, tx pgx.Tx, refund *models.Refund) (*models.Refund, error) {
	const query = `
		INSERT INTO refunds (sale_id, branch_id, total_amount, reason, created_by)
		VALUES ($1,$2,$3,$4,$5)
		RETURNING id, sale_id, branch_id, total_amount::float8, reason, created_by, created_at
	`
	out := &models.Refund{}
	if err := tx.QueryRow(ctx, query, refund.SaleID, refund.BranchID, refund.TotalAmount, refund.Reason, refund.CreatedBy).Scan(
		&out.ID, &out.SaleID, &out.BranchID, &out.TotalAmount, &out.Reason, &out.CreatedBy, &out.CreatedAt,
	); err != nil {
		return nil, err
	}
	return out, nil
}

func (r *refundRepository) CreateRefundItemsTx(ctx context.Context, tx pgx.Tx, refundID string, items []models.RefundItem) ([]models.RefundItem, error) {
	const query = `INSERT INTO refund_items (refund_id, product_id, quantity, line_total) VALUES ($1,$2,$3,$4) RETURNING id, refund_id, product_id, quantity::float8, line_total::float8`
	out := make([]models.RefundItem, 0, len(items))
	for _, i := range items {
		var row models.RefundItem
		if err := tx.QueryRow(ctx, query, refundID, i.ProductID, i.Quantity, i.LineTotal).Scan(&row.ID, &row.RefundID, &row.ProductID, &row.Quantity, &row.LineTotal); err != nil {
			return nil, err
		}
		out = append(out, row)
	}
	return out, nil
}

func (r *refundRepository) GetSaleItemsTx(ctx context.Context, tx pgx.Tx, saleID string) ([]models.SaleItem, error) {
	const query = `SELECT sale_id, product_id, quantity::float8, unit_price::float8, vat_amount::float8, line_total::float8 FROM sale_items WHERE sale_id = $1`
	rows, err := tx.Query(ctx, query, saleID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]models.SaleItem, 0)
	for rows.Next() {
		var i models.SaleItem
		if err := rows.Scan(&i.SaleID, &i.ProductID, &i.Quantity, &i.UnitPrice, &i.VATAmount, &i.LineTotal); err != nil {
			return nil, err
		}
		out = append(out, i)
	}
	return out, rows.Err()
}

func (r *refundRepository) GetRefundedQtyBySaleProductTx(ctx context.Context, tx pgx.Tx, saleID, productID string) (float64, error) {
	const query = `
		SELECT COALESCE(SUM(ri.quantity),0)::float8
		FROM refunds rf
		JOIN refund_items ri ON ri.refund_id = rf.id
		WHERE rf.sale_id = $1 AND ri.product_id = $2
	`
	var qty float64
	if err := tx.QueryRow(ctx, query, saleID, productID).Scan(&qty); err != nil {
		return 0, err
	}
	return qty, nil
}
