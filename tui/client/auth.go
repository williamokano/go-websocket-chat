package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// AuthUser represents a user returned from auth endpoints.
type AuthUser struct {
	ID       string `json:"id"`
	Username string `json:"username"`
}

// AuthResponse is the response from login/register endpoints.
type AuthResponse struct {
	Token string   `json:"token"`
	User  AuthUser `json:"user"`
}

// httpURL converts a WebSocket URL to an HTTP URL.
func httpURL(serverURL string) string {
	u := serverURL
	u = strings.TrimSuffix(u, "/")
	if strings.HasPrefix(u, "ws://") {
		u = "http://" + strings.TrimPrefix(u, "ws://")
	} else if strings.HasPrefix(u, "wss://") {
		u = "https://" + strings.TrimPrefix(u, "wss://")
	}
	return u
}

// Login authenticates with username and password.
func Login(serverURL, username, password string) (*AuthResponse, error) {
	base := httpURL(serverURL)
	body, err := json.Marshal(map[string]string{
		"username": username,
		"password": password,
	})
	if err != nil {
		return nil, err
	}

	resp, err := http.Post(base+"/api/auth/login", "application/json", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("connection failed: %w", err)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		var errResp struct {
			Error string `json:"error"`
		}
		if json.Unmarshal(data, &errResp) == nil && errResp.Error != "" {
			return nil, fmt.Errorf("%s", errResp.Error)
		}
		return nil, fmt.Errorf("login failed (status %d)", resp.StatusCode)
	}

	var authResp AuthResponse
	if err := json.Unmarshal(data, &authResp); err != nil {
		return nil, fmt.Errorf("invalid response: %w", err)
	}
	return &authResp, nil
}

// Register creates a new account.
func Register(serverURL, username, password string) (*AuthResponse, error) {
	base := httpURL(serverURL)
	body, err := json.Marshal(map[string]string{
		"username": username,
		"password": password,
	})
	if err != nil {
		return nil, err
	}

	resp, err := http.Post(base+"/api/auth/register", "application/json", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("connection failed: %w", err)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		var errResp struct {
			Error string `json:"error"`
		}
		if json.Unmarshal(data, &errResp) == nil && errResp.Error != "" {
			return nil, fmt.Errorf("%s", errResp.Error)
		}
		return nil, fmt.Errorf("registration failed (status %d)", resp.StatusCode)
	}

	var authResp AuthResponse
	if err := json.Unmarshal(data, &authResp); err != nil {
		return nil, fmt.Errorf("invalid response: %w", err)
	}
	return &authResp, nil
}

// GetMe validates a token and returns the current user.
func GetMe(serverURL, token string) (*AuthUser, error) {
	base := httpURL(serverURL)
	req, err := http.NewRequest("GET", base+"/api/auth/me", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("connection failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("token invalid (status %d)", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var user AuthUser
	if err := json.Unmarshal(data, &user); err != nil {
		return nil, fmt.Errorf("invalid response: %w", err)
	}
	return &user, nil
}
