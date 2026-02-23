package client

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"net"
	"net/http"
	"os/exec"
	"runtime"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

//go:embed webauthn_login.html
var loginHTML string

//go:embed webauthn_register.html
var registerHTML string

//go:embed webauthn_add_credential.html
var addCredentialHTML string

// WebAuthnLoginResultMsg carries the result of a WebAuthn login flow.
type WebAuthnLoginResultMsg struct {
	Token    string
	UserID   string
	Username string
	Err      error
}

// WebAuthnRegisterResultMsg carries the result of a WebAuthn registration flow.
type WebAuthnRegisterResultMsg struct {
	Token    string
	UserID   string
	Username string
	Err      error
}

// WebAuthnAddCredentialResultMsg carries the result of adding a credential.
type WebAuthnAddCredentialResultMsg struct {
	Err error
}

// callbackResult is the JSON payload sent by the browser to the local callback.
type callbackResult struct {
	Token string    `json:"token"`
	User  *AuthUser `json:"user"`
	Error string    `json:"error"`
}

// templateData holds the values injected into the HTML templates.
type templateData struct {
	ServerURL    string
	CallbackURL  string
	Username     string
	Token        string
	FriendlyName string
}

// WebAuthnBrowserFlow manages browser-based WebAuthn ceremonies.
type WebAuthnBrowserFlow struct {
	serverURL string
}

// NewWebAuthnBrowserFlow creates a new flow targeting the given backend URL.
func NewWebAuthnBrowserFlow(serverURL string) *WebAuthnBrowserFlow {
	return &WebAuthnBrowserFlow{serverURL: httpURL(serverURL)}
}

// Login starts a WebAuthn login ceremony in the browser.
func (f *WebAuthnBrowserFlow) Login() tea.Cmd {
	return func() tea.Msg {
		result, err := f.runBrowserFlow(loginHTML, templateData{
			ServerURL: f.serverURL,
		})
		if err != nil {
			return WebAuthnLoginResultMsg{Err: err}
		}
		if result.Error != "" {
			return WebAuthnLoginResultMsg{Err: fmt.Errorf("%s", result.Error)}
		}
		if result.User == nil {
			return WebAuthnLoginResultMsg{Err: fmt.Errorf("no user returned from login")}
		}
		return WebAuthnLoginResultMsg{
			Token:    result.Token,
			UserID:   result.User.ID,
			Username: result.User.Username,
		}
	}
}

// Register starts a WebAuthn registration ceremony in the browser.
func (f *WebAuthnBrowserFlow) Register(username string) tea.Cmd {
	return func() tea.Msg {
		result, err := f.runBrowserFlow(registerHTML, templateData{
			ServerURL: f.serverURL,
			Username:  username,
		})
		if err != nil {
			return WebAuthnRegisterResultMsg{Err: err}
		}
		if result.Error != "" {
			return WebAuthnRegisterResultMsg{Err: fmt.Errorf("%s", result.Error)}
		}
		if result.User == nil {
			return WebAuthnRegisterResultMsg{Err: fmt.Errorf("no user returned from registration")}
		}
		return WebAuthnRegisterResultMsg{
			Token:    result.Token,
			UserID:   result.User.ID,
			Username: result.User.Username,
		}
	}
}

// AddCredential starts a WebAuthn credential registration ceremony for an
// already-authenticated user.
func (f *WebAuthnBrowserFlow) AddCredential(token, friendlyName string) tea.Cmd {
	return func() tea.Msg {
		result, err := f.runBrowserFlow(addCredentialHTML, templateData{
			ServerURL:    f.serverURL,
			Token:        token,
			FriendlyName: friendlyName,
		})
		if err != nil {
			return WebAuthnAddCredentialResultMsg{Err: err}
		}
		if result.Error != "" {
			return WebAuthnAddCredentialResultMsg{Err: fmt.Errorf("%s", result.Error)}
		}
		return WebAuthnAddCredentialResultMsg{}
	}
}

// runBrowserFlow starts a local HTTP server, opens the browser, and waits for
// the callback. It returns the result posted by the browser page.
func (f *WebAuthnBrowserFlow) runBrowserFlow(htmlTemplate string, data templateData) (*callbackResult, error) {
	// Parse the template
	tmpl, err := template.New("webauthn").Parse(htmlTemplate)
	if err != nil {
		return nil, fmt.Errorf("failed to parse template: %w", err)
	}

	// Listen on a random available port
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return nil, fmt.Errorf("failed to start local server: %w", err)
	}
	port := listener.Addr().(*net.TCPAddr).Port
	callbackURL := fmt.Sprintf("http://127.0.0.1:%d", port)
	data.CallbackURL = callbackURL

	resultCh := make(chan *callbackResult, 1)

	mux := http.NewServeMux()

	// Serve the HTML page
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		if err := tmpl.Execute(w, data); err != nil {
			http.Error(w, "template error", http.StatusInternalServerError)
		}
	})

	// Handle the callback from the browser
	mux.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}
		var result callbackResult
		if err := json.Unmarshal(body, &result); err != nil {
			http.Error(w, "invalid JSON", http.StatusBadRequest)
			return
		}
		w.Header().Set("Access-Control-Allow-Origin", callbackURL)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"ok":true}`))

		// Send result without blocking (buffered channel)
		select {
		case resultCh <- &result:
		default:
		}
	})

	server := &http.Server{Handler: mux}

	// Start the server in a goroutine
	serverErr := make(chan error, 1)
	go func() {
		if err := server.Serve(listener); err != nil && err != http.ErrServerClosed {
			serverErr <- err
		}
	}()

	// Open the browser
	url := callbackURL + "/"
	if err := openBrowser(url); err != nil {
		_ = server.Close()
		return nil, fmt.Errorf("failed to open browser: %w", err)
	}

	// Wait for result with a 2-minute timeout
	timeout := time.NewTimer(2 * time.Minute)
	defer timeout.Stop()

	var result *callbackResult
	select {
	case result = <-resultCh:
		// Got a result from the browser
	case err := <-serverErr:
		return nil, fmt.Errorf("local server error: %w", err)
	case <-timeout.C:
		_ = server.Shutdown(context.Background())
		return nil, fmt.Errorf("timed out waiting for browser response")
	}

	// Gracefully shut down the server
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = server.Shutdown(ctx)

	return result, nil
}

// openBrowser opens a URL in the user's default browser.
func openBrowser(url string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "linux":
		cmd = exec.Command("xdg-open", url)
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", url)
	default:
		return fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}
	return cmd.Start()
}
