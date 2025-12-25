package app

import (
	"net"
	"os"
	"os/signal"
	"secretKeeper/internal/client/logger"
	"secretKeeper/internal/server/config"
	"secretKeeper/internal/server/db"
	"secretKeeper/internal/server/repository"
	"secretKeeper/internal/server/service"
	"secretKeeper/pkg/server/storage"
	"syscall"

	"google.golang.org/grpc"
)

func Run() {
	logger, err := logger.NewLogger()
	if err != nil {
		panic(err)
	}
	defer logger.Sync()

	cfg, err := config.LoadConfig()
	if err != nil {
		logger.Error(err)
	}

	db, err := db.NewDBConnection(&cfg.Postgre)
	if err != nil {
		logger.Error(err)

		return
	}

	// tokenManager

	minioStorage, err := storage.NewClient(&cfg.Storage)
	if err != nil {
		logger.Error(err)

		return
	}

	repos := repository.NewRepositories(db.DB, minioStorage)

	authService := service.NewAuthService(repos.Authorization, &cfg.GRPC, logger)

	services := service.NewServices(authService)

	authHandler := grpc.NewAuthHandler(services.Auth)

	grpcServer := grpc.NewServer()

	listener, err := net.Listen("tcp", cfg.GRPC.Addr)
	if err != nil {
		logger.Error(err)

		return
	}

	if err := grpcServer.Serve(listener); err != nil {
		logger.Error(err)
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)

	<-quit

	grpcServer.GracefulStop()

	if err := db.Disconnect(); err != nil {
		logger.Error(err)
	}
}
