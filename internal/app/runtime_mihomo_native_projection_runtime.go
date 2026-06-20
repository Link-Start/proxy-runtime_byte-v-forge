package app

import (
	"context"

	mihomoapp "github.com/byte-v-forge/proxy-runtime/internal/app/mihomonative/application"
)

func (r *Runtime) projectMihomoNativeSettings(ctx context.Context) error {
	if r == nil || r.settings == nil {
		return nil
	}
	return mihomoapp.Project(ctx, mihomoapp.ProjectionDependencies{
		ConfigDir:    r.cfg.Mihomo.ConfigDir,
		LoadSettings: r.settings.LoadMihomoNative,
		SaveSettings: r.settings.SaveMihomoNative,
	})
}
