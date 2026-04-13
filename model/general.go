package model

const (
	DefaultDatabaseURL             = "postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable"
	DefaultRedisURL                = "redis://localhost:6379"
	DefaultRedisPassword           = ""
	DefaultRedisDB                 = 0
	DefaultRabbitMQURL             = "amqp://guest:guest@localhost:5672/"
	DefaultRabbitMQExchange        = "kaktus.events"
	DefaultRabbitMQExchangeType    = "topic"
	DefaultThreadCreatedRoutingKey = "thread.created"
	DefaultCommentCreatedRoutingKey = "comment.created"
	DefaultThreadLikedRoutingKey    = "thread.liked"
	DefaultServerAddr              = ":3333"
	DefaultJTWSecretKey            = ""
	DefaultJWTIssuer               = "kaktus"
	DefaultJWTAccessTokenTTLMin    = 15
	DefaultJWTRefreshTokenTTLHours = 48
	DefaultThreadListCacheTTLSec   = 60
	DefaultThreadDetailCacheTTLSec = 60
	DefaultThreadListPage          = 1
	DefaultThreadListLimit         = 10
)

type AppConfig struct {
	DatabaseURL   string
	RedisURL      string
	RedisPassword string
	RedisDB       int
	RabbitMQ      RabbitConfig
	ServerAddr    string
	JWT           JWTConfig
}
