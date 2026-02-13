package repositories

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/kcgsperera/texa-multi-pos/backend/internal/models"
)

type SaleRepository interface {
	CreateSaleTx(ctx context.Context, tx pgx.Tx, sale *models.Sale) (*models.Sale, error)
	CreateSaleItemsTx(ctx context.Context, tx pgx.Tx, saleID string, items []models.SaleItem) error
	CreateSalePaymentsTx(ctx context.Context, tx pgx.Tx, saleID string, payments []models.SalePayment) ([]models.SalePayment, error)
	GetByIdempotencyKey(ctx context.Context, branchID, idempotencyKey string) (*models.Sale, error)
	GetByIDTx(ctx context.Context, tx pgx.Tx, branchID, saleID string) (*models.Sale, error)
}

type saleRepository struct{ db *pgxpool.Pool }

func NewSaleRepository(db *pgxpool.Pool) SaleRepository { return &saleRepository{db: db} }

func (r *saleRepository) CreateSaleTx(ctx context.Context, tx pgx.Tx, sale *models.Sale) (*models.Sale, error) {
	const query = `
		INSERT INTO sales (branch_id, user_id, customer_id, total_amount, vat_amount, discount_amount, net_amount, idempotency_key)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, branch_id, user_id, customer_id, total_amount::float8, vat_amount::float8,
		          discount_amount::float8, net_amount::float8, created_at, idempotency_key
	`

	out := &models.Sale{}
	err := tx.QueryRow(ctx, query, sale.BranchID, sale.UserID, sale.CustomerID, sale.TotalAmount, sale.VATAmount, sale.DiscountAmount, sale.NetAmount, sale.IdempotencyKey).Scan(
		&out.ID,
		&out.BranchID,
		&out.UserID,
		&out.CustomerID,
		&out.TotalAmount,
		&out.VATAmount,
		&out.DiscountAmount,
		&out.NetAmount,
		&out.CreatedAt,
		&out.IdempotencyKey,
	)
	if err != nil {
		return nil, err
	}

	return out, nil
}

func (r *saleRepository) CreateSaleItemsTx(ctx context.Context, tx pgx.Tx, saleID string, items []models.SaleItem) error {
	if len(items) == 0 {
		return nil
	}
	rows := make([][]interface{}, 0, len(items))
	for _, item := range items {
		rows = append(rows, []interface{}{saleID, item.ProductID, item.Quantity, item.UnitPrice, item.VATAmount, item.LineTotal})
	}
	_, err := tx.CopyFrom(ctx, pgx.Identifier{"sale_items"}, []string{"sale_id", "product_id", "quantity", "unit_price", "vat_amount", "line_total"}, pgx.CopyFromRows(rows))
	return err
}

func (r *saleRepository) CreateSalePaymentsTx(ctx context.Context, tx pgx.Tx, saleID string, payments []models.SalePayment) ([]models.SalePayment, error) {
	const query = `INSERT INTO sale_payments (sale_id, payment_method, amount) VALUES ($1, $2, $3) RETURNING id, sale_id, payment_method, amount::float8, paid_at`
	result := make([]models.SalePayment, 0, len(payments))
	for _, payment := range payments {
		var out models.SalePayment
		if err := tx.QueryRow(ctx, query, saleID, payment.PaymentMethod, payment.Amount).Scan(&out.ID, &out.SaleID, &out.PaymentMethod, &out.Amount, &out.PaidAt); err != nil {
			return nil, err
		}
		result = append(result, out)
	}
	return result, nil
}

func (r *saleRepository) GetByIDTx(ctx context.Context, tx pgx.Tx, branchID, saleID string) (*models.Sale, error) {
	const query = `
		SELECT id, branch_id, user_id, customer_id, total_amount::float8, vat_amount::float8,
		       discount_amount::float8, net_amount::float8, created_at, idempotency_key
		FROM sales WHERE id = $1 AND branch_id = $2
	`
	sale := &models.Sale{}
	if err := tx.QueryRow(ctx, query, saleID, branchID).Scan(&sale.ID, &sale.BranchID, &sale.UserID, &sale.CustomerID, &sale.TotalAmount, &sale.VATAmount, &sale.DiscountAmount, &sale.NetAmount, &sale.CreatedAt, &sale.IdempotencyKey); err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return sale, nil
}

