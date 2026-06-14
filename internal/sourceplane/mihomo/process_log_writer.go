package mihomo

import (
	"strings"
	"sync"
)

type processLogWriter struct {
	mu      sync.Mutex
	ring    *processLogRing
	stream  string
	pending string
}

func (w *processLogWriter) Write(data []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	text := w.pending + string(data)
	parts := strings.Split(text, "\n")
	w.pending = parts[len(parts)-1]
	for _, line := range parts[:len(parts)-1] {
		w.ring.Append(w.stream, line)
	}
	return len(data), nil
}
