package app

import (
	"fmt"

	leaseapp "github.com/byte-v-forge/proxy-runtime/internal/app/lease"
	"github.com/byte-v-forge/proxy-runtime/internal/provider/accountproxy"
)

func (c leaseCoordinator) newSessionProvider(providerCfg accountproxy.Config) (leaseapp.SessionProvider, error) {
	if c.deps.sessionProviders == nil {
		return nil, fmt.Errorf("provider session factory is required")
	}
	return c.deps.sessionProviders.NewSessionProvider(providerCfg)
}
