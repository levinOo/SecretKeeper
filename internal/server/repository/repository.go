package repository

import (
	"context"
	"database/sql"
	"io"
	pb "secretKeeper/internal/proto"
	"secretKeeper/internal/server/domain"
	"secretKeeper/internal/server/repository/minio"
	"secretKeeper/internal/server/repository/postgres"
	"secretKeeper/pkg/server/storage"
	"time"
)

// Authorization интерфейс авторизации
type Authorization interface {
	CreateUser(ctx context.Context, username, hashedPassword string) (string, error)
	LoginUser(ctx context.Context, username string) (string, string, error)
	CreateSession(ctx context.Context, id, refreshTokenHash string, expiresIn time.Duration) error
}

// User интерфейс работы с данными пользователя
type User interface {
	SaveSecret(ctx context.Context, userID, secretID, idempotencyKey string, meta *pb.SecretMetadata) error
	CheckSecretExists(transactionID, userID string) (bool, error)
	ListUserFiles(ctx context.Context, userID string) ([]domain.File, error)
	FileExists(ctx context.Context, userID, secretID string) (bool, error)
	DeleteSecret(ctx context.Context, userID, secretID string) error
	GetSecretMeta(ctx context.Context, userID, secretID string) (*pb.SecretMetadata, error)
}

// FileStorage интерфейс файлового хранилища
type FileStorage interface {
	UploadFile(ctx context.Context, objectName string, reader io.Reader, objectSize int64, contentType string) error
	GetSecret(ctx context.Context, secretID string) (io.Reader, error)
	DeleteSecret(ctx context.Context, secretID string) error
}

// Repositories структура, объединяющая все репозитории
type Repositories struct {
	Authorization
	User
	FileStorage
}

// NewRepositories инициализирует репозитории
func NewRepositories(db *sql.DB, minioStorage *storage.MinioStorage) *Repositories {
	return &Repositories{
		Authorization: postgres.NewAuthRepo(db),
		User:          postgres.NewUserRepo(db),
		FileStorage:   minio.NewUserStorage(minioStorage),
	}
}
