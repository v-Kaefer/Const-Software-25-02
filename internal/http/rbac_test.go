package http_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/v-Kaefer/Const-Software-25-02/pkg/user"
)

// TestRBAC_AdminOnlyRoutes tests routes that require admin role
func TestRBAC_AdminOnlyRoutes(t *testing.T) {
	ts := newTestServer(t)
	defer ts.Close()

	// Test GET /users (list all users) - Admin only
	t.Run("GET /users requires admin", func(t *testing.T) {
		// Create a user first
		body := []byte(`{"email":"test1@example.com","name":"Test User 1"}`)
		resp, err := http.Post(ts.URL+"/users", "application/json", bytes.NewReader(body))
		if err != nil {
			t.Fatalf("POST /users: %v", err)
		}
		defer resp.Body.Close()
		
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("POST status = %d, want 200", resp.StatusCode)
		}

		// GET /users should work (mock middleware gives admin role)
		getResp, err := http.Get(ts.URL + "/users")
		if err != nil {
			t.Fatalf("GET /users: %v", err)
		}
		defer getResp.Body.Close()
		
		if getResp.StatusCode != http.StatusOK {
			t.Fatalf("GET /users status = %d, want 200", getResp.StatusCode)
		}

		var users []user.User
		if err := json.NewDecoder(getResp.Body).Decode(&users); err != nil {
			t.Fatalf("decode users: %v", err)
		}
		
		if len(users) == 0 {
			t.Fatal("expected at least one user")
		}
	})

	// Test POST /users - Admin only
	t.Run("POST /users requires admin", func(t *testing.T) {
		body := []byte(`{"email":"admin-created@example.com","name":"Admin Created"}`)
		resp, err := http.Post(ts.URL+"/users", "application/json", bytes.NewReader(body))
		if err != nil {
			t.Fatalf("POST /users: %v", err)
		}
		defer resp.Body.Close()
		
		// Should succeed with mock auth (admin role)
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("POST status = %d, want 200", resp.StatusCode)
		}
	})

	// Test DELETE /users/{id} - Admin only
	t.Run("DELETE /users/{id} requires admin", func(t *testing.T) {
		// Create a user to delete
		body := []byte(`{"email":"to-delete@example.com","name":"To Delete"}`)
		resp, err := http.Post(ts.URL+"/users", "application/json", bytes.NewReader(body))
		if err != nil {
			t.Fatalf("POST /users: %v", err)
		}
		defer resp.Body.Close()

		var created user.User
		if err := json.NewDecoder(resp.Body).Decode(&created); err != nil {
			t.Fatalf("decode created: %v", err)
		}

		// Delete the user
		req, err := http.NewRequest("DELETE", fmt.Sprintf("%s/users/%d", ts.URL, created.ID), nil)
		if err != nil {
			t.Fatalf("create DELETE request: %v", err)
		}
		
		delResp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("DELETE /users/%d: %v", created.ID, err)
		}
		defer delResp.Body.Close()

		// Should succeed with admin role (mock auth)
		if delResp.StatusCode != http.StatusNoContent {
			t.Fatalf("DELETE status = %d, want 204", delResp.StatusCode)
		}
	})
}

