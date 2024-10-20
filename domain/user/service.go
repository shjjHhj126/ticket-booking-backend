package user

import "database/sql"

type UserService struct {
	repo *UserRepository
}

func NewUserService(db *sql.DB) *UserService {
	return &UserService{
		repo: NewUserRepository(db),
	}
}

func (s *UserService) CreateUser(user *User) error {
	return s.repo.Create(user)
}
