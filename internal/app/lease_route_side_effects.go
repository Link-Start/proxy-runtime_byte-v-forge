package app

import (
	"context"

	"github.com/byte-v-forge/proxy-runtime/internal/app/proxycheck"
)

type leaseRouteSideEffects struct {
	exitCheckCache         *proxycheck.ExitCheckCache
	closeInUserConnections leaseConnectionCleanupFunc
}

func (e leaseRouteSideEffects) afterRouteChange(ctx context.Context, accountID string) {
	e.clearExitCheckCache()
	if accountID == playgroundProfileID {
		e.closePlaygroundConnections(ctx)
	}
}

func (e leaseRouteSideEffects) clearExitCheckCache() {
	if e.exitCheckCache != nil {
		e.exitCheckCache.Clear()
	}
}

func (e leaseRouteSideEffects) closePlaygroundConnections(ctx context.Context) {
	if e.closeInUserConnections != nil {
		e.closeInUserConnections(ctx, []string{playgroundUsername})
	}
}
