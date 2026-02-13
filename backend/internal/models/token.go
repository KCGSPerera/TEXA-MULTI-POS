package models

import "time"

type RefreshToken struct {
	ID         string     `json:"id"`
	UserID     string     `json:"user_id"`
	TokenHash  string     `json:"-"`
	IssuedAt   time.Time  `json:"issued_at"`
	ExpiresAt  time.Time  `json:"expires_at"`
	RevokedAt  *time.Time `json:"revoked_at,omitempty"`
	ReplacedBy *string    `json:"replaced_by,omitempty"`
	UserAgent  *string    `json:"user_agent,omitempty"`
	IPAddress  *string    `json:"ip_address,omitempty"`
}

type TokenMeta struct {
	UserAgent string
	IPAddress string
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required,min=20"`
}

type LogoutRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required,min=20"`
}
