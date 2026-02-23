package auth

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/williamokano/example-websocket-chat/internal/user"
)

const testJWTSecret = "test-secret-for-handler-tests"

func setupTestHandler() (*Handler, *JWTService, *Middleware) {
	repo := user.NewMockRepository()
	jwtSvc := NewJWTService(testJWTSecret)
	svc := NewService(repo, jwtSvc)
	handler := NewHandler(svc)
	mw := NewMiddleware(jwtSvc)
	return handler, jwtSvc, mw
}

func TestRegister_ValidData_Returns201(t *testing.T) {
	handler, _, _ := setupTestHandler()

	body, _ := json.Marshal(authRequest{Username: "alice", Password: "password123"})
	req := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	handler.Register(rr, req)

	if rr.Code != http.StatusCreated {
		t.Errorf("expected status %d, got %d", http.StatusCreated, rr.Code)
	}

	var resp AuthResponse
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resp.Token == "" {
		t.Error("expected non-empty token")
	}
	if resp.User.Username != "alice" {
		t.Errorf("expected username %q, got %q", "alice", resp.User.Username)
	}
}

func TestRegister_ShortUsername_Returns400(t *testing.T) {
	handler, _, _ := setupTestHandler()

	body, _ := json.Marshal(authRequest{Username: "ab", Password: "password123"})
	req := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	handler.Register(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, rr.Code)
	}
}

func TestRegister_DuplicateUsername_Returns400(t *testing.T) {
	handler, _, _ := setupTestHandler()

	body, _ := json.Marshal(authRequest{Username: "alice", Password: "password123"})

	// First registration
	req1 := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewReader(body))
	req1.Header.Set("Content-Type", "application/json")
	rr1 := httptest.NewRecorder()
	handler.Register(rr1, req1)

	if rr1.Code != http.StatusCreated {
		t.Fatalf("first registration failed: status %d", rr1.Code)
	}

	// Duplicate registration
	body2, _ := json.Marshal(authRequest{Username: "alice", Password: "differentpw"})
	req2 := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewReader(body2))
	req2.Header.Set("Content-Type", "application/json")
	rr2 := httptest.NewRecorder()
	handler.Register(rr2, req2)

	if rr2.Code != http.StatusBadRequest {
		t.Errorf("expected status %d for duplicate, got %d", http.StatusBadRequest, rr2.Code)
	}
}

func TestLogin_ValidCredentials_Returns200(t *testing.T) {
	handler, _, _ := setupTestHandler()

	// Register first
	regBody, _ := json.Marshal(authRequest{Username: "alice", Password: "password123"})
	regReq := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewReader(regBody))
	regReq.Header.Set("Content-Type", "application/json")
	regRR := httptest.NewRecorder()
	handler.Register(regRR, regReq)

	// Login
	loginBody, _ := json.Marshal(authRequest{Username: "alice", Password: "password123"})
	loginReq := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewReader(loginBody))
	loginReq.Header.Set("Content-Type", "application/json")
	loginRR := httptest.NewRecorder()
	handler.Login(loginRR, loginReq)

	if loginRR.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, loginRR.Code)
	}

	var resp AuthResponse
	if err := json.NewDecoder(loginRR.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resp.Token == "" {
		t.Error("expected non-empty token")
	}
}

func TestLogin_WrongPassword_Returns401(t *testing.T) {
	handler, _, _ := setupTestHandler()

	// Register first
	regBody, _ := json.Marshal(authRequest{Username: "alice", Password: "password123"})
	regReq := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewReader(regBody))
	regReq.Header.Set("Content-Type", "application/json")
	regRR := httptest.NewRecorder()
	handler.Register(regRR, regReq)

	// Login with wrong password
	loginBody, _ := json.Marshal(authRequest{Username: "alice", Password: "wrongpassword"})
	loginReq := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewReader(loginBody))
	loginReq.Header.Set("Content-Type", "application/json")
	loginRR := httptest.NewRecorder()
	handler.Login(loginRR, loginReq)

	if loginRR.Code != http.StatusUnauthorized {
		t.Errorf("expected status %d, got %d", http.StatusUnauthorized, loginRR.Code)
	}
}

func TestMe_WithValidToken_ReturnsUser(t *testing.T) {
	handler, _, mw := setupTestHandler()

	// Register to get a token
	regBody, _ := json.Marshal(authRequest{Username: "alice", Password: "password123"})
	regReq := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewReader(regBody))
	regReq.Header.Set("Content-Type", "application/json")
	regRR := httptest.NewRecorder()
	handler.Register(regRR, regReq)

	var regResp AuthResponse
	json.NewDecoder(regRR.Body).Decode(&regResp)

	// Call /me with the token, wrapping handler in middleware
	meReq := httptest.NewRequest(http.MethodGet, "/api/auth/me", nil)
	meReq.Header.Set("Authorization", "Bearer "+regResp.Token)
	meRR := httptest.NewRecorder()

	protected := mw.Authenticate(http.HandlerFunc(handler.Me))
	protected.ServeHTTP(meRR, meReq)

	if meRR.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, meRR.Code)
	}

	var meResp map[string]string
	if err := json.NewDecoder(meRR.Body).Decode(&meResp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if meResp["username"] != "alice" {
		t.Errorf("expected username %q, got %q", "alice", meResp["username"])
	}
}

func TestMe_WithoutToken_Returns401(t *testing.T) {
	handler, _, mw := setupTestHandler()

	req := httptest.NewRequest(http.MethodGet, "/api/auth/me", nil)
	rr := httptest.NewRecorder()

	protected := mw.Authenticate(http.HandlerFunc(handler.Me))
	protected.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected status %d, got %d", http.StatusUnauthorized, rr.Code)
	}
}

func TestRegister_InvalidJSON_Returns400(t *testing.T) {
	handler, _, _ := setupTestHandler()

	req := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewReader([]byte("not json")))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	handler.Register(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, rr.Code)
	}
}

func TestRegister_ShortPassword_Returns400(t *testing.T) {
	handler, _, _ := setupTestHandler()

	body, _ := json.Marshal(authRequest{Username: "alice", Password: "short"})
	req := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	handler.Register(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, rr.Code)
	}
}
