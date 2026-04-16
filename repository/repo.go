package repository

import (
	"database/sql"

	authrepo "final-task/repository/auth"
	threadrepo "final-task/repository/thread"

	"github.com/redis/go-redis/v9"
)

type Repositories struct {
	Auth   authrepo.Repositories
	Thread threadrepo.Repositories
}

func New(databaseConnection *sql.DB, redisClient *redis.Client) Repositories {
	return Repositories{
		Auth:   authrepo.NewRepositories(databaseConnection, redisClient),
		Thread: threadrepo.NewRepositories(databaseConnection, redisClient),
	}
}
