package repository

import (
	"context"
	"database/sql"

	"final-task/dto/response"
)

func GetThreadList(
	requestContext context.Context,
	databaseConnection *sql.DB,
	limit int,
	offset int,
) ([]response.ThreadData, int, error) {
	threadListQuery := `
		SELECT
			threads.id,
			threads.title,
			threads.description,
			users.name AS created_by,
			threads.created_at
		FROM threads
		JOIN users ON users.id = threads.created_by
		ORDER BY threads.created_at DESC
		LIMIT $1 OFFSET $2
	`

	rows, queryError := databaseConnection.QueryContext(requestContext, threadListQuery, limit, offset)
	if queryError != nil {
		return nil, 0, queryError
	}
	defer rows.Close()

	threadList := make([]response.ThreadData, 0)
	for rows.Next() {
		var threadData response.ThreadData
		scanError := rows.Scan(
			&threadData.ID,
			&threadData.Title,
			&threadData.Description,
			&threadData.CreatedBy,
			&threadData.CreatedAt,
		)
		if scanError != nil {
			return nil, 0, scanError
		}

		threadList = append(threadList, threadData)
	}

	if rowsError := rows.Err(); rowsError != nil {
		return nil, 0, rowsError
	}

	var totalForum int
	countQueryError := databaseConnection.QueryRowContext(
		requestContext,
		`SELECT COUNT(1) FROM threads`,
	).Scan(&totalForum)
	if countQueryError != nil {
		return nil, 0, countQueryError
	}

	return threadList, totalForum, nil
}
