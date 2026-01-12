package storage

import "fmt"

// TokensStorage хранилище токенов
type TokensStorage struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

// TokenRepository интерфейс репозитория токенов
type TokenRepository interface {
	SaveTokens(accessToken, refreshToken string) error
	GetToken() string
	DeleteTokens()
}

// NewTokensStorage создает новое хранилище токенов
func NewTokensStorage() *TokensStorage {
	return &TokensStorage{}
}

// SaveTokens сохраняет токены
func (s *TokensStorage) SaveTokens(accessToken, refreshToken string) error {
	if accessToken == "" || refreshToken == "" {
		return fmt.Errorf("токены не могут быть пустыми")
	}

	s.AccessToken = accessToken
	s.RefreshToken = refreshToken

	return nil
}

// GetToken получает токен доступа
func (s *TokensStorage) GetToken() string {
	return s.AccessToken
}

// DeleteTokens удаляет токены
func (s *TokensStorage) DeleteTokens() {
	s.AccessToken = ""
	s.RefreshToken = ""
}
