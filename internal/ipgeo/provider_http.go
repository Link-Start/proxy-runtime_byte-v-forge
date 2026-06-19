package ipgeo

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
)

type httpProvider struct {
	client   *http.Client
	template string
	auth     AuthConfig
}

func newHTTPProvider(client *http.Client, template string, auth AuthConfig) httpProvider {
	return httpProvider{client: client, template: template, auth: auth}
}

func (p httpProvider) Lookup(ctx context.Context, ip string) (*proxyruntimev1.ProxyExitGeo, error) {
	payload, err := p.lookupJSON(ctx, ip)
	if err != nil {
		return nil, err
	}
	geo := parseGeo(payload)
	if geo.GetCountryCode() == "" && geo.GetRegion() == "" && geo.GetCity() == "" {
		return nil, errors.New("IP geo response is empty")
	}
	return geo, nil
}

func (p httpProvider) lookupJSON(ctx context.Context, ip string) (map[string]any, error) {
	keys := []string{""}
	if p.auth.APIKey != nil && len(p.auth.APIKey.Keys) > 0 {
		keys = p.auth.APIKey.Keys
	}
	var lastErr error
	for _, key := range keys {
		rawURL, err := p.requestURL(ip, key)
		if err != nil {
			return nil, err
		}
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
		if err != nil {
			return nil, err
		}
		req.Close = true
		req.Header.Set("Accept", "application/json")
		resp, err := p.client.Do(req)
		if err != nil {
			lastErr = err
			continue
		}
		body, readErr := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
		_ = resp.Body.Close()
		if readErr != nil {
			return nil, readErr
		}
		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			lastErr = fmt.Errorf("IP geo request failed: HTTP %d", resp.StatusCode)
			continue
		}
		var payload map[string]any
		if err := json.Unmarshal(body, &payload); err != nil {
			return nil, fmt.Errorf("decode IP geo response: %w", err)
		}
		return payload, nil
	}
	return nil, lastErr
}

func (p httpProvider) requestURL(ip string, key string) (string, error) {
	value := strings.ReplaceAll(p.template, "{ip}", url.QueryEscape(ip))
	parsed, err := url.Parse(value)
	if err != nil {
		return "", err
	}
	if p.auth.APIKey != nil && p.auth.APIKey.Placement == "query" && key != "" {
		query := parsed.Query()
		query.Set(p.auth.APIKey.Name, key)
		parsed.RawQuery = query.Encode()
	}
	return parsed.String(), nil
}