func (r *saleRepository) GetByIdempotencyKey(ctx context.Context, branchID, idempotencyKey string) (*models.Sale, error) {
	sale, err := r.GetByIDByCondition(ctx, branchID, "idempotency_key = $2", idempotencyKey)
	if err != nil || sale == nil {
		return sale, err
	}
	return r.loadSaleDetails(ctx, sale)
}

func (r *saleRepository) GetByIDByCondition(ctx context.Context, branchID, condition, value string) (*models.Sale, error) {
	query := `
		SELECT id, branch_id, user_id, customer_id, total_amount::float8, vat_amount::float8,
		       discount_amount::float8, net_amount::float8, created_at, idempotency_key
		FROM sales
		WHERE branch_id = $1 AND ` + condition
	sale := &models.Sale{}
	if err := r.db.QueryRow(ctx, query, branchID, value).Scan(&sale.ID, &sale.BranchID, &sale.UserID, &sale.CustomerID, &sale.TotalAmount, &sale.VATAmount, &sale.DiscountAmount, &sale.NetAmount, &sale.CreatedAt, &sale.IdempotencyKey); err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return sale, nil
}

func (r *saleRepository) loadSaleDetails(ctx context.Context, sale *models.Sale) (*models.Sale, error) {
	const itemsQuery = `SELECT sale_id, product_id, quantity::float8, unit_price::float8, vat_amount::float8, line_total::float8 FROM sale_items WHERE sale_id = $1`
	itemRows, err := r.db.Query(ctx, itemsQuery, sale.ID)
	if err != nil {
		return nil, err
	}
	defer itemRows.Close()
	sale.Items = make([]models.SaleItem, 0)
	for itemRows.Next() {
		var item models.SaleItem
		if err := itemRows.Scan(&item.SaleID, &item.ProductID, &item.Quantity, &item.UnitPrice, &item.VATAmount, &item.LineTotal); err != nil {
			return nil, err
		}
		sale.Items = append(sale.Items, item)
	}
	if err := itemRows.Err(); err != nil {
		return nil, err
	}

	const paymentsQuery = `SELECT id, sale_id, payment_method, amount::float8, paid_at FROM sale_payments WHERE sale_id = $1`
	payRows, err := r.db.Query(ctx, paymentsQuery, sale.ID)
	if err != nil {
		return nil, err
	}
	defer payRows.Close()
	sale.Payments = make([]models.SalePayment, 0)
	for payRows.Next() {
		var p models.SalePayment
		if err := payRows.Scan(&p.ID, &p.SaleID, &p.PaymentMethod, &p.Amount, &p.PaidAt); err != nil {
			return nil, err
		}
		sale.Payments = append(sale.Payments, p)
	}
	if err := payRows.Err(); err != nil {
		return nil, err
	}

	const discountsQuery = `
		SELECT discount_rule_id, discount_amount::float8
		FROM sale_discounts WHERE sale_id = $1
	`
	dRows, err := r.db.Query(ctx, discountsQuery, sale.ID)
	if err != nil {
		return nil, err
	}
	defer dRows.Close()
	sale.AppliedDiscounts = make([]models.AppliedDiscount, 0)
	for dRows.Next() {
		var ruleID *string
		var amount float64
		if err := dRows.Scan(&ruleID, &amount); err != nil {
			return nil, err
		}
		sale.AppliedDiscounts = append(sale.AppliedDiscounts, models.AppliedDiscount{DiscountRuleID: ruleID, Amount: amount})
	}
	return sale, dRows.Err()
}
