package helper

import (
	"fmt"

	"kaktus-consumer/model"
)

func BuildThreadDetailCacheKey(threadID int64) string {
	return fmt.Sprintf(model.ThreadDetailCacheKeyPattern, threadID)
}
