package authrepo

import (
	"context"
	"errors"

	"final-task/model"

	"github.com/lib/pq"
)

func (authSQLRepository AuthSQLRepository) CreateUser(
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
	err := authSQLRepository.db.QueryRowContext(
		context.Background(),
		insertUserQuery,
		userName,
		userEmail,
		passwordHash,
	).Scan(&createdUser.ID, &createdUser.Name, &createdUser.Email, &createdUser.CreatedAt)
	if err != nil {
		var postgresErr *pq.Error
		if errors.As(err, &postgresErr) && postgresErr.Code == "23505" {
			return model.User{}, model.ErrEmailAlreadyExists
		}

		return model.User{}, err
	}

	return createdUser, nil
}

func (authSQLRepository AuthSQLRepository) FindUserByEmail(userEmail string) (model.User, error) {
	findUserByEmailQuery := `
		SELECT id, name, email, password_hash, created_at
		FROM users
		WHERE email = $1
	`

	var foundUser model.User
	err := authSQLRepository.db.QueryRowContext(context.Background(), findUserByEmailQuery, userEmail).
		Scan(&foundUser.ID, &foundUser.Name, &foundUser.Email, &foundUser.PasswordHash, &foundUser.CreatedAt)
	if err != nil {
		return model.User{}, err
	}

	return foundUser, nil
}
