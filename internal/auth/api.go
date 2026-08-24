package auth

import (
	"errors"
	"net/http"
	"strings"
)

func GetAPIKey(headers http.Header) (string, error) {
	authHeader := headers.Get("Authorization")
	if authHeader == "" {
		return "", errors.New("api key not found")
	}

	tokens := strings.Split(authHeader, " ")
	if len(tokens) != 2 || tokens[0] != "ApiKey" {
		return "", errors.New("invalid token format")
	}

	return tokens[1], nil
}
