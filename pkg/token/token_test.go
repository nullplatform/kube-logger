package token

import (
	"reflect"
	"testing"
)

func TestEncodeDecode(t *testing.T) {
	testCases := []struct {
		name      string
		tokenData TokenData
	}{
		{
			name:      "empty data",
			tokenData: TokenData{},
		},
		{
			name: "single pod",
			tokenData: TokenData{
				"pod-1": "2025-04-01T15:44:44.534Z",
			},
		},
		{
			name: "multiple pods",
			tokenData: TokenData{
				"pod-1": "2025-04-01T15:44:44.534Z",
				"pod-2": "2025-04-01T15:44:45.123Z",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			encoded, err := Encode(tc.tokenData)
			if err != nil {
				t.Fatalf("Encode error: %v", err)
			}

			if len(tc.tokenData) == 0 {
				if encoded != "" {
					t.Errorf("Expected empty token for empty data, got: %s", encoded)
				}
				return
			}

			if encoded == "" {
				t.Fatal("Expected non-empty token")
			}

			decoded, err := Decode(encoded)
			if err != nil {
				t.Fatalf("Decode error: %v", err)
			}

			if !reflect.DeepEqual(tc.tokenData, decoded) {
				t.Errorf("Expected decoded %v, got %v", tc.tokenData, decoded)
			}
		})
	}
}

func TestDecode_InvalidInput(t *testing.T) {
	testCases := []struct {
		name  string
		token string
	}{
		{
			name:  "empty token",
			token: "",
		},
		{
			name:  "invalid base64",
			token: "not-base64!@#",
		},
		{
			name:  "valid base64 but invalid JSON",
			token: "aW52YWxpZCBqc29u",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			decoded, err := Decode(tc.token)

			if tc.token == "" {
				if len(decoded) != 0 {
					t.Errorf("Expected empty map for empty token, got: %v", decoded)
				}
				return
			}

			if err == nil {
				t.Errorf("Expected error for invalid token, got nil")
			}
		})
	}
}
