package auth

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/MicahParks/keyfunc/v2"
	"github.com/golang-jwt/jwt/v5"
)

// JWTConfig holds JWT validation configuration
type JWTConfig struct {
	Issuer   string
	Audience string
	JWKSURI  string
}

// JWTMiddleware validates JWT tokens from the Authorization header
type JWTMiddleware struct {
	config *JWTConfig
	jwks   *keyfunc.JWKS
}

// Claims represents the JWT claims we expect
type Claims struct {
	jwt.RegisteredClaims
	Groups []string               `json:"cognito:groups"`
	Custom map[string]interface{} `json:"-"`
}

// ContextKey is the type for context keys
type ContextKey string

const (
	// UserClaimsKey is the context key for user claims
	UserClaimsKey ContextKey = "user_claims"
)

var (
	// ErrMissingToken is returned when no token is provided
	ErrMissingToken = errors.New("missing authorization token")
	// ErrInvalidToken is returned when token validation fails
	ErrInvalidToken = errors.New("invalid token")
	// ErrExpiredToken is returned when token has expired
	ErrExpiredToken = errors.New("token has expired")
)

// NewJWTMiddleware creates a new JWT middleware instance
func NewJWTMiddleware(config *JWTConfig) (*JWTMiddleware, error) {
	// Create JWKS client with caching
	options := keyfunc.Options{
		RefreshInterval: 1 * time.Hour,
		RefreshTimeout:  10 * time.Second,
		RefreshErrorHandler: func(err error) {
			// Log error but don't fail - use cached keys
		},
	}

	jwks, err := keyfunc.Get(config.JWKSURI, options)
	if err != nil {
		return nil, err
	}

	return &JWTMiddleware{
		config: config,
		jwks:   jwks,
	}, nil
}

// Middleware returns an HTTP middleware that validates JWT tokens
func (m *JWTMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract token from Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, ErrMissingToken.Error(), http.StatusUnauthorized)
			return
		}

		// Check for Bearer prefix
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			http.Error(w, "invalid authorization header format", http.StatusUnauthorized)
			return
		}

		tokenString := parts[1]

		// Parse and validate token
		claims, err := m.ValidateToken(tokenString)
		if err != nil {
			statusCode := http.StatusUnauthorized
			if errors.Is(err, ErrExpiredToken) {
				statusCode = http.StatusUnauthorized
			}
			http.Error(w, err.Error(), statusCode)
			return
		}

		// Add claims to request context
		ctx := context.WithValue(r.Context(), UserClaimsKey, claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// ValidateToken parses and validates a JWT token string
func (m *JWTMiddleware) ValidateToken(tokenString string) (*Claims, error) {
	// Parse token with JWKS
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, m.jwks.Keyfunc)
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrExpiredToken
		}
		return nil, ErrInvalidToken
	}

	if !token.Valid {
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(*Claims)
	if !ok {
		return nil, ErrInvalidToken
	}

	// Validate issuer
	if claims.Issuer != m.config.Issuer {
		return nil, errors.New("invalid issuer")
	}

	// Validate audience
	validAudience := false
	for _, aud := range claims.Audience {
		if aud == m.config.Audience {
			validAudience = true
			break
		}
	}
	if !validAudience {
		return nil, errors.New("invalid audience")
	}

	// Validate expiration
	if claims.ExpiresAt != nil && claims.ExpiresAt.Time.Before(time.Now()) {
		return nil, ErrExpiredToken
	}

	// Validate not before
	if claims.NotBefore != nil && claims.NotBefore.Time.After(time.Now()) {
		return nil, errors.New("token not yet valid")
	}

	return claims, nil
}

// GetClaimsFromContext extracts claims from request context
func GetClaimsFromContext(ctx context.Context) (*Claims, error) {
	claims, ok := ctx.Value(UserClaimsKey).(*Claims)
	if !ok || claims == nil {
		return nil, errors.New("no claims in context")
	}
	return claims, nil
}

// HasRole checks if the user has a specific role/group
func (c *Claims) HasRole(role string) bool {
	for _, g := range c.Groups {
		if g == role {
			return true
		}
	}
	return false
}

// IsAdmin checks if the user is an admin
func (c *Claims) IsAdmin() bool {
	return c.HasRole("admin-group")
}
