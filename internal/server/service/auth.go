package service

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"secretKeeper/internal/server/config"
	"secretKeeper/pkg/server/auth"
	"strings"
	"time"

	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

const (
	maxPasswordBytes = 72 // лимит на длину пароля
)

// AuthRepository интерфейс репозитория авторизации
type AuthRepository interface {
	CreateUser(ctx context.Context, username, hashedPassword string) (string, error)
	LoginUser(ctx context.Context, username string) (string, string, error)
	CreateSession(ctx context.Context, id, refreshTokenHash string, expiresIn time.Duration) error
}

// AuthService сервис авторизации
type AuthService struct {
	repo         AuthRepository
	cfg          *config.GRPCConfig
	logger       *zap.SugaredLogger
	tokenManager *auth.TokenManager
}

// NewAuthService создает новый сервис авторизации
func NewAuthService(repo AuthRepository, cfg *config.GRPCConfig, logger *zap.SugaredLogger, tokenManager *auth.TokenManager) *AuthService {
	return &AuthService{repo: repo, cfg: cfg, logger: logger, tokenManager: tokenManager}
}

// CreateUser создает пользователя
func (s *AuthService) CreateUser(ctx context.Context, username, password string) (string, string, error) {
	if username == "" {
		return "", "", errors.New("имя пользователя не может быть пустым")
	}

	username = strings.TrimSpace(username)

	if err := validateUsername(username); err != nil {
		return "", "", err
	}

	username = strings.ToLower(username)

	if password == "" {
		return "", "", errors.New("пароль не может быть пустым")
	}

	b := []byte(password)
	if len(b) >= maxPasswordBytes {
		return "", "", fmt.Errorf("пароль не может быть длиннее %d байт", maxPasswordBytes)
	}

	password = password + s.cfg.Auth.Pepper

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", "", fmt.Errorf("ошибка bcrypt: %w", err)
	}

	id, err := s.repo.CreateUser(ctx, username, string(hash))
	if err != nil {
		return "", "", err
	}

	tokens, err := s.createSession(ctx, id)
	if err != nil {
		return "", "", err
	}

	return tokens.AccessToken, tokens.RefreshToken, err
}

// LoginUser выполняет вход пользователя
func (s *AuthService) LoginUser(ctx context.Context, username, password string) (string, string, error) {
	if username == "" {
		return "", "", errors.New("имя пользователя не может быть пустым")
	}

	username = strings.TrimSpace(username)

	if err := validateUsername(username); err != nil {
		return "", "", err
	}

	username = strings.ToLower(username)

	if password == "" {
		return "", "", errors.New("password cannot be empty")
	}

	// Используем перец из конфигурации, как при регистрации
	password = password + s.cfg.Auth.Pepper

	id, hashedPassword, err := s.repo.LoginUser(ctx, username)
	if err != nil {
		return "", "", err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password)); err != nil {
		return "", "", errors.New("неверное имя пользователя или пароль")
	}

	tokens, err := s.createSession(ctx, id)
	if err != nil {
		return "", "", err
	}

	return tokens.AccessToken, tokens.RefreshToken, nil
}

// func CreateNewTokenPair || RefreshTokens если access протухы

func (s *AuthService) createSession(ctx context.Context, id string) (Tokens, error) {
	var (
		res Tokens
		err error
	)

	res.AccessToken, err = s.tokenManager.GenerateAccessToken(id)
	if err != nil {
		return res, err
	}

	res.RefreshToken, err = s.tokenManager.GenerateRefreshToken()
	if err != nil {
		return res, err
	}

	err = s.repo.CreateSession(ctx, id, res.RefreshToken, s.cfg.Auth.JWT.RefreshTokenTTL)
	if err != nil {
		return res, err
	}

	return res, nil
}

// Tokens пара токенов
type Tokens struct {
	AccessToken  string
	RefreshToken string
}

// Функция проверки валидности имени пользователя
func validateUsername(username string) error {
	usernameRegex := regexp.MustCompile(`^[a-zA-Z0-9_]{3,32}$`)

	if !usernameRegex.MatchString(username) {
		return errors.New("имя пользователя должно содержать от 3 до 32 символов и состоять только из букв, цифр или подчеркивания")
	}

	return nil
}
