package protocol

import "time"

// Message types
const (
	TypeSendMessage = "send_message"
	TypeChatMessage = "chat_message"
	TypeUserJoined  = "user_joined"
	TypeUserLeft    = "user_left"
	TypeError       = "error"
)

// User represents a minimal user for protocol messages.
type User struct {
	ID       string `json:"id"`
	Username string `json:"username"`
}

// ClientMessage is sent from client to server.
type ClientMessage struct {
	Type    string `json:"type"`
	Content string `json:"content,omitempty"`
}

// ServerMessage is sent from server to client.
type ServerMessage struct {
	Type        string    `json:"type"`
	ID          string    `json:"id,omitempty"`
	Content     string    `json:"content,omitempty"`
	Sender      *User     `json:"sender,omitempty"`
	User        *User     `json:"user,omitempty"`
	OnlineCount int       `json:"online_count,omitempty"`
	Message     string    `json:"message,omitempty"`
	Timestamp   time.Time `json:"timestamp,omitempty"`
}
