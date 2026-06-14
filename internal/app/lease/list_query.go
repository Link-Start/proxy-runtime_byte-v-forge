package lease

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"
)

func ParseListOptions(query url.Values) (ListOptions, error) {
	options := DefaultListOptions()
	status := strings.TrimSpace(query.Get("status"))
	if status == "" && query.Get("include_inactive") == "true" {
		status = string(ListModeRecent)
	}
	if status != "" {
		mode, err := parseListMode(status)
		if err != nil {
			return ListOptions{}, err
		}
		options.Mode = mode
	}
	if rawLimit := strings.TrimSpace(query.Get("limit")); rawLimit != "" {
		limit, err := strconv.Atoi(rawLimit)
		if err != nil || limit <= 0 {
			return ListOptions{}, fmt.Errorf("lease list limit must be a positive integer")
		}
		options.Limit = limit
	}
	return NormalizeListOptions(options), nil
}

func parseListMode(status string) (ListMode, error) {
	switch mode := ListMode(strings.TrimSpace(status)); mode {
	case ListModeActive, ListModeRecent, ListModeHistory:
		return mode, nil
	default:
		return "", fmt.Errorf("unsupported lease list status")
	}
}
