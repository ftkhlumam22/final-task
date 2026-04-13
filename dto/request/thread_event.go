package request

type ThreadEventRequest struct {
	Event       string `json:"event"`
	RequestID   string `json:"request_id"`
	CreatedBy   int64  `json:"created_by"`
	Title       string `json:"title"`
	Description string `json:"description"`
}
