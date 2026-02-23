package screens

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/williamokano/example-websocket-chat/pkg/protocol"
	"github.com/williamokano/example-websocket-chat/tui/client"
	"github.com/williamokano/example-websocket-chat/tui/components"
)

// LogoutMsg is sent when the user requests to log out from the chat screen.
type LogoutMsg struct{}

// ChatScreen is the main chat view.
type ChatScreen struct {
	header    components.Header
	messages  components.MessageList
	input     components.TextInput
	statusBar components.StatusBar
	wsClient  *client.WSClient
	width     int
	height    int
	menu      components.Menu
	showMenu  bool
}

// NewChatScreen creates a new chat screen.
func NewChatScreen(username, userID string, ws *client.WSClient) ChatScreen {
	header := components.NewHeader()
	header.SetUsername(username)

	ml := components.NewMessageList()
	ml.SetOwnUserID(userID)

	input := components.NewTextInput("Type a message...")
	input.Focus()

	sb := components.NewStatusBar()
	sb.SetUsername(username)
	sb.SetStatus(components.StatusConnecting)

	menu := components.NewMenu("Menu", []string{"Resume", "Logout"})

	return ChatScreen{
		header:    header,
		messages:  ml,
		input:     input,
		statusBar: sb,
		wsClient:  ws,
		menu:      menu,
	}
}

// SetSize sets the terminal size and distributes space.
func (c *ChatScreen) SetSize(w, h int) {
	c.width = w
	c.height = h
	c.header.SetWidth(w)
	c.statusBar.SetWidth(w)
	c.input.SetWidth(w)

	// header=1, statusbar=1, input border=3, remaining goes to messages
	msgHeight := h - 1 - 1 - 3
	if msgHeight < 3 {
		msgHeight = 3
	}
	c.messages.SetSize(w, msgHeight)
}

// SetConnected updates the connection status indicator.
func (c *ChatScreen) SetConnected(connected bool) {
	if connected {
		c.statusBar.SetStatus(components.StatusOnline)
	} else {
		c.statusBar.SetStatus(components.StatusOffline)
	}
}

// Init returns the initial command.
func (c ChatScreen) Init() tea.Cmd {
	return nil
}

// Update handles messages.
func (c *ChatScreen) Update(msg tea.Msg) tea.Cmd {
	// When menu is showing, delegate all input to the menu
	if c.showMenu {
		switch msg := msg.(type) {
		case tea.KeyMsg:
			return c.menu.Update(msg)
		case components.MenuSelectMsg:
			c.showMenu = false
			if msg.Index == 1 { // Logout
				return func() tea.Msg { return LogoutMsg{} }
			}
			// Index 0 = Resume, just close menu
			return nil
		case components.MenuCloseMsg:
			c.showMenu = false
			return nil
		}
		// Still handle non-key messages (like incoming chat) even when menu is open
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.Type == tea.KeyEsc {
			c.showMenu = true
			return nil
		}
		switch msg.Type {
		case tea.KeyEnter:
			text := c.input.Value()
			if text == "" {
				return nil
			}
			c.input.Reset()
			if c.wsClient != nil {
				ws := c.wsClient
				return func() tea.Msg {
					err := ws.Send(protocol.ClientMessage{
						Type:    protocol.TypeSendMessage,
						Content: text,
					})
					if err != nil {
						return client.ConnectionErrorMsg{Err: err}
					}
					return nil
				}
			}
			return nil
		case tea.KeyPgUp, tea.KeyPgDown, tea.KeyUp, tea.KeyDown:
			return c.messages.Update(msg)
		default:
			return c.input.Update(msg)
		}

	case client.IncomingChatMsg:
		c.messages.AddMessage(msg.Msg)
		return nil

	case client.ConnectedMsg:
		c.statusBar.SetStatus(components.StatusOnline)
		return nil

	case client.DisconnectedMsg:
		c.statusBar.SetStatus(components.StatusOffline)
		return nil

	case client.ConnectionErrorMsg:
		c.statusBar.SetStatus(components.StatusOffline)
		return nil
	}

	return nil
}

// View renders the chat screen.
func (c ChatScreen) View() string {
	if c.showMenu {
		return c.menu.View(c.width, c.height)
	}

	header := c.header.View()
	msgs := c.messages.View()
	input := c.input.View()
	status := c.statusBar.View()

	return header + "\n" + msgs + "\n" + input + "\n" + status
}
