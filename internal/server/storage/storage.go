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
	client     *minio.Client
	BucketName string
}

// Функция создания нового экземпляра MinIO хранилища
func NewMinioStorage(cfg *config.StorageConfig) (*MinioStorage, error) {
	minioClient, err := minio.New(cfg.MinIOEndpoint, &minio.Options{
		Creds:        credentials.NewStaticV4(cfg.MinIOAccessKey, cfg.MinIOSecretKey, ""),
		Secure:       cfg.MinIOUseSSL,
		Transport:    nil,
		Trace:        nil,
		Region:       "us-east-1",
		BucketLookup: minio.BucketLookupAuto,
	})
	if err != nil {
		return nil, err
	}

	exist, err := minioClient.BucketExists(context.Background(), cfg.MinIOBucket)
	if err != nil {
		return nil, err
	}
	if !exist {
		err := minioClient.MakeBucket(context.Background(), cfg.MinIOBucket, minio.MakeBucketOptions{})
		if err != nil {
			return nil, err
		}
	}

	return &MinioStorage{
		client: minioClient,
	}, nil
}

// Функция загрузки файла в хранилище (MinIO)
func (m *MinioStorage) UploadFile(ctx context.Context, objectName string, reader io.Reader, objectSize int64, contentType string) error {
	_, err := m.client.PutObject(ctx, m.BucketName, objectName, reader, objectSize, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return fmt.Errorf("minio upload error: %w", err)
	}
	return nil
}
