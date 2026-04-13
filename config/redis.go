package config

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

func NewRedis(redisURL, password string, redisDB int) (*redis.Client, error) {
	var (
		opts *redis.Options
		err  error
	)

	if strings.Contains(redisURL, "://") {
		opts, err = redis.ParseURL(redisURL)
		if err != nil {
			return nil, fmt.Errorf("parse redis url: %w", err)
		}
	} else {
		opts = &redis.Options{Addr: redisURL}
	}

	if strings.TrimSpace(password) != "" {
		opts.Password = password
	}
	opts.DB = redisDB

	client := redis.NewClient(opts)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err = client.Ping(ctx).Err(); err != nil {
		_ = client.Close()
		return nil, fmt.Errorf("ping redis: %w", err)
	}

	return client, nil
}
