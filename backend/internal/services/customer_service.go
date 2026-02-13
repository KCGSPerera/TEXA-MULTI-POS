package services

import (
	"context"
	"strings"

	"github.com/kcgsperera/texa-multi-pos/backend/internal/models"
	"github.com/kcgsperera/texa-multi-pos/backend/internal/repositories"
)

type CustomerService interface {
	Create(ctx context.Context, branchID, userID string, req models.CreateCustomerRequest) (*models.Customer, error)
	GetByID(ctx context.Context, branchID, id string) (*models.Customer, error)
	ListByBranch(ctx context.Context, branchID string, limit, offset int) ([]models.Customer, error)
	Update(ctx context.Context, branchID, userID, id string, req models.UpdateCustomerRequest) (*models.Customer, error)
	Delete(ctx context.Context, branchID, userID, id string) error
}

type customerService struct {
	repo         repositories.CustomerRepository
	auditService AuditService
}

func NewCustomerService(repo repositories.CustomerRepository, auditService AuditService) CustomerService {
	return &customerService{repo: repo, auditService: auditService}
}

func (s *customerService) Create(ctx context.Context, branchID, userID string, req models.CreateCustomerRequest) (*models.Customer, error) {
	loyalty := 0.0
	if req.LoyaltyPoints != nil {
		loyalty = *req.LoyaltyPoints
	}
	out, err := s.repo.Create(ctx, &models.Customer{BranchID: branchID, Name: strings.TrimSpace(req.Name), MobileNumber: req.MobileNumber, Email: req.Email, Address: req.Address, LoyaltyPoints: loyalty})
	if err != nil {
		return nil, err
	}
	_ = s.auditService.Log(ctx, nil, "customers", out.ID, "CREATE", &userID, nil, out)
	return out, nil
}

func (s *customerService) GetByID(ctx context.Context, branchID, id string) (*models.Customer, error) {
	item, err := s.repo.GetByID(ctx, branchID, id)
	if err != nil {
		return nil, err
	}
	if item == nil {
		return nil, ErrNotFound
	}
	return item, nil
}

func (s *customerService) ListByBranch(ctx context.Context, branchID string, limit, offset int) ([]models.Customer, error) {
	limit, offset = normalizePagination(limit, offset)
	return s.repo.ListByBranch(ctx, branchID, limit, offset)
}

func (s *customerService) Update(ctx context.Context, branchID, userID, id string, req models.UpdateCustomerRequest) (*models.Customer, error) {
	oldItem, err := s.repo.GetByID(ctx, branchID, id)
	if err != nil {
		return nil, err
	}
	if oldItem == nil {
		return nil, ErrNotFound
	}
	item, err := s.repo.Update(ctx, &models.Customer{ID: id, BranchID: branchID, Name: strings.TrimSpace(req.Name), MobileNumber: req.MobileNumber, Email: req.Email, Address: req.Address, LoyaltyPoints: req.LoyaltyPoints})
	if err != nil {
		return nil, err
	}
	if item == nil {
		return nil, ErrNotFound
	}
	_ = s.auditService.Log(ctx, nil, "customers", item.ID, "UPDATE", &userID, oldItem, item)
	return item, nil
}

func (s *customerService) Delete(ctx context.Context, branchID, userID, id string) error {
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
	_ = s.auditService.Log(ctx, nil, "customers", id, "DELETE", &userID, oldItem, nil)
	return nil
}
