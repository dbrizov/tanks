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
      var b = make([]byte, 6)
      rand.Read(b)
      return hex.EncodeToString(b)
}
