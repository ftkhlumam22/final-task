package config

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

func NewRedis(redisURL string, redisPassword string, redisDB int) (*redis.Client, error) {
	redisOptions := &redis.Options{
		Addr:     redisURL,
		Password: redisPassword,
		DB:       redisDB,
	}

	if strings.HasPrefix(redisURL, "redis://") || strings.HasPrefix(redisURL, "rediss://") {
		parsedRedisOptions, parseError := redis.ParseURL(redisURL)
		if parseError != nil {
			return nil, fmt.Errorf("parse redis url: %w", parseError)
		}

		if redisPassword != "" {
			parsedRedisOptions.Password = redisPassword
		}
		parsedRedisOptions.DB = redisDB
		redisOptions = parsedRedisOptions
	}

	redisClient := redis.NewClient(redisOptions)

	pingContext, cancelPing := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelPing()

	if pingError := redisClient.Ping(pingContext).Err(); pingError != nil {
		_ = redisClient.Close()
		return nil, fmt.Errorf("ping redis: %w", pingError)
	}

	return redisClient, nil
}
