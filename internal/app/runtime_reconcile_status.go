package app

import "github.com/byte-v-forge/proxy-runtime/internal/dataplane"

type runtimeReconcileState struct {
	pending   bool
	running   bool
	lastError string
}

type runtimeLeaseRestoreState struct {
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
	return r.decorateLeaseRestoreStatus(r.decorateRuntimeStatus(statusString(r.dataPlane.Status())))
}

func statusString(status dataplane.Status) string {
	if !status.Running {
		return firstNonEmpty(status.LastError, "stopped")
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
