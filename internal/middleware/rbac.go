package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/likhithnagaraj79/auth-service/internal/rbac"
)

func RequirePermission(perm rbac.Permission) gin.HandlerFunc {
	return func(c *gin.Context) {
		role := GetUserRole(c)
		if !rbac.HasPermission(role, perm) {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error": "insufficient permissions",
				"required": string(perm),
			})
			return
		}
		c.Next()
	}
}

func RequireRole(roles ...string) gin.HandlerFunc {
	allowed := make(map[string]bool, len(roles))
	for _, r := range roles {
		allowed[r] = true
	}
	return func(c *gin.Context) {
		role := GetUserRole(c)
		if !allowed[string(role)] {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "access denied"})
			return
		}
		c.Next()
	}
}
