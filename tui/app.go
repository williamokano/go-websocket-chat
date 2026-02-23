package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/williamokano/example-websocket-chat/tui/client"
	"github.com/williamokano/example-websocket-chat/tui/screens"
)

type screen int

const (
	screenLogin screen = iota
	screenRegister
	screenChat
)

// tokenCheckResult carries the result of the startup token validation.
type tokenCheckResult struct {
	token    string
	userID   string
	username string
	valid    bool
}

// App is the root Bubble Tea model.
type App struct {
	serverURL     string
	currentScreen screen
	width         int
	height        int

	loginScreen    screens.LoginScreen
	registerScreen screens.RegisterScreen
	chatScreen     screens.ChatScreen

	token    string
	userID   string
	username string

	wsClient *client.WSClient
	program  *tea.Program
}

// NewApp creates a new App with the given server URL.
func NewApp(serverURL string) *App {
	return &App{
		serverURL:      serverURL,
		currentScreen:  screenLogin,
		loginScreen:    screens.NewLoginScreen(serverURL),
		registerScreen: screens.NewRegisterScreen(serverURL),
	}
}

// SetProgram sets the tea.Program reference needed by the WebSocket client.
func (a *App) SetProgram(p *tea.Program) {
	a.program = p
}

// Init checks for a saved token on startup.
func (a *App) Init() tea.Cmd {
	return func() tea.Msg {
		token, err := client.LoadToken()
		if err != nil || token == "" {
			return tokenCheckResult{valid: false}
		}
		user, err := client.GetMe(a.serverURL, token)
		if err != nil {
			return tokenCheckResult{valid: false}
		}
		return tokenCheckResult{
			token:    token,
			userID:   user.ID,
			username: user.Username,
			valid:    true,
		}
	}
}

func (a *App) connectAndSwitchToChat() {
	if a.wsClient != nil {
		a.wsClient.Close()
	}
	if a.program != nil {
		a.wsClient = client.NewWSClient(a.serverURL, a.token, a.program)
		a.chatScreen = screens.NewChatScreen(a.username, a.userID, a.wsClient)
		a.wsClient.Connect()
	} else {
		a.chatScreen = screens.NewChatScreen(a.username, a.userID, nil)
	}
	a.chatScreen.SetSize(a.width, a.height)
	a.currentScreen = screenChat
}

// Update handles all messages.
func (a *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			if a.wsClient != nil {
				a.wsClient.Close()
			}
			return a, tea.Quit
		}

	case tea.WindowSizeMsg:
		a.width = msg.Width
		a.height = msg.Height
		a.loginScreen.SetSize(msg.Width, msg.Height)
		a.registerScreen.SetSize(msg.Width, msg.Height)
		a.chatScreen.SetSize(msg.Width, msg.Height)
		return a, nil

	case tokenCheckResult:
		if msg.valid {
			a.token = msg.token
			a.userID = msg.userID
			a.username = msg.username
			a.connectAndSwitchToChat()
		}
		return a, nil

	case screens.LoginSuccessMsg:
		a.token = msg.Token
		a.userID = msg.UserID
		a.username = msg.Username
		_ = client.SaveToken(msg.Token)
		a.connectAndSwitchToChat()
		return a, nil

	case screens.RegisterSuccessMsg:
		a.token = msg.Token
		a.userID = msg.UserID
		a.username = msg.Username
		_ = client.SaveToken(msg.Token)
		a.connectAndSwitchToChat()
		return a, nil

	case screens.SwitchToRegisterMsg:
		a.currentScreen = screenRegister
		return a, nil

	case screens.SwitchToLoginMsg:
		a.currentScreen = screenLogin
		return a, nil

	case client.ForceReloginMsg, screens.LogoutMsg:
		if a.wsClient != nil {
			a.wsClient.Close()
			a.wsClient = nil
		}
		_ = client.DeleteToken()
		a.token = ""
		a.userID = ""
		a.username = ""
		a.currentScreen = screenLogin
		a.loginScreen = screens.NewLoginScreen(a.serverURL)
		a.loginScreen.SetSize(a.width, a.height)
		return a, nil
	}

	// Delegate to current screen
	var cmd tea.Cmd
	switch a.currentScreen {
	case screenLogin:
		cmd = a.loginScreen.Update(msg)
	case screenRegister:
		cmd = a.registerScreen.Update(msg)
	case screenChat:
		cmd = a.chatScreen.Update(msg)
	}

	return a, cmd
}

// View renders the current screen.
func (a *App) View() string {
	switch a.currentScreen {
	case screenLogin:
		return a.loginScreen.View()
	case screenRegister:
		return a.registerScreen.View()
	case screenChat:
		return a.chatScreen.View()
	default:
		return "Loading..."
	}
}
