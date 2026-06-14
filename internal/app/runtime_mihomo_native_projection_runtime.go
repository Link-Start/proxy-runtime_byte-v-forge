package app

import "context"

func (r *Runtime) projectMihomoNativeSettings(ctx context.Context) error {
	if r == nil || r.settings == nil {
		return nil
	}
	return projectMihomoNativeSettings(ctx, r.settings, r.cfg.Mihomo.ConfigDir)
}
