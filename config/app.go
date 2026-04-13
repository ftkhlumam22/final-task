package config

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"final-task/model"
)

func LoadAppConfig(environmentFactory model.GetEnvFactory) (model.AppConfig, error) {
	redisDB, err := environmentFactory.GetInt("REDIS_DB", model.DefaultRedisDB)
	if err != nil {
		return model.AppConfig{}, fmt.Errorf("REDIS_DB: %w", err)
	}

	accessTTLMin, err := environmentFactory.GetInt("JWT_ACCESS_TTL_MIN", model.DefaultJWTAccessTokenTTLMin)
	if err != nil {
		return model.AppConfig{}, fmt.Errorf("JWT_ACCESS_TTL_MIN: %w", err)
	}

	refreshTTLHour, err := environmentFactory.GetInt("JWT_REFRESH_TTL_HOUR", model.DefaultJWTRefreshTokenTTLHours)
	if err != nil {
		return model.AppConfig{}, fmt.Errorf("JWT_REFRESH_TTL_HOUR: %w", err)
	}

	secret := strings.TrimSpace(environmentFactory.GetString("JWT_SECRET_KEY", model.DefaultJTWSecretKey))
	if secret == "" {
		return model.AppConfig{}, errors.New(model.MessageJWTSecretRequired)
	}

	return model.AppConfig{
		DatabaseURL:   environmentFactory.GetString("DATABASE_URL", model.DefaultDatabaseURL),
		RedisURL:      environmentFactory.GetString("REDIS_URL", model.DefaultRedisURL),
		RedisPassword: environmentFactory.GetString("REDIS_PASSWORD", model.DefaultRedisPassword),
		RedisDB:       redisDB,
		RabbitMQ: model.RabbitConfig{
			URL:                      environmentFactory.GetString("RABBITMQ_URL", model.DefaultRabbitMQURL),
			ExchangeName:             environmentFactory.GetString("RABBITMQ_EXCHANGE", model.DefaultRabbitMQExchange),
			ExchangeType:             environmentFactory.GetString("RABBITMQ_EXCHANGE_TYPE", model.DefaultRabbitMQExchangeType),
			ThreadCreatedRoutingKey:  environmentFactory.GetString("RABBITMQ_THREAD_CREATED_KEY", model.DefaultThreadCreatedRoutingKey),
			CommentCreatedRoutingKey: environmentFactory.GetString("RABBITMQ_COMMENT_CREATED_KEY", model.DefaultCommentCreatedRoutingKey),
			ThreadLikedRoutingKey:    environmentFactory.GetString("RABBITMQ_THREAD_LIKED_KEY", model.DefaultThreadLikedRoutingKey),
		},
		ServerAddr:    environmentFactory.GetString("SERVER_ADDR", model.DefaultServerAddr),
		JWT: model.JWTConfig{
			SecretKey:       secret,
			Issuer:          environmentFactory.GetString("JWT_ISSUER", model.DefaultJWTIssuer),
			AccessTokenTTL:  time.Duration(accessTTLMin) * time.Minute,
			RefreshTokenTTL: time.Duration(refreshTTLHour) * time.Hour,
		},
	}, nil
}
