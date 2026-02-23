package screens

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/williamokano/example-websocket-chat/tui/client"
	"github.com/williamokano/example-websocket-chat/tui/components"
)

// RegisterSuccessMsg is emitted when registration succeeds.
type RegisterSuccessMsg struct {
	Token    string
	UserID   string
	Username string
}

// SwitchToLoginMsg is emitted to switch to the login screen.
type SwitchToLoginMsg struct{}

// registerResultMsg wraps the result of an async register attempt.
type registerResultMsg struct {
	resp *client.AuthResponse
	err  error
}

// RegisterScreen handles user registration.
type RegisterScreen struct {
	dialog       components.Dialog
	serverURL    string
	width        int
	height       int
	loading      bool
	webauthnFlow *client.WebAuthnBrowserFlow
}

// NewRegisterScreen creates a new register screen.
func NewRegisterScreen(serverURL string) RegisterScreen {
	d := components.NewDialog("Register", []string{"Username", "Password", "Confirm Password"})
	d.SetMask(1, '*')
	d.SetMask(2, '*')
	d.SetHint("Ctrl+R: Login | Ctrl+P: Passkey")
	return RegisterScreen{
		dialog:       d,
		serverURL:    serverURL,
		webauthnFlow: client.NewWebAuthnBrowserFlow(serverURL),
	}
}

// SetSize updates terminal dimensions.
func (r *RegisterScreen) SetSize(w, h int) {
	r.width = w
	r.height = h
	r.dialog.SetSize(w, h)
}

// Init returns the initial command.
func (r RegisterScreen) Init() tea.Cmd {
	return nil
}

// Update handles messages.
func (r *RegisterScreen) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.Type == tea.KeyCtrlR {
			return func() tea.Msg { return SwitchToLoginMsg{} }
		}
		if msg.String() == "ctrl+p" {
			if r.loading {
				return nil
			}
			vals := r.dialog.Values()
			if len(vals) < 1 || vals[0] == "" {
				r.dialog.SetError("Enter a username first, then press Ctrl+P")
				return nil
			}
			r.loading = true
			r.dialog.ClearError()
			return r.webauthnFlow.Register(vals[0])
		}
	case client.WebAuthnRegisterResultMsg:
		r.loading = false
		if msg.Err != nil {
			r.dialog.SetError(msg.Err.Error())
			return nil
		}
		token := msg.Token
		userID := msg.UserID
		username := msg.Username
		return func() tea.Msg {
			return RegisterSuccessMsg{
				Token:    token,
				UserID:   userID,
				Username: username,
			}
		}
	case components.DialogSubmitMsg:
		if r.loading {
			return nil
		}
		vals := msg.Values
		if len(vals) < 3 || vals[0] == "" || vals[1] == "" {
			r.dialog.SetError("All fields are required")
			return nil
		}
		if vals[1] != vals[2] {
			r.dialog.SetError("Passwords do not match")
			return nil
		}
		r.loading = true
		r.dialog.ClearError()
		username := vals[0]
		password := vals[1]
		serverURL := r.serverURL
		return func() tea.Msg {
			resp, err := client.Register(serverURL, username, password)
			return registerResultMsg{resp: resp, err: err}
		}
	case registerResultMsg:
		r.loading = false
		if msg.err != nil {
			r.dialog.SetError(msg.err.Error())
			return nil
		}
		token := msg.resp.Token
		userID := msg.resp.User.ID
		username := msg.resp.User.Username
		return func() tea.Msg {
			return RegisterSuccessMsg{
				Token:    token,
				UserID:   userID,
				Username: username,
			}
		}
	}

	cmd, handled := r.dialog.Update(msg)
	if handled {
		return cmd
	}
	return nil
}

// View renders the register screen.
func (r RegisterScreen) View() string {
	return r.dialog.View(r.width, r.height)
}
