package services

import (
	"context"

	"github.com/jackc/pgx/v5"

	"github.com/kcgsperera/texa-multi-pos/backend/internal/repositories"
	"github.com/kcgsperera/texa-multi-pos/backend/internal/requestctx"
)

type AuditService interface {
	Log(ctx context.Context, tx pgx.Tx, entityName, entityID, action string, performedBy *string, oldData, newData interface{}) error
}

type auditService struct {
	repo repositories.AuditRepository
}

func NewAuditService(repo repositories.AuditRepository) AuditService {
	return &auditService{repo: repo}
}

func (s *auditService) Log(ctx context.Context, tx pgx.Tx, entityName, entityID, action string, performedBy *string, oldData, newData interface{}) error {
	requestID := requestctx.RequestIDFromContext(ctx)
	return s.repo.Create(ctx, tx, entityName, entityID, action, performedBy, strPtrOrNil(requestID), oldData, newData)
}
