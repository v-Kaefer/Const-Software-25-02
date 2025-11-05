package jwt

import (
	"context"
	"net/http"
	"strings"
)

type contextKey string

const (
	// ClaimsContextKey é a chave usada para armazenar claims no contexto
	ClaimsContextKey contextKey = "jwt_claims"
)

// Middleware cria um middleware HTTP para validação de JWT
func (g *Generator) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extrai o token do header Authorization
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "authorization header required", http.StatusUnauthorized)
			return
		}

		// Espera formato "Bearer <token>"
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			http.Error(w, "invalid authorization header format", http.StatusUnauthorized)
			return
		}

		tokenString := parts[1]

		// Valida o token
		claims, err := g.ValidateToken(tokenString)
		if err != nil {
			if err == ErrExpiredToken {
				http.Error(w, "token has expired", http.StatusUnauthorized)
			} else {
				http.Error(w, "invalid token", http.StatusUnauthorized)
			}
			return
		}

		// Adiciona claims ao contexto da requisição
		ctx := context.WithValue(r.Context(), ClaimsContextKey, claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetClaimsFromContext extrai as claims do contexto da requisição
func GetClaimsFromContext(ctx context.Context) (*Claims, bool) {
	claims, ok := ctx.Value(ClaimsContextKey).(*Claims)
	return claims, ok
}
