package services

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5"

	"github.com/kcgsperera/texa-multi-pos/backend/internal/models"
)

type refundRepoSimpleStub struct{}

func (r *refundRepoSimpleStub) CreateRefundTx(ctx context.Context, tx pgx.Tx, refund *models.Refund) (*models.Refund, error) {
	refund.ID = "r1"
	return refund, nil
}
func (r *refundRepoSimpleStub) CreateRefundItemsTx(ctx context.Context, tx pgx.Tx, refundID string, items []models.RefundItem) ([]models.RefundItem, error) {
	return items, nil
}
func (r *refundRepoSimpleStub) GetSaleItemsTx(ctx context.Context, tx pgx.Tx, saleID string) ([]models.SaleItem, error) {
	return []models.SaleItem{{SaleID: saleID, ProductID: "p1", Quantity: 2, LineTotal: 50}}, nil
}
func (r *refundRepoSimpleStub) GetRefundedQtyBySaleProductTx(ctx context.Context, tx pgx.Tx, saleID, productID string) (float64, error) {
	return 0, nil
}

type saleRepoSimpleStub struct{}

func (s *saleRepoSimpleStub) CreateSaleTx(ctx context.Context, tx pgx.Tx, sale *models.Sale) (*models.Sale, error) {
	return nil, nil
}
func (s *saleRepoSimpleStub) CreateSaleItemsTx(ctx context.Context, tx pgx.Tx, saleID string, items []models.SaleItem) error {
	return nil
}
func (s *saleRepoSimpleStub) CreateSalePaymentsTx(ctx context.Context, tx pgx.Tx, saleID string, payments []models.SalePayment) ([]models.SalePayment, error) {
	return nil, nil
}
func (s *saleRepoSimpleStub) GetByIdempotencyKey(ctx context.Context, branchID, idempotencyKey string) (*models.Sale, error) {
	return nil, nil
}
func (s *saleRepoSimpleStub) GetByIDTx(ctx context.Context, tx pgx.Tx, branchID, saleID string) (*models.Sale, error) {
	return &models.Sale{ID: saleID, BranchID: branchID, VATAmount: 0}, nil
}

func TestRefundRestoresInventory(t *testing.T) {
	inv := &inventoryRepoStub{inv: &models.Inventory{ProductID: "p1", BranchID: "b1", Quantity: 1, ReservedQuantity: 0}}
	svc := &refundService{
		refundRepo:    &refundRepoSimpleStub{},
		saleRepo:      &saleRepoSimpleStub{},
		inventoryRepo: inv,
		loyaltySvc:    NewLoyaltyService(&loyaltyRepoStub{}),
		ledgerSvc:     NewLedgerService(&ledgerRepoStub{}, noopAudit{}),
		auditService:  noopAudit{},
		txRunner: func(ctx context.Context, fn func(pgx.Tx) error) error {
			return fn(nil)
		},
	}
	_, err := svc.Create(context.Background(), "b1", "u1", models.CreateRefundRequest{SaleID: "s1", Items: []models.CreateRefundItemRequest{{ProductID: "p1", Quantity: 1, LineTotal: 20}}})
	if err != nil {
		t.Fatal(err)
	}
	if !inv.updateCalled || inv.updatedQty != 2 {
		t.Fatalf("expected restored quantity 2, got called=%v qty=%v", inv.updateCalled, inv.updatedQty)
	}
}
