package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/likhithnagaraj79/auth-service/internal/auth"
	"github.com/likhithnagaraj79/auth-service/internal/models"
)

const (
	ctxUserID   = "user_id"
	ctxEmail    = "email"
	ctxUsername = "username"
	ctxRole     = "role"
)

func AuthRequired(jwtSvc *auth.JWTService) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if header == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "authorization header required"})
			return
		}

		parts := strings.SplitN(header, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "bearer") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization header format"})
			return
		}

		claims, err := jwtSvc.ValidateAccessToken(parts[1])
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
			return
		}

		c.Set(ctxUserID, claims.UserID)
		c.Set(ctxEmail, claims.Email)
		c.Set(ctxUsername, claims.Username)
		c.Set(ctxRole, claims.Role)
		c.Next()
	}
}

func GetUserRole(c *gin.Context) models.Role {
	role, _ := c.Get(ctxRole)
	r, _ := role.(models.Role)
	return r
}

func GetUserID(c *gin.Context) string {
	id, _ := c.Get(ctxUserID)
	s, _ := id.(string)
	return s
}
