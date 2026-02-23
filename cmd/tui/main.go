package main

import (
	"flag"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/williamokano/example-websocket-chat/tui"
)

func main() {
	serverURL := flag.String("server", "ws://localhost:8080", "WebSocket server URL")
	flag.Parse()

	app := tui.NewApp(*serverURL)
	p := tea.NewProgram(app, tea.WithAltScreen())
	app.SetProgram(p)

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
