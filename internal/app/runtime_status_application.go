package app

import (
	"context"

	proxygatewayv1 "github.com/byte-v-forge/proxy-gateway/gen/go/byte/v/forge/contracts/proxygateway/v1"
	"github.com/byte-v-forge/proxy-gateway/internal/dataplane"

	"github.com/byte-v-forge/proxy-gateway/internal/app/appcore"
)

type runtimeStatusApplication struct {
	runtimeStatus func() *proxygatewayv1.ProxyGatewayStatus
}

type runtimeStatusApplicationDependencies struct {
	RuntimeStatus func() *proxygatewayv1.ProxyGatewayStatus
}

func newRuntimeStatusApplication(deps runtimeStatusApplicationDependencies) runtimeStatusApplication {
	return runtimeStatusApplication{runtimeStatus: deps.RuntimeStatus}
}

func (a runtimeStatusApplication) GetProxyGatewayStatus(context.Context) (*proxygatewayv1.GetProxyGatewayStatusResponse, error) {
	if a.runtimeStatus == nil {
		return nil, appcore.InternalError("runtime status provider is not configured", nil)
	}
	return &proxygatewayv1.GetProxyGatewayStatusResponse{Status: a.runtimeStatus()}, nil
}

func (r *Runtime) runtimeStatus() *proxygatewayv1.ProxyGatewayStatus {
	dataPlaneStatus := r.dataPlane.Status()
	reconcile := r.currentReconcileState()
	leaseRestore := r.currentLeaseRestoreState()
	leaseWorker := r.currentLeaseWorkerState()
	settingsApply := r.currentSettingsApplyState()
	configStale := dataPlaneConfigStale(dataPlaneStatus)
	ready := dataPlaneStatus.Running && dataPlaneStatus.LastError == "" && !configStale
	return &proxygatewayv1.ProxyGatewayStatus{
		Ready:                ready,
		Status:               runtimeStatusLabel(ready, dataPlaneStatus, configStale, reconcile, leaseRestore, leaseWorker, settingsApply),
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

func (r *Runtime) currentLeaseWorkerState() runtimeLeaseWorkerState {
	r.leaseWorkerMu.RLock()
	defer r.leaseWorkerMu.RUnlock()
	return r.leaseWorker
}

func (r *Runtime) currentSettingsApplyState() runtimeSettingsApplyState {
	r.settingsApplyMu.RLock()
	defer r.settingsApplyMu.RUnlock()
	return r.settingsApply
}

func dataPlaneConfigStale(status dataplane.Status) bool {
	return status.DesiredConfigHash != "" && status.DesiredConfigHash != status.AppliedConfigHash
}

func runtimeStatusLabel(ready bool, dataPlaneStatus dataplane.Status, configStale bool, reconcile runtimeReconcileState, leaseRestore runtimeLeaseRestoreState, leaseWorker runtimeLeaseWorkerState, settingsApply runtimeSettingsApplyState) string {
	switch {
	case reconcile.running || reconcile.pending || leaseRestore.running || leaseWorker.running || settingsApply.running || settingsApply.pending:
		return "applying"
	case !dataPlaneStatus.Running:
		return "stopped"
	case !ready || configStale || reconcile.lastError != "" || leaseRestore.lastError != "" || leaseWorker.lastError != "" || settingsApply.lastError != "":
		return "degraded"
	default:
		return "running"
	}
}
