package http

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/v-Kaefer/Const-Software-25-02/internal/auth"
	"github.com/v-Kaefer/Const-Software-25-02/internal/config"
	"github.com/v-Kaefer/Const-Software-25-02/pkg/user"
	"github.com/v-Kaefer/Const-Software-25-02/pkg/workspace"
	"gorm.io/gorm"
)

const (
	apiPrefix          = "/api/v1"
	defaultHTTPPage    = 1
	defaultHTTPPerPage = 10
	maxHTTPPerPage     = 50
)

// Router simples usando net/http para não adicionar dependências.
type Router struct {
	userSvc        *user.Service
	projectSvc     *workspace.ProjectService
	taskSvc        *workspace.TaskService
	timeSvc        *workspace.TimeEntryService
	authMiddleware *auth.Middleware
	mux            *http.ServeMux
}

func NewRouter(
	userSvc *user.Service,
	projectSvc *workspace.ProjectService,
	taskSvc *workspace.TaskService,
	timeSvc *workspace.TimeEntryService,
	authMiddleware *auth.Middleware,
) *Router {
	r := &Router{
		userSvc:        userSvc,
		projectSvc:     projectSvc,
		taskSvc:        taskSvc,
		timeSvc:        timeSvc,
		authMiddleware: authMiddleware,
		mux:            http.NewServeMux(),
	}
	r.routes()
	return r
}

func (r *Router) routes() {
	// Usuários
	r.mux.Handle("POST "+apiPrefix+"/users", r.authMiddleware.Authenticate(
		r.authMiddleware.RequireRole(auth.RoleAdmin)(http.HandlerFunc(r.handleCreateUser)),
	))
	r.mux.Handle("GET "+apiPrefix+"/users", r.authMiddleware.Authenticate(
		r.authMiddleware.RequireRole(auth.RoleAdmin)(http.HandlerFunc(r.handleListUsers)),
	))
	r.mux.Handle("GET "+apiPrefix+"/users/{id}", r.authMiddleware.Authenticate(
		http.HandlerFunc(r.handleGetUser),
	))
	r.mux.Handle("PUT "+apiPrefix+"/users/{id}", r.authMiddleware.Authenticate(
		http.HandlerFunc(r.handleUpdateUser),
	))
	r.mux.Handle("PATCH "+apiPrefix+"/users/{id}", r.authMiddleware.Authenticate(
		http.HandlerFunc(r.handlePatchUser),
	))
	r.mux.Handle("DELETE "+apiPrefix+"/users/{id}", r.authMiddleware.Authenticate(
		r.authMiddleware.RequireRole(auth.RoleAdmin)(http.HandlerFunc(r.handleDeleteUser)),
	))

	// Projetos
	r.mux.Handle("POST "+apiPrefix+"/projects", r.authMiddleware.Authenticate(
		r.authMiddleware.RequireRole(auth.RoleAdmin, auth.RoleReviewer)(http.HandlerFunc(r.handleCreateProject)),
	))
	r.mux.Handle("GET "+apiPrefix+"/projects", r.authMiddleware.Authenticate(
		r.authMiddleware.RequireRole(auth.RoleAdmin, auth.RoleReviewer)(http.HandlerFunc(r.handleListProjects)),
	))
	r.mux.Handle("GET "+apiPrefix+"/projects/{id}", r.authMiddleware.Authenticate(
		http.HandlerFunc(r.handleGetProject),
	))
	r.mux.Handle("PUT "+apiPrefix+"/projects/{id}", r.authMiddleware.Authenticate(
		http.HandlerFunc(r.handleUpdateProject),
	))
	r.mux.Handle("DELETE "+apiPrefix+"/projects/{id}", r.authMiddleware.Authenticate(
		http.HandlerFunc(r.handleDeleteProject),
	))

	// Tarefas
	r.mux.Handle("POST "+apiPrefix+"/projects/{projectID}/tasks", r.authMiddleware.Authenticate(
		http.HandlerFunc(r.handleCreateTask),
	))
	r.mux.Handle("GET "+apiPrefix+"/projects/{projectID}/tasks", r.authMiddleware.Authenticate(
		http.HandlerFunc(r.handleListProjectTasks),
	))
	r.mux.Handle("GET "+apiPrefix+"/tasks", r.authMiddleware.Authenticate(
		http.HandlerFunc(r.handleListTasks),
	))
	r.mux.Handle("GET "+apiPrefix+"/tasks/{id}", r.authMiddleware.Authenticate(
		http.HandlerFunc(r.handleGetTask),
	))
	r.mux.Handle("PUT "+apiPrefix+"/tasks/{id}", r.authMiddleware.Authenticate(
		http.HandlerFunc(r.handleUpdateTask),
	))

	// Lançamentos de horas
	r.mux.Handle("POST "+apiPrefix+"/tasks/{taskID}/time-entries", r.authMiddleware.Authenticate(
		http.HandlerFunc(r.handleCreateTimeEntry),
	))
	r.mux.Handle("GET "+apiPrefix+"/tasks/{taskID}/time-entries", r.authMiddleware.Authenticate(
		http.HandlerFunc(r.handleListTaskEntries),
	))
	r.mux.Handle("GET "+apiPrefix+"/time-entries", r.authMiddleware.Authenticate(
		http.HandlerFunc(r.handleListTimeEntries),
	))
	r.mux.Handle("GET "+apiPrefix+"/time-entries/{id}", r.authMiddleware.Authenticate(
		http.HandlerFunc(r.handleGetTimeEntry),
	))
	r.mux.Handle("PUT "+apiPrefix+"/time-entries/{id}", r.authMiddleware.Authenticate(
		http.HandlerFunc(r.handleUpdateTimeEntry),
	))
	r.mux.Handle("PATCH "+apiPrefix+"/time-entries/{id}/approve", r.authMiddleware.Authenticate(
		r.authMiddleware.RequireRole(auth.RoleAdmin)(http.HandlerFunc(r.handleApproveTimeEntry)),
	))
}

