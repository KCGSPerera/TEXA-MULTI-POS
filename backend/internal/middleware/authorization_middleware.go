package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/kcgsperera/texa-multi-pos/backend/internal/repositories"
)

type AuthorizationMiddleware struct {
	roleRepo repositories.RoleRepository
}

func NewAuthorizationMiddleware(roleRepo repositories.RoleRepository) *AuthorizationMiddleware {
	return &AuthorizationMiddleware{roleRepo: roleRepo}
}

func (m *AuthorizationMiddleware) RequireRole(roleName string) gin.HandlerFunc {
	required := strings.TrimSpace(strings.ToLower(roleName))

	return func(c *gin.Context) {
		roleID := c.GetString("role_id")
		if roleID == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing role in token context"})
			return
		}

		actualRole, err := m.roleRepo.GetRoleNameByID(c.Request.Context(), roleID)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "failed to resolve user role"})
			return
		}
		if strings.ToLower(strings.TrimSpace(actualRole)) != required {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "insufficient role permissions"})
			return
		}

		c.Next()
	}
}

func (m *AuthorizationMiddleware) RequirePermission(permissionCode string) gin.HandlerFunc {
	required := strings.TrimSpace(permissionCode)

	return func(c *gin.Context) {
		roleID := c.GetString("role_id")
		if roleID == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing role in token context"})
			return
		}

		ok, err := m.roleRepo.HasPermission(c.Request.Context(), roleID, required)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "failed to resolve user permissions"})
			return
		}
		if !ok {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "insufficient permissions"})
			return
		}

		c.Next()
	}
}

func (m *AuthorizationMiddleware) RequireBranchMatch(paramName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		jwtBranchID := c.GetString("branch_id")
		if jwtBranchID == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing branch in token context"})
			return
		}

		resourceBranchID := strings.TrimSpace(c.Param(paramName))
		if resourceBranchID == "" {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "branch path parameter is required"})
			return
		}

		if resourceBranchID != jwtBranchID {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "branch access denied"})
			return
		}

		c.Next()
	}
}
