package repositories

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/kcgsperera/texa-multi-pos/backend/internal/models"
)

type CategoryRepository interface {
	Create(ctx context.Context, category *models.Category) (*models.Category, error)
	GetByID(ctx context.Context, branchID, id string) (*models.Category, error)
	ListByBranch(ctx context.Context, branchID string, limit, offset int) ([]models.Category, error)
	Update(ctx context.Context, category *models.Category) (*models.Category, error)
	Delete(ctx context.Context, branchID, id, userID string) (bool, error)
}

type categoryRepository struct {
	db *pgxpool.Pool
}

func NewCategoryRepository(db *pgxpool.Pool) CategoryRepository {
	return &categoryRepository{db: db}
}

func (r *categoryRepository) Create(ctx context.Context, category *models.Category) (*models.Category, error) {
	const query = `
		INSERT INTO categories (branch_id, name, description, created_by, updated_by)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, branch_id, name, description, created_at, updated_at, deleted_at, created_by, updated_by
	`

	out := &models.Category{}
	err := r.db.QueryRow(ctx, query, category.BranchID, category.Name, category.Description, category.CreatedBy, category.UpdatedBy).Scan(
		&out.ID, &out.BranchID, &out.Name, &out.Description, &out.CreatedAt, &out.UpdatedAt, &out.DeletedAt, &out.CreatedBy, &out.UpdatedBy,
	)
	if err != nil {
		return nil, err
	}

	return out, nil
}

func (r *categoryRepository) GetByID(ctx context.Context, branchID, id string) (*models.Category, error) {
	const query = `
		SELECT id, branch_id, name, description, created_at, updated_at, deleted_at, created_by, updated_by
		FROM categories
		WHERE id = $1 AND branch_id = $2 AND deleted_at IS NULL
	`

	out := &models.Category{}
	err := r.db.QueryRow(ctx, query, id, branchID).Scan(
		&out.ID, &out.BranchID, &out.Name, &out.Description, &out.CreatedAt, &out.UpdatedAt, &out.DeletedAt, &out.CreatedBy, &out.UpdatedBy,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return out, nil
}

func (r *categoryRepository) ListByBranch(ctx context.Context, branchID string, limit, offset int) ([]models.Category, error) {
	const query = `
		SELECT id, branch_id, name, description, created_at, updated_at, deleted_at, created_by, updated_by
		FROM categories
		WHERE branch_id = $1 AND deleted_at IS NULL
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.Query(ctx, query, branchID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]models.Category, 0)
	for rows.Next() {
		var item models.Category
		if err := rows.Scan(&item.ID, &item.BranchID, &item.Name, &item.Description, &item.CreatedAt, &item.UpdatedAt, &item.DeletedAt, &item.CreatedBy, &item.UpdatedBy); err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	return items, rows.Err()
}

func (r *categoryRepository) Update(ctx context.Context, category *models.Category) (*models.Category, error) {
	const query = `
		UPDATE categories
		SET name = $1,
		    description = $2,
		    updated_at = now(),
		    updated_by = $3
		WHERE id = $4 AND branch_id = $5 AND deleted_at IS NULL
		RETURNING id, branch_id, name, description, created_at, updated_at, deleted_at, created_by, updated_by
	`

	out := &models.Category{}
	err := r.db.QueryRow(ctx, query, category.Name, category.Description, category.UpdatedBy, category.ID, category.BranchID).Scan(
		&out.ID, &out.BranchID, &out.Name, &out.Description, &out.CreatedAt, &out.UpdatedAt, &out.DeletedAt, &out.CreatedBy, &out.UpdatedBy,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return out, nil
}

func (r *categoryRepository) Delete(ctx context.Context, branchID, id, userID string) (bool, error) {
	const query = `
		UPDATE categories
		SET deleted_at = now(), updated_at = now(), updated_by = $3
		WHERE id = $1 AND branch_id = $2 AND deleted_at IS NULL
	`

	result, err := r.db.Exec(ctx, query, id, branchID, userID)
	if err != nil {
		return false, err
	}

	return result.RowsAffected() > 0, nil
}
