package app

import "github.com/byte-v-forge/proxy-runtime/internal/dataplane"

type runtimeReconcileState struct {
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
	return r.decorateRuntimeStatus(statusString(r.dataPlane.Status()))
}

func statusString(status dataplane.Status) string {
	if !status.Running {
		return firstNonEmpty(status.LastError, "stopped")
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
