package app

import (
	"context"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
	"github.com/byte-v-forge/proxy-runtime/internal/dataplane"
)

type runtimeStatusApplication struct {
	runtime *Runtime
}

func newRuntimeStatusApplication(runtime *Runtime) runtimeStatusApplication {
	return runtimeStatusApplication{runtime: runtime}
}

func (s *RuntimeService) GetProxyRuntimeStatus(ctx context.Context, _ *proxyruntimev1.GetProxyRuntimeStatusRequest) (*proxyruntimev1.GetProxyRuntimeStatusResponse, error) {
	return s.status.GetProxyRuntimeStatus(ctx)
}

func (a runtimeStatusApplication) GetProxyRuntimeStatus(context.Context) (*proxyruntimev1.GetProxyRuntimeStatusResponse, error) {
	return &proxyruntimev1.GetProxyRuntimeStatusResponse{Status: a.runtime.runtimeStatus()}, nil
}

func (r *Runtime) runtimeStatus() *proxyruntimev1.ProxyRuntimeStatus {
	dataPlaneStatus := r.dataPlane.Status()
	reconcile := r.currentReconcileState()
	leaseRestore := r.currentLeaseRestoreState()
	configStale := dataPlaneConfigStale(dataPlaneStatus)
	ready := dataPlaneStatus.Running && dataPlaneStatus.LastError == "" && !configStale
	return &proxyruntimev1.ProxyRuntimeStatus{
		Ready:                ready,
		Status:               runtimeStatusLabel(ready, dataPlaneStatus, configStale, reconcile, leaseRestore),
		DataPlaneRunning:     dataPlaneStatus.Running,
		DataPlaneConfigStale: configStale,
		ReconcilePending:     reconcile.pending,
		ReconcileRunning:     reconcile.running,
		ReconcileFailed:      reconcile.lastError != "",
		LeaseRestoreRunning:  leaseRestore.running,
		LeaseRestoreFailed:   leaseRestore.lastError != "",
	}
}

func (r *Runtime) currentReconcileState() runtimeReconcileState {
	r.reconcileMu.RLock()
	defer r.reconcileMu.RUnlock()
	return r.reconcileState
}

func (r *Runtime) currentLeaseRestoreState() runtimeLeaseRestoreState {
	r.leaseRestoreMu.RLock()
	defer r.leaseRestoreMu.RUnlock()
	return r.leaseRestore
}

func dataPlaneConfigStale(status dataplane.Status) bool {
	return status.DesiredConfigHash != "" && status.DesiredConfigHash != status.AppliedConfigHash
}

func runtimeStatusLabel(ready bool, dataPlaneStatus dataplane.Status, configStale bool, reconcile runtimeReconcileState, leaseRestore runtimeLeaseRestoreState) string {
	switch {
	case !dataPlaneStatus.Running:
		return "stopped"
	case reconcile.running || reconcile.pending || leaseRestore.running:
		return "applying"
	case !ready || configStale || reconcile.lastError != "" || leaseRestore.lastError != "":
		return "degraded"
	default:
		return "running"
	}
}
