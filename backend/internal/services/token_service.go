package services

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"os"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/kcgsperera/texa-multi-pos/backend/internal/database"
	"github.com/kcgsperera/texa-multi-pos/backend/internal/models"
	"github.com/kcgsperera/texa-multi-pos/backend/internal/repositories"
)

var ErrInvalidRefreshToken = errors.New("invalid refresh token")

type TokenService interface {
	IssueRefreshToken(ctx context.Context, userID string, meta models.TokenMeta) (string, error)
	RotateRefreshToken(ctx context.Context, currentToken string, meta models.TokenMeta) (string, string, error)
	RevokeRefreshToken(ctx context.Context, token string) error
}

type tokenService struct {
	db       *pgxpool.Pool
	txRunner func(context.Context, func(pgx.Tx) error) error
	tokenTTL time.Duration
	repo     repositories.TokenRepository
}

func NewTokenService(db *pgxpool.Pool, repo repositories.TokenRepository) TokenService {
	ttlHours := parseEnvInt("REFRESH_TOKEN_TTL_HOURS", 720)
	s := &tokenService{db: db, tokenTTL: time.Duration(ttlHours) * time.Hour, repo: repo}
	s.txRunner = func(ctx context.Context, fn func(pgx.Tx) error) error {
		return database.WithTx(ctx, db, fn)
	}
	return s
}

func (s *tokenService) IssueRefreshToken(ctx context.Context, userID string, meta models.TokenMeta) (string, error) {
	plainToken, hash, err := generateRefreshToken()
	if err != nil {
		return "", err
	}

	input := repositories.CreateRefreshTokenInput{
		UserID:    userID,
		TokenHash: hash,
		ExpiresAt: time.Now().UTC().Add(s.tokenTTL),
		UserAgent: strPtrOrNil(meta.UserAgent),
		IPAddress: strPtrOrNil(meta.IPAddress),
	}
	if _, err := s.repo.Create(ctx, input); err != nil {
		return "", err
	}

	return plainToken, nil
}

func (s *tokenService) RotateRefreshToken(ctx context.Context, currentToken string, meta models.TokenMeta) (string, string, error) {
	if currentToken == "" {
		return "", "", ErrInvalidRefreshToken
	}

	hash := hashToken(currentToken)
	var userID string
	var nextPlainToken string

	err := s.txRunner(ctx, func(tx pgx.Tx) error {
		existing, err := s.repo.GetActiveByHashForUpdateTx(ctx, tx, hash)
		if err != nil {
			return err
		}
		if existing == nil {
			return ErrInvalidRefreshToken
		}
		userID = existing.UserID

		var nextHash string
		nextPlainToken, nextHash, err = generateRefreshToken()
		if err != nil {
			return err
		}

		created, err := s.repo.CreateTx(ctx, tx, repositories.CreateRefreshTokenInput{
			UserID:    existing.UserID,
			TokenHash: nextHash,
			ExpiresAt: time.Now().UTC().Add(s.tokenTTL),
			UserAgent: strPtrOrNil(meta.UserAgent),
			IPAddress: strPtrOrNil(meta.IPAddress),
		})
		if err != nil {
			return err
		}

		return s.repo.RevokeAndReplaceTx(ctx, tx, existing.ID, created.ID)
	})
	if err != nil {
		return "", "", err
	}

	return nextPlainToken, userID, nil
}

func (s *tokenService) RevokeRefreshToken(ctx context.Context, token string) error {
	if token == "" {
		return ErrInvalidRefreshToken
	}
	return s.repo.RevokeByHash(ctx, hashToken(token))
}

func generateRefreshToken() (plain string, hashed string, err error) {
	buf := make([]byte, 32)
	if _, err = rand.Read(buf); err != nil {
		return "", "", err
	}
	plain = base64.RawURLEncoding.EncodeToString(buf)
	return plain, hashToken(plain), nil
}

func hashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

func parseEnvInt(key string, fallback int) int {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil || parsed <= 0 {
		return fallback
	}
	return parsed
}

func strPtrOrNil(value string) *string {
	if value == "" {
		return nil
	}
	v := value
	return &v
}
