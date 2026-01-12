package myGrpc

import (
	"context"
	pb "secretKeeper/internal/proto"
	"secretKeeper/internal/server/service"
)

// UserHandler обработчик gRPC пользователя
type UserHandler struct {
	pb.UnimplementedSecretServiceServer
	service *service.UserService
}

// NewUserHandler создает новый обработчик
func NewUserHandler(service *service.UserService) *UserHandler {
	return &UserHandler{service: service}
}

// CreateSecret создает секрет
func (h *UserHandler) CreateSecret(stream pb.SecretService_CreateSecretServer) error {
	id, status, err := h.service.CreateSecret(stream)
	if err != nil {
		return err
	}

	return stream.SendAndClose(&pb.CreateSecretResponse{Id: id, Status: status})
}

// GetSecret получает секрет
func (h *UserHandler) GetSecret(req *pb.GetSecretRequest, stream pb.SecretService_GetSecretServer) error {
	return h.service.GetSecret(req, stream)
}

// ListSecrets возвращает список
func (h *UserHandler) ListSecrets(ctx context.Context, req *pb.ListSecretsRequest) (*pb.ListSecretsResponse, error) {
	userSecrets, err := h.service.ListSecrets(ctx, req)
	if err != nil {
		return nil, err
	}

	return &pb.ListSecretsResponse{Secrets: userSecrets}, nil
}

// DeleteSecret удаляет секрет
func (h *UserHandler) DeleteSecret(ctx context.Context, req *pb.DeleteSecretRequest) (*pb.DeleteSecretResponse, error) {
	err := h.service.DeleteSecret(ctx, req)
	if err != nil {
		return nil, err
	}

	return &pb.DeleteSecretResponse{Status: "DELETED"}, nil
}
