package repositories

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/kcgsperera/texa-multi-pos/backend/internal/models"
)

type PurchaseRepository interface {
	ListPurchaseOrdersByBranch(ctx context.Context, branchID string, limit, offset int) ([]models.PurchaseOrder, error)
	CreatePurchaseOrderTx(ctx context.Context, tx pgx.Tx, po *models.PurchaseOrder) (*models.PurchaseOrder, error)
	CreatePurchaseOrderItemsTx(ctx context.Context, tx pgx.Tx, poID string, items []models.PurchaseOrderItem) ([]models.PurchaseOrderItem, error)
	GetPurchaseOrderByIDTx(ctx context.Context, tx pgx.Tx, branchID, poID string) (*models.PurchaseOrder, error)
	CreateGRNTx(ctx context.Context, tx pgx.Tx, grn *models.GRN) (*models.GRN, error)
	CreateGRNItemsTx(ctx context.Context, tx pgx.Tx, grnID string, items []models.GRNItem) ([]models.GRNItem, error)
	GetGRNByIDForUpdateTx(ctx context.Context, tx pgx.Tx, branchID, grnID string) (*models.GRN, error)
	GetGRNItemsTx(ctx context.Context, tx pgx.Tx, grnID string) ([]models.GRNItem, error)
	UpdateGRNApprovedTx(ctx context.Context, tx pgx.Tx, grnID, approvedBy string, approvedAt time.Time) error
	IncrementPOReceivedQtyTx(ctx context.Context, tx pgx.Tx, poID, productID string, qty float64) error
	RecalculatePOStatusTx(ctx context.Context, tx pgx.Tx, poID string) error
}

type purchaseRepository struct {
	db *pgxpool.Pool
}

func NewPurchaseRepository(db *pgxpool.Pool) PurchaseRepository {
	return &purchaseRepository{db: db}
}

