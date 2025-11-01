package persistence

import (
	"encoding/base64"
	"encoding/json"
	"fmt"

	"example.com/go-yippi/internal/domain/entities"
)

// EncodeCursor encodes a cursor to a base64 string
func EncodeCursor(cursor entities.Cursor) (string, error) {
	data, err := json.Marshal(cursor)
	if err != nil {
		return "", fmt.Errorf("failed to marshal cursor: %w", err)
	}
	return base64.StdEncoding.EncodeToString(data), nil
}

// DecodeCursor decodes a base64 cursor string
func DecodeCursor(cursorStr string) (*entities.Cursor, error) {
	data, err := base64.StdEncoding.DecodeString(cursorStr)
	if err != nil {
		return nil, fmt.Errorf("failed to decode cursor: %w", err)
	}

	var cursor entities.Cursor
	if err := json.Unmarshal(data, &cursor); err != nil {
		return nil, fmt.Errorf("failed to unmarshal cursor: %w", err)
	}

	return &cursor, nil
}
