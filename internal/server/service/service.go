package service

type Services struct {
	Auth *AuthService
	User *UserService
}

func NewServices(auth *AuthService, user *UserService) *Services {
	return &Services{Auth: auth, User: user}
}
