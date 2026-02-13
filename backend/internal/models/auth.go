package models

import "time"

type User struct {
	ID                    string    `json:"id"`
	BranchID              string    `json:"branch_id"`
	RoleID                string    `json:"role_id"`
	Name                  string    `json:"name"`
	Email                 string    `json:"email"`
	PasswordHash          string    `json:"-"`
	MobileNumber          string    `json:"mobile_number"`
	SecondaryMobileNumber *string   `json:"secondary_mobile_number,omitempty"`
	NIC                   string    `json:"nic"`
	IsActive              bool      `json:"is_active"`
	CreatedAt             time.Time `json:"created_at"`
	UpdatedAt             time.Time `json:"updated_at"`
	CreatedBy             *string   `json:"created_by,omitempty"`
	UpdatedBy             *string   `json:"updated_by,omitempty"`
}

type CreateUserInput struct {
	BranchID              string
	RoleID                string
	Name                  string
	Email                 string
	PasswordHash          string
	MobileNumber          string
	SecondaryMobileNumber *string
	NIC                   string
	IsActive              bool
	CreatedBy             *string
	UpdatedBy             *string
}

type RegisterRequest struct {
	BranchID              string  `json:"branch_id" binding:"required,uuid"`
	RoleID                string  `json:"role_id" binding:"required,uuid"`
	Name                  string  `json:"name" binding:"required,min=2,max=150"`
	Email                 string  `json:"email" binding:"required,email,max=255"`
	Password              string  `json:"password" binding:"required,min=8,max=72"`
	MobileNumber          string  `json:"mobile_number" binding:"required,min=7,max=20"`
	SecondaryMobileNumber *string `json:"secondary_mobile_number" binding:"omitempty,min=7,max=20"`
	NIC                   string  `json:"nic" binding:"required,min=5,max=20"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email,max=255"`
	Password string `json:"password" binding:"required,min=8,max=72"`
}

type AuthUserResponse struct {
	ID                    string    `json:"id"`
	BranchID              string    `json:"branch_id"`
	RoleID                string    `json:"role_id"`
	Name                  string    `json:"name"`
	Email                 string    `json:"email"`
	MobileNumber          string    `json:"mobile_number"`
	SecondaryMobileNumber *string   `json:"secondary_mobile_number,omitempty"`
	NIC                   string    `json:"nic"`
	IsActive              bool      `json:"is_active"`
	CreatedAt             time.Time `json:"created_at"`
	UpdatedAt             time.Time `json:"updated_at"`
}

type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int64  `json:"expires_in"`
}
