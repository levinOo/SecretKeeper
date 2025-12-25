package minio

import (
	"context"
	"fmt"
	"io"
	"secretKeeper/pkg/server/storage"

	"github.com/minio/minio-go/v7"
)

type UserStorage struct {
	client *storage.MinioStorage
}

func NewUserStorage(client *storage.MinioStorage) *UserStorage {
	return &UserStorage{client: client}
}

func (s *UserStorage) UploadFile(ctx context.Context, objectName string, reader io.Reader, objectSize int64, contentType string) error {
	_, err := s.client.Client.PutObject(ctx, s.client.BucketName, objectName, reader, objectSize, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return fmt.Errorf("minio upload error: %w", err)
	}
	return nil
}
