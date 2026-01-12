package service

// Services структура сервисов
type Services struct {
	Auth *AuthService
	User *UserService
}

// NewServices - конструктор сервисов
func NewServices(auth *AuthService, user *UserService) *Services {
	return &Services{
		Auth: auth,
		User: user,
	}
}
