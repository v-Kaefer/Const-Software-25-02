package http

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/v-Kaefer/Const-Software-25-02/internal/auth"
	"github.com/v-Kaefer/Const-Software-25-02/pkg/user"
)

// Router simples usando net/http para não adicionar dependências.
type Router struct {
	userSvc       *user.Service
	mux           *http.ServeMux
	jwtMiddleware *auth.JWTMiddleware
}

func NewRouter(userSvc *user.Service, jwtMiddleware *auth.JWTMiddleware) *Router {
	r := &Router{
		userSvc:       userSvc,
		mux:           http.NewServeMux(),
		jwtMiddleware: jwtMiddleware,
	}
	r.routes()
	return r
}

func (r *Router) routes() {
	// Apply JWT middleware to all routes
	r.mux.HandleFunc("POST /users", r.handleCreateUser)
	
	// GET /users - admin only
	r.mux.Handle("GET /users", auth.RequireAdmin(http.HandlerFunc(r.handleListUsers)))
	
	// GET /users/{id} - owner or admin
	r.mux.Handle("GET /users/", auth.RequireOwnerOrAdmin(http.HandlerFunc(r.handleGetUser)))
	
	// PUT /users/{id} - owner or admin
	r.mux.Handle("PUT /users/", auth.RequireOwnerOrAdmin(http.HandlerFunc(r.handleUpdateUser)))
	
	// PATCH /users/{id} - owner or admin
	r.mux.Handle("PATCH /users/", auth.RequireOwnerOrAdmin(http.HandlerFunc(r.handlePatchUser)))
	
	// DELETE /users/{id} - admin only
	r.mux.Handle("DELETE /users/", auth.RequireAdmin(http.HandlerFunc(r.handleDeleteUser)))
}

func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	// Wrap the entire mux with JWT middleware if available
	if r.jwtMiddleware != nil {
		r.jwtMiddleware.Middleware(r.mux).ServeHTTP(w, req)
	} else {
		// No JWT middleware - useful for testing or when JWT is not configured
		r.mux.ServeHTTP(w, req)
	}
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
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(u)
}

func (r *Router) handleListUsers(w http.ResponseWriter, req *http.Request) {
	// For now, support email query parameter
	email := req.URL.Query().Get("email")
	if email != "" {
		r.handleGetUserByEmail(w, req)
		return
	}
	
	// TODO: Implement full list with pagination
	http.Error(w, "list all users not yet implemented - use ?email=...", http.StatusNotImplemented)
}

func (r *Router) handleGetUser(w http.ResponseWriter, req *http.Request) {
	// Extract ID from path using utility function
	userID, ok := auth.ExtractUserIDFromPath(req.URL.Path)
	if !ok {
		http.Error(w, "invalid path - expected /users/{id}", http.StatusBadRequest)
		return
	}
	
	// Try to parse as numeric ID
	id, err := strconv.ParseUint(userID, 10, 32)
	if err == nil {
		r.handleGetUserByID(w, req, uint(id))
		return
	}
	
	// Fallback: treat as email
	ctx, cancel := context.WithTimeout(req.Context(), 5*time.Second)
	defer cancel()
	u, err := r.userSvc.GetByEmail(ctx, userID)
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(u)
}

func (r *Router) handleGetUserByID(w http.ResponseWriter, req *http.Request, id uint) {
	// TODO: Implement GetByID in service
	http.Error(w, "get by ID not yet implemented", http.StatusNotImplemented)
}

func (r *Router) handleUpdateUser(w http.ResponseWriter, req *http.Request) {
	// TODO: Implement update
	http.Error(w, "update not yet implemented", http.StatusNotImplemented)
}

func (r *Router) handlePatchUser(w http.ResponseWriter, req *http.Request) {
	// TODO: Implement patch
	http.Error(w, "patch not yet implemented", http.StatusNotImplemented)
}

func (r *Router) handleDeleteUser(w http.ResponseWriter, req *http.Request) {
	// TODO: Implement delete
	http.Error(w, "delete not yet implemented", http.StatusNotImplemented)
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
