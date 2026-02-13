package services

import (
	"context"
	"errors"
	"testing"

	"github.com/jackc/pgx/v5"

	"github.com/kcgsperera/texa-multi-pos/backend/internal/models"
)

type ledgerRepoStub struct{}

func (l *ledgerRepoStub) GetAccountByCodeTx(ctx context.Context, tx pgx.Tx, branchID, code string) (*models.Account, error) {
	return &models.Account{ID: code}, nil
}
func (l *ledgerRepoStub) CreateJournalEntryTx(ctx context.Context, tx pgx.Tx, branchID, referenceType, referenceID string) (*models.JournalEntry, error) {
	return &models.JournalEntry{ID: "je1"}, nil
}
func (l *ledgerRepoStub) CreateJournalLinesTx(ctx context.Context, tx pgx.Tx, entryID string, lines []models.JournalLineInput) error {
	return nil
}

func TestLedgerBalanceEnforcement(t *testing.T) {
	svc := NewLedgerService(&ledgerRepoStub{}, noopAudit{})
	err := svc.Post(context.Background(), nil, "b1", "SALE", "s1", []models.JournalLineInput{
		{AccountID: "a1", Debit: 100, Credit: 0},
		{AccountID: "a2", Debit: 0, Credit: 90},
	}, nil)
	if !errors.Is(err, ErrLedgerImbalance) {
		t.Fatalf("expected ErrLedgerImbalance, got %v", err)
	}
}
