package mihomo

import (
	"context"
	"errors"
	"os"
	"strings"

	"github.com/byte-v-forge/proxy-runtime/internal/processruntime"
)

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
