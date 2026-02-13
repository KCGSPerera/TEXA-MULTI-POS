package services

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/kcgsperera/texa-multi-pos/backend/internal/database"
	"github.com/kcgsperera/texa-multi-pos/backend/internal/logger"
	"github.com/kcgsperera/texa-multi-pos/backend/internal/models"
	"github.com/kcgsperera/texa-multi-pos/backend/internal/repositories"
)

type PurchaseService interface {
	CreatePurchaseOrder(ctx context.Context, branchID, userID string, req models.CreatePurchaseOrderRequest) (*models.PurchaseOrder, error)
	ListPurchaseOrders(ctx context.Context, branchID string, limit, offset int) ([]models.PurchaseOrder, error)
	CreateGRN(ctx context.Context, branchID, userID, poID string, req models.CreateGRNRequest) (*models.GRN, error)
	ApproveGRN(ctx context.Context, branchID, userID, grnID string) (*models.GRN, error)
}

type purchaseService struct {
	db            *pgxpool.Pool
	txRunner      func(context.Context, func(pgx.Tx) error) error
	purchaseRepo  repositories.PurchaseRepository
	supplierRepo  repositories.SupplierRepository
	productRepo   repositories.ProductRepository
	inventoryRepo repositories.InventoryRepository
	ledgerSvc     LedgerService
	auditService  AuditService
}

func NewPurchaseService(db *pgxpool.Pool, purchaseRepo repositories.PurchaseRepository, supplierRepo repositories.SupplierRepository, productRepo repositories.ProductRepository, inventoryRepo repositories.InventoryRepository, ledgerSvc LedgerService, auditService AuditService) PurchaseService {
	s := &purchaseService{db: db, purchaseRepo: purchaseRepo, supplierRepo: supplierRepo, productRepo: productRepo, inventoryRepo: inventoryRepo, ledgerSvc: ledgerSvc, auditService: auditService}
	s.txRunner = func(ctx context.Context, fn func(pgx.Tx) error) error {
		return database.WithTx(ctx, db, fn)
	}
	return s
}

func (s *purchaseService) CreatePurchaseOrder(ctx context.Context, branchID, userID string, req models.CreatePurchaseOrderRequest) (*models.PurchaseOrder, error) {
	supplier, err := s.supplierRepo.GetByID(ctx, branchID, req.SupplierID)
	if err != nil {
		return nil, err
	}
	if supplier == nil {
		return nil, ErrInvalidInput
	}

	total := 0.0
	items := make([]models.PurchaseOrderItem, 0, len(req.Items))
	for _, item := range req.Items {
		exists, err := s.productRepo.ExistsActiveInBranch(ctx, branchID, item.ProductID)
		if err != nil {
			return nil, err
		}
		if !exists {
			return nil, ErrInvalidInput
		}
		total += item.LineTotal
		items = append(items, models.PurchaseOrderItem{ProductID: item.ProductID, Quantity: item.Quantity, UnitCost: item.UnitCost, LineTotal: item.LineTotal})
	}

	var po *models.PurchaseOrder
	err = s.txRunner(ctx, func(tx pgx.Tx) error {
		var err error
		po, err = s.purchaseRepo.CreatePurchaseOrderTx(ctx, tx, &models.PurchaseOrder{BranchID: branchID, SupplierID: req.SupplierID, Status: "OPEN", TotalAmount: total, CreatedBy: &userID})
		if err != nil {
			return err
		}
		po.Items, err = s.purchaseRepo.CreatePurchaseOrderItemsTx(ctx, tx, po.ID, items)
		if err != nil {
			return err
		}
		return s.auditService.Log(ctx, tx, "purchase_orders", po.ID, "CREATE", &userID, nil, po)
	})
	if err != nil {
		return nil, err
	}
	return po, nil
}

func (s *purchaseService) ListPurchaseOrders(ctx context.Context, branchID string, limit, offset int) ([]models.PurchaseOrder, error) {
	limit, offset = normalizePagination(limit, offset)
	return s.purchaseRepo.ListPurchaseOrdersByBranch(ctx, branchID, limit, offset)
}

