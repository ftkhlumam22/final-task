package sqlrepo

import (
	"database/sql"
	"final-task/model"
)

type UserRepository struct {
	db *sql.DB
}

type ThreadReadRepository struct {
	db *sql.DB
}

type ThreadCommentRow = model.ThreadCommentRow
