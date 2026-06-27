package rxtspot

import (
	"encoding/base64"
	"fmt"
	"os"
)

// IsBase64 checks if a string is valid base64 encoding.
// It verifies the string only contains valid base64 characters and can be decoded.
func IsBase64(s string) bool {
	if s == "" {
		return false
	}
	_, err := base64.StdEncoding.DecodeString(s)
	return err == nil
}

// PrepareUserData takes raw user data (string) and returns base64-encoded data.
// If the input is already valid base64, it is returned as-is.
// If the input is empty, an empty string is returned.
func PrepareUserData(input string) string {
	if input == "" {
		return ""
	}
	// If already base64, return as-is
	if IsBase64(input) {
		return input
	}
	// Encode to base64
	return base64.StdEncoding.EncodeToString([]byte(input))
}

// PrepareUserDataFromScript reads a file and returns its content as base64-encoded data.
// Returns an error if the file cannot be read.
func PrepareUserDataFromScript(filePath string) (string, error) {
	if filePath == "" {
		return "", fmt.Errorf("file path is required")
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to read script file '%s': %w", filePath, err)
	}

	return base64.StdEncoding.EncodeToString(data), nil
}
