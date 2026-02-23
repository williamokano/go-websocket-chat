package chat

import (
	"time"

	"github.com/google/uuid"
	"github.com/williamokano/example-websocket-chat/pkg/protocol"
)

func NewChatMessage(content, userID, username string) protocol.ServerMessage {
	return protocol.ServerMessage{
		Type:    protocol.TypeChatMessage,
		ID:      "msg_" + uuid.New().String()[:8],
		Content: content,
		Sender: &protocol.User{
			ID:       userID,
			Username: username,
		},
		Timestamp: time.Now().UTC(),
	}
}

func NewUserJoinedMessage(userID, username string, onlineCount int) protocol.ServerMessage {
	return protocol.ServerMessage{
		Type: protocol.TypeUserJoined,
		User: &protocol.User{
			ID:       userID,
			Username: username,
		},
		OnlineCount: onlineCount,
		Timestamp:   time.Now().UTC(),
	}
}

func NewUserLeftMessage(userID, username string, onlineCount int) protocol.ServerMessage {
	return protocol.ServerMessage{
		Type: protocol.TypeUserLeft,
		User: &protocol.User{
			ID:       userID,
			Username: username,
		},
		OnlineCount: onlineCount,
		Timestamp:   time.Now().UTC(),
	}
}

func NewErrorMessage(message string) protocol.ServerMessage {
	return protocol.ServerMessage{
		Type:    protocol.TypeError,
		Message: message,
	}
}
