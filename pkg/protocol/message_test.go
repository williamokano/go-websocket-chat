package protocol

import (
	"encoding/json"
	"testing"
	"time"
)

func TestMessageTypeConstants(t *testing.T) {
	tests := []struct {
		name     string
		got      string
		expected string
	}{
		{"TypeSendMessage", TypeSendMessage, "send_message"},
		{"TypeChatMessage", TypeChatMessage, "chat_message"},
		{"TypeUserJoined", TypeUserJoined, "user_joined"},
		{"TypeUserLeft", TypeUserLeft, "user_left"},
		{"TypeError", TypeError, "error"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, tt.got)
			}
		})
	}
}

func TestServerMessage_JSONMarshalUnmarshal(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	msg := ServerMessage{
		Type:    TypeChatMessage,
		ID:      "msg_abc123",
		Content: "Hello, world!",
		Sender: &User{
			ID:       "user-1",
			Username: "alice",
		},
		Timestamp: now,
	}

	data, err := json.Marshal(msg)
	if err != nil {
		t.Fatalf("json.Marshal failed: %v", err)
	}

	var decoded ServerMessage
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("json.Unmarshal failed: %v", err)
	}

	if decoded.Type != msg.Type {
		t.Errorf("Type: expected %q, got %q", msg.Type, decoded.Type)
	}
	if decoded.ID != msg.ID {
		t.Errorf("ID: expected %q, got %q", msg.ID, decoded.ID)
	}
	if decoded.Content != msg.Content {
		t.Errorf("Content: expected %q, got %q", msg.Content, decoded.Content)
	}
	if decoded.Sender == nil {
		t.Fatal("Sender is nil after unmarshal")
	}
	if decoded.Sender.ID != msg.Sender.ID {
		t.Errorf("Sender.ID: expected %q, got %q", msg.Sender.ID, decoded.Sender.ID)
	}
	if decoded.Sender.Username != msg.Sender.Username {
		t.Errorf("Sender.Username: expected %q, got %q", msg.Sender.Username, decoded.Sender.Username)
	}
}

func TestClientMessage_JSONMarshalUnmarshal(t *testing.T) {
	msg := ClientMessage{
		Type:    TypeSendMessage,
		Content: "Hello from client",
	}

	data, err := json.Marshal(msg)
	if err != nil {
		t.Fatalf("json.Marshal failed: %v", err)
	}

	var decoded ClientMessage
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("json.Unmarshal failed: %v", err)
	}

	if decoded.Type != msg.Type {
		t.Errorf("Type: expected %q, got %q", msg.Type, decoded.Type)
	}
	if decoded.Content != msg.Content {
		t.Errorf("Content: expected %q, got %q", msg.Content, decoded.Content)
	}
}

func TestServerMessage_RoundTrip(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	original := ServerMessage{
		Type: TypeUserJoined,
		User: &User{
			ID:       "user-42",
			Username: "charlie",
		},
		OnlineCount: 5,
		Timestamp:   now,
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	var roundTripped ServerMessage
	if err := json.Unmarshal(data, &roundTripped); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	if roundTripped.Type != original.Type {
		t.Errorf("Type mismatch: %q vs %q", original.Type, roundTripped.Type)
	}
	if roundTripped.OnlineCount != original.OnlineCount {
		t.Errorf("OnlineCount mismatch: %d vs %d", original.OnlineCount, roundTripped.OnlineCount)
	}
	if roundTripped.User == nil {
		t.Fatal("User is nil after round-trip")
	}
	if roundTripped.User.Username != original.User.Username {
		t.Errorf("User.Username mismatch: %q vs %q", original.User.Username, roundTripped.User.Username)
	}
}

func TestClientMessage_RoundTrip(t *testing.T) {
	original := ClientMessage{
		Type:    TypeSendMessage,
		Content: "test message content",
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	var roundTripped ClientMessage
	if err := json.Unmarshal(data, &roundTripped); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	if roundTripped != original {
		t.Errorf("round-trip mismatch: got %+v, want %+v", roundTripped, original)
	}
}

func TestServerMessage_OmitsEmptyFields(t *testing.T) {
	msg := ServerMessage{
		Type:    TypeError,
		Message: "something went wrong",
	}

	data, err := json.Marshal(msg)
	if err != nil {
		t.Fatalf("json.Marshal failed: %v", err)
	}

	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("json.Unmarshal failed: %v", err)
	}

	// Fields with omitempty should not appear when zero-valued
	if _, exists := raw["id"]; exists {
		t.Error("empty ID should be omitted from JSON")
	}
	if _, exists := raw["content"]; exists {
		t.Error("empty Content should be omitted from JSON")
	}
	if _, exists := raw["sender"]; exists {
		t.Error("nil Sender should be omitted from JSON")
	}
}
