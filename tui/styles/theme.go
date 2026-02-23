package styles

import "github.com/charmbracelet/lipgloss"

// Color palette
var (
	Primary   = lipgloss.Color("#7C3AED")
	Secondary = lipgloss.Color("#6366F1")
	Accent    = lipgloss.Color("#06B6D4")
	Error     = lipgloss.Color("#EF4444")
	Success   = lipgloss.Color("#22C55E")
	Muted     = lipgloss.Color("#6B7280")
	White     = lipgloss.Color("#F9FAFB")
	Dark      = lipgloss.Color("#1F2937")
	DarkBg    = lipgloss.Color("#111827")
	InputBg   = lipgloss.Color("#1E293B")
)

// Header styles
var (
	HeaderStyle = lipgloss.NewStyle().
			Background(Primary).
			Foreground(White).
			Bold(true).
			Padding(0, 1)

	HeaderTitleStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(White)

	HeaderInfoStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#D1D5DB"))
)

// Message styles
var (
	OwnMessageStyle = lipgloss.NewStyle().
			Foreground(Accent).
			Bold(true)

	OtherMessageStyle = lipgloss.NewStyle().
				Foreground(Secondary).
				Bold(true)

	SystemMessageStyle = lipgloss.NewStyle().
				Foreground(Muted).
				Italic(true)

	TimestampStyle = lipgloss.NewStyle().
			Foreground(Muted)

	MessageContentStyle = lipgloss.NewStyle().
				Foreground(White)
)

// StatusBar styles
var (
	StatusBarStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("#374151")).
			Foreground(White).
			Padding(0, 1)

	StatusConnected = lipgloss.NewStyle().
			Foreground(Success)

	StatusConnecting = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#EAB308"))

	StatusDisconnected = lipgloss.NewStyle().
				Foreground(Error)

	StatusHelpStyle = lipgloss.NewStyle().
			Foreground(Muted)
)

// Input styles
var (
	InputStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(Primary).
			Padding(0, 1)

	InputFocusedStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(Accent).
				Padding(0, 1)

	InputBlurredStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(Muted).
				Padding(0, 1)

	InputPlaceholderStyle = lipgloss.NewStyle().
				Foreground(Muted)

	InputLabelStyle = lipgloss.NewStyle().
			Foreground(Primary).
			Bold(true)
)

// Dialog styles
var (
	DialogStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(Primary).
			Padding(1, 2).
			Width(50)

	DialogTitleStyle = lipgloss.NewStyle().
			Foreground(Primary).
			Bold(true).
			Align(lipgloss.Center)

	DialogErrorStyle = lipgloss.NewStyle().
			Foreground(Error).
			Italic(true)

	DialogHintStyle = lipgloss.NewStyle().
			Foreground(Muted).
			Italic(true).
			Align(lipgloss.Center)
)

// Menu styles
var (
	MenuItemStyle = lipgloss.NewStyle().
			Foreground(White).
			Padding(0, 1)

	MenuSelectedItemStyle = lipgloss.NewStyle().
				Foreground(Accent).
				Bold(true).
				Padding(0, 1)
)

// Border style for panels
var (
	PanelBorder = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(Muted)
)
