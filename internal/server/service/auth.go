package service

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"secretKeeper/pkg/server/auth"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

const (
	maxPasswordBytes = 72 // лимит на длину пароля
)

type UserRepository interface {
	LoginUser(ctx context.Context, username string) (string, string, error)
	RegisterUser(ctx context.Context, username, hashedPassword string) (string, error)
	AddToken(ctx context.Context, userID, refreshToken string) error
}

type AuthService struct {
	repo          UserRepository
	tokenManager  *auth.Manager
}

func NewAuthService(repo UserRepository, tokenManager *auth.Manager) *AuthService {
	return &AuthService{
		repo:          repo,
		tokenManager:  tokenManager,
	}
}

// Функция регистрации нового пользователя
func (s *AuthService) Register(ctx context.Context, username, password string) (string, error) {
	if username == "" {
		return "", errors.New("username cannot be empty")
	}

	username = strings.TrimSpace(username)

	if len(username) < 3 || len(username) > 20 {
		return "", fmt.Errorf("username must be between 3 and 20 characters long")
	}

	if err := validateUsername(username); err != nil {
		return "", err
	}

	username = strings.ToLower(username)

	if password == "" {
		return "", errors.New("password cannot be empty")
	}

	b := []byte(password)
	if len(b) >= maxPasswordBytes {
		return "", fmt.Errorf("password cannot be longer than %d bytes", maxPasswordBytes)
	}

	password = password + "peper"

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("bcrypt error: %w", err)
	}

	id, err := s.repo.RegisterUser(ctx, username, string(hash))
	if err != nil {
		return "", err
	}

	return id, nil
}

func (s *AuthService) Login(ctx context.Context, username, password string) (string, error) {
	if username == "" {
		return "", errors.New("username cannot be empty")
	}

	username = strings.TrimSpace(username)

	if err := validateUsername(username); err != nil {
		return "", err
	}

	username = strings.ToLower(username)

	if password == "" {
		return "", errors.New("password cannot be empty")
	}

	password = password + "peper"

	id, hashedPassword, err := s.repo.LoginUser(ctx, username)
	if err != nil {
		return "", err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password)); err != nil {
		return "", errors.New("invalid username or password")
	}

	// генирирукм токены

	return id, nil
}

var usernameRegex = regexp.MustCompile(`^[a-zA-Z0-9_]{3,32}$`)

// Функция проверки валидности имени пользователя
func validateUsername(username string) error {
	if !usernameRegex.MatchString(username) {
		return errors.New("username must be 3-32 chars long and contain only alphanumeric characters or underscore")
	}

	return nil
}

type TokenPair struct {
	AccessToken  string
	RefreshToken string
}

// SecretKey храни в env!
var jwtKey = []byte("my_secret_key")

type Claims struct {
	UserID string `json:"user_id"`
	jwt.RegisteredClaims
}
