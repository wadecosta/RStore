package main

import (
	"crypto/rand"
	"encoding/hex"
	"io"
)

func GenerateAES256Key() (string, error) {
	key := make([]byte, 32)
	_, err := io.ReadFull(rand.Reader, key)
	if (err != nil) {
		return "", err
	}
	return hex.EncodeToString(key), nil
}
