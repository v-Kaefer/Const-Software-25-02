package auth

import (
	"errors"
	"strings"
)

var (
	// ErrInvalidPath is returned when path parsing fails
	ErrInvalidPath = errors.New("invalid path")
)

// ExtractUserIDFromPath extracts the user ID from a path like /users/{id}
// Returns the ID and true if found, empty string and false otherwise
func ExtractUserIDFromPath(path string) (string, bool) {
	// Remove leading/trailing slashes and split
	parts := strings.Split(strings.Trim(path, "/"), "/")
	
	// Expecting pattern: /users/{id}
	if len(parts) != 2 || parts[0] != "users" {
		return "", false
	}
	
	// Return the ID part (second segment)
	return parts[1], true
}
