package config

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func NewDB(databaseURL string) (*sql.DB, error) {
	databaseConnection, err := sql.Open("pgx", databaseURL)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	databaseConnection.SetConnMaxLifetime(5 * time.Minute)
	databaseConnection.SetMaxIdleConns(5)
	databaseConnection.SetMaxOpenConns(20)

	if err = databaseConnection.Ping(); err != nil {
		_ = databaseConnection.Close()
		return nil, fmt.Errorf("ping database: %w", err)
	}

	return databaseConnection, nil
}