func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.mux.ServeHTTP(w, req)
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

// === Handlers: Usuários ===

func (r *Router) handleCreateUser(w http.ResponseWriter, req *http.Request) {
	type in struct {
		Email string `json:"email"`
		Name  string `json:"name"`
	}
	var body in
	if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
		respondError(w, http.StatusBadRequest, "invalid body")
		return
	}
	ctx, cancel := context.WithTimeout(req.Context(), 5*time.Second)
	defer cancel()
	u, err := r.userSvc.Register(ctx, body.Email, body.Name)
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}
	respondJSON(w, http.StatusCreated, u)
}

func (r *Router) handleListUsers(w http.ResponseWriter, req *http.Request) {
	ctx, cancel := context.WithTimeout(req.Context(), 5*time.Second)
	defer cancel()
	users, err := r.userSvc.List(ctx)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to list users")
		return
	}
	respondJSON(w, http.StatusOK, users)
}

func (r *Router) handleGetUser(w http.ResponseWriter, req *http.Request) {
	id, err := parseUintParam(req, "id")
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid user id")
		return
	}

	ctx, cancel := context.WithTimeout(req.Context(), 5*time.Second)
	defer cancel()

	if !r.isAdminOrOwner(ctx, id) {
		respondError(w, http.StatusForbidden, "insufficient permissions")
		return
	}

	u, err := r.userSvc.GetByID(ctx, id)
	if err != nil {
		respondError(w, http.StatusNotFound, "not found")
		return
	}
	respondJSON(w, http.StatusOK, u)
}

func (r *Router) handleUpdateUser(w http.ResponseWriter, req *http.Request) {
	id, err := parseUintParam(req, "id")
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid user id")
		return
	}

	ctx, cancel := context.WithTimeout(req.Context(), 5*time.Second)
	defer cancel()

	if !r.isAdminOrOwner(ctx, id) {
		respondError(w, http.StatusForbidden, "insufficient permissions")
		return
	}

	type in struct {
		Email string `json:"email"`
		Name  string `json:"name"`
	}
	var body in
	if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
		respondError(w, http.StatusBadRequest, "invalid body")
		return
	}

	u, err := r.userSvc.Update(ctx, id, body.Email, body.Name)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			respondError(w, http.StatusNotFound, "not found")
		} else {
			respondError(w, http.StatusBadRequest, err.Error())
		}
		return
	}
	respondJSON(w, http.StatusOK, u)
}

