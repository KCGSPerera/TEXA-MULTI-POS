package services

import "github.com/kcgsperera/texa-multi-pos/backend/internal/models"

func normalizePagination(limit, offset int) (int, int) {
	if limit <= 0 {
		limit = models.DefaultLimit
	}
	if limit > models.MaxLimit {
		limit = models.MaxLimit
	}
	if offset < 0 {
		offset = 0
	}
	return limit, offset
}
