package config

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq"
)

func NewDB(databaseURL string) (*sql.DB, error) {
	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		return nil, fmt.Errorf("open postgres: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err = db.PingContext(ctx)

	if err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("ping postgres: %w", err)
	}

	return db, nil
}
