package helper

import (
	"crypto/rand"
	"encoding/hex"
)

func NewRequestID() (string, error) {
	randomBytes := make([]byte, 16)
	if _, readError := rand.Read(randomBytes); readError != nil {
		return "", readError
	}

	return hex.EncodeToString(randomBytes), nil
}
