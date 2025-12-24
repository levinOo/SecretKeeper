package auth

import (
	"errors"
	"secretKeeper/internal/server/db"
)

type TokenManager interface {
}

type Manager struct {
	signingKey string
}

func NewManager(signingKey string) (*Manager, error) {
	if signingKey == "" {
		return nil, errors.New("empty signing key")
	}

	return &Manager{signingKey: signingKey}, nil
}

// Функция генерации пары токенов (Access + Refresh)
func GenerateTokens(db *db.DB, userID string) (*TokenPair, error) {
	return nil, nil
}
