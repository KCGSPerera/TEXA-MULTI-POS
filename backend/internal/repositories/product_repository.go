package repositories

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/kcgsperera/texa-multi-pos/backend/internal/models"
)

type ProductRepository interface {
	Create(ctx context.Context, product *models.Product) (*models.Product, error)
	GetByID(ctx context.Context, branchID, id string) (*models.Product, error)
	ListByBranch(ctx context.Context, branchID string, limit, offset int) ([]models.Product, error)
	Update(ctx context.Context, product *models.Product) (*models.Product, error)
	Delete(ctx context.Context, branchID, id, userID string) (bool, error)
	ExistsActiveInBranch(ctx context.Context, branchID, id string) (bool, error)
}

type productRepository struct {
	db *pgxpool.Pool
}

func NewProductRepository(db *pgxpool.Pool) ProductRepository {
	return &productRepository{db: db}
}

func (r *productRepository) Create(ctx context.Context, product *models.Product) (*models.Product, error) {
	const query = `
		INSERT INTO products (
			branch_id, category_id, name, sku, barcode, cost_price, selling_price,
			vat_percentage, is_vat_inclusive, is_active, created_by, updated_by
		)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)
		RETURNING id, branch_id, category_id, name, sku, barcode,
		          cost_price::float8, selling_price::float8, vat_percentage::float8,
		          is_vat_inclusive, is_active, created_at, updated_at, deleted_at, created_by, updated_by
	`

	out := &models.Product{}
	err := r.db.QueryRow(
		ctx,
		query,
		product.BranchID,
		product.CategoryID,
		product.Name,
		product.SKU,
		product.Barcode,
		product.CostPrice,
		product.SellingPrice,
		product.VATPercentage,
		product.IsVATInclusive,
		product.IsActive,
		product.CreatedBy,
		product.UpdatedBy,
	).Scan(
		&out.ID,
		&out.BranchID,
		&out.CategoryID,
		&out.Name,
		&out.SKU,
		&out.Barcode,
		&out.CostPrice,
		&out.SellingPrice,
		&out.VATPercentage,
		&out.IsVATInclusive,
		&out.IsActive,
		&out.CreatedAt,
		&out.UpdatedAt,
		&out.DeletedAt,
		&out.CreatedBy,
		&out.UpdatedBy,
	)
	if err != nil {
		return nil, err
	}

	return out, nil
}

func (r *productRepository) GetByID(ctx context.Context, branchID, id string) (*models.Product, error) {
	const query = `
		SELECT id, branch_id, category_id, name, sku, barcode,
		       cost_price::float8, selling_price::float8, vat_percentage::float8,
		       is_vat_inclusive, is_active, created_at, updated_at, deleted_at, created_by, updated_by
		FROM products
		WHERE id = $1 AND branch_id = $2 AND deleted_at IS NULL
	`

	out := &models.Product{}
	err := r.db.QueryRow(ctx, query, id, branchID).Scan(
		&out.ID,
		&out.BranchID,
		&out.CategoryID,
		&out.Name,
		&out.SKU,
		&out.Barcode,
		&out.CostPrice,
		&out.SellingPrice,
		&out.VATPercentage,
		&out.IsVATInclusive,
		&out.IsActive,
		&out.CreatedAt,
		&out.UpdatedAt,
		&out.DeletedAt,
		&out.CreatedBy,
		&out.UpdatedBy,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return out, nil
}

func (r *productRepository) ListByBranch(ctx context.Context, branchID string, limit, offset int) ([]models.Product, error) {
	const query = `
		SELECT id, branch_id, category_id, name, sku, barcode,
		       cost_price::float8, selling_price::float8, vat_percentage::float8,
		       is_vat_inclusive, is_active, created_at, updated_at, deleted_at, created_by, updated_by
		FROM products
		WHERE branch_id = $1 AND deleted_at IS NULL
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.Query(ctx, query, branchID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]models.Product, 0)
	for rows.Next() {
		var item models.Product
		if err := rows.Scan(
			&item.ID,
			&item.BranchID,
			&item.CategoryID,
			&item.Name,
			&item.SKU,
			&item.Barcode,
			&item.CostPrice,
			&item.SellingPrice,
			&item.VATPercentage,
			&item.IsVATInclusive,
			&item.IsActive,
			&item.CreatedAt,
			&item.UpdatedAt,
			&item.DeletedAt,
			&item.CreatedBy,
			&item.UpdatedBy,
		); err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	return items, rows.Err()
}

func (r *productRepository) Update(ctx context.Context, product *models.Product) (*models.Product, error) {
	const query = `
		UPDATE products
		SET category_id = $1,
		    name = $2,
		    sku = $3,
		    barcode = $4,
		    cost_price = $5,
		    selling_price = $6,
		    vat_percentage = $7,
		    is_vat_inclusive = $8,
		    is_active = $9,
		    updated_at = now(),
		    updated_by = $10
		WHERE id = $11 AND branch_id = $12 AND deleted_at IS NULL
		RETURNING id, branch_id, category_id, name, sku, barcode,
		          cost_price::float8, selling_price::float8, vat_percentage::float8,
		          is_vat_inclusive, is_active, created_at, updated_at, deleted_at, created_by, updated_by
	`

	out := &models.Product{}
	err := r.db.QueryRow(
		ctx,
		query,
		product.CategoryID,
		product.Name,
		product.SKU,
		product.Barcode,
		product.CostPrice,
		product.SellingPrice,
		product.VATPercentage,
		product.IsVATInclusive,
		product.IsActive,
		product.UpdatedBy,
		product.ID,
		product.BranchID,
	).Scan(
		&out.ID,
		&out.BranchID,
		&out.CategoryID,
		&out.Name,
		&out.SKU,
		&out.Barcode,
		&out.CostPrice,
		&out.SellingPrice,
		&out.VATPercentage,
		&out.IsVATInclusive,
		&out.IsActive,
		&out.CreatedAt,
		&out.UpdatedAt,
		&out.DeletedAt,
		&out.CreatedBy,
		&out.UpdatedBy,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return out, nil
}

func (r *productRepository) Delete(ctx context.Context, branchID, id, userID string) (bool, error) {
	const query = `
		UPDATE products
		SET deleted_at = now(), updated_at = now(), updated_by = $3
		WHERE id = $1 AND branch_id = $2 AND deleted_at IS NULL
	`

	result, err := r.db.Exec(ctx, query, id, branchID, userID)
	if err != nil {
		return false, err
	}

	return result.RowsAffected() > 0, nil
}

func (r *productRepository) ExistsActiveInBranch(ctx context.Context, branchID, id string) (bool, error) {
	const query = `
		SELECT EXISTS(
			SELECT 1 FROM products
			WHERE id = $1 AND branch_id = $2 AND is_active = true AND deleted_at IS NULL
		)
	`

	var exists bool
	if err := r.db.QueryRow(ctx, query, id, branchID).Scan(&exists); err != nil {
		return false, err
	}

	return exists, nil
}