func (r *purchaseRepository) ListPurchaseOrdersByBranch(ctx context.Context, branchID string, limit, offset int) ([]models.PurchaseOrder, error) {
	const query = `
		SELECT id, branch_id, supplier_id, status, total_amount::float8, created_by, created_at, updated_at
		FROM purchase_orders
		WHERE branch_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.Query(ctx, query, branchID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]models.PurchaseOrder, 0)
	for rows.Next() {
		var po models.PurchaseOrder
		if err := rows.Scan(&po.ID, &po.BranchID, &po.SupplierID, &po.Status, &po.TotalAmount, &po.CreatedBy, &po.CreatedAt, &po.UpdatedAt); err != nil {
			return nil, err
		}
		items = append(items, po)
	}

	return items, rows.Err()
}

func (r *purchaseRepository) CreatePurchaseOrderTx(ctx context.Context, tx pgx.Tx, po *models.PurchaseOrder) (*models.PurchaseOrder, error) {
	const query = `
		INSERT INTO purchase_orders (branch_id, supplier_id, status, total_amount, created_by)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, branch_id, supplier_id, status, total_amount::float8, created_by, created_at, updated_at
	`

	out := &models.PurchaseOrder{}
	if err := tx.QueryRow(ctx, query, po.BranchID, po.SupplierID, po.Status, po.TotalAmount, po.CreatedBy).Scan(
		&out.ID, &out.BranchID, &out.SupplierID, &out.Status, &out.TotalAmount, &out.CreatedBy, &out.CreatedAt, &out.UpdatedAt,
	); err != nil {
		return nil, err
	}

	return out, nil
}

func (r *purchaseRepository) CreatePurchaseOrderItemsTx(ctx context.Context, tx pgx.Tx, poID string, items []models.PurchaseOrderItem) ([]models.PurchaseOrderItem, error) {
	if len(items) == 0 {
		return []models.PurchaseOrderItem{}, nil
	}

	rows := make([][]interface{}, 0, len(items))
	for _, item := range items {
		rows = append(rows, []interface{}{poID, item.ProductID, item.Quantity, item.UnitCost, item.LineTotal})
	}

	_, err := tx.CopyFrom(
		ctx,
		pgx.Identifier{"purchase_order_items"},
		[]string{"purchase_order_id", "product_id", "quantity", "unit_cost", "line_total"},
		pgx.CopyFromRows(rows),
	)
	if err != nil {
		return nil, err
	}

	const fetchQuery = `
		SELECT id, purchase_order_id, product_id, quantity::float8, received_quantity::float8, unit_cost::float8, line_total::float8
		FROM purchase_order_items
		WHERE purchase_order_id = $1
	`
	frows, err := tx.Query(ctx, fetchQuery, poID)
	if err != nil {
		return nil, err
	}
	defer frows.Close()

	out := make([]models.PurchaseOrderItem, 0)
	for frows.Next() {
		var row models.PurchaseOrderItem
		if err := frows.Scan(&row.ID, &row.PurchaseOrderID, &row.ProductID, &row.Quantity, &row.ReceivedQuantity, &row.UnitCost, &row.LineTotal); err != nil {
			return nil, err
		}
		out = append(out, row)
	}
	return out, frows.Err()
}

func (r *purchaseRepository) GetPurchaseOrderByIDTx(ctx context.Context, tx pgx.Tx, branchID, poID string) (*models.PurchaseOrder, error) {
	const query = `
		SELECT id, branch_id, supplier_id, status, total_amount::float8, created_by, created_at, updated_at
		FROM purchase_orders
		WHERE id = $1 AND branch_id = $2
		FOR UPDATE
	`

	po := &models.PurchaseOrder{}
	if err := tx.QueryRow(ctx, query, poID, branchID).Scan(&po.ID, &po.BranchID, &po.SupplierID, &po.Status, &po.TotalAmount, &po.CreatedBy, &po.CreatedAt, &po.UpdatedAt); err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return po, nil
}

func (r *purchaseRepository) CreateGRNTx(ctx context.Context, tx pgx.Tx, grn *models.GRN) (*models.GRN, error) {
	const query = `
		INSERT INTO grns (purchase_order_id, branch_id, supplier_id, status, total_amount, created_by)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, purchase_order_id, branch_id, supplier_id, status, total_amount::float8, created_by, approved_by, created_at, approved_at
	`

	out := &models.GRN{}
	if err := tx.QueryRow(ctx, query, grn.PurchaseOrderID, grn.BranchID, grn.SupplierID, grn.Status, grn.TotalAmount, grn.CreatedBy).Scan(
		&out.ID, &out.PurchaseOrderID, &out.BranchID, &out.SupplierID, &out.Status, &out.TotalAmount, &out.CreatedBy, &out.ApprovedBy, &out.CreatedAt, &out.ApprovedAt,
	); err != nil {
		return nil, err
	}
	return out, nil
}

func (r *purchaseRepository) CreateGRNItemsTx(ctx context.Context, tx pgx.Tx, grnID string, items []models.GRNItem) ([]models.GRNItem, error) {
	if len(items) == 0 {
		return []models.GRNItem{}, nil
	}

	rows := make([][]interface{}, 0, len(items))
	for _, item := range items {
		rows = append(rows, []interface{}{grnID, item.ProductID, item.Quantity, item.UnitCost, item.LineTotal})
	}

	_, err := tx.CopyFrom(
		ctx,
		pgx.Identifier{"grn_items"},
		[]string{"grn_id", "product_id", "quantity", "unit_cost", "line_total"},
		pgx.CopyFromRows(rows),
	)
	if err != nil {
		return nil, err
	}

	const fetchQuery = `
		SELECT id, grn_id, product_id, quantity::float8, unit_cost::float8, line_total::float8
		FROM grn_items
		WHERE grn_id = $1
	`
	frows, err := tx.Query(ctx, fetchQuery, grnID)
	if err != nil {
		return nil, err
	}
	defer frows.Close()

	out := make([]models.GRNItem, 0)
	for frows.Next() {
		var row models.GRNItem
		if err := frows.Scan(&row.ID, &row.GRNID, &row.ProductID, &row.Quantity, &row.UnitCost, &row.LineTotal); err != nil {
			return nil, err
		}
		out = append(out, row)
	}
	return out, frows.Err()
}

func (r *purchaseRepository) GetGRNByIDForUpdateTx(ctx context.Context, tx pgx.Tx, branchID, grnID string) (*models.GRN, error) {
	const query = `
		SELECT id, purchase_order_id, branch_id, supplier_id, status, total_amount::float8, created_by, approved_by, created_at, approved_at
		FROM grns
		WHERE id = $1 AND branch_id = $2
		FOR UPDATE
	`

	grn := &models.GRN{}
	if err := tx.QueryRow(ctx, query, grnID, branchID).Scan(
		&grn.ID, &grn.PurchaseOrderID, &grn.BranchID, &grn.SupplierID, &grn.Status, &grn.TotalAmount, &grn.CreatedBy, &grn.ApprovedBy, &grn.CreatedAt, &grn.ApprovedAt,
	); err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return grn, nil
}

func (r *purchaseRepository) GetGRNItemsTx(ctx context.Context, tx pgx.Tx, grnID string) ([]models.GRNItem, error) {
	const query = `
		SELECT id, grn_id, product_id, quantity::float8, unit_cost::float8, line_total::float8
		FROM grn_items
		WHERE grn_id = $1
	`

	rows, err := tx.Query(ctx, query, grnID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]models.GRNItem, 0)
	for rows.Next() {
		var item models.GRNItem
		if err := rows.Scan(&item.ID, &item.GRNID, &item.ProductID, &item.Quantity, &item.UnitCost, &item.LineTotal); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *purchaseRepository) UpdateGRNApprovedTx(ctx context.Context, tx pgx.Tx, grnID, approvedBy string, approvedAt time.Time) error {
	const query = `
		UPDATE grns
		SET status = 'COMPLETED', approved_by = $2, approved_at = $3
		WHERE id = $1
	`
	_, err := tx.Exec(ctx, query, grnID, approvedBy, approvedAt)
	return err
}

func (r *purchaseRepository) IncrementPOReceivedQtyTx(ctx context.Context, tx pgx.Tx, poID, productID string, qty float64) error {
	const query = `
		UPDATE purchase_order_items
		SET received_quantity = received_quantity + $3
		WHERE purchase_order_id = $1 AND product_id = $2
	`
	_, err := tx.Exec(ctx, query, poID, productID, qty)
	return err
}

func (r *purchaseRepository) RecalculatePOStatusTx(ctx context.Context, tx pgx.Tx, poID string) error {
	const query = `
		WITH stat AS (
			SELECT
				COUNT(*) AS total_items,
				COUNT(*) FILTER (WHERE received_quantity >= quantity) AS completed_items,
				COUNT(*) FILTER (WHERE received_quantity > 0) AS partial_items
			FROM purchase_order_items
			WHERE purchase_order_id = $1
		)
		UPDATE purchase_orders po
		SET status = CASE
			WHEN s.completed_items = s.total_items THEN 'COMPLETED'
			WHEN s.partial_items > 0 THEN 'PARTIAL'
			ELSE 'OPEN'
		END,
		updated_at = now()
		FROM stat s
		WHERE po.id = $1
	`
	_, err := tx.Exec(ctx, query, poID)
	return err
}
