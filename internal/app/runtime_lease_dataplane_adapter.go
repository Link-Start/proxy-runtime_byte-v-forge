package app

import (
	"context"
	"errors"

	leaseapp "github.com/byte-v-forge/proxy-runtime/internal/app/lease"
	"github.com/byte-v-forge/proxy-runtime/internal/dataplane"
)

type leaseRouteDataPlane interface {
	UpsertSessionRoute(context.Context, dataplane.SessionRoute) error
	DeleteSessionRoute(context.Context, dataplane.SessionRoute) error
}

type leaseRuntimeDataPlaneApplier struct {
	dataPlane leaseRouteDataPlane
}

func (a leaseRuntimeDataPlaneApplier) UpsertSessionRoute(ctx context.Context, route leaseapp.SessionRoute) error {
	if a.dataPlane == nil {
		return errors.New("dataplane route applier is required")
	}
	return a.dataPlane.UpsertSessionRoute(ctx, dataPlaneSessionRoute(route))
}

func (a leaseRuntimeDataPlaneApplier) DeleteSessionRoute(ctx context.Context, route leaseapp.SessionRoute) error {
	if a.dataPlane == nil {
		return errors.New("dataplane route applier is required")
	}
	return a.dataPlane.DeleteSessionRoute(ctx, dataPlaneSessionRoute(route))
}

func dataPlaneSessionRoute(route leaseapp.SessionRoute) dataplane.SessionRoute {
	return dataplane.SessionRoute{
		SessionID:   route.SessionID,
		Listener:    dataPlaneLocalService(route.Listener),
		Pool:        route.Pool,
		DialerProxy: route.DialerProxy,
	}
}

func dataPlaneLocalService(service leaseapp.LocalService) dataplane.LocalService {
	return dataplane.LocalService{
		Name:     service.Name,
		Addr:     service.Addr,
		Protocol: service.Protocol,
		Username: service.Username,
		Password: service.Password,
		Route:    service.Route,
	}
}
