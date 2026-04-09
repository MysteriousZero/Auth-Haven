package middleware

import (
	"auth-haven/internal/domain/errors"
	"auth-haven/internal/domain/interfaces"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// AuthMiddleware validates the Bearer token and injects claims into gin context.
func AuthMiddleware(tokenService interfaces.TokenService) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error":   "Unauthenticated",
				"message": "Authorization header required",
			})
			return
		}

		token := strings.TrimPrefix(authHeader, "Bearer ")
		claims, err := tokenService.ValidateAccessToken(c.Request.Context(), token)
		if err != nil {
			status, code := errors.HTTPError(errors.ErrInvalidToken)
			c.AbortWithStatusJSON(status, gin.H{
				"error":   code,
				"message": "Invalid or expired token",
			})
			return
		}

		// Inject claims into context for handlers
		c.Set("user_id", claims.UserID)
		c.Set("tenant_id", claims.TenantID)
		c.Set("tenant_type", claims.TenantType)
		c.Set("role_id", claims.RoleID)
		c.Set("email", claims.Email)
		c.Set("claims", claims)

		c.Next()
	}
}
