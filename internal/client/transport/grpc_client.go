package transport

import (
	"context"
	"fmt"
	"secretKeeper/internal/client/config"
	"secretKeeper/internal/client/logger"
	"secretKeeper/internal/client/repository"
	pb "secretKeeper/internal/server/proto"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

// Структура клиента, объединяющая gRPC, конфиг и токенов
type Client struct {
	grpc         *GRPCClient
	logger       *zap.SugaredLogger
	tokenStorage *repository.UserTokens
	cfg          *config.Config
}

// Структура обертки над gRPC соединением
type GRPCClient struct {
	conn   *grpc.ClientConn
	client pb.ProjectServiceClient
	addr   string
}

// Функция проверки авторизации пользователя
func (g *Client) IsAuthorized() bool {
	if g.tokenStorage.AccessToken != "" {
		return true
	}
	return false
}

// Функция выхода из системы (сброс токена)
func (g *Client) Logout() {
	g.tokenStorage.AccessToken = ""
}

// Функция настройки и инициализации клиента
func SetupClient(ctx context.Context) (*Client, error) {
	logger, err := logger.NewLogger()
	if err != nil {
		return nil, err
	}

	cfg, err := config.LoadConfig()
	if err != nil {
		return nil, err
	}

	tokenStorage := repository.NewUserTokensStorage()

	srv, err := NewGRPCClient(cfg.GRPC.Addr)
	if err != nil {
		return nil, err
	}

	return &Client{
		grpc:         srv,
		logger:       logger,
		tokenStorage: tokenStorage,
		cfg:          cfg,
	}, nil
}

// Функция создания нового gRPC клиента
func NewGRPCClient(addr string) (*GRPCClient, error) {
	conn, err := grpc.NewClient(
		addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create gRPC client for %s: %w", addr, err)
	}

	return &GRPCClient{
		conn:   conn,
		client: pb.NewProjectServiceClient(conn),
		addr:   addr,
	}, nil
}

// Функция регистрации через gRPC
func (g *Client) Register(ctx context.Context, username, password string) error {
	resp, err := g.grpc.client.Register(ctx, &pb.RegisterRequest{
		Username: username,
		Password: password,
	})

	if err != nil {
		return fmt.Errorf("register RPC error: %w", err)
	}
	// Save tokens
	g.tokenStorage.LoadTokens(resp.AccessToken, resp.RefreshToken)
	g.logger.Info("Registered successfully")

	return nil
}

// Функция входа через gRPC
func (g *Client) Login(ctx context.Context, username, password string) error {
	resp, err := g.grpc.client.Login(ctx, &pb.LoginRequest{
		Username: username,
		Password: password,
	})

	if err != nil {
		return fmt.Errorf("login failed: %w", err)
	}

	// Save tokens
	g.tokenStorage.LoadTokens(resp.AccessToken, resp.RefreshToken)
	g.logger.Info("Logged in successfully")

	return nil
}

// Функция создания секрета (отправка метаданных и данных)
func (g *Client) CreateSecret(ctx context.Context, secretMeta *pb.SecretMetadata, data []byte) (*pb.CreateSecretResponse, error) {
	token, err := g.tokenStorage.GetToken()
	if err != nil {
		return nil, fmt.Errorf("auth error: %w", err)
	}

	md := metadata.Pairs(
		"authorization", "Bearer "+token, // JWT Token
		"idempotency-key", uuid.New().String(), // id запроса
	)
	ctx = metadata.NewOutgoingContext(ctx, md)

	// Открываем поток к серверу
	stream, err := g.grpc.client.CreateSecret(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to open stream: %w", err)
	}

	reqMeta := &pb.CreateSecretRequest{
		Data: &pb.CreateSecretRequest_Metadata{
			Metadata: secretMeta,
		},
	}

	if err := stream.Send(reqMeta); err != nil {
		// Пытаемся получить детали ошибки от сервера (например, "invalid token")
		// Игнорируем ошибку Recv, нам важнее ошибка Send
		_ = stream.CloseSend()
		return nil, fmt.Errorf("failed to send meta %w", err)
	}

	const chunkSize = 64 * 1024
	dataLen := len(data)

	for i := 0; i < dataLen; i += chunkSize {
		end := i + chunkSize
		if end > dataLen {
			end = dataLen
		}

		reqChunk := &pb.CreateSecretRequest{
			Data: &pb.CreateSecretRequest_ChunkData{
				ChunkData: data[i:end],
			},
		}

		if err := stream.Send(reqChunk); err != nil {
			_, recvErr := stream.CloseAndRecv()
			if recvErr != nil {
				return nil, fmt.Errorf("stream error: %v (orig: %w)", recvErr, err)
			}
			return nil, fmt.Errorf("failed to send chunk: %w", err)
		}

	}

	return stream.CloseAndRecv()
}

func (g *GRPCClient) GetSecret() {

}

func (g *GRPCClient) DeleteSecret() {

}

// Функция получения списка секретов
func (g *GRPCClient) GetSecretList(ctx context.Context, token string, req *pb.GetListRequest) (*pb.GetListResponse, error) {
	ctx = metadata.AppendToOutgoingContext(ctx, "authorization", "Bearer "+token)

	resp, err := g.client.GetSecretList(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get secret list: %w", err)
	}

	return resp, nil
}
