package grpc

type Transport struct {
	authHandler *AuthHandler
}

func NewTransport(authHandler *AuthHandler) *Transport {
	return &Transport{authHandler: authHandler}
}
