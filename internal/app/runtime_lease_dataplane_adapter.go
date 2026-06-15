package app

import (
	"context"
	"errors"
	"time"

	leaseapp "github.com/byte-v-forge/proxy-runtime/internal/app/lease"
	"github.com/byte-v-forge/proxy-runtime/internal/dataplane"
)

type leaseRouteDataPlane interface {
	UpsertSessionRoute(context.Context, dataplane.SessionRoute) error
	DeleteSessionRoute(context.Context, dataplane.SessionRoute) error
}

type leaseRuntimeDataPlaneApplier struct {
	dataPlane leaseRouteDataPlane
	metrics   *runtimeMetrics
}

func (a leaseRuntimeDataPlaneApplier) UpsertSessionRoute(ctx context.Context, route leaseapp.SessionRoute) error {
	if a.dataPlane == nil {
		return errors.New("dataplane route applier is required")
	}
	startedAt := time.Now()
	err := a.dataPlane.UpsertSessionRoute(ctx, dataPlaneSessionRoute(route))
	a.observe(runtimeMetricDataPlaneUpsertSessionRoute, startedAt, err)
	return err
}

func (a leaseRuntimeDataPlaneApplier) DeleteSessionRoute(ctx context.Context, route leaseapp.SessionRoute) error {
	if a.dataPlane == nil {
		return errors.New("dataplane route applier is required")
	}
	startedAt := time.Now()
	err := a.dataPlane.DeleteSessionRoute(ctx, dataPlaneSessionRoute(route))
	a.observe(runtimeMetricDataPlaneDeleteSessionRoute, startedAt, err)
	return err
}

func (a leaseRuntimeDataPlaneApplier) observe(operation string, startedAt time.Time, err error) {
	if a.metrics != nil {
		a.metrics.Observe(operation, startedAt, err)
	}
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
