package repositories

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/kcgsperera/texa-multi-pos/backend/internal/models"
)

type LedgerRepository interface {
	GetAccountByCodeTx(ctx context.Context, tx pgx.Tx, branchID, code string) (*models.Account, error)
	CreateJournalEntryTx(ctx context.Context, tx pgx.Tx, branchID, referenceType, referenceID string) (*models.JournalEntry, error)
	CreateJournalLinesTx(ctx context.Context, tx pgx.Tx, entryID string, lines []models.JournalLineInput) error
}

type ledgerRepository struct{ db *pgxpool.Pool }

func NewLedgerRepository(db *pgxpool.Pool) LedgerRepository { return &ledgerRepository{db: db} }

func (r *ledgerRepository) GetAccountByCodeTx(ctx context.Context, tx pgx.Tx, branchID, code string) (*models.Account, error) {
	const query = `SELECT id, branch_id, code, name, type FROM accounts WHERE branch_id = $1 AND code = $2`
	acc := &models.Account{}
	if err := tx.QueryRow(ctx, query, branchID, code).Scan(&acc.ID, &acc.BranchID, &acc.Code, &acc.Name, &acc.Type); err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return acc, nil
}

func (r *ledgerRepository) CreateJournalEntryTx(ctx context.Context, tx pgx.Tx, branchID, referenceType, referenceID string) (*models.JournalEntry, error) {
	const query = `INSERT INTO journal_entries (branch_id, reference_type, reference_id) VALUES ($1,$2,$3) RETURNING id, branch_id, reference_type, reference_id, created_at`
	je := &models.JournalEntry{}
	if err := tx.QueryRow(ctx, query, branchID, referenceType, referenceID).Scan(&je.ID, &je.BranchID, &je.ReferenceType, &je.ReferenceID, &je.CreatedAt); err != nil {
		return nil, err
	}
	return je, nil
}

func (r *ledgerRepository) CreateJournalLinesTx(ctx context.Context, tx pgx.Tx, entryID string, lines []models.JournalLineInput) error {
	const query = `INSERT INTO journal_lines (journal_entry_id, account_id, debit, credit) VALUES ($1,$2,$3,$4)`
	for _, l := range lines {
		if _, err := tx.Exec(ctx, query, entryID, l.AccountID, l.Debit, l.Credit); err != nil {
			return err
		}
	}
	return nil
}
