package mihomo

import (
	"io"
	"log/slog"
	"strings"
	"sync"
)

const defaultProcessLogRingLimit = 40

type processLogRing struct {
	mu     sync.Mutex
	logger *slog.Logger
	limit  int
	lines  []string
}

func newProcessLogRing(logger *slog.Logger, limit int) *processLogRing {
	if logger == nil {
		logger = slog.Default()
	}
	if limit <= 0 {
		limit = defaultProcessLogRingLimit
	}
	return &processLogRing{logger: logger, limit: limit}
}

func (r *processLogRing) Writer(stream string) io.Writer {
	return &processLogWriter{ring: r, stream: strings.TrimSpace(stream)}
}

func (r *processLogRing) Append(stream string, line string) {
	if r == nil {
		return
	}
	line = redactProcessLogLine(strings.TrimSpace(line))
	if line == "" {
		return
	}
	stream = strings.TrimSpace(stream)
	if stream == "" {
		stream = "process"
	}
	entry := stream + ": " + line
	r.mu.Lock()
	r.lines = append(r.lines, entry)
	if len(r.lines) > r.limit {
		r.lines = append([]string(nil), r.lines[len(r.lines)-r.limit:]...)
	}
	r.mu.Unlock()
	r.logger.Debug("mihomo process output", "stream", stream, "line", line)
}

func (r *processLogRing) Tail() string {
	if r == nil {
		return ""
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if len(r.lines) == 0 {
		return ""
	}
	start := len(r.lines) - 3
	if start < 0 {
		start = 0
	}
	return strings.Join(r.lines[start:], " | ")
}
