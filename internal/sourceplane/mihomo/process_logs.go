package mihomo

import (
	"io"
	"log/slog"
	"regexp"
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

type processLogWriter struct {
	mu      sync.Mutex
	ring    *processLogRing
	stream  string
	pending string
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

var processLogRedactors = []*regexp.Regexp{
	regexp.MustCompile(`(?i)(authorization:\s*bearer\s+)[^\s]+`),
	regexp.MustCompile(`(?i)([?&](?:token|secret|password|passwd|api_key|apikey)=)[^&\s]+`),
	regexp.MustCompile(`(?i)\b((?:https?|socks5h?)://)([^/\s:@]+):([^@\s/]+)@`),
}

func redactProcessLogLine(line string) string {
	line = processLogRedactors[0].ReplaceAllString(line, `${1}<redacted>`)
	line = processLogRedactors[1].ReplaceAllString(line, `${1}<redacted>`)
	line = processLogRedactors[2].ReplaceAllString(line, `${1}<redacted>:<redacted>@`)
	return line
}
