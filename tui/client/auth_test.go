package client

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHttpURL(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"ws to http", "ws://localhost:8080", "http://localhost:8080"},
		{"wss to https", "wss://example.com/chat", "https://example.com/chat"},
		{"strips trailing slash", "ws://localhost:8080/", "http://localhost:8080"},
		{"http passthrough", "http://localhost:8080", "http://localhost:8080"},
		{"already https", "https://example.com", "https://example.com"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := httpURL(tt.input)
			if got != tt.want {
				t.Errorf("httpURL(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestLogin_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/auth/login" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Method != http.MethodPost {
			t.Errorf("unexpected method: %s", r.Method)
		}

		var body map[string]string
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("failed to decode request body: %v", err)
		}
		if body["username"] != "alice" || body["password"] != "secret" {
			t.Errorf("unexpected credentials: %v", body)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(AuthResponse{
			Token: "jwt-token-123",
			User:  AuthUser{ID: "u1", Username: "alice"},
		})
	}))
	defer srv.Close()

	resp, err := Login(srv.URL, "alice", "secret")
	if err != nil {
		t.Fatalf("Login() error = %v", err)
	}
	if resp.Token != "jwt-token-123" {
		t.Errorf("Token = %q, want %q", resp.Token, "jwt-token-123")
	}
	if resp.User.Username != "alice" {
		t.Errorf("Username = %q, want %q", resp.User.Username, "alice")
	}
}

func TestLogin_Unauthorized(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid credentials"})
	}))
	defer srv.Close()

	_, err := Login(srv.URL, "alice", "wrong")
	if err == nil {
		t.Fatal("Login() expected error for 401, got nil")
	}
	if !strings.Contains(err.Error(), "invalid credentials") {
		t.Errorf("error = %q, want to contain %q", err.Error(), "invalid credentials")
	}
}

func TestRegister_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/auth/register" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(AuthResponse{
			Token: "new-token",
			User:  AuthUser{ID: "u2", Username: "bob"},
		})
	}))
	defer srv.Close()

	resp, err := Register(srv.URL, "bob", "password123")
	if err != nil {
		t.Fatalf("Register() error = %v", err)
	}
	if resp.Token != "new-token" {
		t.Errorf("Token = %q, want %q", resp.Token, "new-token")
	}
	if resp.User.ID != "u2" {
		t.Errorf("User.ID = %q, want %q", resp.User.ID, "u2")
	}
}

func TestGetMe_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/auth/me" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Method != http.MethodGet {
			t.Errorf("unexpected method: %s", r.Method)
		}

		auth := r.Header.Get("Authorization")
		if auth != "Bearer my-token" {
			t.Errorf("Authorization = %q, want %q", auth, "Bearer my-token")
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(AuthUser{ID: "u1", Username: "alice"})
	}))
	defer srv.Close()

	user, err := GetMe(srv.URL, "my-token")
	if err != nil {
		t.Fatalf("GetMe() error = %v", err)
	}
	if user.ID != "u1" {
		t.Errorf("ID = %q, want %q", user.ID, "u1")
	}
	if user.Username != "alice" {
		t.Errorf("Username = %q, want %q", user.Username, "alice")
	}
}

func TestGetMe_InvalidToken(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer srv.Close()

	_, err := GetMe(srv.URL, "bad-token")
	if err == nil {
		t.Fatal("GetMe() expected error for invalid token, got nil")
	}
	if !strings.Contains(err.Error(), "token invalid") {
		t.Errorf("error = %q, want to contain %q", err.Error(), "token invalid")
	}
}
