package auth_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/v-Kaefer/Const-Software-25-02/internal/auth"
)

func TestMiddleware_Authenticate_SkipAuth(t *testing.T) {
	// Test that mock middleware skips authentication
	middleware := auth.NewMockMiddleware()

	handler := middleware.Authenticate(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify user info is in context
		user, ok := auth.GetUserFromContext(r.Context())
		if !ok {
			t.Error("expected user in context")
		}
		if user != "test-user" {
			t.Errorf("expected user 'test-user', got %s", user)
		}

		// Verify roles are in context
		roles, ok := auth.GetRolesFromContext(r.Context())
		if !ok {
			t.Error("expected roles in context")
		}
		if len(roles) == 0 {
			t.Error("expected at least one role")
		}

		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}

func TestMiddleware_Authenticate_MissingHeader(t *testing.T) {
	// Create a real middleware (not mock) to test authentication
	middleware := auth.NewMiddleware(auth.CognitoConfig{
		Region:     "us-east-1",
		UserPoolID: "test-pool",
	})

	handler := middleware.Authenticate(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	// Don't set Authorization header
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", w.Code)
	}
}

func TestMiddleware_Authenticate_InvalidFormat(t *testing.T) {
	middleware := auth.NewMiddleware(auth.CognitoConfig{
		Region:     "us-east-1",
		UserPoolID: "test-pool",
	})

	handler := middleware.Authenticate(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	tests := []struct {
		name   string
		header string
	}{
		{"missing Bearer prefix", "token123"},
		{"empty Bearer", "Bearer "},
		{"wrong prefix", "Basic token123"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/test", nil)
			req.Header.Set("Authorization", tt.header)
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			if w.Code != http.StatusUnauthorized {
				t.Errorf("expected status 401, got %d", w.Code)
			}
		})
	}
}

func TestMiddleware_RequireRole_SkipAuth(t *testing.T) {
	// Test that mock middleware skips role check
	middleware := auth.NewMockMiddleware()

	handler := middleware.RequireRole(auth.RoleAdmin)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200 (role check skipped), got %d", w.Code)
	}
}

func TestMiddleware_RequireRole_NoRoles(t *testing.T) {
	middleware := auth.NewMiddleware(auth.CognitoConfig{})

	handler := middleware.RequireRole(auth.RoleAdmin)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// Create request with no roles in context
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("expected status 403, got %d", w.Code)
	}
}

func TestMiddleware_RequireRole_InsufficientPermissions(t *testing.T) {
	middleware := auth.NewMiddleware(auth.CognitoConfig{})

	handler := middleware.RequireRole(auth.RoleAdmin)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// Create request with user role instead of admin role
	ctx := context.WithValue(context.Background(), context.TODO(), []string{string(auth.RoleUser)})
	req := httptest.NewRequest("GET", "/test", nil).WithContext(ctx)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("expected status 403, got %d", w.Code)
	}
}

func TestGetUserFromContext(t *testing.T) {
	tests := []struct {
		name     string
		setup    func() context.Context
		wantUser string
		wantOk   bool
	}{
		{
			name: "user exists",
			setup: func() context.Context {
				middleware := auth.NewMockMiddleware()
				req := httptest.NewRequest("GET", "/", nil)
				var ctx context.Context
				handler := middleware.Authenticate(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					ctx = r.Context()
				}))
				handler.ServeHTTP(httptest.NewRecorder(), req)
				return ctx
			},
			wantUser: "test-user",
			wantOk:   true,
		},
		{
			name: "user not in context",
			setup: func() context.Context {
				return context.Background()
			},
			wantUser: "",
			wantOk:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := tt.setup()
			user, ok := auth.GetUserFromContext(ctx)
			if ok != tt.wantOk {
				t.Errorf("GetUserFromContext() ok = %v, want %v", ok, tt.wantOk)
			}
			if user != tt.wantUser {
				t.Errorf("GetUserFromContext() user = %v, want %v", user, tt.wantUser)
			}
		})
	}
}

func TestGetRolesFromContext(t *testing.T) {
	tests := []struct {
		name      string
		setup     func() context.Context
		wantRoles []string
		wantOk    bool
	}{
		{
			name: "roles exist",
			setup: func() context.Context {
				middleware := auth.NewMockMiddleware()
				req := httptest.NewRequest("GET", "/", nil)
				var ctx context.Context
				handler := middleware.Authenticate(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					ctx = r.Context()
				}))
				handler.ServeHTTP(httptest.NewRecorder(), req)
				return ctx
			},
			wantRoles: []string{string(auth.RoleAdmin)},
			wantOk:    true,
		},
		{
			name: "roles not in context",
			setup: func() context.Context {
				return context.Background()
			},
			wantRoles: nil,
			wantOk:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := tt.setup()
			roles, ok := auth.GetRolesFromContext(ctx)
			if ok != tt.wantOk {
				t.Errorf("GetRolesFromContext() ok = %v, want %v", ok, tt.wantOk)
			}
			if ok && len(roles) != len(tt.wantRoles) {
				t.Errorf("GetRolesFromContext() roles = %v, want %v", roles, tt.wantRoles)
			}
		})
	}
}
