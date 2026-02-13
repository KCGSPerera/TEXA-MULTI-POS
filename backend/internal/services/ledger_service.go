package services

import (
	"context"
	"math"

	"github.com/jackc/pgx/v5"

	"github.com/kcgsperera/texa-multi-pos/backend/internal/logger"
	"github.com/kcgsperera/texa-multi-pos/backend/internal/models"
	"github.com/kcgsperera/texa-multi-pos/backend/internal/repositories"
)

type LedgerService interface {
	Post(ctx context.Context, tx pgx.Tx, branchID, referenceType, referenceID string, lines []models.JournalLineInput, performedBy *string) error
	PostByCodes(ctx context.Context, tx pgx.Tx, branchID, referenceType, referenceID string, lines []LedgerLineByCode, performedBy *string) error
}

type ledgerService struct {
	repo         repositories.LedgerRepository
	auditService AuditService
}

type LedgerLineByCode struct {
	AccountCode string
	Debit       float64
	Credit      float64
}

func NewLedgerService(repo repositories.LedgerRepository, auditService AuditService) LedgerService {
	return &ledgerService{repo: repo, auditService: auditService}
}

func (s *ledgerService) Post(ctx context.Context, tx pgx.Tx, branchID, referenceType, referenceID string, lines []models.JournalLineInput, performedBy *string) error {
	if len(lines) == 0 {
		return ErrInvalidInput
	}

	debit := 0.0
	credit := 0.0
	for i := range lines {
		debit += lines[i].Debit
		credit += lines[i].Credit
	}
	if math.Abs(debit-credit) > 0.0001 {
		return ErrLedgerImbalance
	}

	entry, err := s.repo.CreateJournalEntryTx(ctx, tx, branchID, referenceType, referenceID)
	if err != nil {
		return err
	}
	if err := s.repo.CreateJournalLinesTx(ctx, tx, entry.ID, lines); err != nil {
		return err
	}
	if err := s.auditService.Log(ctx, tx, "journal_entries", entry.ID, "CREATE", performedBy, nil, entry); err != nil {
		return err
	}

	logger.L().Info().Str("branch_id", branchID).Str("entity_id", entry.ID).Str("action", "ledger_post").Msg("ledger_posted")
	return nil
}

func (s *ledgerService) PostByCodes(ctx context.Context, tx pgx.Tx, branchID, referenceType, referenceID string, lines []LedgerLineByCode, performedBy *string) error {
	resolved := make([]models.JournalLineInput, 0, len(lines))
	for _, l := range lines {
		acc, err := s.repo.GetAccountByCodeTx(ctx, tx, branchID, l.AccountCode)
		if err != nil {
			return err
		}
		if acc == nil {
			return ErrInvalidInput
		}
		resolved = append(resolved, models.JournalLineInput{
			AccountID: acc.ID,
			Debit:     l.Debit,
			Credit:    l.Credit,
		})
	}
	return s.Post(ctx, tx, branchID, referenceType, referenceID, resolved, performedBy)
}
