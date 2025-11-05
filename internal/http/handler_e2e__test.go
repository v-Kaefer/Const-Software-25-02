package http_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	httpapi "github.com/v-Kaefer/Const-Software-25-02/internal/http"
	"github.com/v-Kaefer/Const-Software-25-02/pkg/jwt"
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
	jwtGen := jwt.NewGenerator("test-secret")
	router := httpapi.NewRouter(svc, jwtGen) // seu handler implementa http.Handler

	return httptest.NewServer(router)
}

func TestHTTP_CreateAndGetUser(t *testing.T) {
	ts := newTestServer(t)
	defer ts.Close()

	// 1) registra usuário com senha
	regBody := []byte(`{"email":"alice@example.com","name":"Alice","password":"password123"}`)
	regResp, err := http.Post(ts.URL+"/auth/register", "application/json", bytes.NewReader(regBody))
	if err != nil {
		t.Fatalf("POST /auth/register: %v", err)
	}
	defer regResp.Body.Close()
	if regResp.StatusCode != http.StatusCreated {
		t.Fatalf("POST /auth/register status = %d, want 201", regResp.StatusCode)
	}

	// 2) faz login para obter token
	loginBody := []byte(`{"email":"alice@example.com","password":"password123"}`)
	loginResp, err := http.Post(ts.URL+"/auth/login", "application/json", bytes.NewReader(loginBody))
	if err != nil {
		t.Fatalf("POST /auth/login: %v", err)
	}
	defer loginResp.Body.Close()
	if loginResp.StatusCode != http.StatusOK {
		t.Fatalf("POST /auth/login status = %d, want 200", loginResp.StatusCode)
	}

	var loginResult struct {
		Token string `json:"token"`
	}
	if err := json.NewDecoder(loginResp.Body).Decode(&loginResult); err != nil {
		t.Fatalf("decode login response: %v", err)
	}
	if loginResult.Token == "" {
		t.Fatal("expected token in login response")
	}

	// 3) cria outro usuário usando o endpoint protegido com JWT
	createBody := []byte(`{"email":"bob@example.com","name":"Bob"}`)
	req, err := http.NewRequest("POST", ts.URL+"/users", bytes.NewReader(createBody))
	if err != nil {
		t.Fatalf("create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+loginResult.Token)

	createResp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("POST /users: %v", err)
	}
	defer createResp.Body.Close()
	if createResp.StatusCode != http.StatusOK {
		t.Fatalf("POST /users status = %d, want 200", createResp.StatusCode)
	}

	var created user.User
	if err := json.NewDecoder(createResp.Body).Decode(&created); err != nil {
		t.Fatalf("decode created: %v", err)
	}
	if created.ID == 0 {
		t.Fatalf("expected created user to have ID > 0")
	}

	// 4) busca por email
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
