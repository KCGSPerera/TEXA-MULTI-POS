package services

import (
	"context"
	"time"

	"github.com/kcgsperera/texa-multi-pos/backend/internal/models"
	"github.com/kcgsperera/texa-multi-pos/backend/internal/repositories"
)

type ReportService interface {
	DailySalesSummary(ctx context.Context, branchID string, from, to time.Time) (*models.DailySalesSummary, error)
	ProductSalesReport(ctx context.Context, branchID string, from, to time.Time) ([]models.ProductSalesRow, error)
	PaymentMethodSummary(ctx context.Context, branchID string, from, to time.Time) ([]models.PaymentMethodSummaryRow, error)
	StockValuationReport(ctx context.Context, branchID string) ([]models.StockValuationRow, error)
}

type reportService struct {
	repo repositories.ReportRepository
}

func NewReportService(repo repositories.ReportRepository) ReportService {
	return &reportService{repo: repo}
}

func (s *reportService) DailySalesSummary(ctx context.Context, branchID string, from, to time.Time) (*models.DailySalesSummary, error) {
	return s.repo.DailySalesSummary(ctx, branchID, from, to)
}

func (s *reportService) ProductSalesReport(ctx context.Context, branchID string, from, to time.Time) ([]models.ProductSalesRow, error) {
	return s.repo.ProductSalesReport(ctx, branchID, from, to)
}

func (s *reportService) PaymentMethodSummary(ctx context.Context, branchID string, from, to time.Time) ([]models.PaymentMethodSummaryRow, error) {
	return s.repo.PaymentMethodSummary(ctx, branchID, from, to)
}

func (s *reportService) StockValuationReport(ctx context.Context, branchID string) ([]models.StockValuationRow, error) {
	return s.repo.StockValuationReport(ctx, branchID)
}
