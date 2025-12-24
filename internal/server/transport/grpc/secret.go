package grpc

import (
	"secretKeeper/internal/server/service"
	pb "secretKeeper/internal/server/transport/grpc/proto"
)

type ScretHandler struct {
	pb.UnimplementedProjectServiceServer
	service *service.SecretService
}

func NewSecretHandler(service *service.SecretService) *ScretHandler {
	return &ScretHandler{
		service: service,
	}
}

func (h *ScretHandler) CreateSecret(stream pb.ProjectService_CreateSecretServer) error {
	id, err := h.service.CreateSecret(stream)
	if err != nil {
		return err
	}
	return stream.SendAndClose(&pb.CreateSecretResponse{Id: id})
}
