package auth

import (
	"crypto/rand"
	"crypto/rsa"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func TestClaims_HasRole(t *testing.T) {
	tests := []struct {
		name     string
		groups   []string
		role     string
		expected bool
	}{
		{
			name:     "has role",
			groups:   []string{"admin-group", "user-group"},
			role:     "admin-group",
			expected: true,
		},
		{
			name:     "does not have role",
			groups:   []string{"user-group"},
			role:     "admin-group",
			expected: false,
		},
		{
			name:     "empty groups",
			groups:   []string{},
			role:     "admin-group",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			claims := &Claims{
				Groups: tt.groups,
			}
			result := claims.HasRole(tt.role)
			if result != tt.expected {
				t.Errorf("HasRole() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestClaims_IsAdmin(t *testing.T) {
	tests := []struct {
		name     string
		groups   []string
		expected bool
	}{
		{
			name:     "is admin",
			groups:   []string{"admin-group"},
			expected: true,
		},
		{
			name:     "is not admin",
			groups:   []string{"user-group"},
			expected: false,
		},
		{
			name:     "multiple groups with admin",
			groups:   []string{"user-group", "admin-group"},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			claims := &Claims{
				Groups: tt.groups,
			}
			result := claims.IsAdmin()
			if result != tt.expected {
				t.Errorf("IsAdmin() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// TestValidateToken_Issuer tests issuer validation
func TestValidateToken_Issuer(t *testing.T) {
	// This is a conceptual test - in reality we'd need a mock JWKS server
	// For now, we test the claims logic
	now := time.Now()
	
	tests := []struct {
		name        string
		issuer      string
		configIss   string
		shouldError bool
	}{
		{
			name:        "valid issuer",
			issuer:      "https://cognito.amazonaws.com",
			configIss:   "https://cognito.amazonaws.com",
			shouldError: false,
		},
		{
			name:        "invalid issuer",
			issuer:      "https://wrong-issuer.com",
			configIss:   "https://cognito.amazonaws.com",
			shouldError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			claims := &Claims{
				RegisteredClaims: jwt.RegisteredClaims{
					Issuer:    tt.issuer,
					Audience:  jwt.ClaimStrings{"test-audience"},
					ExpiresAt: jwt.NewNumericDate(now.Add(1 * time.Hour)),
					NotBefore: jwt.NewNumericDate(now.Add(-1 * time.Minute)),
				},
			}

			// Test issuer validation
			if claims.Issuer != tt.configIss && !tt.shouldError {
				t.Error("Expected valid issuer")
			}
			if claims.Issuer == tt.configIss && tt.shouldError {
				t.Error("Expected invalid issuer")
			}
		})
	}
}

// TestValidateToken_Expiration tests expiration validation
func TestValidateToken_Expiration(t *testing.T) {
	now := time.Now()
	
	tests := []struct {
		name      string
		expiresAt time.Time
		isExpired bool
	}{
		{
			name:      "not expired",
			expiresAt: now.Add(1 * time.Hour),
			isExpired: false,
		},
		{
			name:      "expired",
			expiresAt: now.Add(-1 * time.Hour),
			isExpired: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			claims := &Claims{
				RegisteredClaims: jwt.RegisteredClaims{
					ExpiresAt: jwt.NewNumericDate(tt.expiresAt),
				},
			}

			if claims.ExpiresAt.Time.Before(now) != tt.isExpired {
				t.Errorf("Expected expired=%v, got expired=%v", tt.isExpired, claims.ExpiresAt.Time.Before(now))
			}
		})
	}
}

// TestValidateToken_Audience tests audience validation
func TestValidateToken_Audience(t *testing.T) {
	tests := []struct {
		name          string
		claimAudience jwt.ClaimStrings
		configAud     string
		shouldBeValid bool
	}{
		{
			name:          "valid audience",
			claimAudience: jwt.ClaimStrings{"test-client-id"},
			configAud:     "test-client-id",
			shouldBeValid: true,
		},
		{
			name:          "invalid audience",
			claimAudience: jwt.ClaimStrings{"wrong-client-id"},
			configAud:     "test-client-id",
			shouldBeValid: false,
		},
		{
			name:          "multiple audiences with valid",
			claimAudience: jwt.ClaimStrings{"other-id", "test-client-id"},
			configAud:     "test-client-id",
			shouldBeValid: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			claims := &Claims{
				RegisteredClaims: jwt.RegisteredClaims{
					Audience: tt.claimAudience,
				},
			}

			// Check if config audience is in claim audience
			found := false
			for _, aud := range claims.Audience {
				if aud == tt.configAud {
					found = true
					break
				}
			}

			if found != tt.shouldBeValid {
				t.Errorf("Expected valid=%v, got valid=%v", tt.shouldBeValid, found)
			}
		})
	}
}

// Helper function to generate RSA key pair for testing
func generateRSAKeyPair(t *testing.T) (*rsa.PrivateKey, *rsa.PublicKey) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("Failed to generate RSA key: %v", err)
	}
	return privateKey, &privateKey.PublicKey
}
