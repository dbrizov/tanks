package main

import (
	"errors"
	"net/http"
	"os"

	"github.com/golang-jwt/jwt/v5"
)

var jwtSecret = loadJwtSecret()

func loadJwtSecret() []byte {
	var secret = os.Getenv("JWT_SECRET")
	if secret != "" {
		return []byte(secret)
	}

	return []byte("some-very-very-long-random-string-at-least-32-bytes-long")
}

func authenticate(request *http.Request) (PlayerId, error) {
	var tokenString = request.URL.Query().Get("token")
	if tokenString == "" {
		return "", errors.New("missing token")
	}

	var claims jwt.RegisteredClaims
	var _, err = jwt.ParseWithClaims(
		tokenString,
		&claims,
		func(_ *jwt.Token) (any, error) { return jwtSecret, nil },
		jwt.WithValidMethods([]string{"HS256"}),
	)

	if err != nil {
		return "", err
	}

	if claims.Subject == "" {
		return "", errors.New("token has no subject")
	}

	return PlayerId(claims.Subject), nil
}
