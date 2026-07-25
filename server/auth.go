package main

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
)

func authenticate(_ *http.Request) (PlayerId, error) {
	return PlayerId(generateRandomId()), nil
}

func generateRandomId() string {
	var randomBytes = make([]byte, 6)
	rand.Read(randomBytes)
	return hex.EncodeToString(randomBytes)
}
