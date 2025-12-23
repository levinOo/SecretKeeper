package main

import (
	"context"
	"secretKeeper/internal/server/service"
)

// Функция точки входа в приложение сервера
func main() {
	if err := run(); err != nil {
		panic(err)
	}
}

// Функция запуска инициализации и старта сервера
func run() error {
	ctx := context.Background()

	srv, err := service.Setup(ctx)
	if err != nil {
		return err
	}

	return srv.Serve()
}
