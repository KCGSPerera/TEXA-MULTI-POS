package repositories

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/kcgsperera/texa-multi-pos/backend/internal/cache"
)

type RoleRepository interface {
	GetRoleNameByID(ctx context.Context, roleID string) (string, error)
	HasPermission(ctx context.Context, roleID, permissionCode string) (bool, error)
}

type roleRepository struct {
	db    *pgxpool.Pool
	cache cache.Client
}

func NewRoleRepository(db *pgxpool.Pool, cacheClient cache.Client) RoleRepository {
	return &roleRepository{db: db, cache: cacheClient}
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

func (r *roleRepository) HasPermission(ctx context.Context, roleID, permissionCode string) (bool, error) {
	if r.cache != nil && r.cache.Enabled() {
		key := fmt.Sprintf("perm:%s:%s", roleID, permissionCode)
		cached, ok, err := r.cache.Get(ctx, key)
		if err == nil && ok {
			return cached == "1", nil
		}
	}

	const query = `
		SELECT EXISTS(
			SELECT 1
			FROM role_permissions rp
			JOIN permissions p ON p.id = rp.permission_id
			WHERE rp.role_id = $1 AND p.code = $2
		)
	`
	var allowed bool
	if err := r.db.QueryRow(ctx, query, roleID, permissionCode).Scan(&allowed); err != nil {
		return false, err
	}

	if r.cache != nil && r.cache.Enabled() {
		value := "0"
		if allowed {
			value = "1"
		}
		_ = r.cache.Set(ctx, fmt.Sprintf("perm:%s:%s", roleID, permissionCode), value, 5*time.Minute)
	}

	return allowed, nil
}
