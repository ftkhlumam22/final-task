package model

const (
	DefaultDatabaseURL          = "postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable"
	DefaultRedisURL             = "redis://localhost:6379"
	DefaultRedisPassword        = ""
	DefaultRedisDB              = 0
	DefaultRabbitMQURL          = "amqp://guest:guest@localhost:5672/"
	DefaultRabbitMQExchange     = "kaktus.events"
	DefaultRabbitMQExchangeType = "topic"
	DefaultThreadCreatedKey     = "thread.created"
	DefaultThreadCreatedQueue   = "kaktus.thread.created.queue"
	DefaultRabbitMQMaxRetry     = 3
	DefaultServerAddr           = ":3334"
)

type AppConfig struct {
	DatabaseURL   string
	RedisURL      string
	RedisPassword string
	RedisDB       int
	RabbitMQ      RabbitConfig
	ServerAddr    string
}
