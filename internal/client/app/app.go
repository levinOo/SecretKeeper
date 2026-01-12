package app

import (
	"context"
	"log"
	"secretKeeper/internal/client/adapter/cache"
	myGrpc "secretKeeper/internal/client/adapter/grpc"
	"secretKeeper/internal/client/config"
	"secretKeeper/internal/client/usecase"
	pkgCache "secretKeeper/pkg/client/cache"
	"secretKeeper/pkg/client/crypto"
	storage "secretKeeper/pkg/client/token_storage"

	"github.com/spf13/viper"
)

// client интерфейс для grpc клиента
type client interface {
	Login(ctx context.Context, username, password string) (string, string, error)
	Register(ctx context.Context, username, password string) (string, string, error)
	Ping(ctx context.Context) error
	Close() error
}

// App основная структура приложения
type App struct {
	AuthService *usecase.AuthService
	UserService *usecase.UserService
	grpcClient  client
}

// NewApp конструктор приложения
func NewApp(v *viper.Viper) (*App, error) {
	cfg, err := config.LoadConfig(v)
	if err != nil {
		return nil, err
	}

	log.Println(cfg)

	tokenStorage := storage.NewTokensStorage()

	grpcClient, err := myGrpc.NewClient(cfg.ServerAddr, cfg.CertPath)
	if err != nil {
		return nil, err
	}

	authService := usecase.NewAuthService(tokenStorage, grpcClient)

	// Hardcoded master key for development testing
	masterKey := []byte("12345678901234567890123456789012")
	cryptoService := crypto.NewCryptoManager(masterKey)

	fileCache, err := pkgCache.NewFileCache("")
	if err != nil {
		return nil, err
	}
	cacheAdapter := cache.NewFileCacheAdapter(fileCache)

	userClient := myGrpc.NewUserClient(grpcClient.SecretService())

	userService := usecase.NewUserService(tokenStorage, userClient, cryptoService, cacheAdapter)

	return &App{
		AuthService: authService,
		UserService: userService,
		grpcClient:  grpcClient,
	}, nil
}

// Close завершает работу приложения
func (a *App) Close() error {
	return a.grpcClient.Close()
}

// Ping проверяет доступность сервера
func (a *App) Ping(ctx context.Context) error {
	return a.grpcClient.Ping(ctx)
}
