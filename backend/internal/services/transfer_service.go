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

type TransferService interface {
	Create(ctx context.Context, branchID, userID string, req models.CreateBranchTransferRequest) (*models.BranchTransfer, error)
	Complete(ctx context.Context, branchID, userID, transferID string) (*models.BranchTransfer, error)
}

type transferService struct {
	db            *pgxpool.Pool
	txRunner      func(context.Context, func(pgx.Tx) error) error
	transferRepo  repositories.TransferRepository
	inventoryRepo repositories.InventoryRepository
	productRepo   repositories.ProductRepository
	ledgerSvc     LedgerService
	auditService  AuditService
}

func NewTransferService(db *pgxpool.Pool, transferRepo repositories.TransferRepository, inventoryRepo repositories.InventoryRepository, productRepo repositories.ProductRepository, ledgerSvc LedgerService, auditService AuditService) TransferService {
	s := &transferService{db: db, transferRepo: transferRepo, inventoryRepo: inventoryRepo, productRepo: productRepo, ledgerSvc: ledgerSvc, auditService: auditService}
	s.txRunner = func(ctx context.Context, fn func(pgx.Tx) error) error { return database.WithTx(ctx, db, fn) }
	return s
}

func (s *transferService) Create(ctx context.Context, branchID, userID string, req models.CreateBranchTransferRequest) (*models.BranchTransfer, error) {
	if branchID == req.ToBranchID || len(req.Items) == 0 {
		return nil, ErrInvalidInput
	}
	var out *models.BranchTransfer
	err := s.txRunner(ctx, func(tx pgx.Tx) error {
		createdBy := userID
		tr, err := s.transferRepo.CreateTransferTx(ctx, tx, &models.BranchTransfer{FromBranchID: branchID, ToBranchID: req.ToBranchID, Status: "IN_TRANSIT", CreatedBy: &createdBy})
		if err != nil {
			return err
		}
		items := make([]models.BranchTransferItem, 0, len(req.Items))
		for _, i := range req.Items {
			exists, err := s.productRepo.ExistsActiveInBranch(ctx, branchID, i.ProductID)
			if err != nil || !exists {
				return ErrInvalidInput
			}
			if err := s.inventoryRepo.EnsureInventoryRowTx(ctx, tx, branchID, i.ProductID); err != nil {
				return err
			}
			inv, err := s.inventoryRepo.GetForUpdateTx(ctx, tx, branchID, i.ProductID)
			if err != nil {
				return err
			}
			if inv.Quantity-inv.ReservedQuantity < i.Quantity {
				return ErrInsufficientInventory
			}
			if err := s.inventoryRepo.UpdateQuantityTx(ctx, tx, branchID, i.ProductID, inv.Quantity-i.Quantity); err != nil {
				return err
			}
			refType := "TRANSFER"
			refID := tr.ID
			if err := s.inventoryRepo.CreateStockMovementTx(ctx, tx, &models.StockMovement{ProductID: i.ProductID, BranchID: branchID, MovementType: "TRANSFER_OUT", Quantity: i.Quantity, ReferenceType: &refType, ReferenceID: &refID, CreatedBy: &userID}); err != nil {
				return err
			}
			items = append(items, models.BranchTransferItem{ProductID: i.ProductID, Quantity: i.Quantity})
		}
		tr.Items, err = s.transferRepo.CreateTransferItemsTx(ctx, tx, tr.ID, items)
		if err != nil {
			return err
		}
		if err := s.auditService.Log(ctx, tx, "branch_transfers", tr.ID, "CREATE", &userID, nil, tr); err != nil {
			return err
		}
		out = tr
		return nil
	})
	if err != nil {
		return nil, err
	}
	observability.IncTransferCreate()
	logger.L().Info().Str("branch_id", branchID).Str("user_id", userID).Str("entity_id", out.ID).Str("action", "transfer_create").Msg("transfer_create")
	return out, nil
}

func (s *transferService) Complete(ctx context.Context, branchID, userID, transferID string) (*models.BranchTransfer, error) {
	var out *models.BranchTransfer
	err := s.txRunner(ctx, func(tx pgx.Tx) error {
		tr, err := s.transferRepo.GetTransferForUpdateTx(ctx, tx, transferID, branchID)
		if err != nil {
			return err
		}
		if tr == nil || tr.Status == "COMPLETED" {
			return ErrInvalidInput
		}
		items, err := s.transferRepo.GetTransferItemsTx(ctx, tx, transferID)
		if err != nil {
			return err
		}
		for _, i := range items {
			if err := s.inventoryRepo.EnsureInventoryRowTx(ctx, tx, tr.ToBranchID, i.ProductID); err != nil {
				return err
			}
			inv, err := s.inventoryRepo.GetForUpdateTx(ctx, tx, tr.ToBranchID, i.ProductID)
			if err != nil {
				return err
			}
			if err := s.inventoryRepo.UpdateQuantityTx(ctx, tx, tr.ToBranchID, i.ProductID, inv.Quantity+i.Quantity); err != nil {
				return err
			}
			refType := "TRANSFER"
			refID := tr.ID
			if err := s.inventoryRepo.CreateStockMovementTx(ctx, tx, &models.StockMovement{ProductID: i.ProductID, BranchID: tr.ToBranchID, MovementType: "TRANSFER_IN", Quantity: i.Quantity, ReferenceType: &refType, ReferenceID: &refID, CreatedBy: &userID}); err != nil {
				return err
			}
		}
		if err := s.transferRepo.UpdateTransferStatusTx(ctx, tx, transferID, "COMPLETED"); err != nil {
			return err
		}
		tr.Status = "COMPLETED"
		tr.Items = items
		if err := s.auditService.Log(ctx, tx, "branch_transfers", tr.ID, "UPDATE", &userID, nil, tr); err != nil {
			return err
		}
		out = tr
		return nil
	})
	if err != nil {
		return nil, err
	}
	logger.L().Info().Str("branch_id", branchID).Str("user_id", userID).Str("entity_id", out.ID).Str("action", "transfer_complete").Msg("transfer_complete")
	return out, nil
}
