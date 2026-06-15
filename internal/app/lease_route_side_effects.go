package app

import "context"

type leaseRouteSideEffects struct {
	exitCheckCache         *proxyExitCheckCache
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
		e.exitCheckCache.clear()
	}
}

func (e leaseRouteSideEffects) closePlaygroundConnections(ctx context.Context) {
	if e.closeInUserConnections != nil {
		e.closeInUserConnections(ctx, []string{playgroundUsername})
	}
}
