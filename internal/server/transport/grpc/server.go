package grpc

import (
	"net"

	"google.golang.org/grpc"
)

func (t *Transport) Start(addr string) error {
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}

	grpcServer := grpc.NewServer()

	// Здесь нужно зарегистрировать ваши сервисы
	// pbService.RegisterYourServiceServer(grpcServer, &YourServiceImpl{})

	if err = grpcServer.Serve(listener); err != nil {
		return err
	}

	return nil
}
