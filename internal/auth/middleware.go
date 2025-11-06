package auth

import (
	"context"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Role represents user roles in the system
type Role string

const (
	RoleAdmin    Role = "admin-group"
	RoleReviewer Role = "reviewers-group"
	RoleUser     Role = "user-group"
)

// contextKey is a custom type for context keys to avoid collisions
type contextKey string

const (
	userContextKey  contextKey = "user"
	rolesContextKey contextKey = "roles"
)

// CognitoConfig holds Cognito configuration
type CognitoConfig struct {
	Region     string
	UserPoolID string
}

// JWK represents a JSON Web Key
type JWK struct {
	Kid string `json:"kid"`
	Kty string `json:"kty"`
	Alg string `json:"alg"`
	Use string `json:"use"`
	N   string `json:"n"`
	E   string `json:"e"`
}

// JWKS represents a set of JSON Web Keys
type JWKS struct {
	Keys []JWK `json:"keys"`
}

// CognitoClaims represents the claims in a Cognito JWT token
type CognitoClaims struct {
	jwt.RegisteredClaims
	Username      string   `json:"cognito:username"`
	Groups        []string `json:"cognito:groups"`
	TokenUse      string   `json:"token_use"`
	Email         string   `json:"email"`
	EmailVerified bool     `json:"email_verified"`
}

// Middleware handles JWT authentication
type Middleware struct {
	config     CognitoConfig
	jwksCache  *JWKS
	keysCache  map[string]*rsa.PublicKey
	cacheMutex sync.RWMutex
	cacheTime  time.Time
	skipAuth   bool // For testing purposes
}

// NewMiddleware creates a new auth middleware
func NewMiddleware(config CognitoConfig) *Middleware {
	return &Middleware{
		config:    config,
		keysCache: make(map[string]*rsa.PublicKey),
		skipAuth:  false,
	}
}

// NewMockMiddleware creates a middleware that skips authentication (for testing)
func NewMockMiddleware() *Middleware {
	return &Middleware{
		config:    CognitoConfig{},
		keysCache: make(map[string]*rsa.PublicKey),
		skipAuth:  true,
	}
}

// Authenticate is a middleware that validates JWT tokens
func (m *Middleware) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Skip authentication in test mode
		if m.skipAuth {
			// Add mock user info to context for testing
			ctx := r.Context()
			ctx = context.WithValue(ctx, userContextKey, "test-user")
			ctx = context.WithValue(ctx, rolesContextKey, []string{string(RoleAdmin)})
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}

		// Extract token from Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "missing authorization header", http.StatusUnauthorized)
			return
		}

		// Check if it's a Bearer token
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			http.Error(w, "invalid authorization header format", http.StatusUnauthorized)
			return
		}

		tokenString := parts[1]

		// Verify and parse token
		claims, err := m.verifyToken(tokenString)
		if err != nil {
			http.Error(w, fmt.Sprintf("invalid token: %v", err), http.StatusUnauthorized)
			return
		}

		// Add user info and roles to context
		ctx := r.Context()
		ctx = context.WithValue(ctx, userContextKey, claims.Username)
		ctx = context.WithValue(ctx, rolesContextKey, claims.Groups)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RequireRole is a middleware that checks if user has required role