func (r *Router) handlePatchUser(w http.ResponseWriter, req *http.Request) {
	id, err := parseUintParam(req, "id")
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid user id")
		return
	}

	ctx, cancel := context.WithTimeout(req.Context(), 5*time.Second)
	defer cancel()

	if !r.isAdminOrOwner(ctx, id) {
		respondError(w, http.StatusForbidden, "insufficient permissions")
		return
	}

	u, err := r.userSvc.GetByID(ctx, id)
	if err != nil {
		respondError(w, http.StatusNotFound, "not found")
		return
	}

	var body map[string]interface{}
	if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
		respondError(w, http.StatusBadRequest, "invalid body")
		return
	}

	email := u.Email
	name := u.Name
	if val, ok := body["email"].(string); ok {
		email = val
	}
	if val, ok := body["name"].(string); ok {
		name = val
	}

	u, err = r.userSvc.Update(ctx, id, email, name)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			respondError(w, http.StatusNotFound, "not found")
		} else {
			respondError(w, http.StatusBadRequest, err.Error())
		}
		return
	}
	respondJSON(w, http.StatusOK, u)
}

func (r *Router) handleDeleteUser(w http.ResponseWriter, req *http.Request) {
	id, err := parseUintParam(req, "id")
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid user id")
		return
	}

	ctx, cancel := context.WithTimeout(req.Context(), 5*time.Second)
	defer cancel()

	if err := r.userSvc.Delete(ctx, id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			respondError(w, http.StatusNotFound, "not found")
		} else {
			respondError(w, http.StatusInternalServerError, "failed to delete user")
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// === Handlers: Projetos ===

func (r *Router) handleCreateProject(w http.ResponseWriter, req *http.Request) {
	type in struct {
		Name        string  `json:"name"`
		ClientName  string  `json:"clientName"`
		Description string  `json:"description"`
		StartDate   string  `json:"startDate"`
		EndDate     *string `json:"endDate"`
	}
	var body in
	if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
		respondError(w, http.StatusBadRequest, "invalid body")
		return
	}
	start, err := parseTimeISO(body.StartDate)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid startDate")
		return
	}
	endTime, err := parseOptionalTimeISO(body.EndDate)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid endDate")
		return
	}

	ctx, cancel := context.WithTimeout(req.Context(), 5*time.Second)
	defer cancel()
	current, err := r.currentUser(ctx)
	if err != nil {
		respondError(w, http.StatusForbidden, "user not registered in system")
		return
	}

	project, err := r.projectSvc.CreateProject(ctx, workspace.ProjectInput{
		Name:        body.Name,
		ClientName:  body.ClientName,
		Description: body.Description,
		StartDate:   start,
		EndDate:     endTime,
		OwnerID:     current.ID,
	})
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}
	respondJSON(w, http.StatusCreated, project)
}

func (r *Router) handleListProjects(w http.ResponseWriter, req *http.Request) {
	page, pageSize := paginationParams(req)
	statuses := projectStatusesFromQuery(req.URL.Query()["status"])
	client := req.URL.Query().Get("client")

	ctx, cancel := context.WithTimeout(req.Context(), 5*time.Second)
	defer cancel()

	filter := workspace.ProjectFilter{
		Page:     page,
		PageSize: pageSize,
		Status:   statuses,
		Client:   client,
	}

	if !r.hasAnyRole(ctx, auth.RoleAdmin) {
		current, err := r.currentUser(ctx)
		if err != nil {
			respondError(w, http.StatusForbidden, "user not registered in system")
			return
		}
		filter.OwnerID = &current.ID
	}

	result, err := r.projectSvc.ListProjects(ctx, filter)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "unable to list projects")
		return
	}

	respondPaginated(w, result.Items, page, pageSize, result.Total)
}

func (r *Router) handleGetProject(w http.ResponseWriter, req *http.Request) {
	projectID, err := parseUintParam(req, "id")
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid project id")
		return
	}

	ctx, cancel := context.WithTimeout(req.Context(), 5*time.Second)
	defer cancel()

	project, err := r.projectSvc.GetProject(ctx, projectID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			respondError(w, http.StatusNotFound, "project not found")
		} else {
			respondError(w, http.StatusInternalServerError, "failed to load project")
		}
		return
	}

	if !r.canManageProject(ctx, project) {
		respondError(w, http.StatusForbidden, "insufficient permissions")
		return
	}

	respondJSON(w, http.StatusOK, project)
}

