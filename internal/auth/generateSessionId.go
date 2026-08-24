package auth

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
)

func GenerateSessionId()(string, error) {
	b := make([]byte, 32)

	_, err := rand.Read(b)
	if err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %w", err)
	}

	return hex.EncodeToString(b), nil
}