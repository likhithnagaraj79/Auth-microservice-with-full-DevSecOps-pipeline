package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/likhithnagaraj79/auth-service/internal/middleware"
	"github.com/likhithnagaraj79/auth-service/internal/models"
)

type UserHandler struct {
	db *sqlx.DB
}

func NewUserHandler(db *sqlx.DB) *UserHandler {
	return &UserHandler{db: db}
}

func (h *UserHandler) ListUsers(c *gin.Context) {
	var users []models.User
	if err := h.db.Select(&users, `SELECT * FROM users ORDER BY created_at DESC`); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch users"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"users": users, "count": len(users)})
}

func (h *UserHandler) GetUser(c *gin.Context) {
	id := c.Param("id")
	var user models.User
	if err := h.db.Get(&user, `SELECT * FROM users WHERE id=$1`, id); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}
	c.JSON(http.StatusOK, user)
}

func (h *UserHandler) UpdateRole(c *gin.Context) {
	id := c.Param("id")
	var body struct {
		Role models.Role `json:"role" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	validRoles := map[models.Role]bool{
		models.RoleAdmin:  true,
		models.RoleUser:   true,
		models.RoleViewer: true,
	}
	if !validRoles[body.Role] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid role"})
		return
	}

	res, err := h.db.Exec(`UPDATE users SET role=$1, updated_at=NOW() WHERE id=$2`, body.Role, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update role"})
		return
	}
	rows, _ := res.RowsAffected()
	if rows == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "role updated"})
}

func (h *UserHandler) DeleteUser(c *gin.Context) {
	id := c.Param("id")
	callerID := middleware.GetUserID(c)
	if id == callerID {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cannot delete yourself"})
		return
	}

	res, err := h.db.Exec(`DELETE FROM users WHERE id=$1`, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete user"})
		return
	}
	rows, _ := res.RowsAffected()
	if rows == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "user deleted"})
}

func (h *UserHandler) GetAuditLogs(c *gin.Context) {
	var logs []models.AuditLog
	if err := h.db.Select(&logs, `SELECT * FROM audit_logs ORDER BY created_at DESC LIMIT 200`); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch audit logs"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"logs": logs, "count": len(logs)})
}

func (h *UserHandler) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "healthy",
		"service": "auth-service",
		"version": "1.0.0",
	})
}

func (h *UserHandler) Permissions(c *gin.Context) {
	role := middleware.GetUserRole(c)
	c.JSON(http.StatusOK, gin.H{
		"role": role,
	})
}
