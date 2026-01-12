package interceptors

import (
	"context"
	"secretKeeper/pkg/server/auth"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// AuthInterceptor интерцептор авторизации
type AuthInterceptor struct {
	tokenManager *auth.TokenManager
}

// NewAuthInterceptor создает новый интерцептор
func NewAuthInterceptor(tokenManager *auth.TokenManager) *AuthInterceptor {
	return &AuthInterceptor{tokenManager: tokenManager}
}

// Stream интерцептор потоковых вызовов
func (i *AuthInterceptor) Stream() grpc.StreamServerInterceptor {
	return func(
		srv interface{},
		ss grpc.ServerStream,
		info *grpc.StreamServerInfo,
		handler grpc.StreamHandler,
	) error {
		switch info.FullMethod {
		case "/grpc.health.v1.Health/Check",
			"/grpc.health.v1.Health/Watch",
			"/project.AuthService/Login",
			"/project.AuthService/Register",
			"/project.AuthService/Ping":
			return handler(srv, ss)
		}

		ctx := ss.Context()

		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return status.Errorf(codes.Unauthenticated, "метаданные не предоставлены")
		}

		values := md["authorization"]
		if len(values) == 0 {
			return status.Errorf(codes.Unauthenticated, "токен авторизации не предоставлен")
		}

		accesToken := values[0]
		accessToken := strings.TrimPrefix(accesToken, "Bearer ")

		if accessToken == "" {
			return status.Errorf(codes.Unauthenticated, "токен авторизации не предоставлен")
		}

		sub, err := i.tokenManager.Parse(accesToken)
		if err != nil {
			return status.Error(codes.Unauthenticated, "неверный токен авторизации")
		}

		newCtx := context.WithValue(ctx, "user_id", sub)

		wrapped := &wrappedStream{
			ss,
			newCtx,
		}

		return handler(srv, wrapped)
	}
}

type wrappedStream struct {
	grpc.ServerStream
	ctx context.Context
}

// Unary интерцептор унарных вызовов
func (i *AuthInterceptor) Unary() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		// Пропускаем авторизацию для Health Check и публичных методов
		switch info.FullMethod {
		case "/grpc.health.v1.Health/Check",
			"/grpc.health.v1.Health/Watch",
			"/project.AuthService/Login",
			"/project.AuthService/Register":
			return handler(ctx, req)
		}

		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Errorf(codes.Unauthenticated, "метаданные не предоставлены")
		}

		values := md["authorization"]
		if len(values) == 0 {
			return nil, status.Errorf(codes.Unauthenticated, "токен авторизации не предоставлен")
		}

		accessToken := strings.TrimPrefix(values[0], "Bearer ")
		if accessToken == "" {
			return nil, status.Errorf(codes.Unauthenticated, "токен пуст")
		}

		userID, err := i.tokenManager.Parse(accessToken)
		if err != nil {
			return nil, status.Errorf(codes.Unauthenticated, "токен доступа неверен: %v", err)
		}

		newCtx := context.WithValue(ctx, "userID", userID)

		return handler(newCtx, req)
	}

}
