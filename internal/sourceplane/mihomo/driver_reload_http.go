package mihomo

import (
	"context"
	"errors"
	"path/filepath"
	"strings"
)

func (d *Driver) reloadLocked(ctx context.Context, configPath string) error {
	if strings.TrimSpace(d.cfg.APIAddr) == "" {
		return errors.New("mihomo api address is required for hot reload")
	}
	configPath = strings.TrimSpace(configPath)
	if !filepath.IsAbs(configPath) {
		return mihomoReloadConfigPathError(configPath)
	}
	req, err := newReloadConfigRequest(ctx, d.cfg.APIAddr, d.cfg.ControllerSecret, configPath)
	if err != nil {
		return err
	}
	resp, err := d.apiClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return mihomoReloadResponseError(resp)
}
