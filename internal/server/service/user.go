package service

type UserRepository interface {
}

type UserService struct {
	repo UserRepository
	// тут могут быть еще логгер и конфиг
}

func NewUserService(repo UserRepository) *UserService {
	return &UserService{repo: repo}
}
