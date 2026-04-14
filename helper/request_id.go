package helper

import (
	"crypto/rand"
	"encoding/hex"
)

func NewRequestID() (string, error) {
	randomBytes := make([]byte, 16)
	if _, err := rand.Read(randomBytes); err != nil {
		return "", err
	}

	return hex.EncodeToString(randomBytes), nil
}
