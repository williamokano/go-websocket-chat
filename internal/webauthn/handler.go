package webauthn

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-webauthn/webauthn/protocol"

	"github.com/williamokano/example-websocket-chat/internal/auth"
)

// Handler provides HTTP endpoints for WebAuthn registration, login, and credential management.
type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// --- Unauthenticated: passwordless registration ---

type beginRegisterRequest struct {
	Username string `json:"username"`
}

type beginResponse struct {
	SessionID string      `json:"session_id"`
	Options   interface{} `json:"options"`
}

func (h *Handler) BeginRegistration(w http.ResponseWriter, r *http.Request) {
	var req beginRegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	options, sessionID, err := h.service.BeginRegistration(r.Context(), req.Username)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, beginResponse{SessionID: sessionID, Options: options})
}

type finishRegisterRequest struct {
	SessionID string `json:"session_id"`
}

func (h *Handler) FinishRegistration(w http.ResponseWriter, r *http.Request) {
	// Parse the session_id from query params since the body is the attestation response
	sessionID := r.URL.Query().Get("session_id")
	if sessionID == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "missing session_id"})
		return
	}

	parsed, err := protocol.ParseCredentialCreationResponseBody(r.Body)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid credential response: " + err.Error()})
		return
	}

	resp, err := h.service.FinishRegistration(r.Context(), sessionID, parsed)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, http.StatusCreated, resp)
}

// --- Unauthenticated: passkey login ---

func (h *Handler) BeginLogin(w http.ResponseWriter, r *http.Request) {
	options, sessionID, err := h.service.BeginLogin(r.Context())
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, beginResponse{SessionID: sessionID, Options: options})
}

func (h *Handler) FinishLogin(w http.ResponseWriter, r *http.Request) {
	sessionID := r.URL.Query().Get("session_id")
	if sessionID == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "missing session_id"})
		return
	}

	parsed, err := protocol.ParseCredentialRequestResponseBody(r.Body)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid assertion response: " + err.Error()})
		return
	}

	resp, err := h.service.FinishLogin(r.Context(), sessionID, parsed)
	if err != nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, resp)
}

// --- Authenticated: credential management ---

func (h *Handler) BeginAddCredential(w http.ResponseWriter, r *http.Request) {
	claims := auth.GetClaimsFromContext(r.Context())
	if claims == nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	options, sessionID, err := h.service.BeginAddCredential(r.Context(), claims.UserID)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, beginResponse{SessionID: sessionID, Options: options})
}

type finishAddCredentialRequest struct {
	FriendlyName string `json:"friendly_name"`
}

func (h *Handler) FinishAddCredential(w http.ResponseWriter, r *http.Request) {
	claims := auth.GetClaimsFromContext(r.Context())
	if claims == nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	sessionID := r.URL.Query().Get("session_id")
	if sessionID == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "missing session_id"})
		return
	}

	friendlyName := r.URL.Query().Get("friendly_name")

	parsed, err := protocol.ParseCredentialCreationResponseBody(r.Body)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid credential response: " + err.Error()})
		return
	}

	info, err := h.service.FinishAddCredential(r.Context(), claims.UserID, sessionID, friendlyName, parsed)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, http.StatusCreated, info)
}

func (h *Handler) ListCredentials(w http.ResponseWriter, r *http.Request) {
	claims := auth.GetClaimsFromContext(r.Context())
	if claims == nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	creds, err := h.service.ListCredentials(r.Context(), claims.UserID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, creds)
}

func (h *Handler) DeleteCredential(w http.ResponseWriter, r *http.Request) {
	claims := auth.GetClaimsFromContext(r.Context())
	if claims == nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	credentialID := chi.URLParam(r, "id")
	if credentialID == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "missing credential ID"})
		return
	}

	if err := h.service.DeleteCredential(r.Context(), claims.UserID, credentialID); err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": err.Error()})
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}
