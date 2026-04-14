package sqlrepo

import (
	"context"
	"database/sql"

	"final-task/dto/response"
)

func NewThreadReadRepository(databaseConnection *sql.DB) *ThreadReadRepository {
	return &ThreadReadRepository{db: databaseConnection}
}

func (threadReadRepository *ThreadReadRepository) GetThreadList(limit int, offset int) ([]response.ThreadData, int, error) {
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

	rows, err := threadReadRepository.db.QueryContext(context.Background(), threadListQuery, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	threadList := make([]response.ThreadData, 0)
	for rows.Next() {
		var threadData response.ThreadData
		err := rows.Scan(
			&threadData.ID,
			&threadData.Title,
			&threadData.Description,
			&threadData.CreatedBy,
			&threadData.CreatedAt,
		)
		if err != nil {
			return nil, 0, err
		}

		threadList = append(threadList, threadData)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	var totalForum int
	err = threadReadRepository.db.QueryRowContext(
		context.Background(),
		`SELECT COUNT(1) FROM threads`,
	).Scan(&totalForum)
	if err != nil {
		return nil, 0, err
	}

	return threadList, totalForum, nil
}
