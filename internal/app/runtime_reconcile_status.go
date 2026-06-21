package app

import (
	"github.com/byte-v-forge/proxy-gateway/internal/app/appcore"
	"github.com/byte-v-forge/proxy-gateway/internal/dataplane"
)

type runtimeReconcileState struct {
	pending   bool
	running   bool
	lastError string
}

type runtimeLeaseRestoreState struct {
	running   bool
	lastError string
}

type runtimeLeaseWorkerState struct {
	running   bool
	lastError string
}

type runtimeSettingsApplyState struct {
	pending   bool
	running   bool
	lastError string
}

func (r *Runtime) markReconcilePending() {
	r.reconcileMu.Lock()
	defer r.reconcileMu.Unlock()
	r.reconcileState.pending = true
}

func (r *Runtime) markReconcileStarted() {
	r.reconcileMu.Lock()
	defer r.reconcileMu.Unlock()
	r.reconcileState.pending = false
	r.reconcileState.running = true
}

func (r *Runtime) markReconcileFinished(err error) {
	r.reconcileMu.Lock()
	defer r.reconcileMu.Unlock()
	r.reconcileState.running = false
	if err != nil {
		r.reconcileState.lastError = err.Error()
		return
	}
	r.reconcileState.lastError = ""
}

func (r *Runtime) dataPlaneStatus() string {
	return r.decorateSettingsApplyStatus(r.decorateLeaseWorkerStatus(r.decorateLeaseRestoreStatus(r.decorateRuntimeStatus(statusString(r.dataPlane.Status())))))
}

func statusString(status dataplane.Status) string {
	if !status.Running {
		return appcore.FirstNonEmpty(status.LastError, "stopped")
	}
	if status.DesiredConfigHash != "" && status.DesiredConfigHash != status.AppliedConfigHash {
		return "running; config projection stale"
	}
	return "running"
}

func (r *Runtime) decorateRuntimeStatus(status string) string {
	r.reconcileMu.RLock()
	defer r.reconcileMu.RUnlock()
	switch {
	case r.reconcileState.running:
		return status + "; reconciling"
	case r.reconcileState.pending:
		return status + "; reconcile pending"
	case r.reconcileState.lastError != "":
		return status + "; last reconcile failed"
	default:
		return status
	}
}

func (r *Runtime) markLeaseRestoreStarted() {
	r.leaseRestoreMu.Lock()
	defer r.leaseRestoreMu.Unlock()
	r.leaseRestore.running = true
}

func (r *Runtime) markLeaseRestoreFinished(err error) {
	r.leaseRestoreMu.Lock()
	defer r.leaseRestoreMu.Unlock()
	r.leaseRestore.running = false
	if err != nil {
		r.leaseRestore.lastError = "failed"
		return
	}
	r.leaseRestore.lastError = ""
}

func (r *Runtime) decorateLeaseRestoreStatus(status string) string {
	r.leaseRestoreMu.RLock()
	defer r.leaseRestoreMu.RUnlock()
	switch {
	case r.leaseRestore.running:
		return status + "; restoring leases"
	case r.leaseRestore.lastError != "":
		return status + "; lease restore failed"
	default:
		return status
	}
}

func (r *Runtime) markLeaseWorkerStarted() {
	r.leaseWorkerMu.Lock()
	defer r.leaseWorkerMu.Unlock()
	r.leaseWorker.running = true
}

func (r *Runtime) markLeaseWorkerFinished(err error) {
	r.leaseWorkerMu.Lock()
	defer r.leaseWorkerMu.Unlock()
	r.leaseWorker.running = false
	if err != nil {
		r.leaseWorker.lastError = "failed"
		return
	}
	r.leaseWorker.lastError = ""
}

func (r *Runtime) decorateLeaseWorkerStatus(status string) string {
	r.leaseWorkerMu.RLock()
	defer r.leaseWorkerMu.RUnlock()
	switch {
	case r.leaseWorker.running:
		return status + "; running lease worker"
	case r.leaseWorker.lastError != "":
		return status + "; lease worker failed"
	default:
		return status
	}
}

func (r *Runtime) markSettingsApplyPending() {
	r.settingsApplyMu.Lock()
	defer r.settingsApplyMu.Unlock()
	r.settingsApply.pending = true
}

func (r *Runtime) markSettingsApplyStartedIfPending() bool {
	r.settingsApplyMu.Lock()
	defer r.settingsApplyMu.Unlock()
	if !r.settingsApply.pending {
		return false
	}
	r.settingsApply.pending = false
	r.settingsApply.running = true
	return true
}

func (r *Runtime) markSettingsApplyFinished(err error) {
	r.settingsApplyMu.Lock()
	defer r.settingsApplyMu.Unlock()
	r.settingsApply.running = false
	if err != nil {
		r.settingsApply.lastError = "failed"
		return
	}
	r.settingsApply.lastError = ""
}

func (r *Runtime) decorateSettingsApplyStatus(status string) string {
	r.settingsApplyMu.RLock()
	defer r.settingsApplyMu.RUnlock()
	switch {
	case r.settingsApply.running:
		return status + "; applying settings"
	case r.settingsApply.pending:
		return status + "; settings apply pending"
	case r.settingsApply.lastError != "":
		return status + "; settings apply failed"
	default:
		return status
	}
}
