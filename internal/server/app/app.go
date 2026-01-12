package app

import (
	"context"
	"crypto/tls"
	"net"
	"os"
	"os/signal"
	pb "secretKeeper/internal/proto"
	"secretKeeper/internal/server/config"
	database "secretKeeper/internal/server/db"
	"secretKeeper/internal/server/repository"
	"secretKeeper/internal/server/service"
	"secretKeeper/internal/server/transport/interceptors"
	myGrpc "secretKeeper/internal/server/transport/interceptors/grpc"
	"secretKeeper/pkg/server/auth"
	"secretKeeper/pkg/server/logger"
	"secretKeeper/pkg/server/storage"
	"syscall"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
)

func Run() {
	// Инициализация логгера
	logger, err := logger.NewLogger()
	if err != nil {
		os.Exit(1)

		return
	}
	defer logger.Sync()

	// Загрузка конфига
	cfg, err := config.LoadConfig()
	if err != nil {
		logger.Error(err)

		return
	}

	// Подключение к БД
	db, err := database.NewDBConnection(&cfg.Postgre)
	if err != nil {
		logger.Error(err)

		return
	}

	// Миграции
	err = database.RunMigrations(db.DB)
	if err != nil {
		logger.Error(err)

		return
	}

	// Инициализация token manager
	tokenManager := auth.NewTokenManager(cfg.Auth.JWT.SigningKey)

	// Подключение к minio
	minioStorage, err := storage.NewClient(&cfg.Storage)
	if err != nil {
		logger.Error(err)

		return
	}

	// Инициализация dependencies слоев
	repos := repository.NewRepositories(db.DB, minioStorage)

	authService := service.NewAuthService(repos.Authorization, &cfg.GRPC, logger, tokenManager)
	userService := service.NewUserService(repos.User, repos.FileStorage, logger)

	services := service.NewServices(authService, userService)

	authHandler := myGrpc.NewAuthHandler(services.Auth)
	userHandler := myGrpc.NewUserHandler(services.User)

	tlsCred, err := generateTLSCreds()
	if err != nil {
		logger.Error(err)

		return
	}

	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(
			interceptors.NewAuthInterceptor(tokenManager).Unary(),
		),

		grpc.ChainStreamInterceptor(
			interceptors.NewAuthInterceptor(tokenManager).Stream(),
		),

		grpc.Creds(tlsCred),
	)

	// Регистрация gRPC сервисов
	pb.RegisterAuthServiceServer(grpcServer, authHandler)
	pb.RegisterSecretServiceServer(grpcServer, userHandler)

	// Регистрируем Health Check сервис
	healthServer := health.NewServer()
	grpc_health_v1.RegisterHealthServer(grpcServer, healthServer)
	healthServer.SetServingStatus("project.AuthService", grpc_health_v1.HealthCheckResponse_SERVING)
	healthServer.SetServingStatus("project.SecretService", grpc_health_v1.HealthCheckResponse_SERVING)

	listener, err := net.Listen("tcp", cfg.GRPC.Addr)
	if err != nil {
		logger.Error(err)

		return
	}

	// Запуск сервера
	go func() {
		logger.Info("Сервер начал работу")
		if err := grpcServer.Serve(listener); err != nil {
			logger.Error(err)
		}
	}()

	// Graceful shutdown
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer stop()

	<-ctx.Done()

	grpcServer.GracefulStop()

	if err := db.Disconnect(); err != nil {
		logger.Error(err)
	}

	logger.Info("Сервер остановлен")
}

// Функция генерации TLS креденшалов
func generateTLSCreds() (credentials.TransportCredentials, error) {
	cert, err := tls.LoadX509KeyPair(
		"certs/server.crt",
		"certs/server.key",
	)
	if err != nil {
		return nil, err
	}

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
	}

	return credentials.NewTLS(tlsConfig), nil
}
