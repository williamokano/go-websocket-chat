package components

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/williamokano/example-websocket-chat/tui/styles"
)

// Header renders a top header bar.
type Header struct {
	width    int
	title    string
	username string
}

// NewHeader creates a new header.
func NewHeader() Header {
	return Header{
		title: "WebSocket Chat",
	}
}

// SetWidth sets the header width.
func (h *Header) SetWidth(w int) {
	h.width = w
}

// SetUsername sets the user info on the right.
func (h *Header) SetUsername(u string) {
	h.username = u
}

// View renders the header.
func (h Header) View() string {
	title := styles.HeaderTitleStyle.Render(h.title)
	user := styles.HeaderInfoStyle.Render(h.username)

	titleLen := lipgloss.Width(title)
	userLen := lipgloss.Width(user)
	gap := h.width - titleLen - userLen - 2
	if gap < 1 {
		gap = 1
	}

	content := title + strings.Repeat(" ", gap) + user

	return styles.HeaderStyle.Width(h.width).Render(content)
}
