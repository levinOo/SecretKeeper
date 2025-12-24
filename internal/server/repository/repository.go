package repository

import (
	"context"
	"database/sql"
)

type Users interface {
	RegisterUser(ctx context.Context, username, passwordHash string) (string, error)
	LoginUser(ctx context.Context, username string) (string, string, error)
	AddToken(ctx context.Context, userID, refreshToken string) error
}

type Secrets interface {
	CreateSecret(ctx context.Context, userID, secretData string) (string, error)
	GetSecret(ctx context.Context, userID, secretID string) (string, error)
	DeleteSecret(ctx context.Context, userID, secretID string) error
}

type Repositories struct {
	Users   Users
	Secrets Secrets
}

func NewRepositories(db *sql.DB) *Repositories {
	return &Repositories{
		Users:   NewUserRepository(db),
		Secrets: NewSecretRepo(db),
	}
}
