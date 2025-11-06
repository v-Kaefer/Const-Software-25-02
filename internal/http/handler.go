package http

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"time"
	"github.com/v-Kaefer/Const-Software-25-02/internal/auth"
	"github.com/v-Kaefer/Const-Software-25-02/internal/config"
	"github.com/v-Kaefer/Const-Software-25-02/pkg/user"
)

// Router simples usando net/http para não adicionar dependências.
type Router struct {
	userSvc      *user.Service
	authMiddleware *auth.Middleware
	mux          *http.ServeMux
}

func NewRouter(userSvc *user.Service, authMiddleware *auth.Middleware) *Router {
	r := &Router{
		userSvc:        userSvc,
		authMiddleware: authMiddleware,
		mux:            http.NewServeMux(),
	}
	r.routes()
	return r
}

func (r *Router) routes() {
	// POST /users - Admin only (create user)
	r.mux.Handle("POST /users", r.authMiddleware.Authenticate(
		r.authMiddleware.RequireRole(auth.RoleAdmin)(http.HandlerFunc(r.handleCreateUser)),
	))
	
	// GET /users - Admin only (list all users)
	r.mux.Handle("GET /users", r.authMiddleware.Authenticate(
		r.authMiddleware.RequireRole(auth.RoleAdmin)(http.HandlerFunc(r.handleListUsers)),
	))

	// GET /users/{id} - Admin or own user
	r.mux.Handle("GET /users/{id}", r.authMiddleware.Authenticate(
		http.HandlerFunc(r.handleGetUser),
	))

	// PUT /users/{id} - Admin or own user
	r.mux.Handle("PUT /users/{id}", r.authMiddleware.Authenticate(
		http.HandlerFunc(r.handleUpdateUser),
	))

	// PATCH /users/{id} - Admin or own user
	r.mux.Handle("PATCH /users/{id}", r.authMiddleware.Authenticate(
		http.HandlerFunc(r.handlePatchUser),
	))

	// DELETE /users/{id} - Admin only
	r.mux.Handle("DELETE /users/{id}", r.authMiddleware.Authenticate(
		r.authMiddleware.RequireRole(auth.RoleAdmin)(http.HandlerFunc(r.handleDeleteUser)),
	))
}

func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.mux.ServeHTTP(w, req)
}

// isAdminOrOwner checks if the user is admin or owns the resource
func (r *Router) isAdminOrOwner(ctx context.Context, userID uint) bool {
	roles, ok := auth.GetRolesFromContext(ctx)
	if !ok {
		return false
	}
	
	// Check if user is admin
	for _, role := range roles {
		if role == string(auth.RoleAdmin) {
			return true
		}
	}
	
	// Check if user owns the resource
	username, ok := auth.GetUserFromContext(ctx)
	if !ok {
		return false
	}
	
	// Get user by username (email) and check if ID matches
	user, err := r.userSvc.GetByEmail(ctx, username)
	if err != nil {
		return false
	}
	
	return user.ID == userID
}

// NewAuthMiddleware creates a new auth middleware from config
func NewAuthMiddleware(cfg config.CognitoConfig) *auth.Middleware {
	authConfig := auth.CognitoConfig{
		Region:      cfg.Region,
		UserPoolID:  cfg.UserPoolID,
		JWTIssuer:   cfg.JWTIssuer,
		JWTAudience: cfg.JWTAudience,
		JWKSURI:     cfg.JWKSURI,
	}
	return auth.NewMiddleware(authConfig)
}

// NewMockAuthMiddleware creates a mock auth middleware for testing
// that bypasses authentication
func NewMockAuthMiddleware() *auth.Middleware {
	return auth.NewMockMiddleware()
}

func (r *Router) handleCreateUser(w http.ResponseWriter, req *http.Request) {
	type in struct {
		Email string `json:"email"`
		Name  string `json:"name"`
	}
	var body in
	if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}
	ctx, cancel := context.WithTimeout(req.Context(), 5*time.Second)
	defer cancel()
	u, err := r.userSvc.Register(ctx, body.Email, body.Name)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(u)
}

func (r *Router) handleGetUserByEmail(w http.ResponseWriter, req *http.Request) {
	email := req.URL.Query().Get("email")
	if email == "" {
		http.Error(w, "email required", http.StatusBadRequest)
		return
	}
	ctx, cancel := context.WithTimeout(req.Context(), 5*time.Second)
	defer cancel()
	u, err := r.userSvc.GetByEmail(ctx, email)
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(u)
}

func (r *Router) handleListUsers(w http.ResponseWriter, req *http.Request) {
	ctx, cancel := context.WithTimeout(req.Context(), 5*time.Second)
	defer cancel()
	users, err := r.userSvc.List(ctx)
	if err != nil {
		http.Error(w, "failed to list users", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(users)
}

func (r *Router) handleGetUser(w http.ResponseWriter, req *http.Request) {
	idStr := req.PathValue("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		http.Error(w, "invalid user id", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(req.Context(), 5*time.Second)
	defer cancel()

	// Check if user is admin or owns the resource
	if !r.isAdminOrOwner(ctx, uint(id)) {
		http.Error(w, "insufficient permissions", http.StatusForbidden)
		return
	}

	u, err := r.userSvc.GetByID(ctx, uint(id))
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(u)
}

func (r *Router) handleUpdateUser(w http.ResponseWriter, req *http.Request) {
	idStr := req.PathValue("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		http.Error(w, "invalid user id", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(req.Context(), 5*time.Second)
	defer cancel()

	// Check if user is admin or owns the resource
	if !r.isAdminOrOwner(ctx, uint(id)) {
		http.Error(w, "insufficient permissions", http.StatusForbidden)
		return
	}

	type in struct {
		Email string `json:"email"`
		Name  string `json:"name"`
	}
	var body in
	if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}

	u, err := r.userSvc.Update(ctx, uint(id), body.Email, body.Name)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(u)
}

func (r *Router) handlePatchUser(w http.ResponseWriter, req *http.Request) {
	idStr := req.PathValue("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		http.Error(w, "invalid user id", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(req.Context(), 5*time.Second)
	defer cancel()

	// Check if user is admin or owns the resource
	if !r.isAdminOrOwner(ctx, uint(id)) {
		http.Error(w, "insufficient permissions", http.StatusForbidden)
		return
	}

	// Get existing user
	u, err := r.userSvc.GetByID(ctx, uint(id))
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	// Parse partial update
	var body map[string]interface{}
	if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}

	// Apply partial updates
	email := u.Email
	name := u.Name
	if val, ok := body["email"].(string); ok {
		email = val
	}
	if val, ok := body["name"].(string); ok {
		name = val
	}

	u, err = r.userSvc.Update(ctx, uint(id), email, name)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(u)
}

func (r *Router) handleDeleteUser(w http.ResponseWriter, req *http.Request) {
	idStr := req.PathValue("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		http.Error(w, "invalid user id", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(req.Context(), 5*time.Second)
	defer cancel()

	if err := r.userSvc.Delete(ctx, uint(id)); err != nil {
		http.Error(w, "failed to delete user", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
