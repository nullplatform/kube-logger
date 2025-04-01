package token

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
)

type TokenData map[string]string

func Decode(token string) (TokenData, error) {
	if token == "" {
		return make(TokenData), nil
	}

	data, err := base64.StdEncoding.DecodeString(token)
	if err != nil {
		return nil, fmt.Errorf("failed to decode token: %w", err)
	}

	var tokenData TokenData

	err = json.Unmarshal(data, &tokenData)

	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal token data: %w", err)
	}

	return tokenData, nil
}

func Encode(data TokenData) (string, error) {
	if len(data) == 0 {
		return "", nil
	}

	jsonData, err := json.Marshal(data)

	if err != nil {
		return "", fmt.Errorf("failed to marshal token data: %w", err)
	}

	return base64.StdEncoding.EncodeToString(jsonData), nil
}