func (r *Router) handleUpdateProject(w http.ResponseWriter, req *http.Request) {
	projectID, err := parseUintParam(req, "id")
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid project id")
		return
	}
	type in struct {
		Name        string  `json:"name"`
		ClientName  string  `json:"clientName"`
		Description string  `json:"description"`
		Status      string  `json:"status"`
		StartDate   string  `json:"startDate"`
		EndDate     *string `json:"endDate"`
	}
	var body in
	if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
		respondError(w, http.StatusBadRequest, "invalid body")
		return
	}

	start, err := parseTimeISO(body.StartDate)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid startDate")
		return
	}
	endTime, err := parseOptionalTimeISO(body.EndDate)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid endDate")
		return
	}
	status := workspace.ProjectStatus(strings.ToLower(body.Status))

	ctx, cancel := context.WithTimeout(req.Context(), 5*time.Second)
	defer cancel()

	project, err := r.projectSvc.GetProject(ctx, projectID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			respondError(w, http.StatusNotFound, "project not found")
		} else {
			respondError(w, http.StatusInternalServerError, "failed to load project")
		}
		return
	}

	if !r.canManageProject(ctx, project) {
		respondError(w, http.StatusForbidden, "insufficient permissions")
		return
	}

	updated, err := r.projectSvc.UpdateProject(ctx, projectID, workspace.ProjectUpdateInput{
		Name:        body.Name,
		ClientName:  body.ClientName,
		Description: body.Description,
		Status:      status,
		StartDate:   start,
		EndDate:     endTime,
	})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			respondError(w, http.StatusNotFound, "project not found")
		} else {
			respondError(w, http.StatusBadRequest, err.Error())
		}
		return
	}
	respondJSON(w, http.StatusOK, updated)
}

func (r *Router) handleDeleteProject(w http.ResponseWriter, req *http.Request) {
	projectID, err := parseUintParam(req, "id")
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid project id")
		return
	}

	ctx, cancel := context.WithTimeout(req.Context(), 5*time.Second)
	defer cancel()

	project, err := r.projectSvc.GetProject(ctx, projectID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			respondError(w, http.StatusNotFound, "project not found")
		} else {
			respondError(w, http.StatusInternalServerError, "failed to load project")
		}
		return
	}

	if !r.canManageProject(ctx, project) {
		respondError(w, http.StatusForbidden, "insufficient permissions")
		return
	}

	if err := r.projectSvc.DeleteProject(ctx, projectID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			respondError(w, http.StatusNotFound, "project not found")
		} else {
			respondError(w, http.StatusInternalServerError, "failed to delete project")
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// === Handlers: Tarefas ===

func (r *Router) handleCreateTask(w http.ResponseWriter, req *http.Request) {
	projectID, err := parseUintParam(req, "projectID")
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid project id")
		return
	}
	type in struct {
		Title       string  `json:"title"`
		Description string  `json:"description"`
		AssigneeID  uint    `json:"assigneeId"`
		DueDate     *string `json:"dueDate"`
	}
	var body in
	if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
		respondError(w, http.StatusBadRequest, "invalid body")
		return
	}
	due, err := parseOptionalTimeISO(body.DueDate)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid dueDate")
		return
	}

	ctx, cancel := context.WithTimeout(req.Context(), 5*time.Second)
	defer cancel()

	project, err := r.projectSvc.GetProject(ctx, projectID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			respondError(w, http.StatusNotFound, "project not found")
		} else {
			respondError(w, http.StatusInternalServerError, "failed to load project")
		}
		return
	}
	if !r.canManageProject(ctx, project) {
		respondError(w, http.StatusForbidden, "insufficient permissions")
		return
	}

	task, err := r.taskSvc.CreateTask(ctx, workspace.TaskInput{
		ProjectID:   projectID,
		Title:       body.Title,
		Description: body.Description,
		AssigneeID:  body.AssigneeID,
		DueDate:     due,
	})
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}
	respondJSON(w, http.StatusCreated, task)
}

