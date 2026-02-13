package repositories

import (
	"context"
	"errors"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type RoleRepository interface {
	GetRoleNameByID(ctx context.Context, roleID string) (string, error)
}

type roleRepository struct {
	db *pgxpool.Pool
}

func NewRoleRepository(db *pgxpool.Pool) RoleRepository {
	return &roleRepository{db: db}
}

func (r *roleRepository) GetRoleNameByID(ctx context.Context, roleID string) (string, error) {
	const query = `SELECT name FROM roles WHERE id = $1`

	var roleName string
	err := r.db.QueryRow(ctx, query, roleID).Scan(&roleName)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", nil
		}
		return "", err
	}

	return strings.TrimSpace(roleName), nil
}
