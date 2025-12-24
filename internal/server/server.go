package server

import (
	"context"
	"os"
	"os/signal"
	"secretKeeper/internal/client/logger"
	"secretKeeper/internal/server/config"
	"secretKeeper/internal/server/db"
	"secretKeeper/internal/server/repository"
	"secretKeeper/internal/server/service"
	"secretKeeper/internal/server/transport/grpc"
	"secretKeeper/pkg/server/auth"
	"secretKeeper/pkg/server/storage"
	"syscall"
	"time"
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

		return
	}

	// Dependencies
	db, err := db.NewDBConnection(&cfg.Postgre)
	if err != nil {
		logger.Error(err)

		return
	}

	tokenManager, err := auth.NewManager(cfg.Auth.JWT.SigningKey)
	if err != nil {
		logger.Error(err)

		return
	}

	storage, err := storage.NewStorage(&cfg.Storage)
	if err != nil {
		// log err

		return
	}

	// Services, Repositories, Handlers
	repos := repository.NewRepositories(db.DB)

	services := service.NewServices(service.Deps{
		Repos:   repos,
		Storage: storage,
	})

	transport := grpc.NewTransport(services, storage, tokenManager)

	go func() {
		if err := transport.Start(cfg.GRPC.Addr); err != nil {
			logger.Error(err)
			return
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)

	<-quit

	const timeout = 5 * time.Second
	ctx, shutdown := context.WithTimeout(context.Background(), timeout)
	defer shutdown()

	// stop transport
	if err := transport.Stop(ctx); err != nil {
		logger.Error(err)
	}

	if err := db.Disconnect(); err != nil {
		logger.Error(err)
	}
}
