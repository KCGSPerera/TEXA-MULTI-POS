package models

import "github.com/golang-jwt/jwt/v5"

type JWTClaims struct {
	UserID   string `json:"user_id"`
	RoleID   string `json:"role_id"`
	BranchID string `json:"branch_id"`
	Type     string `json:"type"`
	jwt.RegisteredClaims
}