func (r *Router) handleListProjectTasks(w http.ResponseWriter, req *http.Request) {
	projectID, err := parseUintParam(req, "projectID")
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid project id")
		return
	}
	page, pageSize := paginationParams(req)
	statuses := taskStatusesFromQuery(req.URL.Query()["status"])

	ctx, cancel := context.WithTimeout(req.Context(), 5*time.Second)
	defer cancel()

	project, err := r.projectSvc.GetProject(ctx, projectID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			respondError(w, http.StatusNotFound, "project not found")
		} else {
			respondError(w, http.StatusInternalServerError, "failed to load project")
		}
		return
	}
	if !r.canManageProject(ctx, project) {
		respondError(w, http.StatusForbidden, "insufficient permissions")
		return
	}

	result, err := r.taskSvc.ListTasks(ctx, workspace.TaskFilter{
		ProjectID: projectID,
		Status:    statuses,
		Page:      page,
		PageSize:  pageSize,
	})
	if err != nil {
		respondError(w, http.StatusInternalServerError, "unable to list tasks")
		return
	}
	respondPaginated(w, result.Items, page, pageSize, result.Total)
}

func (r *Router) handleListTasks(w http.ResponseWriter, req *http.Request) {
	page, pageSize := paginationParams(req)
	statuses := taskStatusesFromQuery(req.URL.Query()["status"])

	ctx, cancel := context.WithTimeout(req.Context(), 5*time.Second)
	defer cancel()

	filter := workspace.TaskFilter{
		Status:   statuses,
		Page:     page,
		PageSize: pageSize,
	}

	if projectIDStr := req.URL.Query().Get("projectId"); projectIDStr != "" && r.hasAnyRole(ctx, auth.RoleAdmin) {
		if projID, err := strconv.ParseUint(projectIDStr, 10, 32); err == nil {
			filter.ProjectID = uint(projID)
		}
	}

	if r.hasAnyRole(ctx, auth.RoleAdmin) {
		if assigneeStr := req.URL.Query().Get("assigneeId"); assigneeStr != "" {
			if assigneeID, err := strconv.ParseUint(assigneeStr, 10, 32); err == nil {
				id := uint(assigneeID)
				filter.AssigneeID = &id
			}
		}
	} else {
		current, err := r.currentUser(ctx)
		if err != nil {
			respondError(w, http.StatusForbidden, "user not registered in system")
			return
		}
		id := current.ID
		filter.AssigneeID = &id
	}

	result, err := r.taskSvc.ListTasks(ctx, filter)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "unable to list tasks")
		return
	}
	respondPaginated(w, result.Items, page, pageSize, result.Total)
}

func (r *Router) handleGetTask(w http.ResponseWriter, req *http.Request) {
	taskID, err := parseUintParam(req, "id")
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid task id")
		return
	}

	ctx, cancel := context.WithTimeout(req.Context(), 5*time.Second)
	defer cancel()

	task, err := r.taskSvc.GetTask(ctx, taskID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			respondError(w, http.StatusNotFound, "task not found")
		} else {
			respondError(w, http.StatusInternalServerError, "failed to load task")
		}
		return
	}

	if !r.canAccessTask(ctx, task) {
		respondError(w, http.StatusForbidden, "insufficient permissions")
		return
	}

	respondJSON(w, http.StatusOK, task)
}

func (r *Router) handleUpdateTask(w http.ResponseWriter, req *http.Request) {
	taskID, err := parseUintParam(req, "id")
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid task id")
		return
	}
	type in struct {
		Title       string  `json:"title"`
		Description string  `json:"description"`
		Status      string  `json:"status"`
		AssigneeID  uint    `json:"assigneeId"`
		DueDate     *string `json:"dueDate"`
	}
	var body in
	if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
		respondError(w, http.StatusBadRequest, "invalid body")
		return
	}
	due, err := parseOptionalTimeISO(body.DueDate)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid dueDate")
		return
	}

	ctx, cancel := context.WithTimeout(req.Context(), 5*time.Second)
	defer cancel()

	task, err := r.taskSvc.GetTask(ctx, taskID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			respondError(w, http.StatusNotFound, "task not found")
		} else {
			respondError(w, http.StatusInternalServerError, "failed to load task")
		}
		return
	}

	if !r.canManageProject(ctx, &task.Project) {
		respondError(w, http.StatusForbidden, "insufficient permissions")
		return
	}

	updated, err := r.taskSvc.UpdateTask(ctx, taskID, workspace.TaskUpdateInput{
		Title:       body.Title,
		Description: body.Description,
		Status:      workspace.TaskStatus(strings.ToLower(body.Status)),
		AssigneeID:  body.AssigneeID,
		DueDate:     due,
	})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			respondError(w, http.StatusNotFound, "task not found")
		} else {
			respondError(w, http.StatusBadRequest, err.Error())
		}
		return
	}

	respondJSON(w, http.StatusOK, updated)
}

