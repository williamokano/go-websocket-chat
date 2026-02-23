package styles

import (
	"testing"

	"github.com/charmbracelet/lipgloss"
)

func TestColorsDefined(t *testing.T) {
	colors := map[string]lipgloss.Color{
		"Primary":   Primary,
		"Secondary": Secondary,
		"Accent":    Accent,
		"Error":     Error,
		"Success":   Success,
		"Muted":     Muted,
		"White":     White,
		"Dark":      Dark,
		"DarkBg":    DarkBg,
		"InputBg":   InputBg,
	}

	for name, c := range colors {
		t.Run(name, func(t *testing.T) {
			if string(c) == "" {
				t.Errorf("color %s is empty", name)
			}
		})
	}
}

func TestStylesRenderNonEmpty(t *testing.T) {
	tests := []struct {
		name  string
		style lipgloss.Style
	}{
		{"HeaderStyle", HeaderStyle},
		{"HeaderTitleStyle", HeaderTitleStyle},
		{"HeaderInfoStyle", HeaderInfoStyle},
		{"OwnMessageStyle", OwnMessageStyle},
		{"OtherMessageStyle", OtherMessageStyle},
		{"SystemMessageStyle", SystemMessageStyle},
		{"TimestampStyle", TimestampStyle},
		{"MessageContentStyle", MessageContentStyle},
		{"StatusBarStyle", StatusBarStyle},
		{"StatusConnected", StatusConnected},
		{"StatusConnecting", StatusConnecting},
		{"StatusDisconnected", StatusDisconnected},
		{"StatusHelpStyle", StatusHelpStyle},
		{"InputStyle", InputStyle},
		{"InputFocusedStyle", InputFocusedStyle},
		{"InputBlurredStyle", InputBlurredStyle},
		{"InputPlaceholderStyle", InputPlaceholderStyle},
		{"InputLabelStyle", InputLabelStyle},
		{"DialogStyle", DialogStyle},
		{"DialogTitleStyle", DialogTitleStyle},
		{"DialogErrorStyle", DialogErrorStyle},
		{"DialogHintStyle", DialogHintStyle},
		{"PanelBorder", PanelBorder},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rendered := tt.style.Render("test")
			if rendered == "" {
				t.Errorf("style %s rendered empty string", tt.name)
			}
		})
	}
}
