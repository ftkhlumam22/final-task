package helper

import (
	"fmt"
	"net/http"
	"strings"
	"time"
)

func GetRequestID(request *http.Request) string {
	requestID := strings.TrimSpace(request.Header.Get("X-Request-ID"))
	if requestID != "" {
		return requestID
	}

	return fmt.Sprintf("request-%d", time.Now().UnixNano())
}
