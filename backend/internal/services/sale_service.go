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
	"github.com/kcgsperera/texa-multi-pos/backend/internal/discount"
	"github.com/kcgsperera/texa-multi-pos/backend/internal/logger"
	"github.com/kcgsperera/texa-multi-pos/backend/internal/models"
	"github.com/kcgsperera/texa-multi-pos/backend/internal/observability"
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
	discountRepo  repositories.DiscountRepository
	loyaltySvc    LoyaltyService
	ledgerSvc     LedgerService
	auditService  AuditService
}

func NewSaleService(
	db *pgxpool.Pool,
	saleRepo repositories.SaleRepository,
	productRepo repositories.ProductRepository,
	inventoryRepo repositories.InventoryRepository,
	discountRepo repositories.DiscountRepository,
	loyaltySvc LoyaltyService,
	ledgerSvc LedgerService,
	auditService AuditService,
) SaleService {
	s := &saleService{db: db, saleRepo: saleRepo, productRepo: productRepo, inventoryRepo: inventoryRepo, discountRepo: discountRepo, loyaltySvc: loyaltySvc, ledgerSvc: ledgerSvc, auditService: auditService}
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

	itemLevelDiscount := 0.0
	items := make([]models.SaleItem, 0, len(req.Items))
	for _, item := range req.Items {
		exists, err := s.productRepo.ExistsActiveInBranch(ctx, branchID, item.ProductID)
		if err != nil {
			return nil, err
		}
		if !exists {
			return nil, ErrInvalidInput
		}
		itemLevelDiscount += item.ItemDiscountAmount
		items = append(items, models.SaleItem{ProductID: item.ProductID, Quantity: item.Quantity, UnitPrice: item.UnitPrice, VATAmount: item.VATAmount, LineTotal: item.LineTotal})
	}

	rules, err := s.discountRepo.GetActiveRulesByIDs(ctx, branchID, req.DiscountRuleIDs)
	if err != nil {
		return nil, err
	}

	breakdown := discount.Calculate(req.TotalAmount, itemLevelDiscount, rules, req.LoyaltyRedeemPoints)
	discountAmount := breakdown.Total
	if discountAmount > req.TotalAmount {
		discountAmount = req.TotalAmount
	}

	netAmount := (req.TotalAmount - discountAmount) + req.VATAmount
	if netAmount < 0 {
		netAmount = 0
	}

	paymentTotal := 0.0
	payments := make([]models.SalePayment, 0, len(req.Payments))
	for _, payment := range req.Payments {
		paymentTotal += payment.Amount
		payments = append(payments, models.SalePayment{PaymentMethod: strings.ToUpper(payment.PaymentMethod), Amount: payment.Amount})
	}
	if math.Abs(paymentTotal-netAmount) > 0.01 {
		return nil, ErrInvalidPaymentTotal
	}

	sale := &models.Sale{
		BranchID:         branchID,
		UserID:           userID,
		CustomerID:       req.CustomerID,
		TotalAmount:      req.TotalAmount,
		VATAmount:        req.VATAmount,
		DiscountAmount:   discountAmount,
		NetAmount:        netAmount,
		IdempotencyKey:   idemKey,
		AppliedDiscounts: breakdown.Applied,
	}

	var createdSale *models.Sale
	err = s.txRunner(ctx, func(tx pgx.Tx) error {
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
		if err := s.discountRepo.CreateSaleDiscountsTx(ctx, tx, createdSale.ID, breakdown.Applied); err != nil {
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

		if createdSale.CustomerID != nil && req.LoyaltyRedeemPoints > 0 {
			if err := s.loyaltySvc.RedeemTx(ctx, tx, *createdSale.CustomerID, branchID, createdSale.ID, req.LoyaltyRedeemPoints); err != nil {
				return err
			}
		}
		if createdSale.CustomerID != nil {
			if err := s.loyaltySvc.EarnForSaleTx(ctx, tx, *createdSale.CustomerID, branchID, createdSale.ID, createdSale.NetAmount); err != nil {
				return err
			}
		}

		revenueAmount := createdSale.TotalAmount - createdSale.DiscountAmount
		if revenueAmount < 0 {
			revenueAmount = 0
		}
		if err := s.ledgerSvc.PostByCodes(ctx, tx, branchID, "SALE", createdSale.ID, []LedgerLineByCode{
			{AccountCode: "CASH", Debit: createdSale.NetAmount, Credit: 0},
			{AccountCode: "REVENUE", Debit: 0, Credit: revenueAmount},
			{AccountCode: "VAT_PAYABLE", Debit: 0, Credit: createdSale.VATAmount},
		}, &userID); err != nil {
			return err
		}

		createdSale.Items = items
		createdSale.Payments = createdPayments
		createdSale.AppliedDiscounts = breakdown.Applied

		if err := s.auditService.Log(ctx, tx, "sales", createdSale.ID, "CREATE", &userID, nil, createdSale); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	observability.IncSaleCreate()
	logger.L().Info().Str("branch_id", branchID).Str("user_id", userID).Str("entity_id", createdSale.ID).Str("action", "sale_create").Interface("discount_breakdown", createdSale.AppliedDiscounts).Msg("sale_create")
	return createdSale, nil
}
