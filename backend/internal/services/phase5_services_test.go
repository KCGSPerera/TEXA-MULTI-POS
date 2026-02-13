package services

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/kcgsperera/texa-multi-pos/backend/internal/models"
)

type noopAudit struct{}

func (n noopAudit) Log(ctx context.Context, tx pgx.Tx, entityName, entityID, action string, performedBy *string, oldData, newData interface{}) error {
	return nil
}

type saleRepoStub struct {
	existing    *models.Sale
	createCalls int
}

func (s *saleRepoStub) CreateSaleTx(ctx context.Context, tx pgx.Tx, sale *models.Sale) (*models.Sale, error) {
	s.createCalls++
	return &models.Sale{ID: "new-sale", BranchID: sale.BranchID, UserID: sale.UserID, NetAmount: sale.NetAmount}, nil
}
func (s *saleRepoStub) CreateSaleItemsTx(ctx context.Context, tx pgx.Tx, saleID string, items []models.SaleItem) error {
	return nil
}
func (s *saleRepoStub) CreateSalePaymentsTx(ctx context.Context, tx pgx.Tx, saleID string, payments []models.SalePayment) ([]models.SalePayment, error) {
	return payments, nil
}
func (s *saleRepoStub) GetByIdempotencyKey(ctx context.Context, branchID, idempotencyKey string) (*models.Sale, error) {
	return s.existing, nil
}

type productRepoStub struct{ exists bool }

func (p *productRepoStub) Create(ctx context.Context, product *models.Product) (*models.Product, error) {
	return nil, nil
}
func (p *productRepoStub) GetByID(ctx context.Context, branchID, id string) (*models.Product, error) {
	return nil, nil
}
func (p *productRepoStub) ListByBranch(ctx context.Context, branchID string, limit, offset int) ([]models.Product, error) {
	return nil, nil
}
func (p *productRepoStub) Update(ctx context.Context, product *models.Product) (*models.Product, error) {
	return nil, nil
}
func (p *productRepoStub) Delete(ctx context.Context, branchID, id, userID string) (bool, error) {
	return false, nil
}
func (p *productRepoStub) ExistsActiveInBranch(ctx context.Context, branchID, id string) (bool, error) {
	return p.exists, nil
}

type inventoryRepoStub struct {
	inv          *models.Inventory
	updateCalled bool
	updatedQty   float64
}

func (i *inventoryRepoStub) ListByBranch(ctx context.Context, branchID string) ([]models.Inventory, error) {
	return nil, nil
}
func (i *inventoryRepoStub) GetByProduct(ctx context.Context, branchID, productID string) (*models.Inventory, error) {
	return i.inv, nil
}
func (i *inventoryRepoStub) EnsureInventoryRowTx(ctx context.Context, tx pgx.Tx, branchID, productID string) error {
	return nil
}
func (i *inventoryRepoStub) GetForUpdateTx(ctx context.Context, tx pgx.Tx, branchID, productID string) (*models.Inventory, error) {
	return i.inv, nil
}
func (i *inventoryRepoStub) UpdateQuantityTx(ctx context.Context, tx pgx.Tx, branchID, productID string, quantity float64) error {
	i.updateCalled = true
	i.updatedQty = quantity
	return nil
}
func (i *inventoryRepoStub) CreateStockMovementTx(ctx context.Context, tx pgx.Tx, movement *models.StockMovement) error {
	return nil
}

