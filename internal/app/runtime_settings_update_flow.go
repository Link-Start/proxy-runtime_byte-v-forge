package app

import (
	"context"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
)

type runtimeSettingsUpdateOperation func(runtimeSettingsRepository) (*proxyruntimev1.ProxyRuntimeSettings, error)

func (a runtimeSettingsApplication) updateSettingsWithConnectionCleanup(ctx context.Context, errorMessage string, operation runtimeSettingsUpdateOperation) (*proxyruntimev1.ProxyRuntimeSettings, error) {
	repository, err := a.settingsRepository()
	if err != nil {
		return nil, err
	}
	before, err := repository.load(ctx)
	if err != nil {
		return nil, err
	}
	settings, err := operation(repository)
	if err != nil {
		return nil, err
	}
	a.scheduleRuntimeSettingsApply(a.changedInUserConnectionUsernamesAfterUpdate(ctx, repository, before, errorMessage))
	return settings, nil
}

func (a runtimeSettingsApplication) changedInUserConnectionUsernamesAfterUpdate(ctx context.Context, repository runtimeSettingsRepository, before *runtimeSettingsFile, errorMessage string) []string {
	if repository == nil {
		a.warn(errorMessage, "error", internalError("runtime settings repository is not configured", nil))
		return nil
	}
	after, err := repository.load(ctx)
	if err != nil {
		a.warn(errorMessage, "error", err)
		return nil
	}
	return changedInUserConnectionUsernames(before, after)
}
