package request

type CreateThread struct {
	Title       string `json:"title"`
	Description string `json:"description"`
}

type LikeThread struct {
	ThreadID int64 `json:"thread_id"`
}
