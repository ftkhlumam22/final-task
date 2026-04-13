package config

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func NewDB(databaseURL string) (*sql.DB, error) {
	databaseConnection, openError := sql.Open("pgx", databaseURL)
	if openError != nil {
		return nil, fmt.Errorf("open database: %w", openError)
	}

	databaseConnection.SetConnMaxLifetime(5 * time.Minute)
	databaseConnection.SetMaxIdleConns(5)
	databaseConnection.SetMaxOpenConns(20)

	if pingError := databaseConnection.Ping(); pingError != nil {
		_ = databaseConnection.Close()
		return nil, fmt.Errorf("ping database: %w", pingError)
	}

	return databaseConnection, nil
}
