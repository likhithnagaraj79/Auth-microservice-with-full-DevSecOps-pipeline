package auth_test

import (
	"testing"

	"github.com/likhithnagaraj79/auth-service/internal/auth"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHashPassword(t *testing.T) {
	hash, err := auth.HashPassword("securepassword123")
	require.NoError(t, err)
	assert.NotEmpty(t, hash)
	assert.NotEqual(t, "securepassword123", hash)
}

func TestCheckPassword(t *testing.T) {
	password := "myS3cur3P@ss!"
	hash, err := auth.HashPassword(password)
	require.NoError(t, err)

	assert.True(t, auth.CheckPassword(password, hash))
	assert.False(t, auth.CheckPassword("wrongpassword", hash))
	assert.False(t, auth.CheckPassword("", hash))
}

func TestHashPassword_DifferentHashesSameInput(t *testing.T) {
	hash1, _ := auth.HashPassword("samepassword")
	hash2, _ := auth.HashPassword("samepassword")
	// bcrypt salts produce different hashes
	assert.NotEqual(t, hash1, hash2)
	assert.True(t, auth.CheckPassword("samepassword", hash1))
	assert.True(t, auth.CheckPassword("samepassword", hash2))
}
