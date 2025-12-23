package repository

import "errors"

// Структура для хранения токенов пользователя
type UserTokens struct {
	AccessToken  string
	RefreshToken string
}

// Функция создания хранилища токенов
func NewUserTokensStorage() *UserTokens {
	return &UserTokens{}
}

// Функция сохранения токенов в память
func (s *UserTokens) LoadTokens(accessToken, refreshToken string) {
	s.AccessToken = accessToken
	s.RefreshToken = refreshToken
}

// Функция получения текущего токена доступа
func (s *UserTokens) GetToken() (string, error) {
	if s.AccessToken == "" {
		return "", errors.New("")
	}
	return s.AccessToken, nil
}
