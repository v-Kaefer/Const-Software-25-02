package http

import (
	"context"
	"encoding/json"
	"net/http"
	"time"
	"github.com/v-Kaefer/Const-Software-25-02/pkg/user"
)

// Router simples usando net/http para não adicionar dependências.
type Router struct {
	userSvc *user.Service
	mux     *http.ServeMux
}

func NewRouter(userSvc *user.Service) *Router {
	r := &Router{userSvc: userSvc, mux: http.NewServeMux()}
	r.routes()
	return r
}

func (r *Router) routes() {
	r.mux.HandleFunc("POST /users", r.handleCreateUser)
	r.mux.HandleFunc("GET /users", r.handleGetUserByEmail) // /users?email=...
}

func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.mux.ServeHTTP(w, req)
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
