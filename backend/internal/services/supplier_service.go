package services

import (
	"context"
	"strings"

	"github.com/kcgsperera/texa-multi-pos/backend/internal/models"
	"github.com/kcgsperera/texa-multi-pos/backend/internal/repositories"
)

type SupplierService interface {
	Create(ctx context.Context, branchID, userID string, req models.CreateSupplierRequest) (*models.Supplier, error)
	GetByID(ctx context.Context, branchID, id string) (*models.Supplier, error)
	ListByBranch(ctx context.Context, branchID string, limit, offset int) ([]models.Supplier, error)
	Update(ctx context.Context, branchID, userID, id string, req models.UpdateSupplierRequest) (*models.Supplier, error)
	Delete(ctx context.Context, branchID, userID, id string) error
}

type supplierService struct {
	repo         repositories.SupplierRepository
	auditService AuditService
}

func NewSupplierService(repo repositories.SupplierRepository, auditService AuditService) SupplierService {
	return &supplierService{repo: repo, auditService: auditService}
}

func (s *supplierService) Create(ctx context.Context, branchID, userID string, req models.CreateSupplierRequest) (*models.Supplier, error) {
	out, err := s.repo.Create(ctx, &models.Supplier{BranchID: branchID, Name: strings.TrimSpace(req.Name), MobileNumber: req.MobileNumber, Email: req.Email, Address: req.Address})
	if err != nil {
		return nil, err
	}
	_ = s.auditService.Log(ctx, nil, "suppliers", out.ID, "CREATE", &userID, nil, out)
	return out, nil
}

func (s *supplierService) GetByID(ctx context.Context, branchID, id string) (*models.Supplier, error) {
	item, err := s.repo.GetByID(ctx, branchID, id)
	if err != nil {
		return nil, err
	}
	if item == nil {
		return nil, ErrNotFound
	}
	return item, nil
}

func (s *supplierService) ListByBranch(ctx context.Context, branchID string, limit, offset int) ([]models.Supplier, error) {
	limit, offset = normalizePagination(limit, offset)
	return s.repo.ListByBranch(ctx, branchID, limit, offset)
}

func (s *supplierService) Update(ctx context.Context, branchID, userID, id string, req models.UpdateSupplierRequest) (*models.Supplier, error) {
	oldItem, err := s.repo.GetByID(ctx, branchID, id)
	if err != nil {
		return nil, err
	}
	if oldItem == nil {
		return nil, ErrNotFound
	}
	item, err := s.repo.Update(ctx, &models.Supplier{ID: id, BranchID: branchID, Name: strings.TrimSpace(req.Name), MobileNumber: req.MobileNumber, Email: req.Email, Address: req.Address})
	if err != nil {
		return nil, err
	}
	if item == nil {
		return nil, ErrNotFound
	}
	_ = s.auditService.Log(ctx, nil, "suppliers", item.ID, "UPDATE", &userID, oldItem, item)
	return item, nil
}

func (s *supplierService) Delete(ctx context.Context, branchID, userID, id string) error {
	oldItem, err := s.repo.GetByID(ctx, branchID, id)
	if err != nil {
		return err
	}
	if oldItem == nil {
		return ErrNotFound
	}
	deleted, err := s.repo.Delete(ctx, branchID, id)
	if err != nil {
		return err
	}
	if !deleted {
		return ErrNotFound
	}
	_ = s.auditService.Log(ctx, nil, "suppliers", id, "DELETE", &userID, oldItem, nil)
	return nil
}
