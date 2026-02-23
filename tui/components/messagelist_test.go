package components

import (
	"strings"
	"testing"
	"time"

	"github.com/williamokano/example-websocket-chat/pkg/protocol"
)

func TestNewMessageList(t *testing.T) {
	ml := NewMessageList()
	if len(ml.messages) != 0 {
		t.Errorf("new MessageList should have 0 messages, got %d", len(ml.messages))
	}
}

func TestMessageList_AddMessage(t *testing.T) {
	ml := NewMessageList()
	ml.SetSize(80, 20)

	ml.AddMessage(protocol.ServerMessage{
		Type:      protocol.TypeChatMessage,
		Content:   "hello",
		Sender:    &protocol.User{ID: "u1", Username: "alice"},
		Timestamp: time.Now(),
	})

	if len(ml.messages) != 1 {
		t.Errorf("message count = %d, want 1", len(ml.messages))
	}

	ml.AddMessage(protocol.ServerMessage{
		Type:      protocol.TypeChatMessage,
		Content:   "world",
		Sender:    &protocol.User{ID: "u2", Username: "bob"},
		Timestamp: time.Now(),
	})

	if len(ml.messages) != 2 {
		t.Errorf("message count = %d, want 2", len(ml.messages))
	}
}

func TestMessageList_ViewChatMessage(t *testing.T) {
	ml := NewMessageList()
	ml.SetSize(80, 20)
	ml.SetOwnUserID("u1")

	ml.AddMessage(protocol.ServerMessage{
		Type:      protocol.TypeChatMessage,
		Content:   "hello everyone",
		Sender:    &protocol.User{ID: "u1", Username: "alice"},
		Timestamp: time.Date(2025, 1, 15, 14, 30, 0, 0, time.Local),
	})

	view := ml.View()
	if !strings.Contains(view, "hello everyone") {
		t.Errorf("View() should contain message content, got:\n%s", view)
	}
	if !strings.Contains(view, "alice") {
		t.Errorf("View() should contain sender name, got:\n%s", view)
	}
}

func TestMessageList_ViewUserJoined(t *testing.T) {
	ml := NewMessageList()
	ml.SetSize(80, 20)

	ml.AddMessage(protocol.ServerMessage{
		Type: protocol.TypeUserJoined,
		User: &protocol.User{ID: "u2", Username: "bob"},
	})

	view := ml.View()
	if !strings.Contains(view, "bob") {
		t.Errorf("View() should contain joined user name, got:\n%s", view)
	}
	if !strings.Contains(view, "joined") {
		t.Errorf("View() should contain 'joined', got:\n%s", view)
	}
}

func TestMessageList_ViewUserLeft(t *testing.T) {
	ml := NewMessageList()
	ml.SetSize(80, 20)

	ml.AddMessage(protocol.ServerMessage{
		Type: protocol.TypeUserLeft,
		User: &protocol.User{ID: "u2", Username: "bob"},
	})

	view := ml.View()
	if !strings.Contains(view, "bob") {
		t.Errorf("View() should contain left user name, got:\n%s", view)
	}
	if !strings.Contains(view, "left") {
		t.Errorf("View() should contain 'left', got:\n%s", view)
	}
}

func TestMessageList_ViewError(t *testing.T) {
	ml := NewMessageList()
	ml.SetSize(80, 20)

	ml.AddMessage(protocol.ServerMessage{
		Type:    protocol.TypeError,
		Message: "something went wrong",
	})

	view := ml.View()
	if !strings.Contains(view, "something went wrong") {
		t.Errorf("View() should contain error message, got:\n%s", view)
	}
}

func TestMessageList_AutoScroll(t *testing.T) {
	ml := NewMessageList()
	ml.SetSize(80, 5)
	ml.SetOwnUserID("u1")

	// Add more messages than the viewport height
	for i := 0; i < 20; i++ {
		ml.AddMessage(protocol.ServerMessage{
			Type:      protocol.TypeChatMessage,
			Content:   "message",
			Sender:    &protocol.User{ID: "u1", Username: "alice"},
			Timestamp: time.Now(),
		})
	}

	// With auto-scroll, the last messages should be visible
	view := ml.View()
	if view == "" {
		t.Error("View() should not be empty after adding messages")
	}

	// scrollPos should be at the bottom
	maxScroll := len(ml.lines) - ml.height
	if maxScroll < 0 {
		maxScroll = 0
	}
	if ml.scrollPos != maxScroll {
		t.Errorf("scrollPos = %d, want %d (bottom)", ml.scrollPos, maxScroll)
	}
}

func TestMessageList_ViewEmptyHeight(t *testing.T) {
	ml := NewMessageList()
	// height is 0 by default
	view := ml.View()
	if view != "" {
		t.Errorf("View() with zero height should be empty, got %q", view)
	}
}
