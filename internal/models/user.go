package models

import (
	"time"

	"github.com/google/uuid"
)

type Role string

const (
	RoleAdmin  Role = "admin"
	RoleUser   Role = "user"
	RoleViewer Role = "viewer"
)

type User struct {
	ID           uuid.UUID  `db:"id" json:"id"`
	Email        string     `db:"email" json:"email"`
	Username     string     `db:"username" json:"username"`
	PasswordHash string     `db:"password_hash" json:"-"`
	Role         Role       `db:"role" json:"role"`
	OAuthProvider string    `db:"oauth_provider" json:"oauth_provider,omitempty"`
	OAuthID      string     `db:"oauth_id" json:"-"`
	IsActive     bool       `db:"is_active" json:"is_active"`
	CreatedAt    time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt    time.Time  `db:"updated_at" json:"updated_at"`
	LastLoginAt  *time.Time `db:"last_login_at" json:"last_login_at,omitempty"`
}

type RefreshToken struct {
	ID        uuid.UUID `db:"id" json:"id"`
	UserID    uuid.UUID `db:"user_id" json:"user_id"`
	Token     string    `db:"token" json:"-"`
	ExpiresAt time.Time `db:"expires_at" json:"expires_at"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	Revoked   bool      `db:"revoked" json:"revoked"`
}

type AuditLog struct {
	ID        uuid.UUID `db:"id" json:"id"`
	UserID    *uuid.UUID `db:"user_id" json:"user_id,omitempty"`
	Action    string    `db:"action" json:"action"`
	Resource  string    `db:"resource" json:"resource"`
	IPAddress string    `db:"ip_address" json:"ip_address"`
	UserAgent string    `db:"user_agent" json:"user_agent"`
	Success   bool      `db:"success" json:"success"`
	Details   string    `db:"details" json:"details"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
}

// Request/Response DTOs

type RegisterRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Username string `json:"username" binding:"required,min=3,max=50"`
	Password string `json:"password" binding:"required,min=8"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type AuthResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int64  `json:"expires_in"`
	User         *User  `json:"user"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type Claims struct {
	UserID   string `json:"user_id"`
	Email    string `json:"email"`
	Username string `json:"username"`
	Role     Role   `json:"role"`
}
