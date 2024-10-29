package session

import (
	"time"

	redislib "github.com/redis/go-redis/v9"
)

type SessionManager struct {
	RedisClient *redislib.Client
	SessionTTL  time.Duration
}

func NewSessionManager(redisClient *redislib.Client, sessionTTL time.Duration) *SessionManager {
	return &SessionManager{
		RedisClient: redisClient,
		SessionTTL:  sessionTTL,
	}
}

func (s *SessionManager) Close() error {
	if err := s.RedisClient.Close(); err != nil {
		return err
	}
	return nil
}
