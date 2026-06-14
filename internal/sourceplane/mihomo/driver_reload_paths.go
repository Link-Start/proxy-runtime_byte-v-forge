package mihomo

import "path/filepath"

func candidateConfigPath(canonicalPath string) string {
	return filepath.Join(filepath.Dir(canonicalPath), "config.candidate.json")
}
