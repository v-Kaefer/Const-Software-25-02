package http

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"
	"github.com/v-Kaefer/Const-Software-25-02/internal/auth"
	"github.com/v-Kaefer/Const-Software-25-02/internal/config"
	"github.com/v-Kaefer/Const-Software-25-02/pkg/user"
	"github.com/v-Kaefer/Const-Software-25-02/pkg/servico"
	"github.com/v-Kaefer/Const-Software-25-02/pkg/agendamento"
	"gorm.io/gorm"
)

// Router simples usando net/http para não adicionar dependências.
type Router struct {
	userSvc         *user.Service
	servicoSvc      *servico.Service
	agendamentoSvc  *agendamento.Service
	authMiddleware  *auth.Middleware
	mux             *http.ServeMux
}

func NewRouter(
	userSvc *user.Service,
	servicoSvc *servico.Service,
	agendamentoSvc *agendamento.Service,
	authMiddleware *auth.Middleware,
) *Router {
	r := &Router{
		userSvc:        userSvc,
		servicoSvc:     servicoSvc,
		agendamentoSvc: agendamentoSvc,
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

	// ===== SERVICOS ENDPOINTS =====
	
	// POST /api/v1/servicos - Admin only
	r.mux.Handle("POST /api/v1/servicos", r.authMiddleware.Authenticate(
		r.authMiddleware.RequireRole(auth.RoleAdmin)(http.HandlerFunc(r.handleCreateServico)),
	))
	
	// GET /api/v1/servicos - Public (list services)
	r.mux.Handle("GET /api/v1/servicos", http.HandlerFunc(r.handleListServicos))
	
	// GET /api/v1/servicos/{id} - Public
	r.mux.Handle("GET /api/v1/servicos/{id}", http.HandlerFunc(r.handleGetServico))
	
	// PUT /api/v1/servicos/{id} - Admin only
	r.mux.Handle("PUT /api/v1/servicos/{id}", r.authMiddleware.Authenticate(
		r.authMiddleware.RequireRole(auth.RoleAdmin)(http.HandlerFunc(r.handleUpdateServico)),
	))
	
	// DELETE /api/v1/servicos/{id} - Admin only
	r.mux.Handle("DELETE /api/v1/servicos/{id}", r.authMiddleware.Authenticate(
		r.authMiddleware.RequireRole(auth.RoleAdmin)(http.HandlerFunc(r.handleDeleteServico)),
	))

	// ===== AGENDAMENTOS ENDPOINTS =====
	
	// POST /api/v1/agendamentos (Agendar) - Authenticated
	r.mux.Handle("POST /api/v1/agendamentos", r.authMiddleware.Authenticate(
		http.HandlerFunc(r.handleAgendar),
	))
	
	// GET /api/v1/agendamentos - Authenticated (list own or all if admin)
	r.mux.Handle("GET /api/v1/agendamentos", r.authMiddleware.Authenticate(
		http.HandlerFunc(r.handleListAgendamentos),
	))
	
	// GET /api/v1/agendamentos/{id} - Authenticated
	r.mux.Handle("GET /api/v1/agendamentos/{id}", r.authMiddleware.Authenticate(
		http.HandlerFunc(r.handleGetAgendamento),
	))
	
	// PATCH /api/v1/agendamentos/{id}/aprovar - Admin only
	r.mux.Handle("PATCH /api/v1/agendamentos/{id}/aprovar", r.authMiddleware.Authenticate(
		r.authMiddleware.RequireRole(auth.RoleAdmin)(http.HandlerFunc(r.handleAprovarAgendamento)),
	))
	
	// PATCH /api/v1/agendamentos/{id}/cancelar - Admin or owner
	r.mux.Handle("PATCH /api/v1/agendamentos/{id}/cancelar", r.authMiddleware.Authenticate(
		http.HandlerFunc(r.handleCancelarAgendamento),
	))
	
	// PATCH /api/v1/agendamentos/{id}/concluir - Admin only
	r.mux.Handle("PATCH /api/v1/agendamentos/{id}/concluir", r.authMiddleware.Authenticate(
		r.authMiddleware.RequireRole(auth.RoleAdmin)(http.HandlerFunc(r.handleConcluirAgendamento)),
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
		if errors.Is(err, gorm.ErrRecordNotFound) {
			http.Error(w, "not found", http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
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
		if errors.Is(err, gorm.ErrRecordNotFound) {
			http.Error(w, "not found", http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
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
		if errors.Is(err, gorm.ErrRecordNotFound) {
			http.Error(w, "not found", http.StatusNotFound)
		} else {
			http.Error(w, "failed to delete user", http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ===== SERVICO HANDLERS =====

func (r *Router) handleCreateServico(w http.ResponseWriter, req *http.Request) {
	type in struct {
		Nome      string  `json:"nome"`
		Descricao string  `json:"descricao"`
		Duracao   int     `json:"duracao"` // minutos
		Preco     float64 `json:"preco"`
	}
	var body in
	if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}
	ctx, cancel := context.WithTimeout(req.Context(), 5*time.Second)
	defer cancel()
	s, err := r.servicoSvc.Create(ctx, body.Nome, body.Descricao, body.Duracao, body.Preco)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(s)
}

func (r *Router) handleListServicos(w http.ResponseWriter, req *http.Request) {
	ctx, cancel := context.WithTimeout(req.Context(), 5*time.Second)
	defer cancel()
	
	// Parse query params for pagination
	limit := 10
	offset := 0
	
	if l := req.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
			limit = parsed
		}
	}
	
	if o := req.URL.Query().Get("offset"); o != "" {
		if parsed, err := strconv.Atoi(o); err == nil && parsed >= 0 {
			offset = parsed
		}
	}
	
	servicos, total, err := r.servicoSvc.List(ctx, offset, limit)
	if err != nil {
		http.Error(w, "failed to list servicos", http.StatusInternalServerError)
		return
	}
	
	type response struct {
		Data  []servico.Servico `json:"data"`
		Total int64             `json:"total"`
	}
	
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(response{
		Data:  servicos,
		Total: total,
	})
}

func (r *Router) handleGetServico(w http.ResponseWriter, req *http.Request) {
	idStr := req.PathValue("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		http.Error(w, "invalid servico id", http.StatusBadRequest)
		return
	}
	ctx, cancel := context.WithTimeout(req.Context(), 5*time.Second)
	defer cancel()
	s, err := r.servicoSvc.GetByID(ctx, uint(id))
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(s)
}

func (r *Router) handleUpdateServico(w http.ResponseWriter, req *http.Request) {
	idStr := req.PathValue("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		http.Error(w, "invalid servico id", http.StatusBadRequest)
		return
	}
	type in struct {
		Nome      string  `json:"nome"`
		Descricao string  `json:"descricao"`
		Duracao   int     `json:"duracao"`
		Preco     float64 `json:"preco"`
	}
	var body in
	if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}
	ctx, cancel := context.WithTimeout(req.Context(), 5*time.Second)
	defer cancel()
	s, err := r.servicoSvc.Update(ctx, uint(id), body.Nome, body.Descricao, body.Duracao, body.Preco)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			http.Error(w, "not found", http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(s)
}

func (r *Router) handleDeleteServico(w http.ResponseWriter, req *http.Request) {
	idStr := req.PathValue("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		http.Error(w, "invalid servico id", http.StatusBadRequest)
		return
	}
	ctx, cancel := context.WithTimeout(req.Context(), 5*time.Second)
	defer cancel()
	if err := r.servicoSvc.Delete(ctx, uint(id)); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			http.Error(w, "not found", http.StatusNotFound)
		} else {
			http.Error(w, "failed to delete servico", http.StatusInternalServerError)
		}
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// ===== AGENDAMENTO HANDLERS =====

func (r *Router) handleAgendar(w http.ResponseWriter, req *http.Request) {
	username, ok := auth.GetUserFromContext(req.Context())
	if !ok {
		http.Error(w, "unable to determine user", http.StatusUnauthorized)
		return
	}
	
	type in struct {
		ServicoID uint   `json:"servico_id"`
		DataHora  string `json:"data_hora"` // RFC3339 format
	}
	var body in
	if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}
	
	dataHora, err := time.Parse(time.RFC3339, body.DataHora)
	if err != nil {
		http.Error(w, "invalid data_hora format (use RFC3339)", http.StatusBadRequest)
		return
	}
	
	ctx, cancel := context.WithTimeout(req.Context(), 5*time.Second)
	defer cancel()
	
	// Get user by email/username
	user, err := r.userSvc.GetByEmail(ctx, username)
	if err != nil {
		http.Error(w, "user not found", http.StatusUnauthorized)
		return
	}
	
	agendamento, err := r.agendamentoSvc.Agendar(ctx, user.ID, body.ServicoID, dataHora)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(agendamento)
}

func (r *Router) handleListAgendamentos(w http.ResponseWriter, req *http.Request) {
	username, ok := auth.GetUserFromContext(req.Context())
	if !ok {
		http.Error(w, "unable to determine user", http.StatusUnauthorized)
		return
	}
	
	ctx, cancel := context.WithTimeout(req.Context(), 5*time.Second)
	defer cancel()
	
	user, err := r.userSvc.GetByEmail(ctx, username)
	if err != nil {
		http.Error(w, "user not found", http.StatusUnauthorized)
		return
	}
	
	roles, _ := auth.GetRolesFromContext(req.Context())
	isAdmin := false
	for _, role := range roles {
		if role == string(auth.RoleAdmin) {
			isAdmin = true
			break
		}
	}
	
	// Parse query params for pagination
	limit := 10
	offset := 0
	
	if l := req.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
			limit = parsed
		}
	}
	
	if o := req.URL.Query().Get("offset"); o != "" {
		if parsed, err := strconv.Atoi(o); err == nil && parsed >= 0 {
			offset = parsed
		}
	}
	
	var agendamentos []agendamento.Agendamento
	var total int64
	
	if isAdmin {
		// Admin sees all agendamentos
		agendamentos, total, err = r.agendamentoSvc.ListAll(ctx, offset, limit)
	} else {
		// Regular user sees only their agendamentos
		agendamentos, total, err = r.agendamentoSvc.ListByCliente(ctx, user.ID, offset, limit)
	}
	
	if err != nil {
		http.Error(w, "failed to list agendamentos", http.StatusInternalServerError)
		return
	}
	
	type response struct {
		Data  []agendamento.Agendamento `json:"data"`
		Total int64                     `json:"total"`
	}
	
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(response{
		Data:  agendamentos,
		Total: total,
	})
}

func (r *Router) handleGetAgendamento(w http.ResponseWriter, req *http.Request) {
	idStr := req.PathValue("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		http.Error(w, "invalid agendamento id", http.StatusBadRequest)
		return
	}
	
	username, ok := auth.GetUserFromContext(req.Context())
	if !ok {
		http.Error(w, "unable to determine user", http.StatusUnauthorized)
		return
	}
	
	ctx, cancel := context.WithTimeout(req.Context(), 5*time.Second)
	defer cancel()
	
	user, err := r.userSvc.GetByEmail(ctx, username)
	if err != nil {
		http.Error(w, "user not found", http.StatusUnauthorized)
		return
	}
	
	agendamento, err := r.agendamentoSvc.GetByID(ctx, uint(id))
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	
	// Check permission: admin or owner
	roles, _ := auth.GetRolesFromContext(req.Context())
	isAdmin := false
	for _, role := range roles {
		if role == string(auth.RoleAdmin) {
			isAdmin = true
			break
		}
	}
	
	if !isAdmin && agendamento.ClienteID != user.ID {
		http.Error(w, "insufficient permissions", http.StatusForbidden)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(agendamento)
}

func (r *Router) handleAprovarAgendamento(w http.ResponseWriter, req *http.Request) {
	idStr := req.PathValue("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		http.Error(w, "invalid agendamento id", http.StatusBadRequest)
		return
	}
	
	ctx, cancel := context.WithTimeout(req.Context(), 5*time.Second)
	defer cancel()
	
	agendamento, err := r.agendamentoSvc.Aprovar(ctx, uint(id))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(agendamento)
}

func (r *Router) handleCancelarAgendamento(w http.ResponseWriter, req *http.Request) {
	idStr := req.PathValue("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		http.Error(w, "invalid agendamento id", http.StatusBadRequest)
		return
	}
	
	username, ok := auth.GetUserFromContext(req.Context())
	if !ok {
		http.Error(w, "unable to determine user", http.StatusUnauthorized)
		return
	}
	
	ctx, cancel := context.WithTimeout(req.Context(), 5*time.Second)
	defer cancel()
	
	user, err := r.userSvc.GetByEmail(ctx, username)
	if err != nil {
		http.Error(w, "user not found", http.StatusUnauthorized)
		return
	}
	
	agendamento, err := r.agendamentoSvc.GetByID(ctx, uint(id))
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	
	// Check permission: admin or owner
	roles, _ := auth.GetRolesFromContext(req.Context())
	isAdmin := false
	for _, role := range roles {
		if role == string(auth.RoleAdmin) {
			isAdmin = true
			break
		}
	}
	
	if !isAdmin && agendamento.ClienteID != user.ID {
		http.Error(w, "insufficient permissions", http.StatusForbidden)
		return
	}
	
	agendamento, err = r.agendamentoSvc.Cancelar(ctx, uint(id))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(agendamento)
}

func (r *Router) handleConcluirAgendamento(w http.ResponseWriter, req *http.Request) {
	idStr := req.PathValue("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		http.Error(w, "invalid agendamento id", http.StatusBadRequest)
		return
	}
	
	ctx, cancel := context.WithTimeout(req.Context(), 5*time.Second)
	defer cancel()
	
	agendamento, err := r.agendamentoSvc.Concluir(ctx, uint(id))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(agendamento)
}
