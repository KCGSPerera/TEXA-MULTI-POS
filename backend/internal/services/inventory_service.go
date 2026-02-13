package services

import (
	"context"
	"math"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/kcgsperera/texa-multi-pos/backend/internal/database"
	"github.com/kcgsperera/texa-multi-pos/backend/internal/logger"
	"github.com/kcgsperera/texa-multi-pos/backend/internal/models"
	"github.com/kcgsperera/texa-multi-pos/backend/internal/repositories"
)

type InventoryService interface {
	ListByBranch(ctx context.Context, branchID string) ([]models.Inventory, error)
	GetByProduct(ctx context.Context, branchID, productID string) (*models.Inventory, error)
	Adjust(ctx context.Context, branchID, userID string, req models.AdjustInventoryRequest) (*models.Inventory, error)
}

type inventoryService struct {
	db            *pgxpool.Pool
	txRunner      func(context.Context, func(pgx.Tx) error) error
	inventoryRepo repositories.InventoryRepository
	productRepo   repositories.ProductRepository
	auditService  AuditService
}

func NewInventoryService(db *pgxpool.Pool, inventoryRepo repositories.InventoryRepository, productRepo repositories.ProductRepository, auditService AuditService) InventoryService {
	s := &inventoryService{db: db, inventoryRepo: inventoryRepo, productRepo: productRepo, auditService: auditService}
	s.txRunner = func(ctx context.Context, fn func(pgx.Tx) error) error {
		return database.WithTx(ctx, db, fn)
	}
	return s
}

func (s *inventoryService) ListByBranch(ctx context.Context, branchID string) ([]models.Inventory, error) {
	return s.inventoryRepo.ListByBranch(ctx, branchID)
}

func (s *inventoryService) GetByProduct(ctx context.Context, branchID, productID string) (*models.Inventory, error) {
	inv, err := s.inventoryRepo.GetByProduct(ctx, branchID, productID)
	if err != nil {
		return nil, err
	}
	if inv == nil {
		return nil, ErrNotFound
	}
	return inv, nil
}

func (s *inventoryService) Adjust(ctx context.Context, branchID, userID string, req models.AdjustInventoryRequest) (*models.Inventory, error) {
	exists, err := s.productRepo.ExistsActiveInBranch(ctx, branchID, req.ProductID)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, ErrInvalidInput
	}

	var result *models.Inventory
	err = s.txRunner(ctx, func(tx pgx.Tx) error {
		if err := s.inventoryRepo.EnsureInventoryRowTx(ctx, tx, branchID, req.ProductID); err != nil {
			return err
		}

		inv, err := s.inventoryRepo.GetForUpdateTx(ctx, tx, branchID, req.ProductID)
		if err != nil {
			if err == pgx.ErrNoRows {
				return ErrInvalidInput
			}
			return err
		}

		newQty := inv.Quantity + req.AdjustmentQuantity
		if newQty < 0 {
			return ErrInsufficientInventory
		}

		if err := s.inventoryRepo.UpdateQuantityTx(ctx, tx, branchID, req.ProductID, newQty); err != nil {
			return err
		}

		refType := req.ReferenceType
		if refType == "" {
			refType = "MANUAL"
		}

		if err := s.inventoryRepo.CreateStockMovementTx(ctx, tx, &models.StockMovement{ProductID: req.ProductID, BranchID: branchID, MovementType: "ADJUSTMENT", Quantity: math.Abs(req.AdjustmentQuantity), ReferenceType: &refType, ReferenceID: req.ReferenceID, CreatedBy: &userID}); err != nil {
			return err
		}

		result = &models.Inventory{ProductID: inv.ProductID, BranchID: inv.BranchID, Quantity: newQty, ReservedQuantity: inv.ReservedQuantity, AvailableQuantity: newQty - inv.ReservedQuantity, CreatedAt: inv.CreatedAt}
		if err := s.auditService.Log(ctx, tx, "inventory", req.ProductID, "UPDATE", &userID, inv, result); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	logger.L().Info().Str("branch_id", branchID).Str("user_id", userID).Str("entity_id", req.ProductID).Str("action", "inventory_adjustment").Msg("inventory_adjustment")
	return result, nil
}
