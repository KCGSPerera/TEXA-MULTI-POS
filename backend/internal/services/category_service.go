package services

import (
	"context"
	"strings"

	"github.com/kcgsperera/texa-multi-pos/backend/internal/models"
	"github.com/kcgsperera/texa-multi-pos/backend/internal/repositories"
)

type CategoryService interface {
	Create(ctx context.Context, branchID, userID string, req models.CreateCategoryRequest) (*models.Category, error)
	GetByID(ctx context.Context, branchID, id string) (*models.Category, error)
	ListByBranch(ctx context.Context, branchID string, limit, offset int) ([]models.Category, error)
	Update(ctx context.Context, branchID, userID, id string, req models.UpdateCategoryRequest) (*models.Category, error)
	Delete(ctx context.Context, branchID, userID, id string) error
}

type categoryService struct {
	repo         repositories.CategoryRepository
	auditService AuditService
}

func NewCategoryService(repo repositories.CategoryRepository, auditService AuditService) CategoryService {
	return &categoryService{repo: repo, auditService: auditService}
}

func (s *categoryService) Create(ctx context.Context, branchID, userID string, req models.CreateCategoryRequest) (*models.Category, error) {
	category := &models.Category{BranchID: branchID, Name: strings.TrimSpace(req.Name), Description: req.Description, CreatedBy: &userID, UpdatedBy: &userID}
	out, err := s.repo.Create(ctx, category)
	if err != nil {
		if isUniqueViolation(err) {
			return nil, ErrConflict
		}
		return nil, err
	}
	_ = s.auditService.Log(ctx, nil, "categories", out.ID, "CREATE", &userID, nil, out)
	return out, nil
}

func (s *categoryService) GetByID(ctx context.Context, branchID, id string) (*models.Category, error) {
	item, err := s.repo.GetByID(ctx, branchID, id)
	if err != nil {
		return nil, err
	}
	if item == nil {
		return nil, ErrNotFound
	}
	return item, nil
}

func (s *categoryService) ListByBranch(ctx context.Context, branchID string, limit, offset int) ([]models.Category, error) {
	limit, offset = normalizePagination(limit, offset)
	return s.repo.ListByBranch(ctx, branchID, limit, offset)
}

func (s *categoryService) Update(ctx context.Context, branchID, userID, id string, req models.UpdateCategoryRequest) (*models.Category, error) {
	oldItem, err := s.repo.GetByID(ctx, branchID, id)
	if err != nil {
		return nil, err
	}
	if oldItem == nil {
		return nil, ErrNotFound
	}
	item := &models.Category{ID: id, BranchID: branchID, Name: strings.TrimSpace(req.Name), Description: req.Description, UpdatedBy: &userID}
	out, err := s.repo.Update(ctx, item)
	if err != nil {
		if isUniqueViolation(err) {
			return nil, ErrConflict
		}
		return nil, err
	}
	if out == nil {
		return nil, ErrNotFound
	}
	_ = s.auditService.Log(ctx, nil, "categories", out.ID, "UPDATE", &userID, oldItem, out)
	return out, nil
}

func (s *categoryService) Delete(ctx context.Context, branchID, userID, id string) error {
	oldItem, err := s.repo.GetByID(ctx, branchID, id)
	if err != nil {
		return err
	}
	if oldItem == nil {
		return ErrNotFound
	}
	deleted, err := s.repo.Delete(ctx, branchID, id, userID)
	if err != nil {
		return err
	}
	if !deleted {
		return ErrNotFound
	}
	_ = s.auditService.Log(ctx, nil, "categories", id, "DELETE", &userID, oldItem, nil)
	return nil
}
