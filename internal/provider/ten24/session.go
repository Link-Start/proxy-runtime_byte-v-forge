package ten24

import (
	"context"

	proxyruntimev1 "github.com/byte-v-forge/common-lib/gen/go/byte/v/forge/contracts/proxyruntime/v1"
	"github.com/byte-v-forge/proxy-runtime/internal/provider"
)

func (p *Provider) Fetch(ctx context.Context, session *proxyruntimev1.ProxySession) ([]provider.Node, error) {
	if session != nil {
		return nil, provider.ErrUnsupportedCapability
	}
	if p.cfg.APIURL == "" {
		return nil, nil
	}
	return p.fetchAPI(ctx)
}

func (p *Provider) CreateSession(context.Context, *proxyruntimev1.AcquireProxyLeaseRequest) (*proxyruntimev1.ProxySession, error) {
	return nil, provider.ErrUnsupportedCapability
}