// TestRBAC_UserOrAdminRoutes tests routes that allow admin or the user themselves
func TestRBAC_UserOrAdminRoutes(t *testing.T) {
	ts := newTestServer(t)
	defer ts.Close()

	// Create a test user
	body := []byte(`{"email":"ownuser@example.com","name":"Own User"}`)
	resp, err := http.Post(ts.URL+"/users", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("POST /users: %v", err)
	}
	defer resp.Body.Close()

	var created user.User
	if err := json.NewDecoder(resp.Body).Decode(&created); err != nil {
		t.Fatalf("decode created: %v", err)
	}

	// Test GET /users/{id} - should work with admin (mock gives admin role)
	t.Run("GET /users/{id} with admin", func(t *testing.T) {
		getResp, err := http.Get(fmt.Sprintf("%s/users/%d", ts.URL, created.ID))
		if err != nil {
			t.Fatalf("GET /users/%d: %v", created.ID, err)
		}
		defer getResp.Body.Close()

		if getResp.StatusCode != http.StatusOK {
			t.Fatalf("GET status = %d, want 200", getResp.StatusCode)
		}

		var got user.User
		if err := json.NewDecoder(getResp.Body).Decode(&got); err != nil {
			t.Fatalf("decode user: %v", err)
		}

		if got.ID != created.ID {
			t.Errorf("got user ID %d, want %d", got.ID, created.ID)
		}
	})

	// Test PUT /users/{id} - should work with admin
	t.Run("PUT /users/{id} with admin", func(t *testing.T) {
		updateBody := []byte(`{"email":"updated@example.com","name":"Updated Name"}`)
		req, err := http.NewRequest("PUT", fmt.Sprintf("%s/users/%d", ts.URL, created.ID), bytes.NewReader(updateBody))
		if err != nil {
			t.Fatalf("create PUT request: %v", err)
		}
		req.Header.Set("Content-Type", "application/json")

		putResp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("PUT /users/%d: %v", created.ID, err)
		}
		defer putResp.Body.Close()

		if putResp.StatusCode != http.StatusOK {
			t.Fatalf("PUT status = %d, want 200", putResp.StatusCode)
		}

		var updated user.User
		if err := json.NewDecoder(putResp.Body).Decode(&updated); err != nil {
			t.Fatalf("decode updated: %v", err)
		}

		if updated.Email != "updated@example.com" || updated.Name != "Updated Name" {
			t.Errorf("user not updated correctly: %+v", updated)
		}
	})

	// Test PATCH /users/{id} - should work with admin
	t.Run("PATCH /users/{id} with admin", func(t *testing.T) {
		patchBody := []byte(`{"name":"Patched Name"}`)
		req, err := http.NewRequest("PATCH", fmt.Sprintf("%s/users/%d", ts.URL, created.ID), bytes.NewReader(patchBody))
		if err != nil {
			t.Fatalf("create PATCH request: %v", err)
		}
		req.Header.Set("Content-Type", "application/json")

		patchResp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("PATCH /users/%d: %v", created.ID, err)
		}
		defer patchResp.Body.Close()

		if patchResp.StatusCode != http.StatusOK {
			t.Fatalf("PATCH status = %d, want 200", patchResp.StatusCode)
		}

		var patched user.User
		if err := json.NewDecoder(patchResp.Body).Decode(&patched); err != nil {
			t.Fatalf("decode patched: %v", err)
		}

		if patched.Name != "Patched Name" {
			t.Errorf("user name not patched correctly: got %s, want 'Patched Name'", patched.Name)
		}
		
		// Email should remain unchanged
		if patched.Email != "updated@example.com" {
			t.Errorf("email should not change: got %s, want 'updated@example.com'", patched.Email)
		}
	})
}

// TestRBAC_InvalidUserID tests handling of invalid user IDs
func TestRBAC_InvalidUserID(t *testing.T) {
	ts := newTestServer(t)
	defer ts.Close()

	tests := []struct {
		name       string
		method     string
		path       string
		body       string
		wantStatus int
	}{
		{
			name:       "GET with invalid ID",
			method:     "GET",
			path:       "/users/invalid",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "PUT with invalid ID",
			method:     "PUT",
			path:       "/users/abc",
			body:       `{"email":"test@example.com","name":"Test"}`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "PATCH with invalid ID",
			method:     "PATCH",
			path:       "/users/xyz",
			body:       `{"name":"Test"}`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "DELETE with invalid ID",
			method:     "DELETE",
			path:       "/users/notanumber",
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var req *http.Request
			var err error

			if tt.body != "" {
				req, err = http.NewRequest(tt.method, ts.URL+tt.path, bytes.NewReader([]byte(tt.body)))
			} else {
				req, err = http.NewRequest(tt.method, ts.URL+tt.path, nil)
			}
			if err != nil {
				t.Fatalf("create request: %v", err)
			}

			if tt.body != "" {
				req.Header.Set("Content-Type", "application/json")
			}

			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				t.Fatalf("%s %s: %v", tt.method, tt.path, err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != tt.wantStatus {
				t.Errorf("status = %d, want %d", resp.StatusCode, tt.wantStatus)
			}
		})
	}
}

// TestRBAC_NotFoundUser tests handling of non-existent users
func TestRBAC_NotFoundUser(t *testing.T) {
	ts := newTestServer(t)
	defer ts.Close()

	nonExistentID := 9999

	tests := []struct {
		name       string
		method     string
		body       string
		wantStatus int
	}{
		{
			name:       "GET non-existent user",
			method:     "GET",
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "PUT non-existent user",
			method:     "PUT",
			body:       `{"email":"test@example.com","name":"Test"}`,
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "PATCH non-existent user",
			method:     "PATCH",
			body:       `{"name":"Test"}`,
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "DELETE non-existent user",
			method:     "DELETE",
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var req *http.Request
			var err error
			path := fmt.Sprintf("%s/users/%d", ts.URL, nonExistentID)

			if tt.body != "" {
				req, err = http.NewRequest(tt.method, path, bytes.NewReader([]byte(tt.body)))
			} else {
				req, err = http.NewRequest(tt.method, path, nil)
			}
			if err != nil {
				t.Fatalf("create request: %v", err)
			}

			if tt.body != "" {
				req.Header.Set("Content-Type", "application/json")
			}

			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				t.Fatalf("%s: %v", tt.method, err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != tt.wantStatus {
				t.Errorf("status = %d, want %d", resp.StatusCode, tt.wantStatus)
			}
		})
	}
}

