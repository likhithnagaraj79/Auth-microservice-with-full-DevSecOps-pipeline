package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/likhithnagaraj79/auth-service/internal/models"
	"go.uber.org/zap"
)

func AuditLogger(db *sqlx.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		action := c.Request.Method + " " + c.FullPath()
		success := c.Writer.Status() < 400

		var userID *uuid.UUID
		if id := GetUserID(c); id != "" {
			if uid, err := uuid.Parse(id); err == nil {
				userID = &uid
			}
		}

		log := models.AuditLog{
			ID:        uuid.New(),
			UserID:    userID,
			Action:    action,
			Resource:  c.Request.URL.Path,
			IPAddress: c.ClientIP(),
			UserAgent: c.Request.UserAgent(),
			Success:   success,
		}

		_, err := db.Exec(`
			INSERT INTO audit_logs (id, user_id, action, resource, ip_address, user_agent, success)
			VALUES ($1, $2, $3, $4, $5, $6, $7)`,
			log.ID, log.UserID, log.Action, log.Resource, log.IPAddress, log.UserAgent, log.Success,
		)
		if err != nil {
			zap.S().Warnf("audit log insert failed: %v", err)
		}
	}
}
