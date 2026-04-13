package config

import (
	"fmt"

	"kaktus-consumer/model"
)

func LoadAppConfig(environmentFactory model.GetEnvFactory) (model.AppConfig, error) {
	redisDatabase, redisDBError := environmentFactory.GetInt("REDIS_DB", model.DefaultRedisDB)
	if redisDBError != nil {
		return model.AppConfig{}, fmt.Errorf("REDIS_DB: %w", redisDBError)
	}

	rabbitMQMaxRetry, maxRetryError := environmentFactory.GetInt("RABBITMQ_MAX_RETRY", model.DefaultRabbitMQMaxRetry)
	if maxRetryError != nil {
		return model.AppConfig{}, fmt.Errorf("RABBITMQ_MAX_RETRY: %w", maxRetryError)
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
			URL:                     environmentFactory.GetString("RABBITMQ_URL", model.DefaultRabbitMQURL),
			ExchangeName:            environmentFactory.GetString("RABBITMQ_EXCHANGE", model.DefaultRabbitMQExchange),
			ExchangeType:            environmentFactory.GetString("RABBITMQ_EXCHANGE_TYPE", model.DefaultRabbitMQExchangeType),
			ThreadCreatedRoutingKey: environmentFactory.GetString("RABBITMQ_THREAD_CREATED_KEY", model.DefaultThreadCreatedKey),
			ThreadCreatedQueue:      environmentFactory.GetString("RABBITMQ_THREAD_CREATED_QUEUE", model.DefaultThreadCreatedQueue),
			MaxRetry:                rabbitMQMaxRetry,
		},
		ServerAddr: environmentFactory.GetString("SERVER_ADDR", model.DefaultServerAddr),
	}, nil
}
