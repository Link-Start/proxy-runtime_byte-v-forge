package mihomo

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/byte-v-forge/proxy-runtime/internal/dataplane"
	"github.com/byte-v-forge/proxy-runtime/internal/processruntime"
)

func (d *Driver) Stop() {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.stopLocked()
}

func (d *Driver) Status() dataplane.Status {
	d.mu.Lock()
	defer d.mu.Unlock()
	return dataplane.Status{
		Running:           d.running,
		ConfigPath:        d.configPath,
		DesiredConfigHash: d.desiredSig,
		AppliedConfigHash: d.signature,
		LastError:         d.lastError,
	}
}

func (d *Driver) ensureConfigDir() (string, error) {
	if d.configDir != "" {
		return d.configDir, nil
	}
	dir := strings.TrimSpace(d.cfg.ConfigDir)
	if dir == "" {
		created, err := os.MkdirTemp("", "proxy-runtime-mihomo-")
		if err != nil {
			return "", err
		}
		dir = created
	}
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return "", err
	}
	d.configDir = dir
	return dir, nil
}

func (d *Driver) startLocked(ctx context.Context, dir string, configPath string) error {
	path := strings.TrimSpace(d.cfg.Path)
	if path == "" {
		return errors.New("mihomo path is required")
	}
	if d.processLogs == nil {
		d.processLogs = newProcessLogRing(d.logger, defaultProcessLogRingLimit)
	}
	process, err := processruntime.Start(ctx, processruntime.Config{Path: path, Args: []string{"-f", configPath}, Dir: dir, Env: append(os.Environ(), "SAFE_PATHS="+d.safePaths(dir)), Stdout: d.processLogs.Writer("stdout"), Stderr: d.processLogs.Writer("stderr")})
	if err != nil {
		return err
	}
	d.process = process
	d.running = true
	go d.wait(process)
	return nil
}

func (d *Driver) safePaths(configDir string) string {
	paths := []string{configDir, d.cfg.DashboardDir}
	seen := map[string]struct{}{}
	out := make([]string, 0, len(paths))
	for _, value := range paths {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		out = append(out, value)
	}
	return strings.Join(out, string(os.PathListSeparator))
}

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

func (d *Driver) wait(process *processruntime.Process) {
	<-process.Done()
	err := process.ExitError()
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.process != process {
		return
	}
	d.running = false
	if err != nil {
		d.lastError = processExitMessage(err, d.processLogs)
	}
}

func processExitMessage(err error, logs *processLogRing) string {
	message := strings.TrimSpace(err.Error())
	if logs == nil {
		return message
	}
	if tail := logs.Tail(); tail != "" {
		return message + ": " + tail
	}
	return message
}

func (d *Driver) stopLocked() {
	process := d.process
	d.process = nil
	d.running = false
	if process != nil {
		_ = process.Stop(5 * time.Second)
	}
}

func controlURL(apiAddr string, path string) string {
	apiAddr = strings.TrimRight(strings.TrimSpace(apiAddr), "/")
	if strings.HasPrefix(apiAddr, "http://") || strings.HasPrefix(apiAddr, "https://") {
		return apiAddr + path
	}
	return "http://" + apiAddr + path
}
