package minio

import (
	"context"
	"fmt"
	"io"
	"secretKeeper/pkg/server/storage"

	"github.com/minio/minio-go/v7"
)

// UserStorage репозиторий для работы с файлами пользователя в MinIO
type UserStorage struct {
	client *storage.MinioStorage
}

// NewUserStorage создает новый репозиторий UserStorage
func NewUserStorage(client *storage.MinioStorage) *UserStorage {
	return &UserStorage{client: client}
}

// UploadFile загружает файл в хранилище
func (s *UserStorage) UploadFile(ctx context.Context, objectName string, reader io.Reader, objectSize int64, contentType string) error {
	_, err := s.client.Client.PutObject(ctx, s.client.BucketName, objectName, reader, objectSize, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return fmt.Errorf("ошибка загрузки minio: %w", err)
	}
	return nil
}

// DeleteSecret удаляет секрет
func (s *UserStorage) DeleteSecret(ctx context.Context, secretID string) error {
	err := s.client.Client.RemoveObject(ctx, s.client.BucketName, secretID, minio.RemoveObjectOptions{})
	if err != nil {
		return fmt.Errorf("ошибка удаления minio: %w", err)
	}
	return nil
}

// GetSecret получает секрет (файл)
func (s *UserStorage) GetSecret(ctx context.Context, secretID string) (io.Reader, error) {
	object, err := s.client.Client.GetObject(ctx, s.client.BucketName, secretID, minio.GetObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("ошибка получения minio: %w", err)
	}
	return object, nil
}
