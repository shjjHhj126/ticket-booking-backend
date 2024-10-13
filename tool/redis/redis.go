package redis

import (
	"context"
	"fmt"
	"log"
	"ticket-booking-backend/tool/util"

	redislib "github.com/redis/go-redis/v9"
)

var (
	defaultRedisHost     = "localhost:6379"
	defaultRedisPoolSize = 10
)

func InitRedis() *redislib.Client {
	redisHost := util.GetEnvOrDefault("REDIS_HOST", defaultRedisHost)

	redisPoolSize := util.GetEnvIntOrDefault("REDIS_POOL_SIZE", defaultRedisPoolSize)

	redisClient := redislib.NewClient(&redislib.Options{
		Addr:     redisHost,
		Password: "",
		DB:       0,
		PoolSize: redisPoolSize,
	})

	pong, err := redisClient.Ping(context.Background()).Result()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("redis connected. pool size %d. pong: %s\n", redisClient.Options().PoolSize, pong)

	return redisClient
}
