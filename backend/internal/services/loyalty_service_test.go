package services

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5"

	"github.com/kcgsperera/texa-multi-pos/backend/internal/models"
)

type loyaltyRepoStub struct {
	balance float64
}

func (l *loyaltyRepoStub) CreateTx(ctx context.Context, tx pgx.Tx, t *models.LoyaltyTransaction) error {
	if t.Type == "REDEEM" {
		l.balance -= t.Points
	} else {
		l.balance += t.Points
	}
	return nil
}
func (l *loyaltyRepoStub) GetBalanceTx(ctx context.Context, tx pgx.Tx, customerID, branchID string) (float64, error) {
	return l.balance, nil
}
func (l *loyaltyRepoStub) UpdateCustomerPointsTx(ctx context.Context, tx pgx.Tx, customerID, branchID string, points float64) error {
	l.balance = points
	return nil
}

func TestLoyaltyEarnRedeemConsistency(t *testing.T) {
	repo := &loyaltyRepoStub{}
	svc := NewLoyaltyService(repo)

	if err := svc.EarnForSaleTx(context.Background(), nil, "c1", "b1", "s1", 200); err != nil {
		t.Fatal(err)
	}
	if repo.balance <= 0 {
		t.Fatalf("expected earned points > 0")
	}

	earned := repo.balance
	if err := svc.RedeemTx(context.Background(), nil, "c1", "b1", "s1", earned/2); err != nil {
		t.Fatal(err)
	}
	if repo.balance != earned/2 {
		t.Fatalf("expected %v, got %v", earned/2, repo.balance)
	}
}
