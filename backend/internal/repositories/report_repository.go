package repositories

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/kcgsperera/texa-multi-pos/backend/internal/models"
)

type ReportRepository interface {
	DailySalesSummary(ctx context.Context, branchID string, from, to time.Time) (*models.DailySalesSummary, error)
	ProductSalesReport(ctx context.Context, branchID string, from, to time.Time) ([]models.ProductSalesRow, error)
	PaymentMethodSummary(ctx context.Context, branchID string, from, to time.Time) ([]models.PaymentMethodSummaryRow, error)
	StockValuationReport(ctx context.Context, branchID string) ([]models.StockValuationRow, error)
}

type reportRepository struct {
	db *pgxpool.Pool
}

func NewReportRepository(db *pgxpool.Pool) ReportRepository {
	return &reportRepository{db: db}
}

func (r *reportRepository) DailySalesSummary(ctx context.Context, branchID string, from, to time.Time) (*models.DailySalesSummary, error) {
	const query = `
		SELECT
			COALESCE(SUM(total_amount), 0)::float8,
			COALESCE(SUM(net_amount), 0)::float8,
			COALESCE(SUM(vat_amount), 0)::float8,
			COALESCE(SUM(discount_amount), 0)::float8,
			COUNT(*)::bigint
		FROM sales
		WHERE branch_id = $1
		  AND created_at >= $2
		  AND created_at < $3
	`

	out := &models.DailySalesSummary{}
	if err := r.db.QueryRow(ctx, query, branchID, from, to).Scan(
		&out.TotalSales,
		&out.NetSales,
		&out.VATAmount,
		&out.DiscountAmount,
		&out.SaleCount,
	); err != nil {
		return nil, err
	}

	return out, nil
}

func (r *reportRepository) ProductSalesReport(ctx context.Context, branchID string, from, to time.Time) ([]models.ProductSalesRow, error) {
	const query = `
		SELECT
			p.id,
			p.name,
			p.sku,
			COALESCE(SUM(si.quantity), 0)::float8 AS quantity_sold,
			COALESCE(SUM(si.line_total), 0)::float8 AS sales_amount
		FROM sales s
		JOIN sale_items si ON si.sale_id = s.id
		JOIN products p ON p.id = si.product_id
		WHERE s.branch_id = $1
		  AND s.created_at >= $2
		  AND s.created_at < $3
		GROUP BY p.id, p.name, p.sku
		ORDER BY sales_amount DESC
	`

	rows, err := r.db.Query(ctx, query, branchID, from, to)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]models.ProductSalesRow, 0)
	for rows.Next() {
		var row models.ProductSalesRow
		if err := rows.Scan(&row.ProductID, &row.ProductName, &row.SKU, &row.QuantitySold, &row.SalesAmount); err != nil {
			return nil, err
		}
		out = append(out, row)
	}
	return out, rows.Err()
}

func (r *reportRepository) PaymentMethodSummary(ctx context.Context, branchID string, from, to time.Time) ([]models.PaymentMethodSummaryRow, error) {
	const query = `
		SELECT sp.payment_method, COALESCE(SUM(sp.amount), 0)::float8
		FROM sale_payments sp
		JOIN sales s ON s.id = sp.sale_id
		WHERE s.branch_id = $1
		  AND s.created_at >= $2
		  AND s.created_at < $3
		GROUP BY sp.payment_method
		ORDER BY sp.payment_method
	`

	rows, err := r.db.Query(ctx, query, branchID, from, to)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]models.PaymentMethodSummaryRow, 0)
	for rows.Next() {
		var row models.PaymentMethodSummaryRow
		if err := rows.Scan(&row.PaymentMethod, &row.TotalAmount); err != nil {
			return nil, err
		}
		out = append(out, row)
	}
	return out, rows.Err()
}

func (r *reportRepository) StockValuationReport(ctx context.Context, branchID string) ([]models.StockValuationRow, error) {
	const query = `
		SELECT
			p.id,
			p.name,
			p.sku,
			i.quantity::float8,
			p.cost_price::float8,
			(i.quantity * p.cost_price)::float8 AS stock_value
		FROM inventory i
		JOIN products p ON p.id = i.product_id
		WHERE i.branch_id = $1
		  AND p.deleted_at IS NULL
		ORDER BY stock_value DESC
	`

	rows, err := r.db.Query(ctx, query, branchID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]models.StockValuationRow, 0)
	for rows.Next() {
		var row models.StockValuationRow
		if err := rows.Scan(&row.ProductID, &row.ProductName, &row.SKU, &row.Quantity, &row.CostPrice, &row.StockValue); err != nil {
			return nil, err
		}
		out = append(out, row)
	}
	return out, rows.Err()
}
