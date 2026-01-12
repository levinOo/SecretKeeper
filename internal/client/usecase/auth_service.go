package usecase

import (
	"context"
	"fmt"
	storage "secretKeeper/pkg/client/token_storage"
)

// AuthProvider интерфейс для аутентификации
type AuthProvider interface {
	Register(ctx context.Context, username, password string) (accessToken string, refreshToken string, err error)
	Login(ctx context.Context, username, password string) (accessToken string, refreshToken string, err error)
}

// AuthService сервис аутентификации
type AuthService struct {
	tokenStorage storage.TokenRepository
	provider     AuthProvider // Используем интерфейс
}

// NewAuthService создает новый сервис аутентификации
func NewAuthService(tokenStorage storage.TokenRepository, provider AuthProvider) *AuthService {
	return &AuthService{
		tokenStorage: tokenStorage,
		provider:     provider,
	}
}

// Register регистрация нового пользователя
func (s *AuthService) Register(ctx context.Context, username, password string) error {
	if username == "" || password == "" {
		return fmt.Errorf("имя пользователя или пароль не могут быть пустыми")
	}

	accessToken, refreshToken, err := s.provider.Register(ctx, username, password)
	if err != nil {
		return err
	}

	if err := s.tokenStorage.SaveTokens(accessToken, refreshToken); err != nil {
		return err
	}

	return nil
}

// Login вход пользователя
func (s *AuthService) Login(ctx context.Context, username, password string) error {
	if username == "" || password == "" {
		return fmt.Errorf("имя пользователя или пароль не могут быть пустыми")
	}

	accessToken, refreshToken, err := s.provider.Login(ctx, username, password)
	if err != nil {
		return err
	}

	if err := s.tokenStorage.SaveTokens(accessToken, refreshToken); err != nil {
		return err
	}

	return nil
}
