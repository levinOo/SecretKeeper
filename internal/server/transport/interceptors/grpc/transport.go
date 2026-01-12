package myGrpc

// Transport структура транспорта
type Transport struct {
	authHandler *AuthHandler
}

// NewTransport создает новый транспорт
func NewTransport(authHandler *AuthHandler) *Transport {
	return &Transport{authHandler: authHandler}
}
