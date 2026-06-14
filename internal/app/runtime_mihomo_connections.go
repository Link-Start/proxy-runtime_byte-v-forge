package app

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	dashboardapp "github.com/byte-v-forge/proxy-runtime/internal/app/dashboard"
	"github.com/byte-v-forge/proxy-runtime/internal/runtimehttp"
)

type mihomoConnectionsResponse struct {
	Connections []mihomoConnection `json:"connections"`
}

type mihomoConnection struct {
	ID          string                   `json:"id"`
	Rule        string                   `json:"rule"`
	RulePayload string                   `json:"rulePayload"`
	Chains      []string                 `json:"chains"`
	Metadata    mihomoConnectionMetadata `json:"metadata"`
}

type mihomoConnectionMetadata struct {
	InboundUser string `json:"inboundUser"`
}

func (r *Runtime) closeMihomoInUserConnections(ctx context.Context, usernames []string) {
	if err := r.closeMihomoConnections(ctx, mihomoConnectionSelector{inboundUsers: usernames}); err != nil {
		r.logger.Warn("mihomo in-user connection cleanup failed", "error", err)
	}
}

type mihomoConnectionSelector struct {
	inboundUsers []string
	chains       []string
}

func (r *Runtime) closeMihomoConnections(ctx context.Context, selector mihomoConnectionSelector) error {
	targets := normalizedSet(selector.inboundUsers)
	chains := normalizedSet(selector.chains)
	if len(targets) == 0 && len(chains) == 0 {
		return nil
	}
	base, err := dashboardapp.APIURL(r.cfg.Mihomo.APIAddr)
	if err != nil {
		return err
	}
	client := runtimehttp.New(5 * time.Second)
	connections, err := listMihomoConnections(ctx, client, base, r.cfg.ControlAuthToken)
	if err != nil {
		return err
	}
	failureCount := 0
	for _, connection := range connections {
		if !connectionMatches(connection, targets, chains) {
			continue
		}
		if err := deleteMihomoConnection(ctx, client, base, connection.ID, r.cfg.ControlAuthToken); err != nil {
			failureCount++
		}
	}
	if failureCount > 0 {
		return fmt.Errorf("delete mihomo connections failed for %d connection(s)", failureCount)
	}
	return nil
}

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

func connectionMatches(connection mihomoConnection, inboundUsers map[string]struct{}, chains map[string]struct{}) bool {
	if _, ok := inboundUsers[normalizedKey(connection.Metadata.InboundUser)]; ok {
		return true
	}
	if connection.Rule == "InUser" {
		if _, ok := inboundUsers[normalizedKey(connection.RulePayload)]; ok {
			return true
		}
	}
	for _, chain := range connection.Chains {
		if _, ok := chains[normalizedKey(chain)]; ok {
			return true
		}
	}
	return false
}

func normalizedSet(values []string) map[string]struct{} {
	out := map[string]struct{}{}
	for _, value := range values {
		if key := normalizedKey(value); key != "" {
			out[key] = struct{}{}
		}
	}
	return out
}

func normalizedKey(value string) string {
	return strings.TrimSpace(value)
}
