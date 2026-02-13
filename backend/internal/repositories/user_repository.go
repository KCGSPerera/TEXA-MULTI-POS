package repositories

import (
	"context"
	"errors"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/kcgsperera/texa-multi-pos/backend/internal/models"
)

type UserRepository interface {
	BranchExists(ctx context.Context, branchID string) (bool, error)
	RoleExists(ctx context.Context, roleID string) (bool, error)
	GetUserByEmail(ctx context.Context, email string) (*models.User, error)
	GetUserByID(ctx context.Context, userID string) (*models.User, error)
	CreateUser(ctx context.Context, input models.CreateUserInput) (*models.User, error)
}

type userRepository struct {
	db *pgxpool.Pool
}

func NewUserRepository(db *pgxpool.Pool) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) BranchExists(ctx context.Context, branchID string) (bool, error) {
	const query = `SELECT EXISTS(SELECT 1 FROM branches WHERE id = $1)`

	var exists bool
	if err := r.db.QueryRow(ctx, query, branchID).Scan(&exists); err != nil {
		return false, err
	}

	return exists, nil
}

func (r *userRepository) RoleExists(ctx context.Context, roleID string) (bool, error) {
	const query = `SELECT EXISTS(SELECT 1 FROM roles WHERE id = $1)`

	var exists bool
	if err := r.db.QueryRow(ctx, query, roleID).Scan(&exists); err != nil {
		return false, err
	}

	return exists, nil
}

func (r *userRepository) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	const query = `
		SELECT id, branch_id, role_id, name, email, password_hash, mobile_number,
		       secondary_mobile_number, nic, is_active, created_at, updated_at,
		       created_by, updated_by
		FROM users
		WHERE email = $1
	`
	return r.getByQuery(ctx, query, strings.ToLower(strings.TrimSpace(email)))
}

func (r *userRepository) GetUserByID(ctx context.Context, userID string) (*models.User, error) {
	const query = `
		SELECT id, branch_id, role_id, name, email, password_hash, mobile_number,
		       secondary_mobile_number, nic, is_active, created_at, updated_at,
		       created_by, updated_by
		FROM users
		WHERE id = $1
	`
	return r.getByQuery(ctx, query, userID)
}

func (r *userRepository) getByQuery(ctx context.Context, query string, arg string) (*models.User, error) {
	user := &models.User{}
	err := r.db.QueryRow(ctx, query, arg).Scan(
		&user.ID,
		&user.BranchID,
		&user.RoleID,
		&user.Name,
		&user.Email,
		&user.PasswordHash,
		&user.MobileNumber,
		&user.SecondaryMobileNumber,
		&user.NIC,
		&user.IsActive,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.CreatedBy,
		&user.UpdatedBy,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return user, nil
}

func (r *userRepository) CreateUser(ctx context.Context, input models.CreateUserInput) (*models.User, error) {
	const query = `
		INSERT INTO users (
			branch_id,
			role_id,
			name,
			email,
			password_hash,
			mobile_number,
			secondary_mobile_number,
			nic,
			is_active,
			created_by,
			updated_by
		)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)
		RETURNING id, branch_id, role_id, name, email, password_hash, mobile_number,
		          secondary_mobile_number, nic, is_active, created_at, updated_at,
		          created_by, updated_by
	`

	user := &models.User{}
	err := r.db.QueryRow(
		ctx,
		query,
		input.BranchID,
		input.RoleID,
		strings.TrimSpace(input.Name),
		strings.ToLower(strings.TrimSpace(input.Email)),
		input.PasswordHash,
		strings.TrimSpace(input.MobileNumber),
		input.SecondaryMobileNumber,
		strings.TrimSpace(input.NIC),
		input.IsActive,
		input.CreatedBy,
		input.UpdatedBy,
	).Scan(
		&user.ID,
		&user.BranchID,
		&user.RoleID,
		&user.Name,
		&user.Email,
		&user.PasswordHash,
		&user.MobileNumber,
		&user.SecondaryMobileNumber,
		&user.NIC,
		&user.IsActive,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.CreatedBy,
		&user.UpdatedBy,
	)
	if err != nil {
		return nil, err
	}

	return user, nil
}
