package mihomo

import (
	"context"
	"fmt"
	"net"
	"time"
)

func waitForEndpoint(ctx context.Context, addr string, timeout time.Duration) error {
	dialAddr, err := endpointDialAddr(addr)
	if err != nil {
		return err
	}
	waitCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	ticker := time.NewTicker(50 * time.Millisecond)
	defer ticker.Stop()
	for {
		conn, err := (&net.Dialer{Timeout: 100 * time.Millisecond}).DialContext(waitCtx, "tcp", dialAddr)
		if err == nil {
			_ = conn.Close()
			return nil
		}
		select {
		case <-waitCtx.Done():
			return fmt.Errorf("mihomo mixed listener %s is not ready: %w", addr, waitCtx.Err())
		case <-ticker.C:
		}
	}
}
