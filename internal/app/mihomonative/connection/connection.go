package connection

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	dashboardapp "github.com/byte-v-forge/proxy-runtime/internal/app/dashboard"
)

// Selector matches mihomo connections by inbound user or proxy chain.
type Selector struct {
	InboundUsers []string
	Chains       []string
}

type connectionsResponse struct {
	Connections []connectionInfo `json:"connections"`
}

type connectionInfo struct {
	ID          string             `json:"id"`
	Rule        string             `json:"rule"`
	RulePayload string             `json:"rulePayload"`
	Chains      []string           `json:"chains"`
	Metadata    connectionMetadata `json:"metadata"`
}

type connectionMetadata struct {
	InboundUser string `json:"inboundUser"`
}

// CloseMatching deletes every mihomo connection matched by selector through the
// controller API at base. It is a no-op when the selector resolves to nothing.
func CloseMatching(ctx context.Context, client *http.Client, base *url.URL, token string, selector Selector) error {
	targets := normalizedSet(selector.InboundUsers)
	chains := normalizedSet(selector.Chains)
	if len(targets) == 0 && len(chains) == 0 {
		return nil
	}
	connections, err := list(ctx, client, base, token)
	if err != nil {
		return err
	}
	failureCount := 0
	for _, conn := range connections {
		if !matches(conn, targets, chains) {
			continue
		}
		if err := deleteConnection(ctx, client, base, conn.ID, token); err != nil {
			failureCount++
		}
	}
	if failureCount > 0 {
		return fmt.Errorf("delete mihomo connections failed for %d connection(s)", failureCount)
	}
	return nil
}

func list(ctx context.Context, client *http.Client, base *url.URL, token string) ([]connectionInfo, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, controllerURL(base, "/connections"), nil)
	if err != nil {
		return nil, err
	}
	dashboardapp.ApplyControllerAuthorization(req, token)
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("list mihomo connections returned HTTP %d", resp.StatusCode)
	}
	var payload connectionsResponse
	if err := json.NewDecoder(io.LimitReader(resp.Body, 1<<20)).Decode(&payload); err != nil {
		return nil, err
	}
	return payload.Connections, nil
}

func deleteConnection(ctx context.Context, client *http.Client, base *url.URL, id string, token string) error {
	id = strings.TrimSpace(id)
	if id == "" {
		return nil
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, controllerURL(base, "/connections/"+url.PathEscape(id)), nil)
	if err != nil {
		return err
	}
	dashboardapp.ApplyControllerAuthorization(req, token)
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

func controllerURL(base *url.URL, path string) string {
	next := *base
	next.Path = path
	next.RawQuery = ""
	next.Fragment = ""
	return next.String()
}

func matches(conn connectionInfo, inboundUsers map[string]struct{}, chains map[string]struct{}) bool {
	if _, ok := inboundUsers[normalizedKey(conn.Metadata.InboundUser)]; ok {
		return true
	}
	if conn.Rule == "InUser" {
		if _, ok := inboundUsers[normalizedKey(conn.RulePayload)]; ok {
			return true
		}
	}
	for _, chain := range conn.Chains {
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
