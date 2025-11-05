package jwt_test

import (
	"testing"
	"time"

	"github.com/v-Kaefer/Const-Software-25-02/pkg/jwt"
)

func TestGenerateAndValidateToken(t *testing.T) {
	secret := "test-secret-key"
	generator := jwt.NewGenerator(secret)

	userID := uint(123)
	email := "test@example.com"
	duration := 1 * time.Hour

	// Gera token
	token, err := generator.GenerateToken(userID, email, duration)
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}

	if token == "" {
		t.Fatal("token should not be empty")
	}

	// Valida token
	claims, err := generator.ValidateToken(token)
	if err != nil {
		t.Fatalf("failed to validate token: %v", err)
	}

	if claims.UserID != userID {
		t.Errorf("expected user_id %d, got %d", userID, claims.UserID)
	}

	if claims.Email != email {
		t.Errorf("expected email %s, got %s", email, claims.Email)
	}
}

func TestValidateExpiredToken(t *testing.T) {
	secret := "test-secret-key"
	generator := jwt.NewGenerator(secret)

	// Gera token com duração negativa (já expirado)
	token, err := generator.GenerateToken(123, "test@example.com", -1*time.Hour)
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}

	// Tenta validar token expirado
	_, err = generator.ValidateToken(token)
	if err != jwt.ErrExpiredToken {
		t.Errorf("expected ErrExpiredToken, got %v", err)
	}
}

func TestValidateInvalidToken(t *testing.T) {
	secret := "test-secret-key"
	generator := jwt.NewGenerator(secret)

	// Tenta validar token inválido
	_, err := generator.ValidateToken("invalid.token.here")
	if err == nil {
		t.Error("expected error for invalid token, got nil")
	}
}

func TestValidateTokenWithWrongSecret(t *testing.T) {
	secret1 := "secret-1"
	secret2 := "secret-2"

	generator1 := jwt.NewGenerator(secret1)
	generator2 := jwt.NewGenerator(secret2)

	// Gera token com secret1
	token, err := generator1.GenerateToken(123, "test@example.com", 1*time.Hour)
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}

	// Tenta validar com secret2 (diferente)
	_, err = generator2.ValidateToken(token)
	if err == nil {
		t.Error("expected error when validating with different secret, got nil")
	}
}
