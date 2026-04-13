package repository

import (
	"context"
	"database/sql"
	"errors"
	"final-task/model"

	"github.com/lib/pq"
)

func CreateUser(
	requestContext context.Context,
	databaseConnection *sql.DB,
	userName string,
	userEmail string,
	passwordHash string,
) (model.User, error) {
	insertUserQuery := `
		INSERT INTO users (name, email, password_hash)
		VALUES ($1, $2, $3)
		RETURNING id, name, email, created_at
	`

	var createdUser model.User
	queryError := databaseConnection.QueryRowContext(
		requestContext,
		insertUserQuery,
		userName,
		userEmail,
		passwordHash,
	).Scan(&createdUser.ID, &createdUser.Name, &createdUser.Email, &createdUser.CreatedAt)
	if queryError != nil {
		var postgresError *pq.Error
		if errors.As(queryError, &postgresError) && postgresError.Code == "23505" {
			return model.User{}, model.ErrEmailAlreadyExists
		}

		return model.User{}, queryError
	}

	return createdUser, nil
}

func FindUserByEmail(
	requestContext context.Context,
	databaseConnection *sql.DB,
	userEmail string,
) (model.User, error) {
	findUserByEmailQuery := `
		SELECT id, name, email, password_hash, created_at
		FROM users
		WHERE email = $1
	`

	var foundUser model.User
	queryError := databaseConnection.QueryRowContext(requestContext, findUserByEmailQuery, userEmail).
		Scan(&foundUser.ID, &foundUser.Name, &foundUser.Email, &foundUser.PasswordHash, &foundUser.CreatedAt)
	if queryError != nil {
		return model.User{}, queryError
	}

	return foundUser, nil
}
