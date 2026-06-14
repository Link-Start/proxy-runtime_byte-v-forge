package auth

import (
	"encoding/json"
	"net/http"
)

type SessionResponse struct {
	Authenticated bool `json:"authenticated"`
	AuthRequired  bool `json:"authRequired"`
}

type WebSocketTokenResponse struct {
	Token string `json:"token"`
}

func WriteSessionResponse(w http.ResponseWriter, authenticated bool, authRequired bool) {
	writeJSON(w, SessionResponse{Authenticated: authenticated, AuthRequired: authRequired})
}

func WriteWebSocketTokenResponse(w http.ResponseWriter, token string) {
	writeJSON(w, WebSocketTokenResponse{Token: token})
}

func writeJSON(w http.ResponseWriter, value any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(value)
}
