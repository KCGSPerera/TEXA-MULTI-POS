package services

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/kcgsperera/texa-multi-pos/backend/internal/logger"
	"github.com/kcgsperera/texa-multi-pos/backend/internal/models"
	"github.com/kcgsperera/texa-multi-pos/backend/internal/repositories"
)

const (
	accessTokenTTL  = 15 * time.Minute
	refreshTokenTTL = 7 * 24 * time.Hour
)

var (
	ErrInvalidBranch      = errors.New("invalid branch_id")
	ErrInvalidRole        = errors.New("invalid role_id")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrInactiveUser       = errors.New("user account is inactive")
)

type AuthService interface {
	Register(ctx context.Context, req models.RegisterRequest) (*models.AuthUserResponse, error)
	Login(ctx context.Context, req models.LoginRequest) (*models.TokenResponse, error)
}

type authService struct {
	userRepo  repositories.UserRepository
	jwtSecret []byte
}

func NewAuthService(userRepo repositories.UserRepository, jwtSecret string) (AuthService, error) {
	if strings.TrimSpace(jwtSecret) == "" {
		return nil, errors.New("JWT_SECRET is required")
	}
	return &authService{userRepo: userRepo, jwtSecret: []byte(jwtSecret)}, nil
}

func (s *authService) Register(ctx context.Context, req models.RegisterRequest) (*models.AuthUserResponse, error) {
	if _, err := uuid.Parse(req.BranchID); err != nil {
		return nil, ErrInvalidBranch
	}
	if _, err := uuid.Parse(req.RoleID); err != nil {
		return nil, ErrInvalidRole
	}
	branchExists, err := s.userRepo.BranchExists(ctx, req.BranchID)
	if err != nil {
		return nil, err
	}
	if !branchExists {
		return nil, ErrInvalidBranch
	}
	roleExists, err := s.userRepo.RoleExists(ctx, req.RoleID)
	if err != nil {
		return nil, err
	}
	if !roleExists {
		return nil, ErrInvalidRole
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	user, err := s.userRepo.CreateUser(ctx, models.CreateUserInput{BranchID: req.BranchID, RoleID: req.RoleID, Name: req.Name, Email: req.Email, PasswordHash: string(hashedPassword), MobileNumber: req.MobileNumber, SecondaryMobileNumber: req.SecondaryMobileNumber, NIC: req.NIC, IsActive: true})
	if err != nil {
		return nil, err
	}
	return &models.AuthUserResponse{ID: user.ID, BranchID: user.BranchID, RoleID: user.RoleID, Name: user.Name, Email: user.Email, MobileNumber: user.MobileNumber, SecondaryMobileNumber: user.SecondaryMobileNumber, NIC: user.NIC, IsActive: user.IsActive, CreatedAt: user.CreatedAt, UpdatedAt: user.UpdatedAt}, nil
}

func (s *authService) Login(ctx context.Context, req models.LoginRequest) (*models.TokenResponse, error) {
	user, err := s.userRepo.GetUserByEmail(ctx, req.Email)
	if err != nil {
		return nil, err
	}
	if user == nil {
		logger.L().Warn().Str("user_id", "").Str("branch_id", "").Str("action", "login_attempt").Str("email", req.Email).Msg("login_failed")
		return nil, ErrInvalidCredentials
	}
	if !user.IsActive {
		logger.L().Warn().Str("user_id", user.ID).Str("branch_id", user.BranchID).Str("action", "login_attempt").Msg("login_inactive")
		return nil, ErrInactiveUser
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		logger.L().Warn().Str("user_id", user.ID).Str("branch_id", user.BranchID).Str("action", "login_attempt").Msg("login_failed")
		return nil, ErrInvalidCredentials
	}

	accessToken, err := s.signToken(user, accessTokenTTL, "access")
	if err != nil {
		return nil, err
	}
	refreshToken, err := s.signToken(user, refreshTokenTTL, "refresh")
	if err != nil {
		return nil, err
	}

	logger.L().Info().Str("user_id", user.ID).Str("branch_id", user.BranchID).Str("action", "login_attempt").Msg("login_success")
	return &models.TokenResponse{AccessToken: accessToken, RefreshToken: refreshToken, TokenType: "Bearer", ExpiresIn: int64(accessTokenTTL.Seconds())}, nil
}

func (s *authService) signToken(user *models.User, ttl time.Duration, tokenType string) (string, error) {
	now := time.Now().UTC()
	claims := models.JWTClaims{UserID: user.ID, RoleID: user.RoleID, BranchID: user.BranchID, Type: tokenType, RegisteredClaims: jwt.RegisteredClaims{IssuedAt: jwt.NewNumericDate(now), NotBefore: jwt.NewNumericDate(now), ExpiresAt: jwt.NewNumericDate(now.Add(ttl)), Subject: user.ID}}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.jwtSecret)
}
