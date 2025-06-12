package common

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

var (
	redisClient *redis.Client
	redisOnce   sync.Once
)

// GetRedisClient returns a singleton Redis client instance
func GetRedisClient() (*redis.Client, error) {
	var initErr error

	redisOnce.Do(func() {
		host := os.Getenv("REDIS_HOST")
		if host == "" {
			host = "localhost"
		}

		port := os.Getenv("REDIS_PORT")
		if port == "" {
			port = "6379"
		}

		password := os.Getenv("REDIS_PASSWORD")

		db := 0
		if dbStr := os.Getenv("REDIS_DB"); dbStr != "" {
			if parsed, err := strconv.Atoi(dbStr); err == nil {
				db = parsed
			}
		}

		poolSize := 100
		if poolStr := os.Getenv("REDIS_POOL_SIZE"); poolStr != "" {
			if parsed, err := strconv.Atoi(poolStr); err == nil {
				poolSize = parsed
			}
		}

		redisClient = redis.NewClient(&redis.Options{
			Addr:         fmt.Sprintf("%s:%s", host, port),
			Password:     password,
			DB:           db,
			PoolSize:     poolSize,
			MinIdleConns: 10,
			MaxRetries:   3,
		})

		// Test connection
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := redisClient.Ping(ctx).Err(); err != nil {
			initErr = fmt.Errorf("failed to connect to Redis: %w", err)
			redisClient = nil
		}
	})

	if initErr != nil {
		return nil, initErr
	}

	if redisClient == nil {
		return nil, fmt.Errorf("Redis client not initialized")
	}

	return redisClient, nil
}