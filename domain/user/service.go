package user

import (
	"database/sql"
	"ticket-booking-backend/cmd/api/session"

	redislib "github.com/redis/go-redis/v9"
)

type UserService struct {
	repo           *UserRepository
	RedisClient    *redislib.Client
	SessionManager *session.SessionManager
}

func NewUserService(db *sql.DB, redisClient *redislib.Client, SessionManager *session.SessionManager) *UserService {
	return &UserService{
		repo:           NewUserRepository(db),
		RedisClient:    redisClient,
		SessionManager: SessionManager,
	}
}

func (s *UserService) CreateUser(user *User) error {
	return s.repo.Create(user)
}

func (s *UserService) GetUserByEmail(email string) (User, error) {
	return s.repo.GetUserByEmail(email)
}
