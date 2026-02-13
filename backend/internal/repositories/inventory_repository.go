package repositories

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/kcgsperera/texa-multi-pos/backend/internal/models"
)

type InventoryRepository interface {
	ListByBranch(ctx context.Context, branchID string) ([]models.Inventory, error)
	GetByProduct(ctx context.Context, branchID, productID string) (*models.Inventory, error)
	EnsureInventoryRowTx(ctx context.Context, tx pgx.Tx, branchID, productID string) error
	GetForUpdateTx(ctx context.Context, tx pgx.Tx, branchID, productID string) (*models.Inventory, error)
	UpdateQuantityTx(ctx context.Context, tx pgx.Tx, branchID, productID string, quantity float64) error
	CreateStockMovementTx(ctx context.Context, tx pgx.Tx, movement *models.StockMovement) error
}

type inventoryRepository struct {
	db *pgxpool.Pool
}

func NewInventoryRepository(db *pgxpool.Pool) InventoryRepository {
	return &inventoryRepository{db: db}
}

func (r *inventoryRepository) ListByBranch(ctx context.Context, branchID string) ([]models.Inventory, error) {
	const query = `
		SELECT product_id, branch_id, quantity::float8, reserved_quantity::float8,
		       (quantity - reserved_quantity)::float8 AS available_quantity,
		       created_at, updated_at
		FROM inventory
		WHERE branch_id = $1
		ORDER BY updated_at DESC
	`

	rows, err := r.db.Query(ctx, query, branchID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]models.Inventory, 0)
	for rows.Next() {
		var item models.Inventory
		if err := rows.Scan(&item.ProductID, &item.BranchID, &item.Quantity, &item.ReservedQuantity, &item.AvailableQuantity, &item.CreatedAt, &item.UpdatedAt); err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	return items, rows.Err()
}

func (r *inventoryRepository) GetByProduct(ctx context.Context, branchID, productID string) (*models.Inventory, error) {
	const query = `
		SELECT product_id, branch_id, quantity::float8, reserved_quantity::float8,
		       (quantity - reserved_quantity)::float8 AS available_quantity,
		       created_at, updated_at
		FROM inventory
		WHERE branch_id = $1 AND product_id = $2
	`

	item := &models.Inventory{}
	if err := r.db.QueryRow(ctx, query, branchID, productID).Scan(
		&item.ProductID,
		&item.BranchID,
		&item.Quantity,
		&item.ReservedQuantity,
		&item.AvailableQuantity,
		&item.CreatedAt,
		&item.UpdatedAt,
	); err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return item, nil
}

func (r *inventoryRepository) EnsureInventoryRowTx(ctx context.Context, tx pgx.Tx, branchID, productID string) error {
	const query = `
		INSERT INTO inventory (product_id, branch_id)
		VALUES ($1, $2)
		ON CONFLICT (branch_id, product_id) DO NOTHING
	`
	_, err := tx.Exec(ctx, query, productID, branchID)
	return err
}

func (r *inventoryRepository) GetForUpdateTx(ctx context.Context, tx pgx.Tx, branchID, productID string) (*models.Inventory, error) {
	const query = `
		SELECT product_id, branch_id, quantity::float8, reserved_quantity::float8,
		       (quantity - reserved_quantity)::float8 AS available_quantity,
		       created_at, updated_at
		FROM inventory
		WHERE branch_id = $1 AND product_id = $2
		FOR UPDATE
	`

	item := &models.Inventory{}
	if err := tx.QueryRow(ctx, query, branchID, productID).Scan(
		&item.ProductID,
		&item.BranchID,
		&item.Quantity,
		&item.ReservedQuantity,
		&item.AvailableQuantity,
		&item.CreatedAt,
		&item.UpdatedAt,
	); err != nil {
		return nil, err
	}

	return item, nil
}

func (r *inventoryRepository) UpdateQuantityTx(ctx context.Context, tx pgx.Tx, branchID, productID string, quantity float64) error {
	const query = `
		UPDATE inventory
		SET quantity = $1,
		    updated_at = now()
		WHERE branch_id = $2 AND product_id = $3
	`
	_, err := tx.Exec(ctx, query, quantity, branchID, productID)
	return err
}

func (r *inventoryRepository) CreateStockMovementTx(ctx context.Context, tx pgx.Tx, movement *models.StockMovement) error {
	const query = `
		INSERT INTO stock_movements (product_id, branch_id, movement_type, quantity, reference_type, reference_id, created_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	_, err := tx.Exec(ctx, query, movement.ProductID, movement.BranchID, movement.MovementType, movement.Quantity, movement.ReferenceType, movement.ReferenceID, movement.CreatedBy)
	return err
}
