package auth

import (
	"errors"
	"net/http"
	"strings"
)

func GetAPIKey(headers http.Header) (string, error) {
	apiKeyString := headers.Get("Authorization")
	if apiKeyString == "" {
		return "", errors.New("no api key found")
	}
	
	apiKeySplit := strings.Split(apiKeyString, " ")
	if len(apiKeySplit) != 2 || apiKeySplit[0] != "ApiKey" {
		return "", errors.New("malformed authorization header")
	}

	return apiKeySplit[1], nil
}