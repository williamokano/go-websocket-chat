package screens

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/williamokano/example-websocket-chat/tui/client"
	"github.com/williamokano/example-websocket-chat/tui/components"
)

// LoginSuccessMsg is emitted when login succeeds.
type LoginSuccessMsg struct {
	Token    string
	UserID   string
	Username string
}

// SwitchToRegisterMsg is emitted to switch to the register screen.
type SwitchToRegisterMsg struct{}

// loginResultMsg wraps the result of an async login attempt.
type loginResultMsg struct {
	resp *client.AuthResponse
	err  error
}

// LoginScreen handles user login.
type LoginScreen struct {
	dialog       components.Dialog
	serverURL    string
	width        int
	height       int
	loading      bool
	webauthnFlow *client.WebAuthnBrowserFlow
}

// NewLoginScreen creates a new login screen.
func NewLoginScreen(serverURL string) LoginScreen {
	d := components.NewDialog("Login", []string{"Username", "Password"})
	d.SetMask(1, '*')
	d.SetHint("Ctrl+R: Register | Ctrl+P: Passkey")
	return LoginScreen{
		dialog:       d,
		serverURL:    serverURL,
		webauthnFlow: client.NewWebAuthnBrowserFlow(serverURL),
	}
}

// SetSize updates terminal dimensions.
func (l *LoginScreen) SetSize(w, h int) {
	l.width = w
	l.height = h
	l.dialog.SetSize(w, h)
}

// Init returns the initial command.
func (l LoginScreen) Init() tea.Cmd {
	return nil
}

// Update handles messages.
func (l *LoginScreen) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.Type == tea.KeyCtrlR {
			return func() tea.Msg { return SwitchToRegisterMsg{} }
		}
		if msg.String() == "ctrl+p" {
			if l.loading {
				return nil
			}
			l.loading = true
			l.dialog.ClearError()
			return l.webauthnFlow.Login()
		}
	case client.WebAuthnLoginResultMsg:
		l.loading = false
		if msg.Err != nil {
			l.dialog.SetError(msg.Err.Error())
			return nil
		}
		token := msg.Token
		userID := msg.UserID
		username := msg.Username
		return func() tea.Msg {
			return LoginSuccessMsg{
				Token:    token,
				UserID:   userID,
				Username: username,
			}
		}
	case components.DialogSubmitMsg:
		if l.loading {
			return nil
		}
		vals := msg.Values
		if len(vals) < 2 || vals[0] == "" || vals[1] == "" {
			l.dialog.SetError("Username and password are required")
			return nil
		}
		l.loading = true
		l.dialog.ClearError()
		username := vals[0]
		password := vals[1]
		serverURL := l.serverURL
		return func() tea.Msg {
			resp, err := client.Login(serverURL, username, password)
			return loginResultMsg{resp: resp, err: err}
		}
	case loginResultMsg:
		l.loading = false
		if msg.err != nil {
			l.dialog.SetError(msg.err.Error())
			return nil
		}
		token := msg.resp.Token
		userID := msg.resp.User.ID
		username := msg.resp.User.Username
		return func() tea.Msg {
			return LoginSuccessMsg{
				Token:    token,
				UserID:   userID,
				Username: username,
			}
		}
	}

	cmd, handled := l.dialog.Update(msg)
	if handled {
		return cmd
	}
	return nil
}

// View renders the login screen.
func (l LoginScreen) View() string {
	return l.dialog.View(l.width, l.height)
}