func TestSaleServiceIdempotencyReturnsExisting(t *testing.T) {
	existing := &models.Sale{ID: "sale-1", BranchID: "b1", UserID: "u1"}
	sr := &saleRepoStub{existing: existing}
	svc := &saleService{
		saleRepo:      sr,
		productRepo:   &productRepoStub{exists: true},
		inventoryRepo: &inventoryRepoStub{},
		auditService:  noopAudit{},
		txRunner: func(ctx context.Context, fn func(pgx.Tx) error) error {
			return fn(nil)
		},
	}
	key := "idem-1"
	out, err := svc.Create(context.Background(), "b1", "u1", models.CreateSaleRequest{
		NetAmount:      10,
		TotalAmount:    10,
		VATAmount:      0,
		DiscountAmount: 0,
		Items:          []models.CreateSaleItemRequest{{ProductID: "p1", Quantity: 1, UnitPrice: 10, VATAmount: 0, LineTotal: 10}},
		Payments:       []models.CreateSalePaymentRequest{{PaymentMethod: "CASH", Amount: 10}},
		IdempotencyKey: &key,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.ID != existing.ID {
		t.Fatalf("expected existing sale ID %s, got %s", existing.ID, out.ID)
	}
	if sr.createCalls != 0 {
		t.Fatalf("expected no create call, got %d", sr.createCalls)
	}
}

func TestInventoryAdjustRejectsNegativeResult(t *testing.T) {
	ir := &inventoryRepoStub{inv: &models.Inventory{ProductID: "p1", BranchID: "b1", Quantity: 2, ReservedQuantity: 0}}
	svc := &inventoryService{
		inventoryRepo: ir,
		productRepo:   &productRepoStub{exists: true},
		auditService:  noopAudit{},
		txRunner: func(ctx context.Context, fn func(pgx.Tx) error) error {
			return fn(nil)
		},
	}
	_, err := svc.Adjust(context.Background(), "b1", "u1", models.AdjustInventoryRequest{ProductID: "p1", AdjustmentQuantity: -5})
	if !errors.Is(err, ErrInsufficientInventory) {
		t.Fatalf("expected ErrInsufficientInventory, got %v", err)
	}
	if ir.updateCalled {
		t.Fatalf("update should not be called on negative result")
	}
}

type purchaseRepoStub struct {
	grn            *models.GRN
	items          []models.GRNItem
	incPOCalled    bool
	recalcCalled   bool
	updateApproved bool
}

func (p *purchaseRepoStub) ListPurchaseOrdersByBranch(ctx context.Context, branchID string, limit, offset int) ([]models.PurchaseOrder, error) {
	return nil, nil
}
func (p *purchaseRepoStub) CreatePurchaseOrderTx(ctx context.Context, tx pgx.Tx, po *models.PurchaseOrder) (*models.PurchaseOrder, error) {
	return nil, nil
}
func (p *purchaseRepoStub) CreatePurchaseOrderItemsTx(ctx context.Context, tx pgx.Tx, poID string, items []models.PurchaseOrderItem) ([]models.PurchaseOrderItem, error) {
	return nil, nil
}
func (p *purchaseRepoStub) GetPurchaseOrderByIDTx(ctx context.Context, tx pgx.Tx, branchID, poID string) (*models.PurchaseOrder, error) {
	return nil, nil
}
func (p *purchaseRepoStub) CreateGRNTx(ctx context.Context, tx pgx.Tx, grn *models.GRN) (*models.GRN, error) {
	return nil, nil
}
func (p *purchaseRepoStub) CreateGRNItemsTx(ctx context.Context, tx pgx.Tx, grnID string, items []models.GRNItem) ([]models.GRNItem, error) {
	return nil, nil
}
func (p *purchaseRepoStub) GetGRNByIDForUpdateTx(ctx context.Context, tx pgx.Tx, branchID, grnID string) (*models.GRN, error) {
	return p.grn, nil
}
func (p *purchaseRepoStub) GetGRNItemsTx(ctx context.Context, tx pgx.Tx, grnID string) ([]models.GRNItem, error) {
	return p.items, nil
}
func (p *purchaseRepoStub) UpdateGRNApprovedTx(ctx context.Context, tx pgx.Tx, grnID, approvedBy string, approvedAt time.Time) error {
	p.updateApproved = true
	return nil
}
func (p *purchaseRepoStub) IncrementPOReceivedQtyTx(ctx context.Context, tx pgx.Tx, poID, productID string, qty float64) error {
	p.incPOCalled = true
	return nil
}
func (p *purchaseRepoStub) RecalculatePOStatusTx(ctx context.Context, tx pgx.Tx, poID string) error {
	p.recalcCalled = true
	return nil
}

func TestApproveGRNIncrementsInventory(t *testing.T) {
	pr := &purchaseRepoStub{grn: &models.GRN{ID: "g1", PurchaseOrderID: "po1", BranchID: "b1", Status: "OPEN"}, items: []models.GRNItem{{ProductID: "p1", Quantity: 3}}}
	ir := &inventoryRepoStub{inv: &models.Inventory{ProductID: "p1", BranchID: "b1", Quantity: 2, ReservedQuantity: 0}}
	svc := &purchaseService{
		purchaseRepo:  pr,
		inventoryRepo: ir,
		auditService:  noopAudit{},
		txRunner: func(ctx context.Context, fn func(pgx.Tx) error) error {
			return fn(nil)
		},
	}

	out, err := svc.ApproveGRN(context.Background(), "b1", "u1", "g1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ir.updateCalled || ir.updatedQty != 5 {
		t.Fatalf("expected inventory update to 5, got called=%v qty=%v", ir.updateCalled, ir.updatedQty)
	}
	if !pr.incPOCalled || !pr.recalcCalled || !pr.updateApproved {
		t.Fatalf("expected PO increment/recalc and GRN approve to be called")
	}
	if out.Status != "COMPLETED" {
		t.Fatalf("expected COMPLETED status, got %s", out.Status)
	}
}

type categoryRepoPaginationStub struct {
	limitSeen  int
	offsetSeen int
}

func (c *categoryRepoPaginationStub) Create(ctx context.Context, category *models.Category) (*models.Category, error) {
	return nil, nil
}
func (c *categoryRepoPaginationStub) GetByID(ctx context.Context, branchID, id string) (*models.Category, error) {
	return nil, nil
}
func (c *categoryRepoPaginationStub) ListByBranch(ctx context.Context, branchID string, limit, offset int) ([]models.Category, error) {
	c.limitSeen = limit
	c.offsetSeen = offset
	return []models.Category{}, nil
}
func (c *categoryRepoPaginationStub) Update(ctx context.Context, category *models.Category) (*models.Category, error) {
	return nil, nil
}
func (c *categoryRepoPaginationStub) Delete(ctx context.Context, branchID, id, userID string) (bool, error) {
	return false, nil
}

func TestCategoryPaginationNormalization(t *testing.T) {
	tests := []struct {
		name         string
		limit        int
		offset       int
		expectLimit  int
		expectOffset int
	}{
		{name: "default values", limit: 0, offset: 0, expectLimit: 20, expectOffset: 0},
		{name: "max clamp", limit: 999, offset: 10, expectLimit: 100, expectOffset: 10},
		{name: "negative offset", limit: 5, offset: -1, expectLimit: 5, expectOffset: 0},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			repo := &categoryRepoPaginationStub{}
			svc := &categoryService{repo: repo, auditService: noopAudit{}}
			_, err := svc.ListByBranch(context.Background(), "b1", tc.limit, tc.offset)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if repo.limitSeen != tc.expectLimit || repo.offsetSeen != tc.expectOffset {
				t.Fatalf("expected limit=%d offset=%d, got limit=%d offset=%d", tc.expectLimit, tc.expectOffset, repo.limitSeen, repo.offsetSeen)
			}
		})
	}
}
