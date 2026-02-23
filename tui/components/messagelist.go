package components

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/williamokano/example-websocket-chat/pkg/protocol"
	"github.com/williamokano/example-websocket-chat/tui/styles"
)

// ChatMessage wraps a server message for display.
type ChatMessage struct {
	ServerMsg  protocol.ServerMessage
	OwnUserID string
}

// MessageList is a scrollable message viewport.
type MessageList struct {
	messages   []ChatMessage
	width      int
	height     int
	scrollPos  int // index of the first visible line
	lines      []string
	ownUserID  string
	autoScroll bool
}

// NewMessageList creates a new message list.
func NewMessageList() MessageList {
	return MessageList{
		autoScroll: true,
	}
}

// SetSize sets the viewport dimensions.
func (m *MessageList) SetSize(w, h int) {
	m.width = w
	m.height = h
	m.rebuildLines()
}

// SetOwnUserID sets the current user's ID for styling own messages.
func (m *MessageList) SetOwnUserID(id string) {
	m.ownUserID = id
}

// AddMessage appends a message and auto-scrolls.
func (m *MessageList) AddMessage(msg protocol.ServerMessage) {
	m.messages = append(m.messages, ChatMessage{
		ServerMsg: msg,
		OwnUserID: m.ownUserID,
	})
	m.rebuildLines()
	if m.autoScroll {
		m.scrollToBottom()
	}
}

func (m *MessageList) rebuildLines() {
	m.lines = nil
	contentWidth := m.width - 2 // border padding
	if contentWidth < 10 {
		contentWidth = 40
	}

	for _, cm := range m.messages {
		msg := cm.ServerMsg
		switch msg.Type {
		case protocol.TypeChatMessage:
			ts := msg.Timestamp.Local().Format("15:04")
			tsStr := styles.TimestampStyle.Render(ts)

			sender := "unknown"
			isOwn := false
			if msg.Sender != nil {
				sender = msg.Sender.Username
				isOwn = msg.Sender.ID == m.ownUserID
			}

			var nameStr string
			if isOwn {
				nameStr = styles.OwnMessageStyle.Render(sender)
			} else {
				nameStr = styles.OtherMessageStyle.Render(sender)
			}

			content := msg.Content
			line := fmt.Sprintf("%s %s: %s", tsStr, nameStr, styles.MessageContentStyle.Render(content))
			m.lines = append(m.lines, line)

		case protocol.TypeUserJoined:
			username := "someone"
			if msg.User != nil {
				username = msg.User.Username
			}
			line := styles.SystemMessageStyle.Render(fmt.Sprintf("  --> %s joined the chat", username))
			m.lines = append(m.lines, line)

		case protocol.TypeUserLeft:
			username := "someone"
			if msg.User != nil {
				username = msg.User.Username
			}
			line := styles.SystemMessageStyle.Render(fmt.Sprintf("  <-- %s left the chat", username))
			m.lines = append(m.lines, line)

		case protocol.TypeError:
			errMsg := msg.Message
			if errMsg == "" {
				errMsg = msg.Content
			}
			line := lipgloss.NewStyle().Foreground(styles.Error).Render(fmt.Sprintf("  [error] %s", errMsg))
			m.lines = append(m.lines, line)
		}
	}
}

func (m *MessageList) scrollToBottom() {
	maxScroll := len(m.lines) - m.height
	if maxScroll < 0 {
		maxScroll = 0
	}
	m.scrollPos = maxScroll
}

// Update handles scroll key messages.
func (m *MessageList) Update(msg tea.Msg) tea.Cmd {
	keyMsg, ok := msg.(tea.KeyMsg)
	if !ok {
		return nil
	}

	maxScroll := len(m.lines) - m.height
	if maxScroll < 0 {
		maxScroll = 0
	}

	switch keyMsg.Type {
	case tea.KeyUp:
		if m.scrollPos > 0 {
			m.scrollPos--
			m.autoScroll = false
		}
	case tea.KeyDown:
		if m.scrollPos < maxScroll {
			m.scrollPos++
		}
		if m.scrollPos >= maxScroll {
			m.autoScroll = true
		}
	case tea.KeyPgUp:
		m.scrollPos -= m.height / 2
		if m.scrollPos < 0 {
			m.scrollPos = 0
		}
		m.autoScroll = false
	case tea.KeyPgDown:
		m.scrollPos += m.height / 2
		if m.scrollPos > maxScroll {
			m.scrollPos = maxScroll
		}
		if m.scrollPos >= maxScroll {
			m.autoScroll = true
		}
	}

	return nil
}

// View renders the message list.
func (m MessageList) View() string {
	if m.height <= 0 {
		return ""
	}

	visibleLines := make([]string, m.height)
	for i := range visibleLines {
		lineIdx := m.scrollPos + i
		if lineIdx < len(m.lines) {
			visibleLines[i] = m.lines[lineIdx]
		} else {
			visibleLines[i] = ""
		}
	}

	content := strings.Join(visibleLines, "\n")

	return styles.PanelBorder.
		Width(m.width - 2).
		Height(m.height).
		Render(content)
}
