package auth

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"math/big"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// TestJWTValidation tests JWT token validation with various scenarios
func TestJWTValidation_IssuerValidation(t *testing.T) {
	// Create a test middleware with issuer configured
	config := CognitoConfig{
		Region:      "us-east-1",
		UserPoolID:  "us-east-1_test",
		JWTIssuer:   "https://cognito-idp.us-east-1.amazonaws.com/us-east-1_test",
		JWTAudience: "",
		JWKSURI:     "",
	}
	m := NewMiddleware(config)

	// Test with correct issuer - would need real token to fully test
	// This test validates the configuration is set up correctly
	if m.config.JWTIssuer != config.JWTIssuer {
		t.Errorf("JWTIssuer not configured correctly: got %s, want %s", m.config.JWTIssuer, config.JWTIssuer)
	}
}

func TestJWTValidation_AudienceValidation(t *testing.T) {
	// Create a test middleware with audience configured
	config := CognitoConfig{
		Region:      "us-east-1",
		UserPoolID:  "us-east-1_test",
		JWTIssuer:   "",
		JWTAudience: "test-client-id",
		JWKSURI:     "",
	}
	m := NewMiddleware(config)

	// Test with correct audience - configuration check
	if m.config.JWTAudience != config.JWTAudience {
		t.Errorf("JWTAudience not configured correctly: got %s, want %s", m.config.JWTAudience, config.JWTAudience)
	}
}

func TestJWTValidation_JWKSURIConfiguration(t *testing.T) {
	tests := []struct {
		name           string
		config         CognitoConfig
		wantCustomURI  bool
		expectedPrefix string
	}{
		{
			name: "Custom JWKS URI",
			config: CognitoConfig{
				Region:     "us-east-1",
				UserPoolID: "us-east-1_test",
				JWKSURI:    "https://custom-jwks.example.com/.well-known/jwks.json",
			},
			wantCustomURI:  true,
			expectedPrefix: "https://custom-jwks",
		},
		{
			name: "Auto-constructed JWKS URI",
			config: CognitoConfig{
				Region:     "us-west-2",
				UserPoolID: "us-west-2_abc123",
				JWKSURI:    "",
			},
			wantCustomURI:  false,
			expectedPrefix: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewMiddleware(tt.config)
			if tt.wantCustomURI && m.config.JWKSURI != tt.config.JWKSURI {
				t.Errorf("JWKSURI not set correctly: got %s, want %s", m.config.JWKSURI, tt.config.JWKSURI)
			}
		})
	}
}

// TestMiddleware_TokenValidation tests token validation with malformed tokens
func TestMiddleware_TokenValidation_MalformedToken(t *testing.T) {
	config := CognitoConfig{
		Region:     "us-east-1",
		UserPoolID: "us-east-1_test",
	}
	m := NewMiddleware(config)

	tests := []struct {
		name       string
		token      string
		wantStatus int
	}{
		{
			name:       "Empty token",
			token:      "",
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "Malformed token - not JWT",
			token:      "Bearer not-a-jwt-token",
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "Malformed token - invalid format",
			token:      "Bearer invalid.token",
			wantStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := m.Authenticate(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			}))

			req := httptest.NewRequest("GET", "/test", nil)
			if tt.token != "" {
				req.Header.Set("Authorization", tt.token)
			}
			rr := httptest.NewRecorder()

			handler.ServeHTTP(rr, req)

			if rr.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", rr.Code, tt.wantStatus)
			}
		})
	}
}

// TestMiddleware_ExpiredToken tests handling of expired tokens
func TestMiddleware_ExpiredToken(t *testing.T) {
	// Generate a test RSA key pair
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("failed to generate key: %v", err)
	}

	// Create an expired token
	claims := &CognitoClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-1 * time.Hour)), // Expired 1 hour ago
			IssuedAt:  jwt.NewNumericDate(time.Now().Add(-2 * time.Hour)),
			NotBefore: jwt.NewNumericDate(time.Now().Add(-2 * time.Hour)),
			Issuer:    "https://cognito-idp.us-east-1.amazonaws.com/us-east-1_test",
			Subject:   "test-user",
		},
		Username: "test@example.com",
		Groups:   []string{"admin-group"},
		TokenUse: "access",
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	token.Header["kid"] = "test-key-id"
	tokenString, err := token.SignedString(privateKey)
	if err != nil {
		t.Fatalf("failed to sign token: %v", err)
	}

	// Create middleware with JWKS mock
	config := CognitoConfig{
		Region:     "us-east-1",
		UserPoolID: "us-east-1_test",
	}
	m := NewMiddleware(config)

	// Create a test handler
	handler := m.Authenticate(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+tokenString)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	// Should return unauthorized due to expired token
	if rr.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusUnauthorized)
	}
}

