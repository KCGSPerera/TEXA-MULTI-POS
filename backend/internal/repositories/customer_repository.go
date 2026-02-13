package repositories

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/kcgsperera/texa-multi-pos/backend/internal/models"
)

type CustomerRepository interface {
	Create(ctx context.Context, customer *models.Customer) (*models.Customer, error)
	GetByID(ctx context.Context, branchID, id string) (*models.Customer, error)
	ListByBranch(ctx context.Context, branchID string, limit, offset int) ([]models.Customer, error)
	Update(ctx context.Context, customer *models.Customer) (*models.Customer, error)
	Delete(ctx context.Context, branchID, id string) (bool, error)
}

type customerRepository struct {
	db *pgxpool.Pool
}

func NewCustomerRepository(db *pgxpool.Pool) CustomerRepository {
	return &customerRepository{db: db}
}

func (r *customerRepository) Create(ctx context.Context, customer *models.Customer) (*models.Customer, error) {
	const query = `
		INSERT INTO customers (branch_id, name, mobile_number, email, address, loyalty_points)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, branch_id, name, mobile_number, email, address, loyalty_points::float8, created_at, updated_at, deleted_at
	`

	out := &models.Customer{}
	err := r.db.QueryRow(ctx, query, customer.BranchID, customer.Name, customer.MobileNumber, customer.Email, customer.Address, customer.LoyaltyPoints).Scan(
		&out.ID, &out.BranchID, &out.Name, &out.MobileNumber, &out.Email, &out.Address, &out.LoyaltyPoints, &out.CreatedAt, &out.UpdatedAt, &out.DeletedAt,
	)
	if err != nil {
		return nil, err
	}

	return out, nil
}

func (r *customerRepository) GetByID(ctx context.Context, branchID, id string) (*models.Customer, error) {
	const query = `
		SELECT id, branch_id, name, mobile_number, email, address, loyalty_points::float8, created_at, updated_at, deleted_at
		FROM customers
		WHERE id = $1 AND branch_id = $2 AND deleted_at IS NULL
	`

	out := &models.Customer{}
	err := r.db.QueryRow(ctx, query, id, branchID).Scan(
		&out.ID, &out.BranchID, &out.Name, &out.MobileNumber, &out.Email, &out.Address, &out.LoyaltyPoints, &out.CreatedAt, &out.UpdatedAt, &out.DeletedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return out, nil
}

func (r *customerRepository) ListByBranch(ctx context.Context, branchID string, limit, offset int) ([]models.Customer, error) {
	const query = `
		SELECT id, branch_id, name, mobile_number, email, address, loyalty_points::float8, created_at, updated_at, deleted_at
		FROM customers
		WHERE branch_id = $1 AND deleted_at IS NULL
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.Query(ctx, query, branchID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]models.Customer, 0)
	for rows.Next() {
		var item models.Customer
		if err := rows.Scan(&item.ID, &item.BranchID, &item.Name, &item.MobileNumber, &item.Email, &item.Address, &item.LoyaltyPoints, &item.CreatedAt, &item.UpdatedAt, &item.DeletedAt); err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	return items, rows.Err()
}

func (r *customerRepository) Update(ctx context.Context, customer *models.Customer) (*models.Customer, error) {
	const query = `
		UPDATE customers
		SET name = $1,
		    mobile_number = $2,
		    email = $3,
		    address = $4,
		    loyalty_points = $5,
		    updated_at = now()
		WHERE id = $6 AND branch_id = $7 AND deleted_at IS NULL
		RETURNING id, branch_id, name, mobile_number, email, address, loyalty_points::float8, created_at, updated_at, deleted_at
	`

	out := &models.Customer{}
	err := r.db.QueryRow(ctx, query, customer.Name, customer.MobileNumber, customer.Email, customer.Address, customer.LoyaltyPoints, customer.ID, customer.BranchID).Scan(
		&out.ID, &out.BranchID, &out.Name, &out.MobileNumber, &out.Email, &out.Address, &out.LoyaltyPoints, &out.CreatedAt, &out.UpdatedAt, &out.DeletedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return out, nil
}

func (r *customerRepository) Delete(ctx context.Context, branchID, id string) (bool, error) {
	const query = `
		UPDATE customers
		SET deleted_at = now(), updated_at = now()
		WHERE id = $1 AND branch_id = $2 AND deleted_at IS NULL
	`

	result, err := r.db.Exec(ctx, query, id, branchID)
	if err != nil {
		return false, err
	}

	return result.RowsAffected() > 0, nil
}
