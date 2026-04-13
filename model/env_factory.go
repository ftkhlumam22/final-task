package model

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

type GetEnvFactory interface {
	GetString(key string, fallback string) string
	GetInt(key string, fallback int) (int, error)
}

type OSEnvFactory struct{}

func NewOSEnvFactory() OSEnvFactory {
	return OSEnvFactory{}
}

func (OSEnvFactory) GetString(key string, fallback string) string {
	environmentValue := strings.TrimSpace(os.Getenv(key))
	if environmentValue == "" {
		return fallback
	}

	return environmentValue
}

func (OSEnvFactory) GetInt(key string, fallback int) (int, error) {
	environmentValue := strings.TrimSpace(os.Getenv(key))
	if environmentValue == "" {
		return fallback, nil
	}

	parsedValue, parseError := strconv.Atoi(environmentValue)
	if parseError != nil {
		return 0, fmt.Errorf("%s must be integer: %w", key, parseError)
	}

	return parsedValue, nil
}
