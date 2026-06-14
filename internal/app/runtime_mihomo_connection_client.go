package app

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

func listMihomoConnections(ctx context.Context, client *http.Client, base *url.URL, token string) ([]mihomoConnection, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, mihomoControllerURL(base, "/connections"), nil)
	if err != nil {
		return nil, err
	}
	applyMihomoControllerAuth(req, token)
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("list mihomo connections returned HTTP %d", resp.StatusCode)
	}
	var payload mihomoConnectionsResponse
	if err := json.NewDecoder(io.LimitReader(resp.Body, 1<<20)).Decode(&payload); err != nil {
		return nil, err
	}
	return payload.Connections, nil
}

func deleteMihomoConnection(ctx context.Context, client *http.Client, base *url.URL, id string, token string) error {
	id = strings.TrimSpace(id)
	if id == "" {
		return nil
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, mihomoControllerURL(base, "/connections/"+url.PathEscape(id)), nil)
	if err != nil {
		return err
	}
	applyMihomoControllerAuth(req, token)
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return nil
	}
	return fmt.Errorf("delete mihomo connection returned HTTP %d", resp.StatusCode)
}

func mihomoControllerURL(base *url.URL, path string) string {
	next := *base
	next.Path = path
	next.RawQuery = ""
	next.Fragment = ""
	return next.String()
}
