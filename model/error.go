package model

import "errors"

var (
	ErrInvalidEventPayload = errors.New(MessageInvalidEventPayload)
	ErrConsumeEvent        = errors.New(MessageFailedConsumeEvent)
)
