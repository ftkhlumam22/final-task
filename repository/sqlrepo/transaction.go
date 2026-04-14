package sqlrepo

import (
	"context"
	"fmt"

	"kaktus-consumer/model"
	"kaktus-consumer/repository"
)

func (repository *SQLRepository) WithTransaction(
	requestContext context.Context,
	operation func(sqlTransactionRepository repository.SQLTransactionRepository) error,
) error {
	transaction, err := repository.databaseConnection.BeginTx(requestContext, nil)
	if err != nil {
		return fmt.Errorf("%w: begin transaction: %v", model.ErrConsumeEvent, err)
	}
	defer transaction.Rollback()

	if err = operation(&sqlTransactionRepository{transaction: transaction}); err != nil {
		return err
	}

	if err = transaction.Commit(); err != nil {
		return fmt.Errorf("%w: commit transaction: %v", model.ErrConsumeEvent, err)
	}

	return nil
}
