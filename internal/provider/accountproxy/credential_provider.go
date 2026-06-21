package accountproxy

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	proxygatewayv1 "github.com/byte-v-forge/proxy-gateway/gen/go/byte/v/forge/contracts/proxygateway/v1"
	"github.com/byte-v-forge/proxy-gateway/internal/clock"
	"github.com/byte-v-forge/proxy-gateway/internal/provider"
	"github.com/byte-v-forge/proxy-gateway/internal/random"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type CredentialProvider struct {
	cfg        Config
	definition Definition
	clock      clock.Clock
}

func NewCredentialProvider(cfg Config, definition Definition, clk clock.Clock) *CredentialProvider {
	return &CredentialProvider{cfg: cfg, definition: definition, clock: clk}
}

func (p *CredentialProvider) Name() string { return p.definition.ProviderID }

func (p *CredentialProvider) FetchSession(_ context.Context, session *proxygatewayv1.ProxySession) ([]provider.Node, error) {
	node, err := p.node(session)
	if err != nil {
		return nil, err
	}
	return []provider.Node{node}, nil
}

func (p *CredentialProvider) CreateSession(_ context.Context, req *proxygatewayv1.AcquireProxyLeaseRequest) (*proxygatewayv1.ProxySession, error) {
	policy := p.sessionPolicy(req.GetPolicy())
	if policy.Mode != proxygatewayv1.ProxySessionMode_PROXY_SESSION_MODE_STICKY && policy.Mode != proxygatewayv1.ProxySessionMode_PROXY_SESSION_MODE_ROTATING {
		return nil, provider.ErrUnsupportedCapability
	}
	policy.UpstreamKind = proxygatewayv1.ProxyUpstreamKind_PROXY_UPSTREAM_KIND_DYNAMIC_IP
	sessionID := requestedSessionID(policy)
	if sessionID == "" {
		generated, err := p.sessionID()
		if err != nil {
			return nil, fmt.Errorf("generate proxy session id: %w", err)
		}
		sessionID = generated
	}
	now := p.clock.Now().UTC()
	return &proxygatewayv1.ProxySession{SessionId: sessionID, ProviderId: p.Name(), Policy: policy, CreatedAt: timestamppb.New(now), ExpiresAt: timestamppb.New(now.Add(policyStickyTTL(policy))), AccountId: strings.TrimSpace(req.GetAccountId()), Purpose: strings.TrimSpace(req.GetPurpose()), Labels: sessionLabels(p.definition)}, nil
}

func (p *CredentialProvider) ReleaseSession(context.Context, *proxygatewayv1.ProxySession) error {
	return nil
}

func (p *CredentialProvider) sessionID() (string, error) {
	if p.definition.GenerateSessionID != nil {
		return p.definition.GenerateSessionID()
	}
	return random.Hex(8)
}

func (p *CredentialProvider) node(session *proxygatewayv1.ProxySession) (provider.Node, error) {
	gateway, ok := gatewayForPolicy(p.definition, session.GetPolicy())
	if !ok {
		return provider.Node{}, provider.ErrUnsupportedCapability
	}
	policy := p.sessionPolicy(nil)
	if session != nil {
		policy = p.sessionPolicy(session.Policy)
	}
	protocol := GatewayProtocol(gateway, p.definition.DefaultProtocol)
	proxyURL := url.URL{Scheme: protocol, Host: gatewayEndpointHost(gateway.EndpointURL), User: url.UserPassword(p.username(session), p.cfg.Password)}
	sessionID := ""
	if session != nil {
		sessionID = session.GetSessionId()
	}
	rotationMode := policy.GetRotationMode()
	if rotationMode == proxygatewayv1.ProxyRotationMode_PROXY_ROTATION_MODE_UNSPECIFIED {
		rotationMode = proxygatewayv1.ProxyRotationMode_PROXY_ROTATION_MODE_STICKY_SESSION
	}
	return provider.Node{ID: p.Name() + "-dynamic-" + gateway.ID, URL: &proxyURL, ProviderID: p.Name(), SessionID: sessionID, UpstreamKind: proxygatewayv1.ProxyUpstreamKind_PROXY_UPSTREAM_KIND_DYNAMIC_IP, RotationMode: rotationMode, Labels: map[string]string{"gateway": gateway.ID, "mode": "credential", "network_kind": "residential", "protocol": protocol, "session_mode": policy.GetMode().String(), "rotation_mode": rotationMode.String()}}, nil
}

func (p *CredentialProvider) username(session *proxygatewayv1.ProxySession) string {
	policy := p.sessionPolicy(nil)
	sessionID := ""
	if session != nil {
		policy = p.sessionPolicy(session.Policy)
		sessionID = session.GetSessionId()
	}
	if p.definition.BuildUsername == nil {
		return p.cfg.Username
	}
	return p.definition.BuildUsername(p.cfg.Username, policy, sessionID)
}

func sessionLabels(definition Definition) map[string]string {
	mode := "provider_configured"
	if definition.UsernameParameterSession {
		mode = "username_parameter"
	}
	return map[string]string{"provider": definition.ProviderID, "session_mode": mode}
}
