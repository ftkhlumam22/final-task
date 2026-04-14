package sqlrepo

import "database/sql"

type SQLRepository struct {
	databaseConnection *sql.DB
}

type sqlTransactionRepository struct {
	transaction *sql.Tx
}

func NewSQLRepository(databaseConnection *sql.DB) *SQLRepository {
	return &SQLRepository{
		databaseConnection: databaseConnection,
	}
}
