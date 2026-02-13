package services

import (
	"context"
	"strings"

	"github.com/kcgsperera/texa-multi-pos/backend/internal/models"
	"github.com/kcgsperera/texa-multi-pos/backend/internal/repositories"
)

type ProductService interface {
	Create(ctx context.Context, branchID, userID string, req models.CreateProductRequest) (*models.Product, error)
	GetByID(ctx context.Context, branchID, id string) (*models.Product, error)
	ListByBranch(ctx context.Context, branchID string, limit, offset int) ([]models.Product, error)
	Update(ctx context.Context, branchID, userID, id string, req models.UpdateProductRequest) (*models.Product, error)
	Delete(ctx context.Context, branchID, userID, id string) error
}

type productService struct {
	productRepo  repositories.ProductRepository
	categoryRepo repositories.CategoryRepository
	auditService AuditService
}

func NewProductService(productRepo repositories.ProductRepository, categoryRepo repositories.CategoryRepository, auditService AuditService) ProductService {
	return &productService{productRepo: productRepo, categoryRepo: categoryRepo, auditService: auditService}
}

func (s *productService) Create(ctx context.Context, branchID, userID string, req models.CreateProductRequest) (*models.Product, error) {
	category, err := s.categoryRepo.GetByID(ctx, branchID, req.CategoryID)
	if err != nil {
		return nil, err
	}
	if category == nil {
		return nil, ErrInvalidInput
	}
	vat := 18.0
	if req.VATPercentage != nil {
		vat = *req.VATPercentage
	}
	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}
	product := &models.Product{BranchID: branchID, CategoryID: req.CategoryID, Name: strings.TrimSpace(req.Name), SKU: strings.TrimSpace(req.SKU), Barcode: req.Barcode, CostPrice: req.CostPrice, SellingPrice: req.SellingPrice, VATPercentage: vat, IsVATInclusive: req.IsVATInclusive, IsActive: isActive, CreatedBy: &userID, UpdatedBy: &userID}
	out, err := s.productRepo.Create(ctx, product)
	if err != nil {
		if isUniqueViolation(err) {
			return nil, ErrConflict
		}
		return nil, err
	}
	_ = s.auditService.Log(ctx, nil, "products", out.ID, "CREATE", &userID, nil, out)
	return out, nil
}

func (s *productService) GetByID(ctx context.Context, branchID, id string) (*models.Product, error) {
	item, err := s.productRepo.GetByID(ctx, branchID, id)
	if err != nil {
		return nil, err
	}
	if item == nil {
		return nil, ErrNotFound
	}
	return item, nil
}

func (s *productService) ListByBranch(ctx context.Context, branchID string, limit, offset int) ([]models.Product, error) {
	limit, offset = normalizePagination(limit, offset)
	return s.productRepo.ListByBranch(ctx, branchID, limit, offset)
}

func (s *productService) Update(ctx context.Context, branchID, userID, id string, req models.UpdateProductRequest) (*models.Product, error) {
	oldItem, err := s.productRepo.GetByID(ctx, branchID, id)
	if err != nil {
		return nil, err
	}
	if oldItem == nil {
		return nil, ErrNotFound
	}
	category, err := s.categoryRepo.GetByID(ctx, branchID, req.CategoryID)
	if err != nil {
		return nil, err
	}
	if category == nil {
		return nil, ErrInvalidInput
	}
	vat := 18.0
	if req.VATPercentage != nil {
		vat = *req.VATPercentage
	}
	item := &models.Product{ID: id, BranchID: branchID, CategoryID: req.CategoryID, Name: strings.TrimSpace(req.Name), SKU: strings.TrimSpace(req.SKU), Barcode: req.Barcode, CostPrice: req.CostPrice, SellingPrice: req.SellingPrice, VATPercentage: vat, IsVATInclusive: req.IsVATInclusive, IsActive: req.IsActive, UpdatedBy: &userID}
	out, err := s.productRepo.Update(ctx, item)
	if err != nil {
		if isUniqueViolation(err) {
			return nil, ErrConflict
		}
		return nil, err
	}
	if out == nil {
		return nil, ErrNotFound
	}
	_ = s.auditService.Log(ctx, nil, "products", out.ID, "UPDATE", &userID, oldItem, out)
	return out, nil
}

func (s *productService) Delete(ctx context.Context, branchID, userID, id string) error {
	oldItem, err := s.productRepo.GetByID(ctx, branchID, id)
	if err != nil {
		return err
	}
	if oldItem == nil {
		return ErrNotFound
	}
	deleted, err := s.productRepo.Delete(ctx, branchID, id, userID)
	if err != nil {
		return err
	}
	if !deleted {
		return ErrNotFound
	}
	_ = s.auditService.Log(ctx, nil, "products", id, "DELETE", &userID, oldItem, nil)
	return nil
}
