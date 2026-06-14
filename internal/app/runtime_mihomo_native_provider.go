package app

import (
	"fmt"
	"net/url"
	"path/filepath"
	"strings"
)

func mihomoNativeSubscriptionProvider(item mihomoNativeSubscription) (mihomoNativeProvider, mihomoNativeSubscription, error) {
	id := runtimeSafeID(item.ID)
	name := strings.TrimSpace(item.Name)
	rawURL := strings.TrimSpace(item.URL)
	if id == "" {
		id = nativeStableID("sub", rawURL)
	}
	if name == "" {
		return mihomoNativeProvider{}, mihomoNativeSubscription{}, invalidArgument("subscription name is required", nil)
	}
	if rawURL == "" {
		return mihomoNativeProvider{}, mihomoNativeSubscription{}, invalidArgument(fmt.Sprintf("subscription %q url is required", name), nil)
	}
	parsed, err := url.Parse(rawURL)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return mihomoNativeProvider{}, mihomoNativeSubscription{}, invalidArgument(fmt.Sprintf("subscription %q url is invalid", name), nil)
	}
	if !strings.EqualFold(parsed.Scheme, "http") && !strings.EqualFold(parsed.Scheme, "https") {
		return mihomoNativeProvider{}, mihomoNativeSubscription{}, invalidArgument(fmt.Sprintf("subscription %q url must use http or https", name), nil)
	}
	return mihomoNativeProvider{
		Type:     "http",
		URL:      rawURL,
		Path:     filepath.Join("providers", id+".yaml"),
		Interval: defaultProviderHealthPeriod,
		Header:   map[string][]string{"User-Agent": {defaultProviderUserAgent}},
		HealthCheck: &mihomoNativeHealthCheck{
			Enable:         true,
			URL:            defaultProviderHealthURL,
			Interval:       defaultProviderHealthPeriod,
			Timeout:        defaultProviderHealthWait,
			Lazy:           true,
			ExpectedStatus: 204,
		},
	}, mihomoNativeSubscription{ID: id, Name: name, URL: rawURL}, nil
}
