package response

import "time"

type GeneralSuccess struct {
	Status     string    `json:"status"`
	Data       any       `json:"data"`
	Code       int       `json:"code"`
	AccessTime time.Time `json:"access_time"`
}

type GeneralError struct {
	Status     string    `json:"status"`
	Message    string    `json:"message"`
	Code       int       `json:"code"`
	AccessTime time.Time `json:"access_time"`
}
