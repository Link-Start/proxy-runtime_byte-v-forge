package mihomo

import (
	"fmt"
	"net/http"
)

func mihomoReloadConfigPathError(configPath string) error {
	return fmt.Errorf("mihomo reload config path must be absolute: %s", configPath)
}

func mihomoReloadResponseError(resp *http.Response) error {
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return nil
	}
	return fmt.Errorf("mihomo config reload returned HTTP %d", resp.StatusCode)
}
