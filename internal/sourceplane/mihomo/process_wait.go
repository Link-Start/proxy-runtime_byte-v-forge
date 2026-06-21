package mihomo

import (
	"strings"
	"time"

	"github.com/byte-v-forge/proxy-gateway/internal/processruntime"
)

func (d *Driver) wait(process *processruntime.Process) {
	<-process.Done()
	err := process.ExitError()
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.process != process {
		return
	}
	d.running = false
	if err != nil {
		d.lastError = processExitMessage(err, d.processLogs)
	}
}

func processExitMessage(err error, logs *processLogRing) string {
	message := strings.TrimSpace(err.Error())
	if logs == nil {
		return message
	}
	if tail := logs.Tail(); tail != "" {
		return message + ": " + tail
	}
	return message
}

func (d *Driver) stopLocked() {
	process := d.process
	d.process = nil
	d.running = false
	if process != nil {
		_ = process.Stop(5 * time.Second)
	}
}
