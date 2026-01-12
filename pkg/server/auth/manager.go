package auth

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// TokenManager менеджер токенов
type TokenManager struct {
	SigningKey string
}

// NewTokenManager создает новый менеджер токенов
func NewTokenManager(signingKey string) *TokenManager {
	return &TokenManager{
		SigningKey: signingKey,
	}
}

// GenerateAccessToken генерирует токен доступа
func (t *TokenManager) GenerateAccessToken(userId string) (string, error) {
	claims := jwt.MapClaims{
		"sub": userId,
		"exp": time.Now().Add(time.Minute * 15).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString([]byte(t.SigningKey))
}

func (t *TokenManager) Parse(accessToken string) (string, error) {
	token, err := jwt.Parse(accessToken, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("неожиданный метод подписи: %v", token.Header["alg"])
		}
		return []byte(t.SigningKey), nil
	})

	if err != nil {
		return "", err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", fmt.Errorf("неверные данные токена")
	}

	return claims["sub"].(string), nil

}

// GenerateRefreshToken генерирует токен обновления
func (t *TokenManager) GenerateRefreshToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}

	return base64.URLEncoding.EncodeToString(b), nil
}