// === Handlers: Lançamento de horas ===

func (r *Router) handleCreateTimeEntry(w http.ResponseWriter, req *http.Request) {
	taskID, err := parseUintParam(req, "taskID")
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid task id")
		return
	}

	type in struct {
		EntryDate string  `json:"entryDate"`
		Hours     float64 `json:"hours"`
		Notes     string  `json:"notes"`
	}
	var body in
	if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
		respondError(w, http.StatusBadRequest, "invalid body")
		return
	}
	entryDate, err := parseTimeISO(body.EntryDate)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid entryDate")
		return
	}

	ctx, cancel := context.WithTimeout(req.Context(), 5*time.Second)
	defer cancel()

	task, err := r.taskSvc.GetTask(ctx, taskID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			respondError(w, http.StatusNotFound, "task not found")
		} else {
			respondError(w, http.StatusInternalServerError, "failed to load task")
		}
		return
	}

	current, err := r.currentUser(ctx)
	if err != nil {
		respondError(w, http.StatusForbidden, "user not registered in system")
		return
	}
	if !r.canLogTimeOnTask(ctx, task, current) {
		respondError(w, http.StatusForbidden, "insufficient permissions")
		return
	}

	entry, err := r.timeSvc.LogTime(ctx, workspace.TimeEntryInput{
		TaskID:    taskID,
		UserID:    current.ID,
		EntryDate: entryDate,
		Hours:     body.Hours,
		Notes:     body.Notes,
	})
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	respondJSON(w, http.StatusCreated, entry)
}

func (r *Router) handleListTaskEntries(w http.ResponseWriter, req *http.Request) {
	taskID, err := parseUintParam(req, "taskID")
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid task id")
		return
	}
	page, pageSize := paginationParams(req)

	ctx, cancel := context.WithTimeout(req.Context(), 5*time.Second)
	defer cancel()

	task, err := r.taskSvc.GetTask(ctx, taskID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			respondError(w, http.StatusNotFound, "task not found")
		} else {
			respondError(w, http.StatusInternalServerError, "failed to load task")
		}
		return
	}

	current, err := r.currentUser(ctx)
	if err != nil {
		respondError(w, http.StatusForbidden, "user not registered in system")
		return
	}

	filter := workspace.TimeEntryFilter{
		TaskID:   &taskID,
		Page:     page,
		PageSize: pageSize,
	}

	if r.canManageProject(ctx, &task.Project) {
		// nothing, managers can see all entries
	} else if task.AssigneeID == current.ID {
		filter.UserID = &current.ID
	} else {
		respondError(w, http.StatusForbidden, "insufficient permissions")
		return
	}

	result, err := r.timeSvc.ListEntries(ctx, filter)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "unable to list time entries")
		return
	}
	respondPaginated(w, result.Items, page, pageSize, result.Total)
}

func (r *Router) handleListTimeEntries(w http.ResponseWriter, req *http.Request) {
	page, pageSize := paginationParams(req)
	ctx, cancel := context.WithTimeout(req.Context(), 5*time.Second)
	defer cancel()

	filter := workspace.TimeEntryFilter{
		Page:     page,
		PageSize: pageSize,
	}

	if approvedParam := req.URL.Query().Get("approved"); approvedParam != "" {
		if approved, err := strconv.ParseBool(approvedParam); err == nil {
			filter.Approved = &approved
		}
	}

	if taskIDStr := req.URL.Query().Get("taskId"); taskIDStr != "" && r.hasAnyRole(ctx, auth.RoleAdmin, auth.RoleReviewer) {
		if taskID, err := strconv.ParseUint(taskIDStr, 10, 32); err == nil {
			id := uint(taskID)
			filter.TaskID = &id
		}
	}

	if r.hasAnyRole(ctx, auth.RoleAdmin) {
		if userIDStr := req.URL.Query().Get("userId"); userIDStr != "" {
			if userID, err := strconv.ParseUint(userIDStr, 10, 32); err == nil {
				id := uint(userID)
				filter.UserID = &id
			}
		}
	} else {
		current, err := r.currentUser(ctx)
		if err != nil {
			respondError(w, http.StatusForbidden, "user not registered in system")
			return
		}
		id := current.ID
		filter.UserID = &id
	}

	result, err := r.timeSvc.ListEntries(ctx, filter)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "unable to list time entries")
		return
	}
	respondPaginated(w, result.Items, page, pageSize, result.Total)
}

