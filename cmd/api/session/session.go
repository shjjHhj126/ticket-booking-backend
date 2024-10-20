package session

import (
	"time"

	redislib "github.com/redis/go-redis/v9"
)

type SessionManager struct {
	redisClient *redislib.Client
	sessionTTL  time.Duration
}

func NewSessionManager(redisClient *redislib.Client, sessionTTL time.Duration) *SessionManager {
	return &SessionManager{
		redisClient: redisClient,
		sessionTTL:  sessionTTL,
	}
}

func (s *SessionManager) Close() error {
	if err := s.redisClient.Close(); err != nil {
		return err
	}
	return nil
}
