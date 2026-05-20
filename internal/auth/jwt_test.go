package auth_test

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/likhithnagaraj79/auth-service/internal/auth"
	"github.com/likhithnagaraj79/auth-service/internal/models"
	"github.com/likhithnagaraj79/auth-service/pkg/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestJWTService() *auth.JWTService {
	return auth.NewJWTService(&config.JWTConfig{
		AccessSecret:  "test-access-secret-32chars-longenough",
		RefreshSecret: "test-refresh-secret-32chars-longenough",
		AccessTTL:     15 * time.Minute,
		RefreshTTL:    168 * time.Hour,
	})
}

func TestGenerateTokenPair(t *testing.T) {
	svc := newTestJWTService()
	user := &models.User{
		ID:       uuid.New(),
		Email:    "test@example.com",
		Username: "testuser",
		Role:     models.RoleUser,
	}

	tokens, err := svc.GenerateTokenPair(user)
	require.NoError(t, err)
	assert.NotEmpty(t, tokens.AccessToken)
	assert.NotEmpty(t, tokens.RefreshToken)
	assert.Greater(t, tokens.ExpiresIn, int64(0))
}

func TestValidateAccessToken(t *testing.T) {
	svc := newTestJWTService()
	user := &models.User{
		ID:       uuid.New(),
		Email:    "test@example.com",
		Username: "testuser",
		Role:     models.RoleAdmin,
	}

	tokens, err := svc.GenerateTokenPair(user)
	require.NoError(t, err)

	claims, err := svc.ValidateAccessToken(tokens.AccessToken)
	require.NoError(t, err)
	assert.Equal(t, user.ID.String(), claims.UserID)
	assert.Equal(t, user.Email, claims.Email)
	assert.Equal(t, models.RoleAdmin, claims.Role)
}

func TestValidateAccessToken_Invalid(t *testing.T) {
	svc := newTestJWTService()

	_, err := svc.ValidateAccessToken("not.a.valid.token")
	assert.Error(t, err)

	_, err = svc.ValidateAccessToken("")
	assert.Error(t, err)
}
