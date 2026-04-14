package model

import "errors"

var (
	ErrEmailAlreadyExists = errors.New(MessageEmailAlreadyExists)
	ErrInvalidCredential  = errors.New(MessageInvalidCredential)
	ErrInvalidRefresh     = errors.New(MessageInvalidRefresh)
	ErrPublishEvent       = errors.New(MessageFailedPublishEvent)
	ErrThreadNotFound     = errors.New(MessageThreadNotFound)
	ErrConsumeEvent       = errors.New(MessageFailedConsumeEvent)
)
