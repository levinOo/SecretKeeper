package myGrpc

import (
	"context"
	pb "secretKeeper/internal/proto"
	"secretKeeper/internal/server/service"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// AuthHandler обработчик gRPC авторизации
type AuthHandler struct {
	pb.UnimplementedAuthServiceServer
	service *service.AuthService
}

// NewAuthHandler создает новый обработчик
func NewAuthHandler(service *service.AuthService) *AuthHandler {
	return &AuthHandler{service: service}
}

// Register регистрирует пользователя
func (h *AuthHandler) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	username := req.GetUsername()
	password := req.GetPassword()

	accessToken, refreshToken, err := h.service.CreateUser(ctx, username, password)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.RegisterResponse{AccessToken: accessToken, RefreshToken: refreshToken}, nil
}

// Login выполняет вход
func (a *AuthHandler) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	accessToken, refreshToken, err := a.service.LoginUser(ctx, req.Username, req.Password)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}
	return &pb.LoginResponse{AccessToken: accessToken, RefreshToken: refreshToken}, nil
}
