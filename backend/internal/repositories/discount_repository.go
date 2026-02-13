package repositories

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/kcgsperera/texa-multi-pos/backend/internal/cache"
	"github.com/kcgsperera/texa-multi-pos/backend/internal/models"
)

type DiscountRepository interface {
	GetActiveRulesByIDs(ctx context.Context, branchID string, ids []string) ([]models.DiscountRule, error)
	CreateSaleDiscountsTx(ctx context.Context, tx pgx.Tx, saleID string, discounts []models.AppliedDiscount) error
}

type discountRepository struct {
	db    *pgxpool.Pool
	cache cache.Client
}

func NewDiscountRepository(db *pgxpool.Pool, cacheClient cache.Client) DiscountRepository {
	return &discountRepository{db: db, cache: cacheClient}
}

func (r *discountRepository) GetActiveRulesByIDs(ctx context.Context, branchID string, ids []string) ([]models.DiscountRule, error) {
	if len(ids) == 0 {
		return []models.DiscountRule{}, nil
	}

	foundByID := make(map[string]models.DiscountRule)
	missing := make([]string, 0, len(ids))

	for _, id := range ids {
		if r.cache == nil || !r.cache.Enabled() {
			missing = append(missing, id)
			continue
		}

		key := fmt.Sprintf("discount_rule:%s:%s", branchID, id)
		cached, ok, err := r.cache.Get(ctx, key)
		if err != nil || !ok {
			missing = append(missing, id)
			continue
		}

		var row models.DiscountRule
		if err := json.Unmarshal([]byte(cached), &row); err != nil || !row.IsActive || row.BranchID != branchID {
			missing = append(missing, id)
			continue
		}
		foundByID[row.ID] = row
	}

	if len(missing) > 0 {
		const query = `
			SELECT id, branch_id, name, type, value::float8, min_order_amount::float8, max_discount_amount::float8, is_active, created_at
			FROM discount_rules
			WHERE branch_id = $1 AND id = ANY($2) AND is_active = true
		`
		rows, err := r.db.Query(ctx, query, branchID, missing)
		if err != nil {
			return nil, err
		}
		defer rows.Close()

		for rows.Next() {
			var row models.DiscountRule
			if err := rows.Scan(&row.ID, &row.BranchID, &row.Name, &row.Type, &row.Value, &row.MinOrderAmount, &row.MaxDiscountAmount, &row.IsActive, &row.CreatedAt); err != nil {
				return nil, err
			}
			foundByID[row.ID] = row

			if r.cache != nil && r.cache.Enabled() {
				payload, err := json.Marshal(row)
				if err == nil {
					_ = r.cache.Set(ctx, fmt.Sprintf("discount_rule:%s:%s", branchID, row.ID), string(payload), 2*time.Minute)
				}
			}
		}
		if err := rows.Err(); err != nil {
			return nil, err
		}
	}

	out := make([]models.DiscountRule, 0, len(ids))
	for _, id := range ids {
		if row, ok := foundByID[id]; ok && row.IsActive {
			out = append(out, row)
		}
	}
	return out, nil
}

func (r *discountRepository) CreateSaleDiscountsTx(ctx context.Context, tx pgx.Tx, saleID string, discounts []models.AppliedDiscount) error {
	if len(discounts) == 0 {
		return nil
	}
	const query = `INSERT INTO sale_discounts (sale_id, discount_rule_id, discount_amount) VALUES ($1, $2, $3)`
	for _, d := range discounts {
		if _, err := tx.Exec(ctx, query, saleID, d.DiscountRuleID, d.Amount); err != nil {
			return err
		}
	}
	return nil
}