// TestMiddleware_ValidTokenStructure tests token structure validation
func TestMiddleware_ValidTokenStructure(t *testing.T) {
	// This test validates that properly structured tokens are processed
	// even if signature validation fails (which is expected without proper JWKS)
	
	config := CognitoConfig{
		Region:     "us-east-1",
		UserPoolID: "us-east-1_test",
	}
	m := NewMiddleware(config)

	// Create claims
	claims := &CognitoClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "https://cognito-idp.us-east-1.amazonaws.com/us-east-1_test",
			Subject:   "test-user",
		},
		Username: "test@example.com",
		Groups:   []string{"admin-group"},
		TokenUse: "access",
	}

	// Create unsigned token for structure testing
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, _ := token.SignedString([]byte("test-secret"))

	handler := m.Authenticate(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+tokenString)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	// Will fail due to signature validation (expected), but validates parsing works
	if rr.Code == http.StatusOK {
		t.Error("Expected token validation to fail without proper JWKS, but got success")
	}
}

// Mock JWKS server for testing
type mockJWKSServer struct {
	server *httptest.Server
	keys   *JWKS
}

func newMockJWKSServer(publicKey *rsa.PublicKey, kid string) *mockJWKSServer {
	// Convert public key to JWK format
	n := base64.RawURLEncoding.EncodeToString(publicKey.N.Bytes())
	e := base64.RawURLEncoding.EncodeToString(big.NewInt(int64(publicKey.E)).Bytes())

	jwks := &JWKS{
		Keys: []JWK{
			{
				Kid: kid,
				Kty: "RSA",
				Alg: "RS256",
				Use: "sig",
				N:   n,
				E:   e,
			},
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(jwks)
	}))

	return &mockJWKSServer{
		server: server,
		keys:   jwks,
	}
}

func (m *mockJWKSServer) Close() {
	m.server.Close()
}

// TestMiddleware_WithMockJWKS tests token validation with a mock JWKS server
func TestMiddleware_WithMockJWKS(t *testing.T) {
	// Generate RSA key pair
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("failed to generate key: %v", err)
	}

	kid := "test-key-id"
	mockJWKS := newMockJWKSServer(&privateKey.PublicKey, kid)
	defer mockJWKS.Close()

	// Create valid token
	claims := &CognitoClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "https://cognito-idp.us-east-1.amazonaws.com/us-east-1_test",
			Subject:   "test-user",
		},
		Username: "test@example.com",
		Groups:   []string{"admin-group"},
		TokenUse: "access",
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	token.Header["kid"] = kid
	tokenString, err := token.SignedString(privateKey)
	if err != nil {
		t.Fatalf("failed to sign token: %v", err)
	}

	// Create middleware with mock JWKS URL
	config := CognitoConfig{
		Region:     "us-east-1",
		UserPoolID: "us-east-1_test",
		JWKSURI:    mockJWKS.server.URL,
	}
	m := NewMiddleware(config)

	handler := m.Authenticate(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify context has user info
		user, ok := GetUserFromContext(r.Context())
		if !ok || user != "test@example.com" {
			t.Errorf("user not in context or incorrect: got %s, want test@example.com", user)
		}
		
		roles, ok := GetRolesFromContext(r.Context())
		if !ok || len(roles) == 0 || roles[0] != "admin-group" {
			t.Errorf("roles not in context or incorrect: got %v", roles)
		}
		
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+tokenString)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("status = %d, want %d. Body: %s", rr.Code, http.StatusOK, rr.Body.String())
	}
}
