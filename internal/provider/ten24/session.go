package ten24

import (
	"context"

	"github.com/byte-v-forge/proxy-gateway/internal/provider"
)

func (p *Provider) Fetch(ctx context.Context) ([]provider.Node, error) {
	if p.cfg.APIURL == "" {
		return nil, nil
	}
	return p.fetchAPI(ctx)
}
