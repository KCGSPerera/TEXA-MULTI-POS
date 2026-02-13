package utils

import (
	"errors"
	"net/http"

	"github.com/kcgsperera/texa-multi-pos/backend/internal/services"
)

func MapServiceError(err error) (int, string) {
	switch {
	case errors.Is(err, services.ErrNotFound):
		return http.StatusNotFound, err.Error()
	case errors.Is(err, services.ErrInvalidInput):
		return http.StatusBadRequest, err.Error()
	case errors.Is(err, services.ErrConflict):
		return http.StatusConflict, err.Error()
	case errors.Is(err, services.ErrInsufficientInventory):
		return http.StatusConflict, err.Error()
	case errors.Is(err, services.ErrForbiddenOperation):
		return http.StatusForbidden, err.Error()
	case errors.Is(err, services.ErrLedgerImbalance):
		return http.StatusBadRequest, err.Error()
	default:
		return http.StatusInternalServerError, "internal server error"
	}
}
