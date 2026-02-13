package models

const (
	DefaultLimit = 20
	MaxLimit     = 100
)

type PaginationQuery struct {
	Limit  int `form:"limit"`
	Offset int `form:"offset"`
}
