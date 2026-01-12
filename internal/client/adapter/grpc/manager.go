package myGrpc // grpcadapter

import (
	"context"
	"fmt"
	"secretKeeper/internal/client/domain"
	pb "secretKeeper/internal/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

type AuthProvider interface {
	Register(ctx context.Context, username, password string) (string, string, error)
	Login(ctx context.Context, username, password string) (string, string, error)
}

type UserProvider interface {
	CreateSecret(ctx context.Context, secret *domain.Secret) (string, string, error)
	GetList(ctx context.Context) ([]domain.SecretMeta, error)
	GetSecret(ctx context.Context, id string) (*domain.Secret, error)
	DeleteSecret(ctx context.Context, id string) (string, error)
}

// Manager управляет gRPC соединениями
type Manager struct {
	conn *grpc.ClientConn

	Auth AuthProvider
	User UserProvider
	// Health
}

// NewManager создает нового менеджера соединений
func NewManager(addr, certFile string) (*Manager, error) {
	opts := []grpc.DialOption{}

	if certFile != "" {
		creds, err := credentials.NewClientTLSFromFile(certFile, "")
		if err != nil {
			return nil, err
		}

		opts = append(opts, grpc.WithTransportCredentials(creds))
	} else {
		return nil, fmt.Errorf("требуется файл TLS сертификата")
	}

	conn, err := grpc.NewClient(addr, opts...)
	if err != nil {
		return nil, err
	}

	return &Manager{
		conn: conn,
		Auth: NewAuthClient(pb.NewAuthServiceClient(conn)),
		User: NewUserClient(pb.NewSecretServiceClient(conn)),
		// Health
	}, nil
}

func (c *Manager) Close() error {
	return c.conn.Close()
}
