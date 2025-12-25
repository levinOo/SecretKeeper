package grpc

import (
	"context"
	"secretKeeper/internal/server/service"
	pb "secretKeeper/internal/server/transport/grpc/proto"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type AuthHandler struct {
	pb.UnimplementedProjectServiceServer
	service *service.AuthService
}

func NewAuthHandler(service *service.AuthService) *AuthHandler {
	return &AuthHandler{service: service}
}

func (h *AuthHandler) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	username := req.GetUsername()
	password := req.GetPassword()

	token, err := h.service.CreateUser(ctx, username, password) // access and refresh tokens
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.RegisterResponse{Token: token}, nil
}

func (a *AuthHandler) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	id, token, err := a.service.LoginUser(ctx, req.Username, req.Password) // access and refresh tokens
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}
	return &pb.LoginResponse{Token: token}, nil
}
