package services

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/kcgsperera/texa-multi-pos/backend/internal/models"
	"github.com/kcgsperera/texa-multi-pos/backend/internal/repositories"
)

type tokenRepoStub struct {
	mu     sync.Mutex
	byHash map[string]*models.RefreshToken
}

func newTokenRepoStub() *tokenRepoStub {
	return &tokenRepoStub{byHash: map[string]*models.RefreshToken{}}
}

func (s *tokenRepoStub) CreateTx(ctx context.Context, tx pgx.Tx, input repositories.CreateRefreshTokenInput) (*models.RefreshToken, error) {
	row := &models.RefreshToken{ID: "new-token", UserID: input.UserID, TokenHash: input.TokenHash, ExpiresAt: input.ExpiresAt}
	s.mu.Lock()
	s.byHash[input.TokenHash] = row
	s.mu.Unlock()
	return row, nil
}

func (s *tokenRepoStub) Create(ctx context.Context, input repositories.CreateRefreshTokenInput) (*models.RefreshToken, error) {
	row := &models.RefreshToken{ID: "initial-token", UserID: input.UserID, TokenHash: input.TokenHash, ExpiresAt: input.ExpiresAt}
	s.mu.Lock()
	s.byHash[input.TokenHash] = row
	s.mu.Unlock()
	return row, nil
}

func (s *tokenRepoStub) GetActiveByHashForUpdateTx(ctx context.Context, tx pgx.Tx, tokenHash string) (*models.RefreshToken, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	row, ok := s.byHash[tokenHash]
	if !ok {
		return nil, nil
	}
	if row.RevokedAt != nil || row.ExpiresAt.Before(time.Now().UTC()) {
		return nil, nil
	}
	return row, nil
}

func (s *tokenRepoStub) RevokeAndReplaceTx(ctx context.Context, tx pgx.Tx, tokenID, replacedBy string) error {
	now := time.Now().UTC()
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, row := range s.byHash {
		if row.ID == tokenID {
			row.RevokedAt = &now
			row.ReplacedBy = &replacedBy
			return nil
		}
	}
	return nil
}

func (s *tokenRepoStub) RevokeByHash(ctx context.Context, tokenHash string) error {
	now := time.Now().UTC()
	s.mu.Lock()
	defer s.mu.Unlock()
	if row, ok := s.byHash[tokenHash]; ok {
		row.RevokedAt = &now
	}
	return nil
}

func TestTokenServiceRotateRefreshToken(t *testing.T) {
	tests := []struct {
		name        string
		seedToken   string
		seedUserID  string
		lookupToken string
		wantErr     error
	}{
		{name: "invalid token", lookupToken: "missing", wantErr: ErrInvalidRefreshToken},
		{name: "rotate success", seedToken: "seed-token", seedUserID: "u1", lookupToken: "seed-token", wantErr: nil},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			repo := newTokenRepoStub()
			svc := &tokenService{repo: repo, tokenTTL: 24 * time.Hour}
			svc.txRunner = func(ctx context.Context, fn func(pgx.Tx) error) error {
				var tx pgx.Tx
				return fn(tx)
			}

			if tc.seedToken != "" {
				h := hashToken(tc.seedToken)
				expires := time.Now().UTC().Add(time.Hour)
				repo.byHash[h] = &models.RefreshToken{ID: "old-token", UserID: tc.seedUserID, TokenHash: h, ExpiresAt: expires}
			}

			next, userID, err := svc.RotateRefreshToken(context.Background(), tc.lookupToken, models.TokenMeta{})
			if tc.wantErr != nil {
				if err == nil || err != tc.wantErr {
					t.Fatalf("expected error %v, got %v", tc.wantErr, err)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if next == "" {
				t.Fatal("expected new refresh token")
			}
			if userID != tc.seedUserID {
				t.Fatalf("expected userID %s got %s", tc.seedUserID, userID)
			}
		})
	}
}