func (m *Middleware) RequireRole(roles ...Role) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Skip role check in test mode
			if m.skipAuth {
				next.ServeHTTP(w, r)
				return
			}

			userRoles, ok := r.Context().Value(rolesContextKey).([]string)
			if !ok {
				http.Error(w, "no roles in context", http.StatusForbidden)
				return
			}

			// Check if user has any of the required roles
			hasRole := false
			for _, role := range roles {
				for _, userRole := range userRoles {
					if userRole == string(role) {
						hasRole = true
						break
					}
				}
				if hasRole {
					break
				}
			}

			if !hasRole {
				http.Error(w, "insufficient permissions", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// verifyToken verifies and parses a Cognito JWT token
func (m *Middleware) verifyToken(tokenString string) (*CognitoClaims, error) {
	// Parse token to get header
	token, err := jwt.ParseWithClaims(tokenString, &CognitoClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Verify signing method
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		// Get key ID from token header
		kid, ok := token.Header["kid"].(string)
		if !ok {
			return nil, errors.New("kid header not found")
		}

		// Get public key for this kid
		publicKey, err := m.getPublicKey(kid)
		if err != nil {
			return nil, err
		}

		return publicKey, nil
	})

	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*CognitoClaims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token claims")
	}

	// Verify token_use claim
	if claims.TokenUse != "access" && claims.TokenUse != "id" {
		return nil, errors.New("invalid token_use claim")
	}

	return claims, nil
}

// getPublicKey retrieves the public key for a given key ID
func (m *Middleware) getPublicKey(kid string) (*rsa.PublicKey, error) {
	m.cacheMutex.RLock()
	if key, ok := m.keysCache[kid]; ok && time.Since(m.cacheTime) < 24*time.Hour {
		m.cacheMutex.RUnlock()
		return key, nil
	}
	m.cacheMutex.RUnlock()

	// Fetch JWKS
	if err := m.fetchJWKS(); err != nil {
		return nil, err
	}

	// Find the key with matching kid
	m.cacheMutex.RLock()
	defer m.cacheMutex.RUnlock()

	if key, ok := m.keysCache[kid]; ok {
		return key, nil
	}

	return nil, fmt.Errorf("key with kid %s not found", kid)
}

// fetchJWKS fetches the JWKS from Cognito
func (m *Middleware) fetchJWKS() error {
	m.cacheMutex.Lock()
	defer m.cacheMutex.Unlock()

	// Check if we need to refresh
	if time.Since(m.cacheTime) < 24*time.Hour && m.jwksCache != nil {
		return nil
	}

	// Build JWKS URL
	jwksURL := fmt.Sprintf("https://cognito-idp.%s.amazonaws.com/%s/.well-known/jwks.json",
		m.config.Region, m.config.UserPoolID)

	// Fetch JWKS
	resp, err := http.Get(jwksURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to fetch JWKS: status %d", resp.StatusCode)
	}

	var jwks JWKS
	if err := json.NewDecoder(resp.Body).Decode(&jwks); err != nil {
		return err
	}

	m.jwksCache = &jwks
	m.cacheTime = time.Now()

	// Convert JWKs to RSA public keys
	for _, jwk := range jwks.Keys {
		key, err := m.jwkToRSAPublicKey(jwk)
		if err != nil {
			continue
		}
		m.keysCache[jwk.Kid] = key
	}

	return nil
}

// jwkToRSAPublicKey converts a JWK to an RSA public key
func (m *Middleware) jwkToRSAPublicKey(jwk JWK) (*rsa.PublicKey, error) {
	// Decode N (modulus)
	nBytes, err := base64.RawURLEncoding.DecodeString(jwk.N)
	if err != nil {
		return nil, err
	}

	// Decode E (exponent)
	eBytes, err := base64.RawURLEncoding.DecodeString(jwk.E)
	if err != nil {
		return nil, err
	}

	// Convert E to int
	var eInt int
	for _, b := range eBytes {
		eInt = eInt<<8 + int(b)
	}

	return &rsa.PublicKey{
		N: new(big.Int).SetBytes(nBytes),
		E: eInt,
	}, nil
}

// GetUserFromContext retrieves the username from the request context
func GetUserFromContext(ctx context.Context) (string, bool) {
	user, ok := ctx.Value(userContextKey).(string)
	return user, ok
}

// GetRolesFromContext retrieves the roles from the request context
func GetRolesFromContext(ctx context.Context) ([]string, bool) {
	roles, ok := ctx.Value(rolesContextKey).([]string)
	return roles, ok
}
