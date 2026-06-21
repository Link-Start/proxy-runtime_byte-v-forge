package mihomo

import "github.com/byte-v-forge/proxy-gateway/internal/dataplane"

func (d *Driver) Stop() {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.stopLocked()
}

func (d *Driver) Status() dataplane.Status {
	d.mu.Lock()
	defer d.mu.Unlock()
	return dataplane.Status{
		Running:           d.running,
		ConfigPath:        d.configPath,
		DesiredConfigHash: d.desiredSig,
		AppliedConfigHash: d.signature,
		LastError:         d.lastError,
	}
}