func (r *Router) handleGetTimeEntry(w http.ResponseWriter, req *http.Request) {
	entryID, err := parseUintParam(req, "id")
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid entry id")
		return
	}

	ctx, cancel := context.WithTimeout(req.Context(), 5*time.Second)
	defer cancel()

	entry, err := r.timeSvc.GetEntry(ctx, entryID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			respondError(w, http.StatusNotFound, "entry not found")
		} else {
			respondError(w, http.StatusInternalServerError, "failed to load entry")
		}
		return
	}

	if ok := r.canViewEntry(ctx, entry); !ok {
		respondError(w, http.StatusForbidden, "insufficient permissions")
		return
	}

	respondJSON(w, http.StatusOK, entry)
}

func (r *Router) handleUpdateTimeEntry(w http.ResponseWriter, req *http.Request) {
	entryID, err := parseUintParam(req, "id")
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid entry id")
		return
	}
	type in struct {
		EntryDate string  `json:"entryDate"`
		Hours     float64 `json:"hours"`
		Notes     string  `json:"notes"`
	}
	var body in
	if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
		respondError(w, http.StatusBadRequest, "invalid body")
		return
	}
	entryDate, err := parseTimeISO(body.EntryDate)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid entryDate")
		return
	}

	ctx, cancel := context.WithTimeout(req.Context(), 5*time.Second)
	defer cancel()

	entry, err := r.timeSvc.GetEntry(ctx, entryID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			respondError(w, http.StatusNotFound, "entry not found")
		} else {
			respondError(w, http.StatusInternalServerError, "failed to load entry")
		}
		return
	}

	current, err := r.currentUser(ctx)
	if err != nil {
		respondError(w, http.StatusForbidden, "user not registered in system")
		return
	}
	if entry.UserID != current.ID && !r.hasAnyRole(ctx, auth.RoleAdmin) {
		respondError(w, http.StatusForbidden, "insufficient permissions")
		return
	}

	updated, err := r.timeSvc.UpdateEntry(ctx, entryID, workspace.TimeEntryUpdateInput{
		EntryDate: entryDate,
		Hours:     body.Hours,
		Notes:     body.Notes,
	})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			respondError(w, http.StatusNotFound, "entry not found")
		} else {
			respondError(w, http.StatusBadRequest, err.Error())
		}
		return
	}

	respondJSON(w, http.StatusOK, updated)
}

func (r *Router) handleApproveTimeEntry(w http.ResponseWriter, req *http.Request) {
	entryID, err := parseUintParam(req, "id")
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid entry id")
		return
	}

	ctx, cancel := context.WithTimeout(req.Context(), 5*time.Second)
	defer cancel()

	current, err := r.currentUser(ctx)
	if err != nil {
		respondError(w, http.StatusForbidden, "user not registered in system")
		return
	}

	entry, err := r.timeSvc.ApproveEntry(ctx, entryID, current.ID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			respondError(w, http.StatusNotFound, "entry not found")
		} else {
			respondError(w, http.StatusBadRequest, err.Error())
		}
		return
	}

	respondJSON(w, http.StatusOK, entry)
}

// === Helpers ===

func respondJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func respondError(w http.ResponseWriter, status int, message string) {
	respondJSON(w, status, map[string]string{"error": message})
}

func respondPaginated(w http.ResponseWriter, data interface{}, page, pageSize int, total int64) {
	respondJSON(w, http.StatusOK, map[string]interface{}{
		"data": data,
		"pagination": map[string]interface{}{
			"page":     page,
			"pageSize": pageSize,
			"total":    total,
		},
	})
}

