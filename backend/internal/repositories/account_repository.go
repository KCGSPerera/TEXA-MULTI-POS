package repositories

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type AccountRepository interface {
	EnsureDefaultAccountsTx(ctx context.Context, tx pgx.Tx, branchID string) error
	MarkBranchBootstrapTx(ctx context.Context, tx pgx.Tx, branchID string) error
	ListBranchIDs(ctx context.Context) ([]string, error)
}

type accountRepository struct {
	db *pgxpool.Pool
}

func NewAccountRepository(db *pgxpool.Pool) AccountRepository {
	return &accountRepository{db: db}
}

func (r *accountRepository) EnsureDefaultAccountsTx(ctx context.Context, tx pgx.Tx, branchID string) error {
	const query = `
		INSERT INTO accounts (branch_id, code, name, type)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (branch_id, code) DO UPDATE
		SET name = EXCLUDED.name,
		    type = EXCLUDED.type
	`

	defaults := []struct {
		Code string
		Name string
		Type string
	}{
		{Code: "CASH", Name: "Cash", Type: "ASSET"},
		{Code: "INVENTORY", Name: "Inventory", Type: "ASSET"},
		{Code: "PAYABLE", Name: "Accounts Payable", Type: "LIABILITY"},
		{Code: "VAT_PAYABLE", Name: "VAT Payable", Type: "LIABILITY"},
		{Code: "REVENUE", Name: "Sales Revenue", Type: "INCOME"},
	}

	for _, account := range defaults {
		if _, err := tx.Exec(ctx, query, branchID, account.Code, account.Name, account.Type); err != nil {
			return err
		}
	}
	return nil
}

func (r *accountRepository) MarkBranchBootstrapTx(ctx context.Context, tx pgx.Tx, branchID string) error {
	const query = `
		INSERT INTO branch_bootstrap (branch_id, accounts_seeded)
		VALUES ($1, true)
		ON CONFLICT (branch_id)
		DO UPDATE SET accounts_seeded = true, updated_at = now()
	`
	_, err := tx.Exec(ctx, query, branchID)
	return err
}

func (r *accountRepository) ListBranchIDs(ctx context.Context) ([]string, error) {
	const query = `SELECT id FROM branches ORDER BY created_at ASC`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]string, 0)
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		items = append(items, id)
	}
	return items, rows.Err()
}
