package storage

import (
	"context"
	"fmt"
	"io"
	"secretKeeper/internal/server/config"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// Структура хранилища на базе MinIO
type MinioStorage struct {
	Client     *minio.Client
	BucketName string
}

func NewClient(cfg *config.StorageConfig) (*MinioStorage, error) {
	minioClient, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure: cfg.UseSSL,
	})

	if err != nil {
		return nil, err
	}

	err = minioClient.MakeBucket(context.Background(), cfg.Bucket, minio.MakeBucketOptions{})
	if err != nil {
		exist, errExist := minioClient.BucketExists(context.Background(), cfg.Bucket)
		if errExist == nil && exist {

		} else {
			return nil, err
		}
	}

	return &MinioStorage{
		Client:     minioClient,
		BucketName: cfg.Bucket,
	}, nil
}

// Функция загрузки файла в хранилище (MinIO)
func (m *MinioStorage) UploadFile(ctx context.Context, objectName string, reader io.Reader, objectSize int64, contentType string) error {
	_, err := m.Client.PutObject(ctx, m.BucketName, objectName, reader, objectSize, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return fmt.Errorf("ошибка загрузки minio: %w", err)
	}
	return nil
}
