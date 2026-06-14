package app

import (
	"context"
	"fmt"
	"time"

	dashboardapp "github.com/byte-v-forge/proxy-runtime/internal/app/dashboard"
	"github.com/byte-v-forge/proxy-runtime/internal/runtimehttp"
)

func (r *Runtime) closeMihomoInUserConnections(ctx context.Context, usernames []string) {
	if err := r.closeMihomoConnections(ctx, mihomoConnectionSelector{inboundUsers: usernames}); err != nil {
		r.logger.Warn("mihomo in-user connection cleanup failed", "error", err)
	}
}

func (r *Runtime) closeMihomoConnections(ctx context.Context, selector mihomoConnectionSelector) error {
	targets := normalizedSet(selector.inboundUsers)
	chains := normalizedSet(selector.chains)
	if len(targets) == 0 && len(chains) == 0 {
		return nil
	}
	base, err := dashboardapp.APIURL(r.cfg.Mihomo.APIAddr)
	if err != nil {
		return err
	}
	client := runtimehttp.New(5 * time.Second)
	connections, err := listMihomoConnections(ctx, client, base, r.cfg.ControlAuthToken)
	if err != nil {
		return err
	}
	failureCount := 0
	for _, connection := range connections {
		if !connectionMatches(connection, targets, chains) {
			continue
		}
		if err := deleteMihomoConnection(ctx, client, base, connection.ID, r.cfg.ControlAuthToken); err != nil {
			failureCount++
		}
	}
	if failureCount > 0 {
		return fmt.Errorf("delete mihomo connections failed for %d connection(s)", failureCount)
	}
	return nil
}
