package mihomo

import (
	"fmt"
	"io"
	"net/http"
	"strings"
)

func mihomoReloadConfigPathError(configPath string) error {
	return fmt.Errorf("mihomo reload config path must be absolute: %s", configPath)
}

func mihomoReloadResponseError(resp *http.Response) error {
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return nil
	}
	data, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
	return fmt.Errorf("mihomo config reload returned HTTP %d: %s", resp.StatusCode, strings.TrimSpace(string(data)))
}
