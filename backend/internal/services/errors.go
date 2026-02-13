package services

import "errors"

var (
	ErrNotFound              = errors.New("resource not found")
	ErrConflict              = errors.New("resource conflict")
	ErrInvalidInput          = errors.New("invalid input")
	ErrForbiddenOperation    = errors.New("forbidden operation")
	ErrInsufficientInventory = errors.New("insufficient inventory")
	ErrInvalidPaymentTotal   = errors.New("invalid payment total")
)
