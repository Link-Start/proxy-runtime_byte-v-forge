package app

import (
	"sync"
	"time"

	"github.com/byte-v-forge/proxy-runtime/internal/provider"
)

type poolSnapshotState struct {
	mu          sync.RWMutex
	nodes       []provider.Node
	refreshedAt time.Time
}

func (s *poolSnapshotState) record(nodes []provider.Node) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.nodes = cloneNodes(nodes)
	s.refreshedAt = time.Now().UTC()
}

func (s *poolSnapshotState) current() ([]provider.Node, time.Time) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return cloneNodes(s.nodes), s.refreshedAt
}

func (s *poolSnapshotState) currentNodes() []provider.Node {
	nodes, _ := s.current()
	return nodes
}
