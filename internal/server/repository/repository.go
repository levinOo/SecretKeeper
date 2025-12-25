package repository

import (
	"context"
	"database/sql"
	"io"
	"secretKeeper/internal/server/repository/minio"
	"secretKeeper/internal/server/repository/postgres"
	"secretKeeper/pkg/server/storage"
)

type Authorization interface {
	CreateUser(ctx context.Context, username, hashedPassword string) (string, error)
	LoginUser(ctx context.Context, username string) (string, string, error)
}

type FileStorage interface {
	UploadFile(ctx context.Context, objectName string, reader io.Reader, objectSize int64, contentType string) error
}

type Repositories struct {
	Authorization
	FileStorage
}

func NewRepositories(db *sql.DB, minioStorage *storage.MinioStorage) *Repositories {
	return &Repositories{
		Authorization: postgres.NewAuthRepo(db),
		FileStorage:   minio.NewUserStorage(minioStorage),
	}
}
