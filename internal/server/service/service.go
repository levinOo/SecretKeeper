package service

import (
	"context"
	"secretKeeper/internal/server/repository"
	"secretKeeper/pkg/server/auth"
	"secretKeeper/pkg/server/storage"
)

type Users interface {
	Login(ctx context.Context, username, password string) (string, error)
	Register(ctx context.Context, username, password string) (string, error)
}

type Secrets interface {
	CreateSecret(ctx context.Context, userID, secretData string) (string, error)
	GetSecret(ctx context.Context, userID, secretID string) (string, error)
	DeleteSecret(ctx context.Context, userID, secretID string) error
}

type Services struct {
	Users   Users
	Secrets Secrets
}

type Deps struct {
	Repos        *repository.Repositories
	Storage      *storage.MinioStorage
	TokenManager *auth.Manager
}

func NewServices(deps Deps) *Services {
	return &Services{
		Users:   NewAuthService(deps.Repos.Users, deps.TokenManager),
		Secrets: NewSecretService(deps.Repos.Secrets, deps.Storage),
	}
}
