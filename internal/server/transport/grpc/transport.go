package grpc

import (
	"secretKeeper/internal/server/service"
	"secretKeeper/pkg/server/auth"
	"secretKeeper/pkg/server/storage"
)

type Transport struct {
	service      *service.Services
	storage      *storage.MinioStorage
	tokenManager *auth.Manager
}

func NewTransport(service *service.Services, storage *storage.MinioStorage, tokenManager *auth.Manager) *Transport {
	return &Transport{
		service:      service,
		storage:      storage,
		tokenManager: tokenManager,
	}
}
