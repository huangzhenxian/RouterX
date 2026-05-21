package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/routex/routex/internal/auth"
)

const (
	CtxAdminID   = "admin_id"
	CtxAdminName = "admin_username"
	CtxAdminRole = "admin_role"
)

func JWTAuth(secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		h := c.GetHeader("Authorization")
		if !strings.HasPrefix(h, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing bearer token"})
			return
		}
		claims, err := auth.Verify(secret, strings.TrimPrefix(h, "Bearer "))
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return
		}
		c.Set(CtxAdminID, claims.AdminID)
		c.Set(CtxAdminName, claims.Username)
		c.Set(CtxAdminRole, claims.Role)
		c.Next()
	}
}
