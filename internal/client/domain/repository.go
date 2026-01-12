package domain

// TokenRepository репозиторий для работы с токенами
type TokenRepository interface {
	SaveToken(token string) error
	GetToken() string
	ClearToken() error
}
