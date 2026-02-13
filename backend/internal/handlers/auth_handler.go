package handlers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/kcgsperera/texa-multi-pos/backend/internal/models"
	"github.com/kcgsperera/texa-multi-pos/backend/internal/services"
)

type AuthHandler struct{ authService services.AuthService }

func NewAuthHandler(authService services.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req models.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, err.Error())
		return
	}
	res, err := h.authService.Register(c.Request.Context(), req)
	if err != nil {
		if errors.Is(err, services.ErrInvalidBranch) || errors.Is(err, services.ErrInvalidRole) {
			respondError(c, http.StatusBadRequest, err.Error())
			return
		}
		respondMappedError(c, err, "failed to register user")
		return
	}
	respondSuccess(c, http.StatusCreated, res)
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, err.Error())
		return
	}
	res, err := h.authService.Login(c.Request.Context(), req)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrInvalidCredentials):
			respondError(c, http.StatusUnauthorized, "invalid email or password")
		case errors.Is(err, services.ErrInactiveUser):
			respondError(c, http.StatusForbidden, err.Error())
		default:
			respondMappedError(c, err, "failed to login")
		}
		return
	}
	respondSuccess(c, http.StatusOK, res)
}

func (h *AuthHandler) Me(c *gin.Context) {
	respondSuccess(c, http.StatusOK, gin.H{"user_id": c.GetString("user_id"), "role_id": c.GetString("role_id"), "branch_id": c.GetString("branch_id")})
}
