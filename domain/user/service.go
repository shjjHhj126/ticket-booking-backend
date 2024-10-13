package user

import "github.com/jmoiron/sqlx"

type UserService struct {
	repo *UserRepository
}

func NewUserService(db *sqlx.DB) *UserService {
	return &UserService{
		repo: NewUserRepository(db),
	}
}

func (s *UserService) CreateUser(user *User) error {
	return s.repo.Create(user)
}
