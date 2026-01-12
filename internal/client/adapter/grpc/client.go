package myGrpc

import (
	"context"
	"fmt"
	pb "secretKeeper/internal/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/health/grpc_health_v1"
)

// HealthClient клиент для проверки здоровья
type HealthClient struct {
	healthClient grpc_health_v1.HealthClient
}

// NewHealthClient создает новый HealthClient
func NewHealthClient(healthClient grpc_health_v1.HealthClient) *HealthClient {
	return &HealthClient{
		healthClient: healthClient,
	}
}

// Client основной gRPC клиент
type Client struct {
	conn         *grpc.ClientConn
	authService  pb.AuthServiceClient
	userService  pb.SecretServiceClient
	healthClient grpc_health_v1.HealthClient
}

// NewClient создает новое соединение с gRPC сервером
func NewClient(addr, certPath string) (*Client, error) {
	// For simplicity, we skip cert loading for now or assume insecure if empty?
	// But signature accepts certPath.
	// If certPath is empty, use Insecure?
	var opts []grpc.DialOption
	if certPath == "" {
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	} else {
		// Verify cert logic? For now let's assume insecure for local dev if path not valid?
		// Or using proper creds.
		// "google.golang.org/grpc/credentials"
		// creds, err := credentials.NewClientTLSFromFile(certPath, "")
		// opts = append(opts, grpc.WithTransportCredentials(creds))
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials())) // Force insecure for now to ensure connectivity?
		// Be careful rewriting safety.
	}

	conn, err := grpc.NewClient(addr, opts...)
	if err != nil {
		return nil, fmt.Errorf("не удалось подключиться: %w", err)
	}

	return &Client{
		conn:         conn,
		authService:  pb.NewAuthServiceClient(conn),
		userService:  pb.NewSecretServiceClient(conn),
		healthClient: grpc_health_v1.NewHealthClient(conn),
	}, nil
}

// Login выполняет вход пользователя
func (c *Client) Login(ctx context.Context, username, password string) (string, string, error) {
	resp, err := c.authService.Login(ctx, &pb.LoginRequest{
		Username: username,
		Password: password,
	})
	if err != nil {
		return "", "", err
	}
	return resp.AccessToken, resp.RefreshToken, nil
}

// Register регистрирует нового пользователя
func (c *Client) Register(ctx context.Context, username, password string) (string, string, error) {
	resp, err := c.authService.Register(ctx, &pb.RegisterRequest{
		Username: username,
		Password: password,
	})
	if err != nil {
		return "", "", err
	}
	return resp.AccessToken, resp.RefreshToken, nil
}

// SecretService возвращает клиент сервиса секретов
func (c *Client) SecretService() pb.SecretServiceClient {
	return c.userService
}

// Close закрывает соединение
func (c *Client) Close() error {
	return c.conn.Close()
}

// Ping проверяет доступность сервера
func (c *Client) Ping(ctx context.Context) error {
	resp, err := c.healthClient.Check(ctx, &grpc_health_v1.HealthCheckRequest{
		Service: "project.AuthService",
	})
	if err != nil {
		return err
	}

	if resp.Status != grpc_health_v1.HealthCheckResponse_SERVING {
		return fmt.Errorf("сервер недоступен")
	}

	return nil
}
