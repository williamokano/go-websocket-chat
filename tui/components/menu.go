package components

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/williamokano/example-websocket-chat/tui/styles"
)

// MenuSelectMsg is sent when the user selects a menu item.
type MenuSelectMsg struct {
	Index int
	Label string
}

// MenuCloseMsg is sent when the user closes the menu without selecting.
type MenuCloseMsg struct{}

// Menu is a centered overlay menu with selectable items.
type Menu struct {
	title  string
	items  []string
	cursor int
	width  int
}

// NewMenu creates a new menu with a title and items.
func NewMenu(title string, items []string) Menu {
	return Menu{
		title: title,
		items: items,
		width: 30,
	}
}

// Update handles key messages for the menu.
func (m *Menu) Update(msg tea.Msg) tea.Cmd {
	keyMsg, ok := msg.(tea.KeyMsg)
	if !ok {
		return nil
	}

	switch keyMsg.Type {
	case tea.KeyUp:
		m.cursor--
		if m.cursor < 0 {
			m.cursor = len(m.items) - 1
		}
		return nil
	case tea.KeyDown:
		m.cursor++
		if m.cursor >= len(m.items) {
			m.cursor = 0
		}
		return nil
	case tea.KeyEnter:
		idx := m.cursor
		label := m.items[idx]
		return func() tea.Msg {
			return MenuSelectMsg{Index: idx, Label: label}
		}
	case tea.KeyEsc:
		return func() tea.Msg { return MenuCloseMsg{} }
	}

	return nil
}

// View renders the menu centered in the terminal.
func (m Menu) View(termWidth, termHeight int) string {
	var b strings.Builder

	// Title
	title := styles.DialogTitleStyle.Width(m.width - 4).Render(m.title)
	b.WriteString(title)
	b.WriteString("\n\n")

	// Items
	for i, item := range m.items {
		if i == m.cursor {
			b.WriteString(styles.MenuSelectedItemStyle.Width(m.width - 4).Render("▸ " + item))
		} else {
			b.WriteString(styles.MenuItemStyle.Width(m.width - 4).Render("  " + item))
		}
		if i < len(m.items)-1 {
			b.WriteString("\n")
		}
	}

	content := b.String()
	dialog := styles.DialogStyle.Width(m.width).Render(content)

	return lipgloss.Place(termWidth, termHeight, lipgloss.Center, lipgloss.Center, dialog)
}
