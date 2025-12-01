package http_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	httpapi "github.com/v-Kaefer/Const-Software-25-02/internal/http"
	"github.com/v-Kaefer/Const-Software-25-02/pkg/user"
	"github.com/v-Kaefer/Const-Software-25-02/pkg/workspace"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func newTestServer(t *testing.T) *httptest.Server {
	t.Helper()

	// DB SQLite em memória para testes (rápido e isolado)
	dsn := fmt.Sprintf("file:testdb_%d?mode=memory&cache=shared", time.Now().UnixNano())
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("gorm open: %v", err)
	}
	if err := db.AutoMigrate(&user.User{}, &workspace.Project{}, &workspace.Task{}, &workspace.TimeEntry{}); err != nil {
		t.Fatalf("automigrate: %v", err)
	}

	repo := user.NewRepo(db)
	svc := user.NewService(db, repo)
	if _, err := svc.Register(context.Background(), "test-user", "Mock Admin"); err != nil {
		t.Fatalf("seed admin user: %v", err)
	}
	projectSvc := workspace.NewProjectService(db)
	taskSvc := workspace.NewTaskService(db)
	timeSvc := workspace.NewTimeEntryService(db)

	// Create a mock auth middleware for testing (empty config is fine for tests without actual auth)
	mockAuthMiddleware := httpapi.NewMockAuthMiddleware()
	router := httpapi.NewRouter(svc, projectSvc, taskSvc, timeSvc, mockAuthMiddleware)

	return httptest.NewServer(router)
}

