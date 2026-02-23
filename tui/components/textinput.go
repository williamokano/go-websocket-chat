package components

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/williamokano/example-websocket-chat/tui/styles"
)

// TextInput is a styled text input component.
type TextInput struct {
	value       []rune
	placeholder string
	focused     bool
	cursor      int
	width       int
	mask        rune // if non-zero, characters are masked (for passwords)
	offset      int  // horizontal scroll offset
}

// NewTextInput creates a new text input with the given placeholder.
func NewTextInput(placeholder string) TextInput {
	return TextInput{
		placeholder: placeholder,
		width:       40,
	}
}

// WithMask sets a mask character for the input (e.g. '*' for passwords).
func (t TextInput) WithMask(mask rune) TextInput {
	t.mask = mask
	return t
}

// Focus gives focus to the input.
func (t *TextInput) Focus() {
	t.focused = true
}

// Blur removes focus from the input.
func (t *TextInput) Blur() {
	t.focused = false
}

// Focused returns whether the input is focused.
func (t TextInput) Focused() bool {
	return t.focused
}

// Value returns the current input value.
func (t TextInput) Value() string {
	return string(t.value)
}

// SetValue sets the input value.
func (t *TextInput) SetValue(s string) {
	t.value = []rune(s)
	t.cursor = len(t.value)
	t.updateOffset()
}

// Reset clears the input.
func (t *TextInput) Reset() {
	t.value = nil
	t.cursor = 0
	t.offset = 0
}

// SetWidth sets the display width.
func (t *TextInput) SetWidth(w int) {
	if w > 6 {
		t.width = w - 4 // account for border + padding
	}
}

func (t *TextInput) updateOffset() {
	visibleWidth := t.width
	if visibleWidth < 1 {
		visibleWidth = 1
	}
	if t.cursor < t.offset {
		t.offset = t.cursor
	}
	if t.cursor >= t.offset+visibleWidth {
		t.offset = t.cursor - visibleWidth + 1
	}
}

// Update handles key messages.
func (t *TextInput) Update(msg tea.Msg) tea.Cmd {
	if !t.focused {
		return nil
	}

	keyMsg, ok := msg.(tea.KeyMsg)
	if !ok {
		return nil
	}

	switch keyMsg.Type {
	case tea.KeyBackspace:
		if t.cursor > 0 {
			t.value = append(t.value[:t.cursor-1], t.value[t.cursor:]...)
			t.cursor--
			t.updateOffset()
		}
	case tea.KeyDelete:
		if t.cursor < len(t.value) {
			t.value = append(t.value[:t.cursor], t.value[t.cursor+1:]...)
		}
	case tea.KeyLeft:
		if t.cursor > 0 {
			t.cursor--
			t.updateOffset()
		}
	case tea.KeyRight:
		if t.cursor < len(t.value) {
			t.cursor++
			t.updateOffset()
		}
	case tea.KeyHome, tea.KeyCtrlA:
		t.cursor = 0
		t.updateOffset()
	case tea.KeyEnd, tea.KeyCtrlE:
		t.cursor = len(t.value)
		t.updateOffset()
	case tea.KeyCtrlU:
		t.value = t.value[t.cursor:]
		t.cursor = 0
		t.updateOffset()
	case tea.KeyCtrlK:
		t.value = t.value[:t.cursor]
	case tea.KeyRunes:
		for _, r := range keyMsg.Runes {
			t.value = append(t.value[:t.cursor], append([]rune{r}, t.value[t.cursor:]...)...)
			t.cursor++
		}
		t.updateOffset()
	}

	return nil
}

// View renders the text input.
func (t TextInput) View() string {
	var display string
	visibleWidth := t.width
	if visibleWidth < 1 {
		visibleWidth = 40
	}

	if len(t.value) == 0 && !t.focused {
		display = styles.InputPlaceholderStyle.Render(t.placeholder)
	} else if len(t.value) == 0 && t.focused {
		display = styles.InputPlaceholderStyle.Render(t.placeholder)
		// Show cursor at start
		display = lipgloss.NewStyle().Reverse(true).Render(" ") + display
	} else {
		var displayRunes []rune
		if t.mask != 0 {
			displayRunes = []rune(strings.Repeat(string(t.mask), len(t.value)))
		} else {
			displayRunes = t.value
		}

		end := t.offset + visibleWidth
		if end > len(displayRunes) {
			end = len(displayRunes)
		}
		start := t.offset
		if start > len(displayRunes) {
			start = len(displayRunes)
		}

		visible := displayRunes[start:end]

		if t.focused {
			cursorPos := t.cursor - t.offset
			if cursorPos < 0 {
				cursorPos = 0
			}
			if cursorPos > len(visible) {
				cursorPos = len(visible)
			}
			before := string(visible[:cursorPos])
			cursor := " "
			if cursorPos < len(visible) {
				cursor = string(visible[cursorPos])
			}
			after := ""
			if cursorPos+1 < len(visible) {
				after = string(visible[cursorPos+1:])
			}
			display = before + lipgloss.NewStyle().Reverse(true).Render(cursor) + after
		} else {
			display = string(visible)
		}
	}

	style := styles.InputBlurredStyle
	if t.focused {
		style = styles.InputFocusedStyle
	}

	return style.Width(visibleWidth).Render(display)
}
