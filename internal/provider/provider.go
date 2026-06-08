package provider

import (
	"context"
	"errors"
	"net"
	"net/url"
	"strconv"
	"strings"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
)

var ErrUnsupportedCapability = errors.New("proxy provider does not support requested capability")

type Node struct {
	ID           string
	URL          *url.URL
	ProviderID   string
	SessionID    string
	UpstreamKind proxyruntimev1.ProxyUpstreamKind
	RotationMode proxyruntimev1.ProxyRotationMode
	Labels       map[string]string
}

func (n Node) RedactedURL() string {
	if n.URL == nil {
		return ""
	}
	redacted := *n.URL
	if redacted.User != nil {
		redacted.User = url.UserPassword(redacted.User.Username(), "xxxxx")
	}
	return redacted.String()
}

func (n Node) Endpoint() *proxyruntimev1.ProxyEndpoint {
	host, port := splitHostPort(n.URL)
	return &proxyruntimev1.ProxyEndpoint{
		Id:           n.ID,
		ProviderId:   n.ProviderID,
		Protocol:     protocolFromURL(n.URL),
		Host:         host,
		Port:         port,
		SessionId:    n.SessionID,
		UpstreamKind: n.UpstreamKind,
		RotationMode: n.RotationMode,
		Labels:       cloneLabels(n.Labels),
	}
}

type PoolProvider interface {
	Name() string
	Descriptor() *proxyruntimev1.ProxyProviderDescriptor
	Fetch(ctx context.Context) ([]Node, error)
}

type SessionProvider interface {
	Name() string
	CreateSession(ctx context.Context, req *proxyruntimev1.AcquireProxyLeaseRequest) (*proxyruntimev1.ProxySession, error)
	FetchSession(ctx context.Context, session *proxyruntimev1.ProxySession) ([]Node, error)
	ReleaseSession(ctx context.Context, session *proxyruntimev1.ProxySession) error
}

type Empty struct{}

const EmptyProviderID = "none"

func (Empty) Name() string {
	return EmptyProviderID
}

func (Empty) Descriptor() *proxyruntimev1.ProxyProviderDescriptor {
	return &proxyruntimev1.ProxyProviderDescriptor{
		ProviderId:  EmptyProviderID,
		DisplayName: "No provider",
		Capabilities: []proxyruntimev1.ProxyCapability{
			proxyruntimev1.ProxyCapability_PROXY_CAPABILITY_UNIFIED_EGRESS_GATEWAY,
		},
		RotationModes: []proxyruntimev1.ProxyRotationMode{
			proxyruntimev1.ProxyRotationMode_PROXY_ROTATION_MODE_NONE,
		},
	}
}

func (Empty) Fetch(context.Context) ([]Node, error) {
	return nil, nil
}

func splitHostPort(proxyURL *url.URL) (string, uint32) {
	if proxyURL == nil {
		return "", 0
	}
	host, portValue, err := net.SplitHostPort(proxyURL.Host)
	if err != nil {
		return proxyURL.Hostname(), 0
	}
	port, err := strconv.Atoi(portValue)
	if err != nil || port < 0 {
		return host, 0
	}
	return host, uint32(port)
}

func protocolFromURL(proxyURL *url.URL) proxyruntimev1.ProxyProtocol {
	if proxyURL == nil {
		return proxyruntimev1.ProxyProtocol_PROXY_PROTOCOL_UNSPECIFIED
	}
	switch strings.ToLower(proxyURL.Scheme) {
	case "socks5", "socks5h":
		return proxyruntimev1.ProxyProtocol_PROXY_PROTOCOL_SOCKS5
	case "http", "https":
		return proxyruntimev1.ProxyProtocol_PROXY_PROTOCOL_HTTP
	default:
		return proxyruntimev1.ProxyProtocol_PROXY_PROTOCOL_UNSPECIFIED
	}
}

func cloneLabels(labels map[string]string) map[string]string {
	if len(labels) == 0 {
		return nil
	}
	cloned := make(map[string]string, len(labels))
	for key, value := range labels {
		cloned[key] = value
	}
	return cloned
}
