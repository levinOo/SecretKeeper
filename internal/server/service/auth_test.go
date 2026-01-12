package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"secretKeeper/internal/server/config"
	"secretKeeper/internal/server/repository/mocks"
	"secretKeeper/pkg/server/auth"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

func TestAuthService_CreateUser(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockAuthorization(ctrl)
	tokenManager := auth.NewTokenManager("secret_signing_key")
	cfg := &config.GRPCConfig{
		Auth: config.AuthConfig{
			Pepper: "test_pepper",
			JWT: config.JWTConfig{
				RefreshTokenTTL: time.Hour,
			},
		},
	}
	logger := zap.NewNop().Sugar()

	service := NewAuthService(mockRepo, cfg, logger, tokenManager)

	ctx := context.Background()
	username := "testuser"
	password := "password123"

	t.Run("Success", func(t *testing.T) {
		userID := "user_id_1"
		mockRepo.EXPECT().CreateUser(ctx, username, gomock.Any()).Return(userID, nil)
		mockRepo.EXPECT().CreateSession(ctx, userID, gomock.Any(), gomock.Any()).Return(nil)

		accessToken, refreshToken, err := service.CreateUser(ctx, username, password)
		require.NoError(t, err)
		assert.NotEmpty(t, accessToken)
		assert.NotEmpty(t, refreshToken)
	})

	t.Run("CreateUser Repo Error", func(t *testing.T) {
		mockRepo.EXPECT().CreateUser(ctx, username, gomock.Any()).Return("", errors.New("db error"))

		_, _, err := service.CreateUser(ctx, username, password)
		assert.Error(t, err)
		assert.Equal(t, "db error", err.Error())
	})

	t.Run("Invalid Username", func(t *testing.T) {
		_, _, err := service.CreateUser(ctx, "ab", password) // too short
		assert.Error(t, err)
	})
}

func TestAuthService_LoginUser(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockAuthorization(ctrl)
	tokenManager := auth.NewTokenManager("secret_signing_key")
	cfg := &config.GRPCConfig{
		Auth: config.AuthConfig{
			Pepper: "test_pepper",
			JWT: config.JWTConfig{
				RefreshTokenTTL: time.Hour,
			},
		},
	}
	logger := zap.NewNop().Sugar()

	service := NewAuthService(mockRepo, cfg, logger, tokenManager)

	ctx := context.Background()
	username := "testuser"
	password := "password123"
	userID := "user_id_1"

	// Hashed password setup
	passwordWithPepper := password + cfg.Auth.Pepper
	hashedPasswordBytes, _ := bcrypt.GenerateFromPassword([]byte(passwordWithPepper), bcrypt.DefaultCost)
	hashedPassword := string(hashedPasswordBytes)

	t.Run("Success", func(t *testing.T) {
		mockRepo.EXPECT().LoginUser(ctx, username).Return(userID, hashedPassword, nil)
		mockRepo.EXPECT().CreateSession(ctx, userID, gomock.Any(), gomock.Any()).Return(nil)

		accessToken, refreshToken, err := service.LoginUser(ctx, username, password)
		require.NoError(t, err)
		assert.NotEmpty(t, accessToken)
		assert.NotEmpty(t, refreshToken)
	})

	t.Run("Invalid Password", func(t *testing.T) {
		mockRepo.EXPECT().LoginUser(ctx, username).Return(userID, hashedPassword, nil)

		_, _, err := service.LoginUser(ctx, username, "wrong_password")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "неверное имя пользователя или пароль")
	})

	t.Run("User Not Found", func(t *testing.T) {
		mockRepo.EXPECT().LoginUser(ctx, username).Return("", "", errors.New("user not found"))

		_, _, err := service.LoginUser(ctx, username, password)
		assert.Error(t, err)
	})
}
