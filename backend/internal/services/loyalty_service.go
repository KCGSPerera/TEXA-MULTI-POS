package services

import (
	"context"

	"github.com/jackc/pgx/v5"

	"github.com/kcgsperera/texa-multi-pos/backend/internal/models"
	"github.com/kcgsperera/texa-multi-pos/backend/internal/repositories"
)

const LoyaltyEarnPerCurrency = 100.0

type LoyaltyService interface {
	EarnForSaleTx(ctx context.Context, tx pgx.Tx, customerID, branchID, saleID string, netAmount float64) error
	RedeemTx(ctx context.Context, tx pgx.Tx, customerID, branchID, referenceID string, points float64) error
	AdjustTx(ctx context.Context, tx pgx.Tx, customerID, branchID, referenceID string, points float64) error
}

type loyaltyService struct {
	repo repositories.LoyaltyRepository
}

func NewLoyaltyService(repo repositories.LoyaltyRepository) LoyaltyService {
	return &loyaltyService{repo: repo}
}

func (s *loyaltyService) EarnForSaleTx(ctx context.Context, tx pgx.Tx, customerID, branchID, saleID string, netAmount float64) error {
	if customerID == "" || netAmount <= 0 {
		return nil
	}
	points := netAmount / LoyaltyEarnPerCurrency
	if points <= 0 {
		return nil
	}
	refType := "SALE"
	refID := saleID
	if err := s.repo.CreateTx(ctx, tx, &models.LoyaltyTransaction{CustomerID: customerID, BranchID: branchID, Type: "EARN", Points: points, ReferenceType: &refType, ReferenceID: &refID}); err != nil {
		return err
	}
	balance, err := s.repo.GetBalanceTx(ctx, tx, customerID, branchID)
	if err != nil {
		return err
	}
	return s.repo.UpdateCustomerPointsTx(ctx, tx, customerID, branchID, balance)
}

func (s *loyaltyService) RedeemTx(ctx context.Context, tx pgx.Tx, customerID, branchID, referenceID string, points float64) error {
	if customerID == "" || points <= 0 {
		return nil
	}
	balance, err := s.repo.GetBalanceTx(ctx, tx, customerID, branchID)
	if err != nil {
		return err
	}
	if balance < points {
		return ErrInvalidInput
	}
	refType := "SALE"
	if err := s.repo.CreateTx(ctx, tx, &models.LoyaltyTransaction{CustomerID: customerID, BranchID: branchID, Type: "REDEEM", Points: points, ReferenceType: &refType, ReferenceID: &referenceID}); err != nil {
		return err
	}
	balance, err = s.repo.GetBalanceTx(ctx, tx, customerID, branchID)
	if err != nil {
		return err
	}
	return s.repo.UpdateCustomerPointsTx(ctx, tx, customerID, branchID, balance)
}

func (s *loyaltyService) AdjustTx(ctx context.Context, tx pgx.Tx, customerID, branchID, referenceID string, points float64) error {
	if customerID == "" || points <= 0 {
		return nil
	}
	refType := "ADJUST"
	if err := s.repo.CreateTx(ctx, tx, &models.LoyaltyTransaction{CustomerID: customerID, BranchID: branchID, Type: "ADJUST", Points: points, ReferenceType: &refType, ReferenceID: &referenceID}); err != nil {
		return err
	}
	balance, err := s.repo.GetBalanceTx(ctx, tx, customerID, branchID)
	if err != nil {
		return err
	}
	return s.repo.UpdateCustomerPointsTx(ctx, tx, customerID, branchID, balance)
}
