package components

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/williamokano/example-websocket-chat/tui/styles"
)

// ConnectionStatus represents the WebSocket connection state.
type ConnectionStatus int

const (
	StatusOffline    ConnectionStatus = iota
	StatusConnecting ConnectionStatus = iota
	StatusOnline     ConnectionStatus = iota
)

// StatusBar renders a bottom status bar.
type StatusBar struct {
	width    int
	status   ConnectionStatus
	username string
	message  string
}

// NewStatusBar creates a new status bar.
func NewStatusBar() StatusBar {
	return StatusBar{
		status: StatusOffline,
	}
}

// SetWidth sets the bar width.
func (s *StatusBar) SetWidth(w int) {
	s.width = w
}

// SetStatus sets the connection status.
func (s *StatusBar) SetStatus(st ConnectionStatus) {
	s.status = st
}

// SetUsername sets the displayed username.
func (s *StatusBar) SetUsername(u string) {
	s.username = u
}

// SetMessage sets a transient status message displayed alongside the status.
func (s *StatusBar) SetMessage(msg string) {
	s.message = msg
}

// ClearMessage clears the transient status message.
func (s *StatusBar) ClearMessage() {
	s.message = ""
}

// View renders the status bar.
func (s StatusBar) View() string {
	var indicator string
	switch s.status {
	case StatusOnline:
		indicator = styles.StatusConnected.Render("● connected")
	case StatusConnecting:
		indicator = styles.StatusConnecting.Render("● connecting")
	case StatusOffline:
		indicator = styles.StatusDisconnected.Render("● disconnected")
	}

	left := indicator
	if s.username != "" {
		left = fmt.Sprintf("%s  %s", indicator, lipgloss.NewStyle().Foreground(styles.White).Bold(true).Render(s.username))
	}
	if s.message != "" {
		left = fmt.Sprintf("%s  %s", left, lipgloss.NewStyle().Foreground(styles.Accent).Italic(true).Render(s.message))
	}

	help := styles.StatusHelpStyle.Render("esc menu | pgup/pgdn scroll | ctrl+c quit")

	leftLen := lipgloss.Width(left)
	helpLen := lipgloss.Width(help)
	gap := s.width - leftLen - helpLen - 2 // 2 for padding
	if gap < 1 {
		gap = 1
	}

	bar := left + strings.Repeat(" ", gap) + help

	return styles.StatusBarStyle.Width(s.width).Render(bar)
}
