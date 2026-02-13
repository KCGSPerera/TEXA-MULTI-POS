package repositories

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/kcgsperera/texa-multi-pos/backend/internal/models"
)

type CreateRefreshTokenInput struct {
	UserID    string
	TokenHash string
	ExpiresAt time.Time
	UserAgent *string
	IPAddress *string
}

type TokenRepository interface {
	CreateTx(ctx context.Context, tx pgx.Tx, input CreateRefreshTokenInput) (*models.RefreshToken, error)
	Create(ctx context.Context, input CreateRefreshTokenInput) (*models.RefreshToken, error)
	GetActiveByHashForUpdateTx(ctx context.Context, tx pgx.Tx, tokenHash string) (*models.RefreshToken, error)
	RevokeAndReplaceTx(ctx context.Context, tx pgx.Tx, tokenID, replacedBy string) error
	RevokeByHash(ctx context.Context, tokenHash string) error
}

type tokenRepository struct {
	db *pgxpool.Pool
}

func NewTokenRepository(db *pgxpool.Pool) TokenRepository {
	return &tokenRepository{db: db}
}

func (r *tokenRepository) CreateTx(ctx context.Context, tx pgx.Tx, input CreateRefreshTokenInput) (*models.RefreshToken, error) {
	const query = `
		INSERT INTO refresh_tokens (user_id, token_hash, expires_at, user_agent, ip_address)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, user_id, token_hash, issued_at, expires_at, revoked_at, replaced_by, user_agent, ip_address
	`

	row := &models.RefreshToken{}
	if err := tx.QueryRow(ctx, query, input.UserID, input.TokenHash, input.ExpiresAt, input.UserAgent, input.IPAddress).Scan(
		&row.ID,
		&row.UserID,
		&row.TokenHash,
		&row.IssuedAt,
		&row.ExpiresAt,
		&row.RevokedAt,
		&row.ReplacedBy,
		&row.UserAgent,
		&row.IPAddress,
	); err != nil {
		return nil, err
	}
	return row, nil
}

func (r *tokenRepository) Create(ctx context.Context, input CreateRefreshTokenInput) (*models.RefreshToken, error) {
	const query = `
		INSERT INTO refresh_tokens (user_id, token_hash, expires_at, user_agent, ip_address)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, user_id, token_hash, issued_at, expires_at, revoked_at, replaced_by, user_agent, ip_address
	`

	row := &models.RefreshToken{}
	if err := r.db.QueryRow(ctx, query, input.UserID, input.TokenHash, input.ExpiresAt, input.UserAgent, input.IPAddress).Scan(
		&row.ID,
		&row.UserID,
		&row.TokenHash,
		&row.IssuedAt,
		&row.ExpiresAt,
		&row.RevokedAt,
		&row.ReplacedBy,
		&row.UserAgent,
		&row.IPAddress,
	); err != nil {
		return nil, err
	}
	return row, nil
}

func (r *tokenRepository) GetActiveByHashForUpdateTx(ctx context.Context, tx pgx.Tx, tokenHash string) (*models.RefreshToken, error) {
	const query = `
		SELECT id, user_id, token_hash, issued_at, expires_at, revoked_at, replaced_by, user_agent, ip_address
		FROM refresh_tokens
		WHERE token_hash = $1 AND revoked_at IS NULL AND expires_at > now()
		FOR UPDATE
	`

	row := &models.RefreshToken{}
	if err := tx.QueryRow(ctx, query, tokenHash).Scan(
		&row.ID,
		&row.UserID,
		&row.TokenHash,
		&row.IssuedAt,
		&row.ExpiresAt,
		&row.RevokedAt,
		&row.ReplacedBy,
		&row.UserAgent,
		&row.IPAddress,
	); err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return row, nil
}

func (r *tokenRepository) RevokeAndReplaceTx(ctx context.Context, tx pgx.Tx, tokenID, replacedBy string) error {
	const query = `
		UPDATE refresh_tokens
		SET revoked_at = now(), replaced_by = $2
		WHERE id = $1 AND revoked_at IS NULL
	`
	_, err := tx.Exec(ctx, query, tokenID, replacedBy)
	return err
}

func (r *tokenRepository) RevokeByHash(ctx context.Context, tokenHash string) error {
	const query = `
		UPDATE refresh_tokens
		SET revoked_at = now()
		WHERE token_hash = $1 AND revoked_at IS NULL
	`
	_, err := r.db.Exec(ctx, query, tokenHash)
	return err
}
