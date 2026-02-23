package components

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/williamokano/example-websocket-chat/tui/styles"
)

// DialogField represents a single input field in a dialog.
type DialogField struct {
	Label string
	Input TextInput
}

// DialogSubmitMsg is sent when the user submits the dialog.
type DialogSubmitMsg struct {
	Values []string
}

// DialogCancelMsg is sent when the user cancels the dialog.
type DialogCancelMsg struct{}

// Dialog is a centered modal dialog with multiple input fields.
type Dialog struct {
	title      string
	fields     []DialogField
	focusIndex int
	errMsg     string
	hint       string
	width      int
	height     int
}

// NewDialog creates a new dialog with a title and field labels.
func NewDialog(title string, labels []string) Dialog {
	fields := make([]DialogField, len(labels))
	for i, label := range labels {
		fields[i] = DialogField{
			Label: label,
			Input: NewTextInput(label),
		}
	}
	if len(fields) > 0 {
		fields[0].Input.Focus()
	}
	return Dialog{
		title:  title,
		fields: fields,
		width:  46,
	}
}

// SetMask sets a mask character on a specific field (e.g. for password fields).
func (d *Dialog) SetMask(index int, mask rune) {
	if index >= 0 && index < len(d.fields) {
		d.fields[index].Input = d.fields[index].Input.WithMask(mask)
	}
}

// SetHint sets the hint text displayed at the bottom.
func (d *Dialog) SetHint(hint string) {
	d.hint = hint
}

// SetError sets the error message.
func (d *Dialog) SetError(msg string) {
	d.errMsg = msg
}

// ClearError clears the error message.
func (d *Dialog) ClearError() {
	d.errMsg = ""
}

// SetSize sets the terminal size for centering.
func (d *Dialog) SetSize(w, h int) {
	d.width = 46
	d.height = h
}

// Values returns all field values.
func (d Dialog) Values() []string {
	vals := make([]string, len(d.fields))
	for i, f := range d.fields {
		vals[i] = f.Input.Value()
	}
	return vals
}

// Reset clears all fields and errors.
func (d *Dialog) Reset() {
	for i := range d.fields {
		d.fields[i].Input.Reset()
	}
	d.errMsg = ""
	d.focusIndex = 0
	if len(d.fields) > 0 {
		d.fields[0].Input.Focus()
	}
}

func (d *Dialog) nextField() {
	d.fields[d.focusIndex].Input.Blur()
	d.focusIndex = (d.focusIndex + 1) % len(d.fields)
	d.fields[d.focusIndex].Input.Focus()
}

func (d *Dialog) prevField() {
	d.fields[d.focusIndex].Input.Blur()
	d.focusIndex--
	if d.focusIndex < 0 {
		d.focusIndex = len(d.fields) - 1
	}
	d.fields[d.focusIndex].Input.Focus()
}

// Update handles key messages for the dialog.
func (d *Dialog) Update(msg tea.Msg) (tea.Cmd, bool) {
	keyMsg, ok := msg.(tea.KeyMsg)
	if !ok {
		return nil, false
	}

	switch keyMsg.Type {
	case tea.KeyEsc:
		return func() tea.Msg { return DialogCancelMsg{} }, true
	case tea.KeyEnter:
		vals := d.Values()
		return func() tea.Msg { return DialogSubmitMsg{Values: vals} }, true
	case tea.KeyTab:
		d.nextField()
		return nil, true
	case tea.KeyShiftTab:
		d.prevField()
		return nil, true
	default:
		cmd := d.fields[d.focusIndex].Input.Update(msg)
		return cmd, true
	}
}

// View renders the dialog centered in the terminal.
func (d Dialog) View(termWidth, termHeight int) string {
	var b strings.Builder

	// Title
	title := styles.DialogTitleStyle.Width(d.width - 4).Render(d.title)
	b.WriteString(title)
	b.WriteString("\n\n")

	// Fields
	for i, f := range d.fields {
		label := styles.InputLabelStyle.Render(f.Label)
		b.WriteString(label)
		b.WriteString("\n")
		inp := f.Input
		inp.SetWidth(d.width - 4)
		_ = i
		b.WriteString(inp.View())
		b.WriteString("\n")
	}

	// Error
	if d.errMsg != "" {
		b.WriteString("\n")
		b.WriteString(styles.DialogErrorStyle.Width(d.width - 4).Render(d.errMsg))
	}

	// Hint
	if d.hint != "" {
		b.WriteString("\n")
		b.WriteString(styles.DialogHintStyle.Width(d.width - 4).Render(d.hint))
	}

	content := b.String()
	dialog := styles.DialogStyle.Render(content)

	// Center in terminal
	return lipgloss.Place(termWidth, termHeight, lipgloss.Center, lipgloss.Center, dialog)
}
