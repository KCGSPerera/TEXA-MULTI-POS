package services

import (
	"context"
	"errors"
	"math"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/kcgsperera/texa-multi-pos/backend/internal/database"
	"github.com/kcgsperera/texa-multi-pos/backend/internal/logger"
	"github.com/kcgsperera/texa-multi-pos/backend/internal/models"
	"github.com/kcgsperera/texa-multi-pos/backend/internal/repositories"
)

type SaleService interface {
	Create(ctx context.Context, branchID, userID string, req models.CreateSaleRequest) (*models.Sale, error)
}

type saleService struct {
	db            *pgxpool.Pool
	txRunner      func(context.Context, func(pgx.Tx) error) error
	saleRepo      repositories.SaleRepository
	productRepo   repositories.ProductRepository
	inventoryRepo repositories.InventoryRepository
	auditService  AuditService
}

func NewSaleService(
	db *pgxpool.Pool,
	saleRepo repositories.SaleRepository,
	productRepo repositories.ProductRepository,
	inventoryRepo repositories.InventoryRepository,
	auditService AuditService,
) SaleService {
	s := &saleService{db: db, saleRepo: saleRepo, productRepo: productRepo, inventoryRepo: inventoryRepo, auditService: auditService}
	s.txRunner = func(ctx context.Context, fn func(pgx.Tx) error) error {
		return database.WithTx(ctx, db, fn)
	}
	return s
}

func (s *saleService) Create(ctx context.Context, branchID, userID string, req models.CreateSaleRequest) (*models.Sale, error) {
	if len(req.Items) == 0 || len(req.Payments) == 0 {
		return nil, ErrInvalidInput
	}

	var idemKey *string
	if req.IdempotencyKey != nil {
		trimmed := strings.TrimSpace(*req.IdempotencyKey)
		if trimmed != "" {
			idemKey = &trimmed
			existing, err := s.saleRepo.GetByIdempotencyKey(ctx, branchID, trimmed)
			if err != nil {
				return nil, err
			}
			if existing != nil {
				logger.L().Info().Str("branch_id", branchID).Str("user_id", userID).Str("entity_id", existing.ID).Str("action", "sale_idempotent_return").Msg("sale_create")
				return existing, nil
			}
		}
	}

	paymentTotal := 0.0
	for _, payment := range req.Payments {
		paymentTotal += payment.Amount
	}
	if math.Abs(paymentTotal-req.NetAmount) > 0.01 {
		return nil, ErrInvalidPaymentTotal
	}

	sale := &models.Sale{BranchID: branchID, UserID: userID, TotalAmount: req.TotalAmount, VATAmount: req.VATAmount, DiscountAmount: req.DiscountAmount, NetAmount: req.NetAmount, IdempotencyKey: idemKey}
	items := make([]models.SaleItem, 0, len(req.Items))
	for _, item := range req.Items {
		exists, err := s.productRepo.ExistsActiveInBranch(ctx, branchID, item.ProductID)
		if err != nil {
			return nil, err
		}
		if !exists {
			return nil, ErrInvalidInput
		}
		items = append(items, models.SaleItem{ProductID: item.ProductID, Quantity: item.Quantity, UnitPrice: item.UnitPrice, VATAmount: item.VATAmount, LineTotal: item.LineTotal})
	}
	payments := make([]models.SalePayment, 0, len(req.Payments))
	for _, payment := range req.Payments {
		payments = append(payments, models.SalePayment{PaymentMethod: strings.ToUpper(payment.PaymentMethod), Amount: payment.Amount})
	}

	var createdSale *models.Sale
	err := s.txRunner(ctx, func(tx pgx.Tx) error {
		var err error
		createdSale, err = s.saleRepo.CreateSaleTx(ctx, tx, sale)
		if err != nil {
			var pgErr *pgconn.PgError
			if idemKey != nil && (isUniqueViolation(err) || (errors.As(err, &pgErr) && pgErr.ConstraintName == "uq_sales_branch_idempotency_key")) {
				existing, getErr := s.saleRepo.GetByIdempotencyKey(ctx, branchID, *idemKey)
				if getErr != nil {
					return getErr
				}
				if existing != nil {
					createdSale = existing
					return nil
				}
			}
			return err
		}

		if err := s.saleRepo.CreateSaleItemsTx(ctx, tx, createdSale.ID, items); err != nil {
			return err
		}
		createdPayments, err := s.saleRepo.CreateSalePaymentsTx(ctx, tx, createdSale.ID, payments)
		if err != nil {
			return err
		}

		for _, item := range items {
			if err := s.inventoryRepo.EnsureInventoryRowTx(ctx, tx, branchID, item.ProductID); err != nil {
				return err
			}
			inv, err := s.inventoryRepo.GetForUpdateTx(ctx, tx, branchID, item.ProductID)
			if err != nil {
				if err == pgx.ErrNoRows {
					return ErrInvalidInput
				}
				return err
			}
			if inv.Quantity-inv.ReservedQuantity < item.Quantity {
				return ErrInsufficientInventory
			}
			newQty := inv.Quantity - item.Quantity
			if newQty < 0 {
				return ErrInsufficientInventory
			}
			if err := s.inventoryRepo.UpdateQuantityTx(ctx, tx, branchID, item.ProductID, newQty); err != nil {
				return err
			}
			refType := "SALE"
			refID := createdSale.ID
			if err := s.inventoryRepo.CreateStockMovementTx(ctx, tx, &models.StockMovement{ProductID: item.ProductID, BranchID: branchID, MovementType: "OUT", Quantity: item.Quantity, ReferenceType: &refType, ReferenceID: &refID, CreatedBy: &userID}); err != nil {
				return err
			}
		}

		createdSale.Items = items
		createdSale.Payments = createdPayments
		if err := s.auditService.Log(ctx, tx, "sales", createdSale.ID, "CREATE", &userID, nil, createdSale); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	logger.L().Info().Str("branch_id", branchID).Str("user_id", userID).Str("entity_id", createdSale.ID).Str("action", "sale_create").Msg("sale_create")
	return createdSale, nil
}
