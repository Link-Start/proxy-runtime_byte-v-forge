package app

import (
	"context"
	"strings"
	"sync"

	proxyruntimev1 "github.com/byte-v-forge/common-lib/gen/go/byte/v/forge/contracts/proxyruntime/v1"
	"google.golang.org/protobuf/proto"
)

type sourceNodeObservation struct {
	mu    sync.RWMutex
	nodes []*proxyruntimev1.ProxySourceNode
	err   string
}

func (r *Runtime) sourceNodeSnapshot(ctx context.Context, sourceID string) ([]*proxyruntimev1.ProxySourceNode, error) {
	sources, err := r.listSources(ctx)
	if err != nil {
		return nil, err
	}
	return r.sourceNodeSnapshotForSources(ctx, sourceID, sources)
}

func (r *Runtime) sourceNodeSnapshotForSources(ctx context.Context, sourceID string, sources []*proxyruntimev1.ProxySourceDescriptor) ([]*proxyruntimev1.ProxySourceNode, error) {
	nodes, err := r.dataPlane.SourceNodes(ctx, "")
	if err == nil {
		r.sourceObservation.recordNodes(nodes)
		return filterDesiredSourceNodes(nodes, sourceID, sources), nil
	}
	r.sourceObservation.recordError(err)
	cached := r.sourceObservation.cachedNodes()
	if len(cached) == 0 {
		return nil, err
	}
	return filterDesiredSourceNodes(cached, sourceID, sources), nil
}

func (r *Runtime) refreshSourceNodeObservation(ctx context.Context) {
	nodes, err := r.dataPlane.SourceNodes(ctx, "")
	if err != nil {
		r.sourceObservation.recordError(err)
		r.logger.Warn("observe proxy source nodes failed", "error", err)
		return
	}
	r.sourceObservation.recordNodes(nodes)
}

func (o *sourceNodeObservation) recordNodes(nodes []*proxyruntimev1.ProxySourceNode) {
	o.mu.Lock()
	defer o.mu.Unlock()
	o.nodes = cloneSourceNodes(nodes)
	o.err = ""
}

func (o *sourceNodeObservation) recordError(err error) {
	if err == nil {
		return
	}
	o.mu.Lock()
	defer o.mu.Unlock()
	o.err = strings.TrimSpace(err.Error())
}

func (o *sourceNodeObservation) cachedNodes() []*proxyruntimev1.ProxySourceNode {
	o.mu.RLock()
	defer o.mu.RUnlock()
	return cloneSourceNodes(o.nodes)
}

func (o *sourceNodeObservation) errorText() string {
	o.mu.RLock()
	defer o.mu.RUnlock()
	return o.err
}

func filterDesiredSourceNodes(nodes []*proxyruntimev1.ProxySourceNode, sourceID string, sources []*proxyruntimev1.ProxySourceDescriptor) []*proxyruntimev1.ProxySourceNode {
	allowed := map[string]struct{}{}
	for _, source := range sources {
		if !source.GetEnabled() {
			continue
		}
		switch source.GetKind() {
		case proxyruntimev1.ProxySourceKind_PROXY_SOURCE_KIND_FIXED_PROXY,
			proxyruntimev1.ProxySourceKind_PROXY_SOURCE_KIND_SUBSCRIPTION:
		default:
			continue
		}
		if id := sourceSafeID(source.GetSourceId()); id != "" {
			allowed[id] = struct{}{}
		}
	}
	filterID := sourceSafeID(sourceID)
	if filterID != "" {
		if _, ok := allowed[filterID]; !ok {
			return nil
		}
	}
	out := make([]*proxyruntimev1.ProxySourceNode, 0, len(nodes))
	for _, node := range nodes {
		if node == nil {
			continue
		}
		id := sourceSafeID(node.GetSourceId())
		if _, ok := allowed[id]; !ok {
			continue
		}
		if filterID != "" && id != filterID {
			continue
		}
		out = append(out, cloneSourceNode(node))
	}
	return out
}

func cloneSourceNodes(nodes []*proxyruntimev1.ProxySourceNode) []*proxyruntimev1.ProxySourceNode {
	if len(nodes) == 0 {
		return nil
	}
	out := make([]*proxyruntimev1.ProxySourceNode, 0, len(nodes))
	for _, node := range nodes {
		if node == nil {
			continue
		}
		out = append(out, cloneSourceNode(node))
	}
	return out
}

func cloneSourceNode(node *proxyruntimev1.ProxySourceNode) *proxyruntimev1.ProxySourceNode {
	if node == nil {
		return nil
	}
	cloned, _ := proto.Clone(node).(*proxyruntimev1.ProxySourceNode)
	return cloned
}
