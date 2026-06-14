package mihomo

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"strings"
)

func newReloadConfigRequest(ctx context.Context, apiAddr string, secret string, configPath string) (*http.Request, error) {
	body, err := json.Marshal(map[string]string{"path": configPath})
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, controlURL(apiAddr, "/configs?force=true"), bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	setMihomoControllerAuthorization(req, secret)
	return req, nil
}

func setMihomoControllerAuthorization(req *http.Request, secret string) {
	secret = strings.TrimSpace(secret)
	if req == nil || secret == "" {
		return
	}
	req.Header.Set("Authorization", "Bearer "+secret)
}

func controlURL(apiAddr string, path string) string {
	apiAddr = strings.TrimRight(strings.TrimSpace(apiAddr), "/")
	if strings.HasPrefix(apiAddr, "http://") || strings.HasPrefix(apiAddr, "https://") {
		return apiAddr + path
	}
	return "http://" + apiAddr + path
}
