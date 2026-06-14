package app

import (
	"fmt"

	"github.com/byte-v-forge/proxy-runtime/internal/provider"
	"github.com/byte-v-forge/proxy-runtime/internal/provider/accountproxy"
)

func (c leaseCoordinator) newSessionProvider(providerCfg accountproxy.Config) (provider.SessionProvider, error) {
	if c.deps.sessionProviders == nil {
		return nil, fmt.Errorf("provider session factory is required")
	}
	return c.deps.sessionProviders.NewSessionProvider(providerCfg)
}
