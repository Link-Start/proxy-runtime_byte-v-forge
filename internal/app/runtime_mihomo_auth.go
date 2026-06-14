package app

import (
	"net/http"
	"strings"
)

func applyMihomoControllerAuth(req *http.Request, token string) {
	token = strings.TrimSpace(token)
	if req == nil || token == "" {
		return
	}
	req.Header.Set("Authorization", "Bearer "+token)
}
