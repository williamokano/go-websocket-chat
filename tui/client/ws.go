package client

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"strings"
	"sync"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/coder/websocket"
	"github.com/williamokano/example-websocket-chat/pkg/protocol"
)

// Tea message types for WebSocket events.

// ConnectedMsg is sent when the WebSocket connection is established.
type ConnectedMsg struct{}

// DisconnectedMsg is sent when the WebSocket connection drops.
type DisconnectedMsg struct {
	Err error
}

// IncomingChatMsg wraps a server message received over WebSocket.
type IncomingChatMsg struct {
	Msg protocol.ServerMessage
}

// ConnectionErrorMsg is sent on connection errors.
type ConnectionErrorMsg struct {
	Err error
}

// ForceReloginMsg is sent when the server rejects the JWT (close code 4001).
type ForceReloginMsg struct{}

// WSClient manages a WebSocket connection to the chat server.
type WSClient struct {
	serverURL string
	token     string
	conn      *websocket.Conn
	program   *tea.Program
	cancel    context.CancelFunc
	mu        sync.Mutex
	closed    bool
}

// NewWSClient creates a new WebSocket client.
func NewWSClient(serverURL, token string, p *tea.Program) *WSClient {
	return &WSClient{
		serverURL: serverURL,
		token:     token,
		program:   p,
	}
}

// Connect establishes the WebSocket connection and starts the read loop.
func (c *WSClient) Connect() {
	go c.connectWithBackoff()
}

func (c *WSClient) connectWithBackoff() {
	attempt := 0
	for {
		c.mu.Lock()
		if c.closed {
			c.mu.Unlock()
			return
		}
		c.mu.Unlock()

		err := c.dial()
		if err == nil {
			attempt = 0
			c.program.Send(ConnectedMsg{})
			c.readLoop()
			// readLoop exited — either error or closure
			c.mu.Lock()
			if c.closed {
				c.mu.Unlock()
				return
			}
			c.mu.Unlock()
		}

		attempt++
		delay := time.Duration(math.Min(float64(time.Second)*math.Pow(2, float64(attempt)), float64(30*time.Second)))
		c.program.Send(DisconnectedMsg{Err: err})
		time.Sleep(delay)
	}
}

func (c *WSClient) dial() error {
	wsURL := c.serverURL
	if !strings.HasPrefix(wsURL, "ws://") && !strings.HasPrefix(wsURL, "wss://") {
		wsURL = "ws://" + wsURL
	}
	wsURL = strings.TrimSuffix(wsURL, "/")
	wsURL = fmt.Sprintf("%s/ws?token=%s", wsURL, c.token)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	conn, _, err := websocket.Dial(ctx, wsURL, nil)
	if err != nil {
		return err
	}

	c.mu.Lock()
	c.conn = conn
	readCtx, readCancel := context.WithCancel(context.Background())
	c.cancel = readCancel
	c.mu.Unlock()
	_ = readCtx
	return nil
}

func (c *WSClient) readLoop() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	c.mu.Lock()
	conn := c.conn
	c.mu.Unlock()

	if conn == nil {
		return
	}

	for {
		_, data, err := conn.Read(ctx)
		if err != nil {
			status := websocket.CloseStatus(err)
			if status == 4001 {
				c.program.Send(ForceReloginMsg{})
				c.Close()
				return
			}
			c.program.Send(DisconnectedMsg{Err: err})
			return
		}

		var msg protocol.ServerMessage
		if err := json.Unmarshal(data, &msg); err != nil {
			continue
		}

		c.program.Send(IncomingChatMsg{Msg: msg})
	}
}

// Send sends a message to the server.
func (c *WSClient) Send(msg protocol.ClientMessage) error {
	c.mu.Lock()
	conn := c.conn
	c.mu.Unlock()

	if conn == nil {
		return fmt.Errorf("not connected")
	}

	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return conn.Write(ctx, websocket.MessageText, data)
}

// Close shuts down the WebSocket connection.
func (c *WSClient) Close() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.closed = true
	if c.cancel != nil {
		c.cancel()
	}
	if c.conn != nil {
		c.conn.Close(websocket.StatusNormalClosure, "bye")
		c.conn = nil
	}
}

// SetToken updates the token for reconnection.
func (c *WSClient) SetToken(token string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.token = token
}
