package config

import (
	"fmt"

	"kaktus-consumer/model"
)

func LoadAppConfig(environmentFactory model.GetEnvFactory) (model.AppConfig, error) {
	redisDatabase, err := environmentFactory.GetInt("REDIS_DB", model.DefaultRedisDB)
	if err != nil {
		return model.AppConfig{}, fmt.Errorf("REDIS_DB: %w", err)
	}

	rabbitMQMaxRetry, err := environmentFactory.GetInt("RABBITMQ_MAX_RETRY", model.DefaultRabbitMQMaxRetry)
	if err != nil {
		return model.AppConfig{}, fmt.Errorf("RABBITMQ_MAX_RETRY: %w", err)
	}
	if rabbitMQMaxRetry < 0 {
		return model.AppConfig{}, fmt.Errorf("RABBITMQ_MAX_RETRY: must be >= 0")
	}

	return model.AppConfig{
		DatabaseURL:   environmentFactory.GetString("DATABASE_URL", model.DefaultDatabaseURL),
		RedisURL:      environmentFactory.GetString("REDIS_URL", model.DefaultRedisURL),
		RedisPassword: environmentFactory.GetString("REDIS_PASSWORD", model.DefaultRedisPassword),
		RedisDB:       redisDatabase,
		RabbitMQ: model.RabbitConfig{
			URL:      environmentFactory.GetString("RABBITMQ_URL", model.DefaultRabbitMQURL),
			MaxRetry: rabbitMQMaxRetry,
		},
		ServerAddr: environmentFactory.GetString("SERVER_ADDR", model.DefaultServerAddr),
	}, nil
}
