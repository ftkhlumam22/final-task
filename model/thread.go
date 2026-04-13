package model

import "time"

type Thread struct {
	ID          int64
	Title       string
	Description string
	CreatedBy   int64
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
