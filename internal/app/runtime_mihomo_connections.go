package app

import (
	"context"
	"time"

	dashboardapp "github.com/byte-v-forge/proxy-runtime/internal/app/dashboard"
	"github.com/byte-v-forge/proxy-runtime/internal/runtimehttp"

	"github.com/byte-v-forge/proxy-runtime/internal/app/appcore"
	"github.com/byte-v-forge/proxy-runtime/internal/app/mihomonative/connection"
)

func (r *Runtime) closeMihomoInUserConnections(ctx context.Context, usernames []string) {
	if err := r.closeMihomoConnections(ctx, connection.Selector{InboundUsers: usernames}); err != nil {
		r.logger.Warn("mihomo in-user connection cleanup failed", "error_type", appcore.ErrorLogType(err))
	}
}

func (r *Runtime) closeMihomoConnections(ctx context.Context, selector connection.Selector) error {
	base, err := dashboardapp.APIURL(r.cfg.Mihomo.APIAddr)
	if err != nil {
		return err
	}
	client := runtimehttp.New(5 * time.Second)
	return connection.CloseMatching(ctx, client, base, r.cfg.ControlAuthToken, selector)
}
