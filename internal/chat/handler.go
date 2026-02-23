package chat

import (
	"log/slog"
	"net/http"

	"github.com/coder/websocket"
	"github.com/williamokano/example-websocket-chat/internal/auth"
)

type Handler struct {
	hub            *Hub
	jwtService     *auth.JWTService
	originPatterns []string
}

func NewHandler(hub *Hub, jwtService *auth.JWTService, originPatterns []string) *Handler {
	return &Handler{hub: hub, jwtService: jwtService, originPatterns: originPatterns}
}

func (h *Handler) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	if token == "" {
		http.Error(w, "missing token", http.StatusUnauthorized)
		return
	}

	claims, err := h.jwtService.ValidateToken(token)
	if err != nil {
		http.Error(w, "invalid token", http.StatusUnauthorized)
		return
	}

	conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		OriginPatterns: h.originPatterns,
	})
	if err != nil {
		slog.Error("websocket accept error", "error", err)
		return
	}

	client := NewClient(h.hub, conn, claims.UserID, claims.Username)
	h.hub.register <- client

	go client.WritePump(r.Context())
	client.ReadPump(r.Context())
}