func paginationParams(req *http.Request) (int, int) {
	page := defaultHTTPPage
	if value := req.URL.Query().Get("page"); value != "" {
		if v, err := strconv.Atoi(value); err == nil && v > 0 {
			page = v
		}
	}
	pageSize := defaultHTTPPerPage
	if value := req.URL.Query().Get("pageSize"); value != "" {
		if v, err := strconv.Atoi(value); err == nil && v > 0 {
			pageSize = v
		}
	}
	if pageSize > maxHTTPPerPage {
		pageSize = maxHTTPPerPage
	}
	return page, pageSize
}

func parseUintParam(req *http.Request, key string) (uint, error) {
	value := req.PathValue(key)
	id, err := strconv.ParseUint(value, 10, 32)
	if err != nil {
		return 0, err
	}
	return uint(id), nil
}

func parseTimeISO(value string) (time.Time, error) {
	return time.Parse(time.RFC3339, value)
}

func parseOptionalTimeISO(value *string) (*time.Time, error) {
	if value == nil || strings.TrimSpace(*value) == "" {
		return nil, nil
	}
	t, err := time.Parse(time.RFC3339, *value)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func projectStatusesFromQuery(values []string) []workspace.ProjectStatus {
	var statuses []workspace.ProjectStatus
	for _, raw := range values {
		if raw == "" {
			continue
		}
		statuses = append(statuses, workspace.ProjectStatus(strings.ToLower(raw)))
	}
	return statuses
}

func taskStatusesFromQuery(values []string) []workspace.TaskStatus {
	var statuses []workspace.TaskStatus
	for _, raw := range values {
		if raw == "" {
			continue
		}
		statuses = append(statuses, workspace.TaskStatus(strings.ToLower(raw)))
	}
	return statuses
}

func (r *Router) currentUser(ctx context.Context) (*user.User, error) {
	username, ok := auth.GetUserFromContext(ctx)
	if !ok {
		return nil, errors.New("no user in context")
	}
	return r.userSvc.GetByEmail(ctx, username)
}

func (r *Router) hasAnyRole(ctx context.Context, roles ...auth.Role) bool {
	userRoles, ok := auth.GetRolesFromContext(ctx)
	if !ok {
		return false
	}
	for _, target := range roles {
		for _, actual := range userRoles {
			if actual == string(target) {
				return true
			}
		}
	}
	return false
}

func (r *Router) isAdmin(ctx context.Context) bool {
	return r.hasAnyRole(ctx, auth.RoleAdmin)
}

// isAdminOrOwner checks if the user is admin or owns the resource
func (r *Router) isAdminOrOwner(ctx context.Context, userID uint) bool {
	if r.isAdmin(ctx) {
		return true
	}
	current, err := r.currentUser(ctx)
	if err != nil {
		return false
	}
	return current.ID == userID
}

func (r *Router) canManageProject(ctx context.Context, project *workspace.Project) bool {
	if r.isAdmin(ctx) {
		return true
	}
	current, err := r.currentUser(ctx)
	if err != nil {
		return false
	}
	return project.OwnerID == current.ID
}

func (r *Router) canAccessTask(ctx context.Context, task *workspace.Task) bool {
	if r.isAdmin(ctx) {
		return true
	}
	current, err := r.currentUser(ctx)
	if err != nil {
		return false
	}
	if task.AssigneeID == current.ID {
		return true
	}
	return task.Project.OwnerID == current.ID
}

func (r *Router) canLogTimeOnTask(ctx context.Context, task *workspace.Task, current *user.User) bool {
	if r.isAdmin(ctx) {
		return true
	}
	if task.AssigneeID == current.ID {
		return true
	}
	return task.Project.OwnerID == current.ID && r.hasAnyRole(ctx, auth.RoleReviewer)
}

func (r *Router) canViewEntry(ctx context.Context, entry *workspace.TimeEntry) bool {
	if r.isAdmin(ctx) {
		return true
	}
	current, err := r.currentUser(ctx)
	if err != nil {
		return false
	}
	if entry.UserID == current.ID {
		return true
	}
	task, err := r.taskSvc.GetTask(ctx, entry.TaskID)
	if err != nil {
		return false
	}
	return task.Project.OwnerID == current.ID
}
