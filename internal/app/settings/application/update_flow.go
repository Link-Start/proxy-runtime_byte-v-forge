package application

import (
	"context"

	proxygatewayv1 "github.com/byte-v-forge/proxy-gateway/gen/go/byte/v/forge/contracts/proxygateway/v1"
)

type updateOperation func(Repository) (*proxygatewayv1.ProxyGatewaySettings, error)

func (a Application) updateWithConnectionCleanup(ctx context.Context, errorMessage string, operation updateOperation) (*proxygatewayv1.ProxyGatewaySettings, error) {
	repository, err := a.repositoryOrError()
	if err != nil {
		return nil, err
	}
	before, err := repository.Load(ctx)
	if err != nil {
		return nil, err
	}
	settings, err := operation(repository)
	if err != nil {
		return nil, err
	}
	a.schedule(a.changedInUserConnectionUsernamesAfterUpdate(ctx, repository, before, errorMessage))
	return settings, nil
}

func (a Application) updateAndSchedule(ctx context.Context, operation updateOperation) (*proxygatewayv1.ProxyGatewaySettings, error) {
	repository, err := a.repositoryOrError()
	if err != nil {
		return nil, err
	}
	settings, err := operation(repository)
	if err != nil {
		return nil, err
	}
	a.schedule(nil)
	return settings, nil
}

func (a Application) changedInUserConnectionUsernamesAfterUpdate(ctx context.Context, repository Repository, before *proxygatewayv1.ProxyGatewayPersistentSettings, errorMessage string) []string {
	if repository == nil {
		a.warn(errorMessage, "error_type", errorType(ErrRepositoryRequired))
		return nil
	}
	after, err := repository.Load(ctx)
	if err != nil {
		a.warn(errorMessage, "error_type", errorType(err))
		return nil
	}
	return ChangedInUserConnectionUsernames(before, after)
}

func (a Application) warn(message string, args ...any) {
	if a.logger != nil {
		a.logger.Warn(message, args...)
	}
}
