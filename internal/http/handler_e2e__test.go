package http_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	httpapi "github.com/v-Kaefer/Const-Software-25-02/internal/http"
	"github.com/v-Kaefer/Const-Software-25-02/pkg/user"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func newTestServer(t *testing.T) *httptest.Server {
	t.Helper()

	// DB SQLite em memória para testes (rápido e isolado)
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("gorm open: %v", err)
	}
	if err := db.AutoMigrate(&user.User{}); err != nil {
		t.Fatalf("automigrate: %v", err)
	}

	repo := user.NewRepo(db)
	svc := user.NewService(db, repo)
	
	// Create a mock auth middleware for testing (empty config is fine for tests without actual auth)
	mockAuthMiddleware := httpapi.NewMockAuthMiddleware()
	router := httpapi.NewRouter(svc, mockAuthMiddleware) // seu handler implementa http.Handler

	return httptest.NewServer(router)
}

func TestHTTP_CreateAndGetUser(t *testing.T) {
	ts := newTestServer(t)
	defer ts.Close()

	// 1) cria usuário
	body := []byte(`{"email":"alice@example.com","name":"Alice"}`)
	resp, err := http.Post(ts.URL+"/users", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("POST /users: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("POST status = %d, want 200", resp.StatusCode)
	}

	var created user.User
	if err := json.NewDecoder(resp.Body).Decode(&created); err != nil {
		t.Fatalf("decode created: %v", err)
	}
	if created.ID == 0 {
		t.Fatalf("expected created user to have ID > 0")
	}

	// 2) busca por email
	getResp, err := http.Get(ts.URL + "/users?email=alice@example.com")
	if err != nil {
		t.Fatalf("GET /users: %v", err)
	}
	defer getResp.Body.Close()
	if getResp.StatusCode != http.StatusOK {
		t.Fatalf("GET status = %d, want 200", getResp.StatusCode)
	}

	var got user.User
	if err := json.NewDecoder(getResp.Body).Decode(&got); err != nil {
		t.Fatalf("decode got: %v", err)
	}
	if got.Email != "alice@example.com" || got.Name != "Alice" {
		t.Fatalf("unexpected user: %+v", got)
	}
}
