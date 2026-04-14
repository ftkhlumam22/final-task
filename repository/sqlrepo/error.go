package sqlrepo

import (
	"errors"
	"fmt"

	"kaktus-consumer/model"

	"github.com/jackc/pgx/v5/pgconn"
)

func wrapSQLError(action string, err error) error {
	if isInvalidPayloadSQLError(err) {
		return fmt.Errorf("%w: %s: %v", model.ErrInvalidEventPayload, action, err)
	}

	return fmt.Errorf("%w: %s: %v", model.ErrConsumeEvent, action, err)
}

func isInvalidPayloadSQLError(err error) bool {
	var postgresError *pgconn.PgError
	if !errors.As(err, &postgresError) {
		return false
	}

	switch postgresError.Code {
	case "23503", // foreign_key_violation
		"23502", // not_null_violation
		"22001", // string_data_right_truncation
		"22P02": // invalid_text_representation
		return true
	default:
		return false
	}
}
