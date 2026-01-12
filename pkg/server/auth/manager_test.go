package auth

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewTokenManager(t *testing.T) {
	signingKey := "secret"
	manager := NewTokenManager(signingKey)
	assert.NotNil(t, manager)
	assert.Equal(t, signingKey, manager.SigningKey)
}

func TestTokenManager_GenerateAccessToken(t *testing.T) {
	signingKey := "secret"
	manager := NewTokenManager(signingKey)
	userID := "user123"

	token, err := manager.GenerateAccessToken(userID)
	require.NoError(t, err)
	assert.NotEmpty(t, token)

	// Verify the token manually to ensure it's valid
	parsedToken, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		return []byte(signingKey), nil
	})
	require.NoError(t, err)
	assert.True(t, parsedToken.Valid)

	claims, ok := parsedToken.Claims.(jwt.MapClaims)
	require.True(t, ok)
	assert.Equal(t, userID, claims["sub"])
	assert.NotZero(t, claims["exp"])
}

func TestTokenManager_Parse(t *testing.T) {
	signingKey := "secret"
	manager := NewTokenManager(signingKey)
	userID := "user123"

	t.Run("Valid Token", func(t *testing.T) {
		token, err := manager.GenerateAccessToken(userID)
		require.NoError(t, err)

		parsedUserID, err := manager.Parse(token)
		require.NoError(t, err)
		assert.Equal(t, userID, parsedUserID)
	})

	t.Run("Invalid Token", func(t *testing.T) {
		_, err := manager.Parse("invalid_token")
		assert.Error(t, err)
	})

	t.Run("Invalid Signature", func(t *testing.T) {
		otherManager := NewTokenManager("wrong_secret")
		token, err := otherManager.GenerateAccessToken(userID)
		require.NoError(t, err)

		_, err = manager.Parse(token)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "signature is invalid")
	})

	t.Run("Expired Token", func(t *testing.T) {
		// Manually create an expired token
		claims := jwt.MapClaims{
			"sub": userID,
			"exp": time.Now().Add(-time.Minute).Unix(),
		}
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		signedToken, err := token.SignedString([]byte(signingKey))
		require.NoError(t, err)

		_, err = manager.Parse(signedToken)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "token is expired")
	})
}

func TestTokenManager_GenerateRefreshToken(t *testing.T) {
	manager := NewTokenManager("secret")

	token, err := manager.GenerateRefreshToken()
	require.NoError(t, err)
	assert.NotEmpty(t, token)
	assert.Greater(t, len(token), 32) // Base64 encoding increases size
}
