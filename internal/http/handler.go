package http

import (
	"context"
	"encoding/json"
	"net/http"
	"time"
	"github.com/v-Kaefer/Const-Software-25-02/pkg/jwt"
	"github.com/v-Kaefer/Const-Software-25-02/pkg/user"
)

// Router simples usando net/http para não adicionar dependências.
type Router struct {
	userSvc      *user.Service
	jwtGenerator *jwt.Generator
	mux          *http.ServeMux
}

func NewRouter(userSvc *user.Service, jwtGenerator *jwt.Generator) *Router {
	r := &Router{
		userSvc:      userSvc,
		jwtGenerator: jwtGenerator,
		mux:          http.NewServeMux(),
	}
	r.routes()
	return r
}

func (r *Router) routes() {
	// Rotas públicas
	r.mux.HandleFunc("POST /auth/register", r.handleRegister)
	r.mux.HandleFunc("POST /auth/login", r.handleLogin)
	
	// Rotas protegidas (com JWT)
	r.mux.Handle("POST /users", r.jwtGenerator.Middleware(http.HandlerFunc(r.handleCreateUser)))
	r.mux.HandleFunc("GET /users", r.handleGetUserByEmail) // /users?email=...
}

func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.mux.ServeHTTP(w, req)
}

func (r *Router) handleRegister(w http.ResponseWriter, req *http.Request) {
	type in struct {
		Email    string `json:"email"`
		Name     string `json:"name"`
		Password string `json:"password"`
	}
	var body in
	if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}

	if body.Email == "" || body.Name == "" || body.Password == "" {
		http.Error(w, "email, name and password are required", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(req.Context(), 5*time.Second)
	defer cancel()

	u, err := r.userSvc.RegisterWithPassword(ctx, body.Email, body.Name, body.Password)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"id":    u.ID,
		"email": u.Email,
		"name":  u.Name,
	})
}

func (r *Router) handleLogin(w http.ResponseWriter, req *http.Request) {
	type in struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	var body in
	if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}

	if body.Email == "" || body.Password == "" {
		http.Error(w, "email and password are required", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(req.Context(), 5*time.Second)
	defer cancel()

	u, err := r.userSvc.Authenticate(ctx, body.Email, body.Password)
	if err != nil {
		http.Error(w, "invalid credentials", http.StatusUnauthorized)
		return
	}

	// Gera token JWT válido por 24 horas
	token, err := r.jwtGenerator.GenerateToken(u.ID, u.Email, 24*time.Hour)
	if err != nil {
		http.Error(w, "failed to generate token", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"token": token,
		"user": map[string]interface{}{
			"id":    u.ID,
			"email": u.Email,
			"name":  u.Name,
		},
	})
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
