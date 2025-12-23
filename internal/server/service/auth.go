package service

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"regexp"
	"secretKeeper/internal/server/db"
	pb "secretKeeper/internal/server/proto"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

const (
	maxPasswordBytes = 72 // лимит на длину пароля
)

// Функция регистрации нового пользователя
func (s *Server) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	// получаем username
	username := req.GetUsername()

	// проверка на пустоту
	if username == "" {
		s.logger.Warnw("Register failed: empty username")
		return nil, fmt.Errorf("username can not be empty")
	}

	// убираем лишние пробелы
	username = strings.TrimSpace(username)

	// проверка на длину username
	if len(username) < 3 || len(username) > 20 {
		s.logger.Warnw("Register failed: invalid username length", "username", username)
		return nil, fmt.Errorf("username must be between 3 and 20 characters long")
	}

	// проверка на валидность username
	if err := ValidateUsername(username); err != nil {
		return nil, err
	}

	// приведение к нижнему регистру
	username = strings.ToLower(username)

	// получаем пароль
	password := req.GetPassword()

	// проверка на пустоту
	if password == "" {
		s.logger.Warnw("Register failed: empty password")
		return nil, fmt.Errorf("password cannot be empty")
	}

	// проверка на длину пароля
	b := []byte(password)
	if len(b) >= maxPasswordBytes {
		s.logger.Warnw("Register failed: password too long", "password length", len(b))
		return nil, fmt.Errorf("password cannot be longer than %d bytes", maxPasswordBytes)
	}

	// формируем пароль с pepper
	password = password + s.config.GRPC.Pepper

	// хешируем пароль
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		s.logger.Warnw("Register failed: bcrypt error", "error", err)
		return nil, fmt.Errorf("bcrypt error: %w", err)
	}

	// регистрируем пользователя
	id, err := s.db.RegisterUser(ctx, username, string(hash))
	if err != nil {
		return nil, err
	}

	// генерируем токены
	tokens, err := GenerateTokens(s.db, id)
	if err != nil {
		return nil, err
	}

	return &pb.RegisterResponse{
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
	}, nil
}

// Функция аутентификации пользователя (Login)
func (s *Server) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	// Получаем username
	username := req.GetUsername()

	// проверка на пустоту
	if username == "" {
		s.logger.Warnw("Login failed: empty username")
		return nil, fmt.Errorf("username cannot be empty")
	}

	// убираем лишние пробелы
	username = strings.TrimSpace(username)

	// проверка на длину username
	if len(username) < 3 || len(username) > 20 {
		s.logger.Warnw("Login failed: invalid username length", "username", username)
		return nil, fmt.Errorf("username must be between 3 and 20 characters long")
	}

	// проверка на валидность username
	if err := ValidateUsername(username); err != nil {
		return nil, err
	}

	// приведение к нижнему регистру
	username = strings.ToLower(username)

	// Получаем пароль
	password := req.GetPassword()

	b := []byte(password)
	if len(b) >= maxPasswordBytes {
		s.logger.Warnw("Login failed: password too long", "password length", len(b))
		return nil, fmt.Errorf("password cannot be longer than %d bytes", maxPasswordBytes)
	}

	// формируем пароль с pepper
	password = password + s.config.GRPC.Pepper

	// получаем хеш пароля
	id, hashedPassword, err := s.db.GetUserByUsername(ctx, username)
	if err != nil {
		return nil, err
	}

	// сравниваем хеши
	if err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password)); err != nil {
		s.logger.Warnw("Login failed: bcrypt error", "error", err)
		return nil, fmt.Errorf("bcrypt error: %w", err)
	}

	// генерируем токены
	tokens, err := GenerateTokens(s.db, id)
	if err != nil {
		return nil, err
	}

	return &pb.LoginResponse{
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
	}, nil
}

var usernameRegex = regexp.MustCompile(`^[a-zA-Z0-9_]{3,32}$`)

// Функция проверки валидности имени пользователя
func ValidateUsername(username string) error {
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

// Функция генерации пары токенов (Access + Refresh)
func GenerateTokens(db *db.DB, userID string) (*TokenPair, error) {
	expirationTime := time.Now().Add(15 * time.Minute) // вынести 15 minutes в const
	claims := &Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	accessTokenString, err := token.SignedString(jwtKey)
	if err != nil {
		return nil, err
	}
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return nil, err
	}

	refreshToken := base64.URLEncoding.EncodeToString(buf)

	db.AddToken(context.Background(), userID, refreshToken)

	return &TokenPair{
		AccessToken:  accessTokenString,
		RefreshToken: refreshToken,
	}, nil
}
