package auth

import "time"

const WebSocketTokenTTL = 2 * time.Minute

func NewWebSocketToken(secret string, now time.Time) (string, error) {
	return SignSession(secret, now.Add(WebSocketTokenTTL))
}
