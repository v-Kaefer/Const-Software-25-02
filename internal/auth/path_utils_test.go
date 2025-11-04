package auth

import "testing"

func TestExtractUserIDFromPath(t *testing.T) {
	tests := []struct {
		name       string
		path       string
		expectedID string
		shouldFind bool
	}{
		{
			name:       "valid numeric ID",
			path:       "/users/123",
			expectedID: "123",
			shouldFind: true,
		},
		{
			name:       "valid email as ID",
			path:       "/users/user@example.com",
			expectedID: "user@example.com",
			shouldFind: true,
		},
		{
			name:       "path with trailing slash",
			path:       "/users/456/",
			expectedID: "456",
			shouldFind: true,
		},
		{
			name:       "invalid path - missing ID",
			path:       "/users",
			expectedID: "",
			shouldFind: false,
		},
		{
			name:       "invalid path - wrong resource",
			path:       "/posts/123",
			expectedID: "",
			shouldFind: false,
		},
		{
			name:       "invalid path - too many segments",
			path:       "/users/123/extra",
			expectedID: "",
			shouldFind: false,
		},
		{
			name:       "path without leading slash",
			path:       "users/789",
			expectedID: "789",
			shouldFind: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id, found := ExtractUserIDFromPath(tt.path)
			
			if found != tt.shouldFind {
				t.Errorf("ExtractUserIDFromPath() found = %v, want %v", found, tt.shouldFind)
			}
			
			if found && id != tt.expectedID {
				t.Errorf("ExtractUserIDFromPath() id = %v, want %v", id, tt.expectedID)
			}
		})
	}
}
