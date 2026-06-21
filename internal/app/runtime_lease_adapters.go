package app

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/byte-v-forge/proxy-gateway/internal/app/appcore"
	leaseapp "github.com/byte-v-forge/proxy-gateway/internal/app/lease"
	"github.com/byte-v-forge/proxy-gateway/internal/app/store"
	"github.com/byte-v-forge/proxy-gateway/internal/dataplane"
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

func writeLeaseHTTPError(w http.ResponseWriter, err error, fallbackStatus int) {
	if errors.Is(err, leaseapp.ErrLeaseIDRequired) {
		writeHTTPError(w, appcore.InvalidArgument(err.Error(), err), http.StatusBadRequest)
		return
	}
	if store.IsNotFound(err) {
		writeHTTPError(w, errors.New("lease not found"), http.StatusNotFound)
		return
	}
	writeHTTPError(w, err, fallbackStatus)
}

func leaseProfilePolicyError(err error) error {
	switch {
	case errors.Is(err, leaseapp.ErrProfileDynamicIPNotConfigured), errors.Is(err, leaseapp.ErrProfileLeaseRequiresSticky):
		return appcore.FailedPrecondition(err.Error(), err)
	case errors.Is(err, leaseapp.ErrRequestRequiresStickyDynamicIP):
		return appcore.InvalidArgument(err.Error(), err)
	default:
		return err
	}
}
