package service

import (
	"testing"

	"secretKeeper/internal/server/config"
	"secretKeeper/pkg/server/auth"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestNewServices(t *testing.T) {
	cfg := &config.GRPCConfig{
		Auth: config.AuthConfig{
			JWT: config.JWTConfig{
				SigningKey: "secret",
			},
		},
	}

	logger := zap.NewNop().Sugar()
	tokenManager := auth.NewTokenManager(cfg.Auth.JWT.SigningKey)

	// We can pass nil repositories for this structural test
	authService := NewAuthService(nil, cfg, logger, tokenManager)
	userService := NewUserService(nil, nil, logger)

	services := NewServices(authService, userService)

	assert.NotNil(t, services)
	assert.Equal(t, authService, services.Auth)
	assert.Equal(t, userService, services.User)
}
