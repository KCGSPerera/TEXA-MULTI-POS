package repositories

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/kcgsperera/texa-multi-pos/backend/internal/models"
)

type SupplierRepository interface {
	Create(ctx context.Context, supplier *models.Supplier) (*models.Supplier, error)
	GetByID(ctx context.Context, branchID, id string) (*models.Supplier, error)
	ListByBranch(ctx context.Context, branchID string, limit, offset int) ([]models.Supplier, error)
	Update(ctx context.Context, supplier *models.Supplier) (*models.Supplier, error)
	Delete(ctx context.Context, branchID, id string) (bool, error)
}

type supplierRepository struct {
	db *pgxpool.Pool
}

func NewSupplierRepository(db *pgxpool.Pool) SupplierRepository {
	return &supplierRepository{db: db}
}

func (r *supplierRepository) Create(ctx context.Context, supplier *models.Supplier) (*models.Supplier, error) {
	const query = `
		INSERT INTO suppliers (branch_id, name, mobile_number, email, address)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, branch_id, name, mobile_number, email, address, created_at, updated_at, deleted_at
	`

	out := &models.Supplier{}
	err := r.db.QueryRow(ctx, query, supplier.BranchID, supplier.Name, supplier.MobileNumber, supplier.Email, supplier.Address).Scan(
		&out.ID, &out.BranchID, &out.Name, &out.MobileNumber, &out.Email, &out.Address, &out.CreatedAt, &out.UpdatedAt, &out.DeletedAt,
	)
	if err != nil {
		return nil, err
	}

	return out, nil
}

func (r *supplierRepository) GetByID(ctx context.Context, branchID, id string) (*models.Supplier, error) {
	const query = `
		SELECT id, branch_id, name, mobile_number, email, address, created_at, updated_at, deleted_at
		FROM suppliers
		WHERE id = $1 AND branch_id = $2 AND deleted_at IS NULL
	`

	out := &models.Supplier{}
	err := r.db.QueryRow(ctx, query, id, branchID).Scan(
		&out.ID, &out.BranchID, &out.Name, &out.MobileNumber, &out.Email, &out.Address, &out.CreatedAt, &out.UpdatedAt, &out.DeletedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return out, nil
}

func (r *supplierRepository) ListByBranch(ctx context.Context, branchID string, limit, offset int) ([]models.Supplier, error) {
	const query = `
		SELECT id, branch_id, name, mobile_number, email, address, created_at, updated_at, deleted_at
		FROM suppliers
		WHERE branch_id = $1 AND deleted_at IS NULL
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.Query(ctx, query, branchID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]models.Supplier, 0)
	for rows.Next() {
		var item models.Supplier
		if err := rows.Scan(&item.ID, &item.BranchID, &item.Name, &item.MobileNumber, &item.Email, &item.Address, &item.CreatedAt, &item.UpdatedAt, &item.DeletedAt); err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	return items, rows.Err()
}

func (r *supplierRepository) Update(ctx context.Context, supplier *models.Supplier) (*models.Supplier, error) {
	const query = `
		UPDATE suppliers
		SET name = $1,
		    mobile_number = $2,
		    email = $3,
		    address = $4,
		    updated_at = now()
		WHERE id = $5 AND branch_id = $6 AND deleted_at IS NULL
		RETURNING id, branch_id, name, mobile_number, email, address, created_at, updated_at, deleted_at
	`

	out := &models.Supplier{}
	err := r.db.QueryRow(ctx, query, supplier.Name, supplier.MobileNumber, supplier.Email, supplier.Address, supplier.ID, supplier.BranchID).Scan(
		&out.ID, &out.BranchID, &out.Name, &out.MobileNumber, &out.Email, &out.Address, &out.CreatedAt, &out.UpdatedAt, &out.DeletedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return out, nil
}

func (r *supplierRepository) Delete(ctx context.Context, branchID, id string) (bool, error) {
	const query = `
		UPDATE suppliers
		SET deleted_at = now(), updated_at = now()
		WHERE id = $1 AND branch_id = $2 AND deleted_at IS NULL
	`

	result, err := r.db.Exec(ctx, query, id, branchID)
	if err != nil {
		return false, err
	}

	return result.RowsAffected() > 0, nil
}
