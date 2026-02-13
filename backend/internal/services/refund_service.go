package services

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/kcgsperera/texa-multi-pos/backend/internal/database"
	"github.com/kcgsperera/texa-multi-pos/backend/internal/logger"
	"github.com/kcgsperera/texa-multi-pos/backend/internal/models"
	"github.com/kcgsperera/texa-multi-pos/backend/internal/observability"
	"github.com/kcgsperera/texa-multi-pos/backend/internal/repositories"
)

type RefundService interface {
	Create(ctx context.Context, branchID, userID string, req models.CreateRefundRequest) (*models.Refund, error)
}

type refundService struct {
	db            *pgxpool.Pool
	txRunner      func(context.Context, func(pgx.Tx) error) error
	refundRepo    repositories.RefundRepository
	saleRepo      repositories.SaleRepository
	inventoryRepo repositories.InventoryRepository
	loyaltySvc    LoyaltyService
	ledgerSvc     LedgerService
	auditService  AuditService
}

func NewRefundService(db *pgxpool.Pool, refundRepo repositories.RefundRepository, saleRepo repositories.SaleRepository, inventoryRepo repositories.InventoryRepository, loyaltySvc LoyaltyService, ledgerSvc LedgerService, auditService AuditService) RefundService {
	s := &refundService{db: db, refundRepo: refundRepo, saleRepo: saleRepo, inventoryRepo: inventoryRepo, loyaltySvc: loyaltySvc, ledgerSvc: ledgerSvc, auditService: auditService}
	s.txRunner = func(ctx context.Context, fn func(pgx.Tx) error) error { return database.WithTx(ctx, db, fn) }
	return s
}

func (s *refundService) Create(ctx context.Context, branchID, userID string, req models.CreateRefundRequest) (*models.Refund, error) {
	if len(req.Items) == 0 {
		return nil, ErrInvalidInput
	}

	var out *models.Refund
	err := s.txRunner(ctx, func(tx pgx.Tx) error {
		sale, err := s.saleRepo.GetByIDTx(ctx, tx, branchID, req.SaleID)
		if err != nil {
			return err
		}
		if sale == nil {
			return ErrNotFound
		}
		saleItems, err := s.refundRepo.GetSaleItemsTx(ctx, tx, req.SaleID)
		if err != nil {
			return err
		}
		soldMap := map[string]float64{}
		for _, i := range saleItems {
			soldMap[i.ProductID] += i.Quantity
		}

		items := make([]models.RefundItem, 0, len(req.Items))
		total := 0.0
		for _, i := range req.Items {
			refunded, err := s.refundRepo.GetRefundedQtyBySaleProductTx(ctx, tx, req.SaleID, i.ProductID)
			if err != nil {
				return err
			}
			if soldMap[i.ProductID]-refunded < i.Quantity {
				return ErrInvalidInput
			}
			items = append(items, models.RefundItem{ProductID: i.ProductID, Quantity: i.Quantity, LineTotal: i.LineTotal})
			total += i.LineTotal
		}

		createdBy := userID
		refund, err := s.refundRepo.CreateRefundTx(ctx, tx, &models.Refund{SaleID: req.SaleID, BranchID: branchID, TotalAmount: total, Reason: req.Reason, CreatedBy: &createdBy})
		if err != nil {
			return err
		}
		refund.Items, err = s.refundRepo.CreateRefundItemsTx(ctx, tx, refund.ID, items)
		if err != nil {
			return err
		}

		for _, i := range refund.Items {
			if err := s.inventoryRepo.EnsureInventoryRowTx(ctx, tx, branchID, i.ProductID); err != nil {
				return err
			}
			inv, err := s.inventoryRepo.GetForUpdateTx(ctx, tx, branchID, i.ProductID)
			if err != nil {
				return err
			}
			if err := s.inventoryRepo.UpdateQuantityTx(ctx, tx, branchID, i.ProductID, inv.Quantity+i.Quantity); err != nil {
				return err
			}
			refType := "REFUND"
			refID := refund.ID
			if err := s.inventoryRepo.CreateStockMovementTx(ctx, tx, &models.StockMovement{ProductID: i.ProductID, BranchID: branchID, MovementType: "IN", Quantity: i.Quantity, ReferenceType: &refType, ReferenceID: &refID, CreatedBy: &userID}); err != nil {
				return err
			}
		}

		if sale.CustomerID != nil {
			if err := s.loyaltySvc.RedeemTx(ctx, tx, *sale.CustomerID, branchID, refund.ID, total/LoyaltyEarnPerCurrency); err != nil {
				return err
			}
		}

		if err := s.ledgerSvc.PostByCodes(ctx, tx, branchID, "REFUND", refund.ID, []LedgerLineByCode{
			{AccountCode: "REVENUE", Debit: total, Credit: 0},
			{AccountCode: "VAT_PAYABLE", Debit: sale.VATAmount, Credit: 0},
			{AccountCode: "CASH", Debit: 0, Credit: total + sale.VATAmount},
		}, &userID); err != nil {
			return err
		}

		if err := s.auditService.Log(ctx, tx, "refunds", refund.ID, "CREATE", &userID, nil, refund); err != nil {
			return err
		}
		out = refund
		return nil
	})
	if err != nil {
		return nil, err
	}

	observability.IncRefundCreate()
	logger.L().Info().Str("branch_id", branchID).Str("user_id", userID).Str("entity_id", out.ID).Str("action", "refund_create").Msg("refund_create")
	return out, nil
}
