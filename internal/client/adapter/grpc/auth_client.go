package myGrpc

import (
	"context"
	pb "secretKeeper/internal/proto"
)

// AuthClient клиент авторизации
type AuthClient struct {
	client pb.AuthServiceClient
}

// NewAuthClient создает новый AuthClient
func NewAuthClient(client pb.AuthServiceClient) *AuthClient {
	return &AuthClient{
		client: client,
	}
}

// Register регистрирует пользователя
func (c *AuthClient) Register(ctx context.Context, username, password string) (string, string, error) {
	req := &pb.RegisterRequest{
		Username: username,
		Password: password,
	}

	resp, err := c.client.Register(ctx, req)
	if err != nil {
		return "", "", err
	}

	return resp.AccessToken, resp.RefreshToken, nil
}

// Login выполняет вход
func (c *AuthClient) Login(ctx context.Context, username, password string) (string, string, error) {
	req := &pb.LoginRequest{
		Username: username,
		Password: password,
	}

	resp, err := c.client.Login(ctx, req)
	if err != nil {
		return "", "", err
	}

	return resp.AccessToken, resp.RefreshToken, nil
}