// TestRBAC_CRUDWorkflow tests a complete CRUD workflow with RBAC
func TestRBAC_CRUDWorkflow(t *testing.T) {
	ts := newTestServer(t)
	defer ts.Close()

	// 1. Create user (admin only)
	createBody := []byte(`{"email":"workflow@example.com","name":"Workflow User"}`)
	createResp, err := http.Post(ts.URL+"/users", "application/json", bytes.NewReader(createBody))
	if err != nil {
		t.Fatalf("POST /users: %v", err)
	}
	defer createResp.Body.Close()

	var created user.User
	if err := json.NewDecoder(createResp.Body).Decode(&created); err != nil {
		t.Fatalf("decode created: %v", err)
	}
	t.Logf("Created user ID: %d", created.ID)

	// 2. Read user (admin or own user)
	getResp, err := http.Get(fmt.Sprintf("%s/users/%d", ts.URL, created.ID))
	if err != nil {
		t.Fatalf("GET /users/%d: %v", created.ID, err)
	}
	defer getResp.Body.Close()

	if getResp.StatusCode != http.StatusOK {
		t.Fatalf("GET status = %d, want 200", getResp.StatusCode)
	}

	// 3. Update user (admin or own user)
	updateBody := []byte(`{"email":"workflow-updated@example.com","name":"Updated Workflow User"}`)
	updateReq, err := http.NewRequest("PUT", fmt.Sprintf("%s/users/%d", ts.URL, created.ID), bytes.NewReader(updateBody))
	if err != nil {
		t.Fatalf("create PUT request: %v", err)
	}
	updateReq.Header.Set("Content-Type", "application/json")

	updateResp, err := http.DefaultClient.Do(updateReq)
	if err != nil {
		t.Fatalf("PUT /users/%d: %v", created.ID, err)
	}
	defer updateResp.Body.Close()

	if updateResp.StatusCode != http.StatusOK {
		t.Fatalf("PUT status = %d, want 200", updateResp.StatusCode)
	}

	// 4. Partial update user (admin or own user)
	patchBody := []byte(`{"name":"Final Name"}`)
	patchReq, err := http.NewRequest("PATCH", fmt.Sprintf("%s/users/%d", ts.URL, created.ID), bytes.NewReader(patchBody))
	if err != nil {
		t.Fatalf("create PATCH request: %v", err)
	}
	patchReq.Header.Set("Content-Type", "application/json")

	patchResp, err := http.DefaultClient.Do(patchReq)
	if err != nil {
		t.Fatalf("PATCH /users/%d: %v", created.ID, err)
	}
	defer patchResp.Body.Close()

	if patchResp.StatusCode != http.StatusOK {
		t.Fatalf("PATCH status = %d, want 200", patchResp.StatusCode)
	}

	// 5. Delete user (admin only)
	deleteReq, err := http.NewRequest("DELETE", fmt.Sprintf("%s/users/%d", ts.URL, created.ID), nil)
	if err != nil {
		t.Fatalf("create DELETE request: %v", err)
	}

	deleteResp, err := http.DefaultClient.Do(deleteReq)
	if err != nil {
		t.Fatalf("DELETE /users/%d: %v", created.ID, err)
	}
	defer deleteResp.Body.Close()

	if deleteResp.StatusCode != http.StatusNoContent {
		t.Fatalf("DELETE status = %d, want 204", deleteResp.StatusCode)
	}

	// 6. Verify user is deleted
	verifyResp, err := http.Get(fmt.Sprintf("%s/users/%d", ts.URL, created.ID))
	if err != nil {
		t.Fatalf("GET /users/%d after delete: %v", created.ID, err)
	}
	defer verifyResp.Body.Close()

	if verifyResp.StatusCode != http.StatusNotFound {
		t.Fatalf("GET after DELETE status = %d, want 404", verifyResp.StatusCode)
	}
}
