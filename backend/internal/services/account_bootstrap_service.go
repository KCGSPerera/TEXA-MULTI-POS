package services

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/kcgsperera/texa-multi-pos/backend/internal/database"
	"github.com/kcgsperera/texa-multi-pos/backend/internal/repositories"
)

type AccountBootstrapService interface {
	EnsureBranchAccounts(ctx context.Context, branchID string) error
	EnsureAllBranchesSeeded(ctx context.Context) error
}

type accountBootstrapService struct {
	db      *pgxpool.Pool
	account repositories.AccountRepository
}

func NewAccountBootstrapService(db *pgxpool.Pool, account repositories.AccountRepository) AccountBootstrapService {
	return &accountBootstrapService{db: db, account: account}
}

func (s *accountBootstrapService) EnsureBranchAccounts(ctx context.Context, branchID string) error {
	return database.WithTx(ctx, s.db, func(tx pgx.Tx) error {
		if err := s.account.EnsureDefaultAccountsTx(ctx, tx, branchID); err != nil {
			return err
		}
		return s.account.MarkBranchBootstrapTx(ctx, tx, branchID)
	})
}

func (s *accountBootstrapService) EnsureAllBranchesSeeded(ctx context.Context) error {
	branchIDs, err := s.account.ListBranchIDs(ctx)
	if err != nil {
		return err
	}

	for _, branchID := range branchIDs {
		if err := s.EnsureBranchAccounts(ctx, branchID); err != nil {
			return err
		}
	}
	return nil
}
