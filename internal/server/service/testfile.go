package service

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"secretKeeper/internal/client/logger"
	"secretKeeper/internal/server/config"
	"secretKeeper/internal/server/db"
	pb "secretKeeper/internal/server/proto"
	"secretKeeper/internal/server/storage"
	"syscall"
	"time"

	"strings"

	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// Структура gRPC сервера
type Server struct {
	pb.UnimplementedProjectServiceServer
	grpcServer *grpc.Server
	logger     *zap.SugaredLogger
	config     *config.Config
	db         *db.DB
	storage    *storage.MinioStorage
}

// Структура зависимостей сервера
type Dependencies struct {
	Logger  *zap.SugaredLogger
	Config  *config.Config
	DB      *db.DB
	Storage *storage.MinioStorage
}

// Функция создания нового экземпляра сервера
func NewServer(deps Dependencies) *Server {
	srv := &Server{
		logger:  deps.Logger,
		config:  deps.Config,
		db:      deps.DB,
		storage: deps.Storage,
	}

	opts := []grpc.ServerOption{
		grpc.UnaryInterceptor(srv.UnaryAuthInterceptor),
		grpc.StreamInterceptor(srv.StreamAuthInterceptor),
	}

	grpcServer := grpc.NewServer(opts...)
	srv.grpcServer = grpcServer

	// Регистрация gRPC сервиса
	pb.RegisterProjectServiceServer(grpcServer, srv)

	return srv
}

// Функция интерцептора для унарных RPC вызовов (проверка авторизации)
func (s *Server) UnaryAuthInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	// Skip auth for public methods
	if info.FullMethod == "/project.ProjectService/Login" || info.FullMethod == "/project.ProjectService/Register" {
		return handler(ctx, req)
	}

	userID, err := s.authorize(ctx)
	if err != nil {
		return nil, err
	}

	ctx = context.WithValue(ctx, "user_id", userID)
	return handler(ctx, req)
}

// Функция интерцептора для потоковых RPC вызовов (проверка авторизации)
func (s *Server) StreamAuthInterceptor(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	// Skip auth for public methods
	if info.FullMethod == "/project.ProjectService/Login" || info.FullMethod == "/project.ProjectService/Register" {
		return handler(srv, ss)
	}

	userID, err := s.authorize(ss.Context())
	if err != nil {
		return err
	}

	wrapped := &wrappedStream{ss, context.WithValue(ss.Context(), "user_id", userID)}
	return handler(srv, wrapped)
}

func (s *Server) authorize(ctx context.Context) (string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", status.Error(codes.Unauthenticated, "metadata is not provided")
	}

	values := md["authorization"]
	if len(values) == 0 {
		return "", status.Error(codes.Unauthenticated, "authorization token is not provided")
	}

	accessToken := values[0]
	if strings.HasPrefix(accessToken, "Bearer ") {
		accessToken = strings.TrimPrefix(accessToken, "Bearer ")
	}

	// Use key from Config
	key := []byte(s.config.JWT.SecretKey)

	token, err := jwt.ParseWithClaims(accessToken, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return key, nil
	})

	if err != nil {
		return "", status.Errorf(codes.Unauthenticated, "invalid token: %v", err)
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return "", status.Error(codes.Unauthenticated, "invalid token claims")
	}

	return claims.UserID, nil
}

type wrappedStream struct {
	grpc.ServerStream
	ctx context.Context
}

func (w *wrappedStream) Context() context.Context {
	return w.ctx
}

// Функция инициализации и настройки сервера
func Setup(ctx context.Context) (*Server, error) {
	logger, err := logger.NewLogger()
	if err != nil {
		return nil, fmt.Errorf("failed to init logger: %w", err)
	}
	defer logger.Sync()

	cfg, err := config.LoadConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	database, err := db.NewDBConnection(cfg.Postgre)
	if err != nil {
		return nil, fmt.Errorf("failed to init database: %w", err)
	}

	for i := 0; i < 5; i++ {
		if err := database.HealthCheck(ctx); err != nil {
			time.Sleep(2 * time.Second)
			continue
		}
	}

	minioClient, err := storage.NewMinioStorage(cfg.Storage)
	if err != nil {
		return nil, fmt.Errorf("failed to init minio: %w", err)
	}

	deps := Dependencies{
		Logger:  logger,
		Config:  cfg,
		DB:      database,
		Storage: minioClient,
	}

	srv := NewServer(deps)

	logger.Info("Server setup completed successfully")

	return srv, nil
}

// Функция запуска gRPC сервера
func (s *Server) Serve() error {

	go func() {
		if err := s.Start(); err != nil && err != grpc.ErrServerStopped {
			s.logger.Errorf("Server failed to start: %v", err)
			os.Exit(1)
		}
	}()

	s.logger.Infof("Server started on port %s", s.config.GRPC.Addr)

	quit := make(chan os.Signal, 1)

	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	sig := <-quit
	s.logger.Infof("Received shutdown signal: %v. Starting graceful shutdown...", sig)

	if err := s.Stop(); err != nil {
		return err
	}
	s.logger.Infof("Server stopped successfully")

	return nil
}

// Функция начала прослушивания порта
func (s *Server) Start() error {
	lis, err := net.Listen("tcp", s.config.GRPC.Addr)
	if err != nil {
		return err
	}

	if err := s.grpcServer.Serve(lis); err != nil {
		return err
	}

	return nil
}

// Функция остановки сервера и закрытия соединений
func (s *Server) Stop() error {
	if err := s.db.Close(); err != nil {
		return err
	}
	return nil
}
