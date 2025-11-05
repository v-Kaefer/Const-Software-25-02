package jwt_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/v-Kaefer/Const-Software-25-02/pkg/jwt"
)

func TestMiddleware_ValidToken(t *testing.T) {
	secret := "test-secret"
	generator := jwt.NewGenerator(secret)

	// Gera um token válido
	token, err := generator.GenerateToken(123, "test@example.com", 1*time.Hour)
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}

	// Handler que será protegido pelo middleware
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims, ok := jwt.GetClaimsFromContext(r.Context())
		if !ok {
			t.Error("claims not found in context")
			return
		}
		if claims.UserID != 123 {
			t.Errorf("expected user_id 123, got %d", claims.UserID)
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Aplica middleware
	protectedHandler := generator.Middleware(handler)

	// Cria requisição com token no header
	req := httptest.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	protectedHandler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}
}

func TestMiddleware_MissingToken(t *testing.T) {
	secret := "test-secret"
	generator := jwt.NewGenerator(secret)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	protectedHandler := generator.Middleware(handler)

	// Requisição sem token
	req := httptest.NewRequest("GET", "/protected", nil)
	rec := httptest.NewRecorder()

	protectedHandler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", rec.Code)
	}
}

func TestMiddleware_InvalidTokenFormat(t *testing.T) {
	secret := "test-secret"
	generator := jwt.NewGenerator(secret)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	protectedHandler := generator.Middleware(handler)

	// Requisição com formato inválido de Authorization
	req := httptest.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "InvalidFormat")
	rec := httptest.NewRecorder()

	protectedHandler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", rec.Code)
	}
}

func TestMiddleware_ExpiredToken(t *testing.T) {
	secret := "test-secret"
	generator := jwt.NewGenerator(secret)

	// Gera token expirado
	token, err := generator.GenerateToken(123, "test@example.com", -1*time.Hour)
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	protectedHandler := generator.Middleware(handler)

	req := httptest.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	protectedHandler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", rec.Code)
	}
}
