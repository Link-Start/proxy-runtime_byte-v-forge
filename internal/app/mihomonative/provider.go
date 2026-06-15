package mihomonative

import (
	"errors"
	"fmt"
	"net/url"
	"path/filepath"
	"strings"
)

func SubscriptionProvider(item Subscription) (Provider, Subscription, error) {
	id := safeID(item.ID)
	name := strings.TrimSpace(item.Name)
	rawURL := strings.TrimSpace(item.URL)
	if id == "" {
		id = StableID("sub", rawURL)
	}
	if name == "" {
		return Provider{}, Subscription{}, errors.New("subscription name is required")
	}
	if rawURL == "" {
		return Provider{}, Subscription{}, fmt.Errorf("subscription %q url is required", name)
	}
	parsed, err := url.Parse(rawURL)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return Provider{}, Subscription{}, fmt.Errorf("subscription %q url is invalid", name)
	}
	if !strings.EqualFold(parsed.Scheme, "http") && !strings.EqualFold(parsed.Scheme, "https") {
		return Provider{}, Subscription{}, fmt.Errorf("subscription %q url must use http or https", name)
	}
	return Provider{
		Type:     "http",
		URL:      rawURL,
		Path:     filepath.Join("providers", id+".yaml"),
		Interval: defaultProviderHealthPeriod,
		Header:   map[string][]string{"User-Agent": {defaultProviderUserAgent}},
		HealthCheck: &HealthCheck{
			Enable:         true,
			URL:            defaultProviderHealthURL,
			Interval:       defaultProviderHealthPeriod,
			Timeout:        defaultProviderHealthWait,
			Lazy:           true,
			ExpectedStatus: 204,
		},
	}, Subscription{ID: id, Name: name, URL: rawURL}, nil
}
