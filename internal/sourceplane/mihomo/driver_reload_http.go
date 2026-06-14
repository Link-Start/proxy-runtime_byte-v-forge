package mihomo

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strings"
)

func (d *Driver) reloadLocked(ctx context.Context, configPath string) error {
	if strings.TrimSpace(d.cfg.APIAddr) == "" {
		return errors.New("mihomo api address is required for hot reload")
	}
	configPath = strings.TrimSpace(configPath)
	if !filepath.IsAbs(configPath) {
		return fmt.Errorf("mihomo reload config path must be absolute: %s", configPath)
	}
	body, err := json.Marshal(map[string]string{"path": configPath})
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, controlURL(d.cfg.APIAddr, "/configs?force=true"), bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	setMihomoControllerAuthorization(req, d.cfg.ControllerSecret)
	resp, err := d.apiClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return nil
	}
	data, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
	return fmt.Errorf("mihomo config reload returned HTTP %d: %s", resp.StatusCode, strings.TrimSpace(string(data)))
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
