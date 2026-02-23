package chat

import (
	"strings"
	"testing"

	"github.com/williamokano/example-websocket-chat/pkg/protocol"
)

func TestNewChatMessage(t *testing.T) {
	msg := NewChatMessage("hello world", "user-1", "alice")

	if msg.Type != protocol.TypeChatMessage {
		t.Errorf("expected type %q, got %q", protocol.TypeChatMessage, msg.Type)
	}
	if msg.Content != "hello world" {
		t.Errorf("expected content %q, got %q", "hello world", msg.Content)
	}
	if msg.Sender == nil {
		t.Fatal("Sender should not be nil")
	}
	if msg.Sender.ID != "user-1" {
		t.Errorf("expected sender ID %q, got %q", "user-1", msg.Sender.ID)
	}
	if msg.Sender.Username != "alice" {
		t.Errorf("expected sender username %q, got %q", "alice", msg.Sender.Username)
	}
	if !strings.HasPrefix(msg.ID, "msg_") {
		t.Errorf("expected ID to start with %q, got %q", "msg_", msg.ID)
	}
	if msg.Timestamp.IsZero() {
		t.Error("Timestamp should not be zero")
	}
}

func TestNewUserJoinedMessage(t *testing.T) {
	msg := NewUserJoinedMessage("user-2", "bob", 3)

	if msg.Type != protocol.TypeUserJoined {
		t.Errorf("expected type %q, got %q", protocol.TypeUserJoined, msg.Type)
	}
	if msg.User == nil {
		t.Fatal("User should not be nil")
	}
	if msg.User.ID != "user-2" {
		t.Errorf("expected user ID %q, got %q", "user-2", msg.User.ID)
	}
	if msg.User.Username != "bob" {
		t.Errorf("expected username %q, got %q", "bob", msg.User.Username)
	}
	if msg.OnlineCount != 3 {
		t.Errorf("expected online count %d, got %d", 3, msg.OnlineCount)
	}
	if msg.Timestamp.IsZero() {
		t.Error("Timestamp should not be zero")
	}
}

func TestNewUserLeftMessage(t *testing.T) {
	msg := NewUserLeftMessage("user-3", "charlie", 2)

	if msg.Type != protocol.TypeUserLeft {
		t.Errorf("expected type %q, got %q", protocol.TypeUserLeft, msg.Type)
	}
	if msg.User == nil {
		t.Fatal("User should not be nil")
	}
	if msg.User.ID != "user-3" {
		t.Errorf("expected user ID %q, got %q", "user-3", msg.User.ID)
	}
	if msg.User.Username != "charlie" {
		t.Errorf("expected username %q, got %q", "charlie", msg.User.Username)
	}
	if msg.OnlineCount != 2 {
		t.Errorf("expected online count %d, got %d", 2, msg.OnlineCount)
	}
	if msg.Timestamp.IsZero() {
		t.Error("Timestamp should not be zero")
	}
}

func TestNewErrorMessage(t *testing.T) {
	msg := NewErrorMessage("something went wrong")

	if msg.Type != protocol.TypeError {
		t.Errorf("expected type %q, got %q", protocol.TypeError, msg.Type)
	}
	if msg.Message != "something went wrong" {
		t.Errorf("expected message %q, got %q", "something went wrong", msg.Message)
	}
}

func TestNewChatMessage_UniqueIDs(t *testing.T) {
	msg1 := NewChatMessage("first", "user-1", "alice")
	msg2 := NewChatMessage("second", "user-1", "alice")

	if msg1.ID == msg2.ID {
		t.Errorf("expected different IDs for different messages, both got %q", msg1.ID)
	}
}
