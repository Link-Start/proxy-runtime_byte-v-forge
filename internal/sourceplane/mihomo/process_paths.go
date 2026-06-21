package mihomo

import (
	"os"
	"strings"
)

func (d *Driver) ensureConfigDir() (string, error) {
	if d.configDir != "" {
		return d.configDir, nil
	}
	dir := strings.TrimSpace(d.cfg.ConfigDir)
	if dir == "" {
		created, err := os.MkdirTemp("", "proxy-gateway-mihomo-")
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
