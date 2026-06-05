package app

import (
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	proxyruntimev1 "github.com/byte-v-forge/common-lib/gen/go/byte/v/forge/contracts/proxyruntime/v1"
	"github.com/byte-v-forge/proxy-runtime/internal/provider/accountproxy"
	"github.com/byte-v-forge/proxy-runtime/internal/sourceplane"
	"google.golang.org/protobuf/types/known/durationpb"
)

func (s *PostgresStore) dynamicSource(account *proxyruntimev1.ProxyProviderAccount, gateways []accountproxy.Gateway) (*proxyruntimev1.ProxySourceDescriptor, error) {
	return s.accountProviders.DynamicSource(account.GetProviderId(), account.GetDisplayName(), account.GetAccountId(), gateways)
}

func (r proxySourceRecord) toDescriptor(payload proxySourcePayload) *proxyruntimev1.ProxySourceDescriptor {
	switch r.SourceKind {
	case proxySourceKindSubscription:
		return subscriptionSourceDescriptor(r, payload.Subscription)
	case proxySourceKindFixed:
		return fixedProxySourceDescriptor(r, payload.FixedProxy)
	default:
		return &proxyruntimev1.ProxySourceDescriptor{SourceId: r.SourceID, ProviderId: proxySourceProviderID, DisplayName: r.DisplayName, Enabled: r.Enabled}
	}
}

func subscriptionSourceDescriptor(record proxySourceRecord, item *sourceplane.SubscriptionProvider) *proxyruntimev1.ProxySourceDescriptor {
	source := &proxyruntimev1.ProxySourceDescriptor{SourceId: record.SourceID, ProviderId: proxySourceProviderID, DisplayName: firstNonEmpty(record.DisplayName, record.SourceID), Kind: proxyruntimev1.ProxySourceKind_PROXY_SOURCE_KIND_SUBSCRIPTION, Enabled: record.Enabled, Capabilities: []proxyruntimev1.ProxyCapability{proxyruntimev1.ProxyCapability_PROXY_CAPABILITY_SUBSCRIPTION_PROVIDER, proxyruntimev1.ProxyCapability_PROXY_CAPABILITY_POOL_REFRESH}, Model: &proxyruntimev1.ProxySourceDescriptor_Subscription{Subscription: &proxyruntimev1.ProxySubscriptionSourceDescriptor{}}}
	if item == nil {
		return source
	}
	source.DisplayName = firstNonEmpty(record.DisplayName, item.DisplayName, item.ID)
	source.Model = &proxyruntimev1.ProxySourceDescriptor_Subscription{Subscription: &proxyruntimev1.ProxySubscriptionSourceDescriptor{Url: item.URL, Interval: durationpb.New(defaultDuration(item.Interval, time.Hour)), Filter: item.Filter, ExcludeFilter: item.ExcludeFilter, HealthCheckUrl: item.HealthCheckURL, HealthInterval: durationpb.New(defaultDuration(item.HealthInterval, 300*time.Second)), HealthTimeout: durationpb.New(defaultDuration(item.HealthTimeout, 5*time.Second)), HealthLazy: item.HealthLazy, ExpectedStatus: defaultExpectedStatus(item.ExpectedStatus), RegionCodes: cleanRegionCodes(item.RegionCodes)}}
	return source
}

func fixedProxySourceDescriptor(record proxySourceRecord, item *sourceplane.FixedProxy) *proxyruntimev1.ProxySourceDescriptor {
	source := &proxyruntimev1.ProxySourceDescriptor{SourceId: record.SourceID, ProviderId: proxySourceProviderID, DisplayName: firstNonEmpty(record.DisplayName, record.SourceID), Kind: proxyruntimev1.ProxySourceKind_PROXY_SOURCE_KIND_FIXED_PROXY, Enabled: record.Enabled, Model: &proxyruntimev1.ProxySourceDescriptor_FixedProxy{FixedProxy: &proxyruntimev1.ProxyFixedSourceDescriptor{}}}
	if item == nil {
		return source
	}
	source.DisplayName = firstNonEmpty(record.DisplayName, item.DisplayName, item.ID)
	source.Model = &proxyruntimev1.ProxySourceDescriptor_FixedProxy{FixedProxy: &proxyruntimev1.ProxyFixedSourceDescriptor{EndpointCount: configuredEndpointCount(item.URI), RegionCodes: cleanRegionCodes(item.RegionCodes), Uri: item.URI}}
	return source
}

func configuredEndpointCount(value string) uint32 {
	if strings.TrimSpace(value) == "" {
		return 0
	}
	return 1
}

func defaultExpectedStatus(status uint32) uint32 {
	if status == 0 {
		return 204
	}
	return status
}

func fixedSourceName(rawURI string) string {
	parsed, err := url.Parse(strings.TrimSpace(rawURI))
	if err != nil {
		return ""
	}
	return strings.TrimSpace(parsed.Fragment)
}

func validateFixedProxyURI(rawURI string) error {
	parsed, err := url.Parse(strings.TrimSpace(rawURI))
	if err != nil {
		return fmt.Errorf("parse fixed proxy uri: %w", err)
	}
	if parsed == nil || parsed.Scheme == "" || parsed.Hostname() == "" || parsed.Port() == "" {
		return errors.New("fixed proxy uri requires scheme, host and port")
	}
	if strings.ToLower(parsed.Scheme) != "vless" {
		return fmt.Errorf("unsupported fixed proxy scheme %q", parsed.Scheme)
	}
	if parsed.User == nil {
		return errors.New("vless uri requires uuid")
	}
	if _, err := parsePort(parsed.Port()); err != nil {
		return err
	}
	return nil
}
