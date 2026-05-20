package handlers

import (
	"crypto/rand"
	"encoding/base64"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/likhithnagaraj79/auth-service/internal/auth"
	"github.com/likhithnagaraj79/auth-service/internal/middleware"
	"github.com/likhithnagaraj79/auth-service/internal/models"
	"github.com/likhithnagaraj79/auth-service/internal/oauth"
	"github.com/likhithnagaraj79/auth-service/pkg/config"
	"go.uber.org/zap"
)

type AuthHandler struct {
	db           *sqlx.DB
	jwtSvc       *auth.JWTService
	googleOAuth  *oauth.GoogleProvider
	cfg          *config.Config
}

func NewAuthHandler(db *sqlx.DB, jwtSvc *auth.JWTService, googleOAuth *oauth.GoogleProvider, cfg *config.Config) *AuthHandler {
	return &AuthHandler{db: db, jwtSvc: jwtSvc, googleOAuth: googleOAuth, cfg: cfg}
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req models.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	hash, err := auth.HashPassword(req.Password)
	if err != nil {
		zap.S().Errorf("hash password: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	user := models.User{
		ID:           uuid.New(),
		Email:        req.Email,
		Username:     req.Username,
		PasswordHash: hash,
		Role:         models.RoleUser,
		IsActive:     true,
	}

	_, err = h.db.Exec(`
		INSERT INTO users (id, email, username, password_hash, role, is_active)
		VALUES ($1, $2, $3, $4, $5, $6)`,
		user.ID, user.Email, user.Username, user.PasswordHash, user.Role, user.IsActive,
	)
	if err != nil {
		zap.S().Errorf("insert user: %v", err)
		c.JSON(http.StatusConflict, gin.H{"error": "email or username already exists"})
		return
	}

	tokens, err := h.jwtSvc.GenerateTokenPair(&user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate tokens"})
		return
	}

	h.storeRefreshToken(user.ID, tokens.RefreshToken)

	c.JSON(http.StatusCreated, models.AuthResponse{
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    tokens.ExpiresIn,
		User:         &user,
	})
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var user models.User
	err := h.db.Get(&user, `SELECT * FROM users WHERE email=$1 AND is_active=TRUE`, req.Email)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}

	if user.OAuthProvider != "" && user.OAuthProvider != "local" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "please use OAuth login"})
		return
	}

	if !auth.CheckPassword(req.Password, user.PasswordHash) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}

	tokens, err := h.jwtSvc.GenerateTokenPair(&user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate tokens"})
		return
	}

	h.storeRefreshToken(user.ID, tokens.RefreshToken)
	h.db.Exec(`UPDATE users SET last_login_at=NOW() WHERE id=$1`, user.ID)

	c.JSON(http.StatusOK, models.AuthResponse{
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    tokens.ExpiresIn,
		User:         &user,
	})
}

func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var req models.RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var rt models.RefreshToken
	err := h.db.Get(&rt, `
		SELECT * FROM refresh_tokens
		WHERE token=$1 AND revoked=FALSE AND expires_at > NOW()`,
		req.RefreshToken,
	)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired refresh token"})
		return
	}

	var user models.User
	if err := h.db.Get(&user, `SELECT * FROM users WHERE id=$1 AND is_active=TRUE`, rt.UserID); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not found"})
		return
	}

	h.db.Exec(`UPDATE refresh_tokens SET revoked=TRUE WHERE id=$1`, rt.ID)

	tokens, err := h.jwtSvc.GenerateTokenPair(&user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate tokens"})
		return
	}

	h.storeRefreshToken(user.ID, tokens.RefreshToken)

	c.JSON(http.StatusOK, models.AuthResponse{
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    tokens.ExpiresIn,
		User:         &user,
	})
}

func (h *AuthHandler) Logout(c *gin.Context) {
	var req models.RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	h.db.Exec(`UPDATE refresh_tokens SET revoked=TRUE WHERE token=$1`, req.RefreshToken)
	c.JSON(http.StatusOK, gin.H{"message": "logged out successfully"})
}

func (h *AuthHandler) GoogleLogin(c *gin.Context) {
	state, _ := generateState()
	url := h.googleOAuth.GetAuthURL(state)
	c.JSON(http.StatusOK, gin.H{"url": url})
}

func (h *AuthHandler) GoogleCallback(c *gin.Context) {
	code := c.Query("code")
	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing code"})
		return
	}

	userInfo, err := h.googleOAuth.ExchangeCode(c.Request.Context(), code)
	if err != nil {
		zap.S().Errorf("google exchange: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "oauth exchange failed"})
		return
	}

	var user models.User
	err = h.db.Get(&user, `SELECT * FROM users WHERE oauth_provider='google' AND oauth_id=$1`, userInfo.ID)
	if err != nil {
		user = models.User{
			ID:            uuid.New(),
			Email:         userInfo.Email,
			Username:      userInfo.Name,
			OAuthProvider: "google",
			OAuthID:       userInfo.ID,
			Role:          models.RoleUser,
			IsActive:      true,
		}
		_, err = h.db.Exec(`
			INSERT INTO users (id, email, username, oauth_provider, oauth_id, role, is_active)
			VALUES ($1, $2, $3, $4, $5, $6, $7)
			ON CONFLICT (email) DO UPDATE SET oauth_provider=$4, oauth_id=$5, updated_at=NOW()`,
			user.ID, user.Email, user.Username, user.OAuthProvider, user.OAuthID, user.Role, user.IsActive,
		)
		if err != nil {
			zap.S().Errorf("upsert oauth user: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create user"})
			return
		}
	}

	tokens, err := h.jwtSvc.GenerateTokenPair(&user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate tokens"})
		return
	}

	h.storeRefreshToken(user.ID, tokens.RefreshToken)

	c.JSON(http.StatusOK, models.AuthResponse{
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    tokens.ExpiresIn,
		User:         &user,
	})
}

func (h *AuthHandler) Me(c *gin.Context) {
	userID := middleware.GetUserID(c)
	var user models.User
	if err := h.db.Get(&user, `SELECT * FROM users WHERE id=$1`, userID); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}
	c.JSON(http.StatusOK, user)
}

func (h *AuthHandler) storeRefreshToken(userID uuid.UUID, token string) {
	h.db.Exec(`
		INSERT INTO refresh_tokens (id, user_id, token, expires_at)
		VALUES ($1, $2, $3, $4)`,
		uuid.New(), userID, token, time.Now().Add(h.cfg.JWT.RefreshTTL),
	)
}

func generateState() (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	return base64.URLEncoding.EncodeToString(b), err
}