func (s *purchaseService) CreateGRN(ctx context.Context, branchID, userID, poID string, req models.CreateGRNRequest) (*models.GRN, error) {
	var grn *models.GRN
	err := s.txRunner(ctx, func(tx pgx.Tx) error {
		po, err := s.purchaseRepo.GetPurchaseOrderByIDTx(ctx, tx, branchID, poID)
		if err != nil {
			return err
		}
		if po == nil || po.Status == "CANCELLED" {
			return ErrInvalidInput
		}

		total := 0.0
		items := make([]models.GRNItem, 0, len(req.Items))
		for _, item := range req.Items {
			exists, err := s.productRepo.ExistsActiveInBranch(ctx, branchID, item.ProductID)
			if err != nil {
				return err
			}
			if !exists {
				return ErrInvalidInput
			}
			total += item.LineTotal
			items = append(items, models.GRNItem{ProductID: item.ProductID, Quantity: item.Quantity, UnitCost: item.UnitCost, LineTotal: item.LineTotal})
		}

		grn, err = s.purchaseRepo.CreateGRNTx(ctx, tx, &models.GRN{PurchaseOrderID: poID, BranchID: branchID, SupplierID: po.SupplierID, Status: "OPEN", TotalAmount: total, CreatedBy: &userID})
		if err != nil {
			return err
		}
		grn.Items, err = s.purchaseRepo.CreateGRNItemsTx(ctx, tx, grn.ID, items)
		if err != nil {
			return err
		}
		return s.auditService.Log(ctx, tx, "grns", grn.ID, "CREATE", &userID, nil, grn)
	})
	if err != nil {
		return nil, err
	}
	return grn, nil
}

func (s *purchaseService) ApproveGRN(ctx context.Context, branchID, userID, grnID string) (*models.GRN, error) {
	var grn *models.GRN
	err := s.txRunner(ctx, func(tx pgx.Tx) error {
		var err error
		grn, err = s.purchaseRepo.GetGRNByIDForUpdateTx(ctx, tx, branchID, grnID)
		if err != nil {
			return err
		}
		if grn == nil || grn.Status == "CANCELLED" || grn.Status == "COMPLETED" {
			return ErrInvalidInput
		}
		items, err := s.purchaseRepo.GetGRNItemsTx(ctx, tx, grn.ID)
		if err != nil {
			return err
		}
		if len(items) == 0 {
			return ErrInvalidInput
		}

		for _, item := range items {
			if err := s.inventoryRepo.EnsureInventoryRowTx(ctx, tx, branchID, item.ProductID); err != nil {
				return err
			}
			inv, err := s.inventoryRepo.GetForUpdateTx(ctx, tx, branchID, item.ProductID)
			if err != nil {
				return err
			}
			if err := s.inventoryRepo.UpdateQuantityTx(ctx, tx, branchID, item.ProductID, inv.Quantity+item.Quantity); err != nil {
				return err
			}
			refType := "GRN"
			refID := grn.ID
			if err := s.inventoryRepo.CreateStockMovementTx(ctx, tx, &models.StockMovement{ProductID: item.ProductID, BranchID: branchID, MovementType: "IN", Quantity: item.Quantity, ReferenceType: &refType, ReferenceID: &refID, CreatedBy: &userID}); err != nil {
				return err
			}
			if err := s.purchaseRepo.IncrementPOReceivedQtyTx(ctx, tx, grn.PurchaseOrderID, item.ProductID, item.Quantity); err != nil {
				return err
			}
		}
		approvedAt := time.Now().UTC()
		if err := s.purchaseRepo.UpdateGRNApprovedTx(ctx, tx, grn.ID, userID, approvedAt); err != nil {
			return err
		}
		if err := s.purchaseRepo.RecalculatePOStatusTx(ctx, tx, grn.PurchaseOrderID); err != nil {
			return err
		}
		if err := s.ledgerSvc.PostByCodes(ctx, tx, branchID, "GRN", grn.ID, []LedgerLineByCode{
			{AccountCode: "INVENTORY", Debit: grn.TotalAmount, Credit: 0},
			{AccountCode: "PAYABLE", Debit: 0, Credit: grn.TotalAmount},
		}, &userID); err != nil {
			return err
		}
		grn.Status = "COMPLETED"
		grn.Items = items
		grn.ApprovedBy = &userID
		grn.ApprovedAt = &approvedAt
		if err := s.auditService.Log(ctx, tx, "grns", grn.ID, "UPDATE", &userID, nil, grn); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	logger.L().Info().Str("branch_id", branchID).Str("user_id", userID).Str("entity_id", grn.ID).Str("action", "grn_approve").Msg("grn_approve")
	return grn, nil
}