func TestHTTP_CreateAndGetUser(t *testing.T) {
	ts := newTestServer(t)
	defer ts.Close()

	// 1) cria usuário
	body := []byte(`{"email":"alice@example.com","name":"Alice"}`)
	resp, err := http.Post(ts.URL+"/api/v1/users", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("POST /users: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("POST status = %d, want 201", resp.StatusCode)
	}

	var created user.User
	if err := json.NewDecoder(resp.Body).Decode(&created); err != nil {
		t.Fatalf("decode created: %v", err)
	}
	if created.ID == 0 {
		t.Fatalf("expected created user to have ID > 0")
	}

	// 2) lista todos os usuários (GET /users agora retorna array)
	getResp, err := http.Get(ts.URL + "/api/v1/users")
	if err != nil {
		t.Fatalf("GET /users: %v", err)
	}
	defer getResp.Body.Close()
	if getResp.StatusCode != http.StatusOK {
		t.Fatalf("GET status = %d, want 200", getResp.StatusCode)
	}

	var users []user.User
	if err := json.NewDecoder(getResp.Body).Decode(&users); err != nil {
		t.Fatalf("decode users: %v", err)
	}
	if len(users) == 0 {
		t.Fatalf("expected at least one user")
	}

	// Verifica se o usuário criado está na lista
	found := false
	for _, u := range users {
		if u.Email == "alice@example.com" && u.Name == "Alice" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("created user not found in list")
	}
}

func TestHTTP_ProjectTaskTimeEntryFlow(t *testing.T) {
	ts := newTestServer(t)
	defer ts.Close()

	// cria usuário para ser responsável
	userBody := []byte(`{"email":"dev@example.com","name":"Dev"}`)
	userResp, err := http.Post(ts.URL+"/api/v1/users", "application/json", bytes.NewReader(userBody))
	if err != nil {
		t.Fatalf("POST /api/v1/users: %v", err)
	}
	defer userResp.Body.Close()
	if userResp.StatusCode != http.StatusCreated {
		t.Fatalf("POST user status = %d", userResp.StatusCode)
	}
	var assignee user.User
	if err := json.NewDecoder(userResp.Body).Decode(&assignee); err != nil {
		t.Fatalf("decode user: %v", err)
	}

	start := time.Now().UTC()
	projectPayload := fmt.Sprintf(`{"name":"New Project","clientName":"ACME","description":"Desc","startDate":"%s"}`, start.Format(time.RFC3339))
	projectResp, err := http.Post(ts.URL+"/api/v1/projects", "application/json", bytes.NewReader([]byte(projectPayload)))
	if err != nil {
		t.Fatalf("POST /api/v1/projects: %v", err)
	}
	defer projectResp.Body.Close()
	if projectResp.StatusCode != http.StatusCreated {
		t.Fatalf("POST /api/v1/projects status = %d", projectResp.StatusCode)
	}
	var project workspace.Project
	if err := json.NewDecoder(projectResp.Body).Decode(&project); err != nil {
		t.Fatalf("decode project: %v", err)
	}

	taskPayload := fmt.Sprintf(`{"title":"Task 1","description":"desc","assigneeId":%d}`, assignee.ID)
	taskURL := fmt.Sprintf("%s/api/v1/projects/%d/tasks", ts.URL, project.ID)
	taskResp, err := http.Post(taskURL, "application/json", bytes.NewReader([]byte(taskPayload)))
	if err != nil {
		t.Fatalf("POST %s: %v", taskURL, err)
	}
	defer taskResp.Body.Close()
	if taskResp.StatusCode != http.StatusCreated {
		t.Fatalf("POST task status = %d", taskResp.StatusCode)
	}
	var task workspace.Task
	if err := json.NewDecoder(taskResp.Body).Decode(&task); err != nil {
		t.Fatalf("decode task: %v", err)
	}

	entryPayload := fmt.Sprintf(`{"entryDate":"%s","hours":2.5,"notes":"dev work"}`, start.Add(time.Hour).Format(time.RFC3339))
	entryURL := fmt.Sprintf("%s/api/v1/tasks/%d/time-entries", ts.URL, task.ID)
	entryResp, err := http.Post(entryURL, "application/json", bytes.NewReader([]byte(entryPayload)))
	if err != nil {
		t.Fatalf("POST %s: %v", entryURL, err)
	}
	defer entryResp.Body.Close()
	if entryResp.StatusCode != http.StatusCreated {
		t.Fatalf("POST time entry status = %d", entryResp.StatusCode)
	}
	var entry workspace.TimeEntry
	if err := json.NewDecoder(entryResp.Body).Decode(&entry); err != nil {
		t.Fatalf("decode entry: %v", err)
	}
	if entry.ID == 0 {
		t.Fatal("expected entry ID")
	}

	approveReq, err := http.NewRequest(http.MethodPatch, fmt.Sprintf("%s/api/v1/time-entries/%d/approve", ts.URL, entry.ID), nil)
	if err != nil {
		t.Fatalf("create approve request: %v", err)
	}
	approveResp, err := http.DefaultClient.Do(approveReq)
	if err != nil {
		t.Fatalf("PATCH approve: %v", err)
	}
	defer approveResp.Body.Close()
	if approveResp.StatusCode != http.StatusOK {
		t.Fatalf("approve status = %d", approveResp.StatusCode)
	}
	var approved workspace.TimeEntry
	if err := json.NewDecoder(approveResp.Body).Decode(&approved); err != nil {
		t.Fatalf("decode approved entry: %v", err)
	}
	if approved.ApprovedAt == nil {
		t.Fatal("expected approvedAt timestamp")
	}
}

func TestHTTP_TrailingSlash(t *testing.T) {
ts := newTestServer(t)
defer ts.Close()

// Teste POST com barra final
body := []byte(`{"email":"trailing@example.com","name":"Trailing"}`)
resp, err := http.Post(ts.URL+"/api/v1/users/", "application/json", bytes.NewReader(body))
if err != nil {
t.Fatalf("POST /users/: %v", err)
}
defer resp.Body.Close()
if resp.StatusCode != http.StatusCreated {
t.Fatalf("POST /users/ status = %d, want 201", resp.StatusCode)
}

// Teste GET com barra final
getResp, err := http.Get(ts.URL + "/api/v1/users/")
if err != nil {
t.Fatalf("GET /users/: %v", err)
}
defer getResp.Body.Close()
if getResp.StatusCode != http.StatusOK {
t.Fatalf("GET /users/ status = %d, want 200", getResp.StatusCode)
}
}
