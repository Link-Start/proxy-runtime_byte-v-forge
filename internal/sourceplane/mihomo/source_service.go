package mihomo

import (
	"context"
	"errors"
	"sort"
	"strings"

	proxyruntimev1 "github.com/byte-v-forge/common-lib/gen/go/byte/v/forge/contracts/proxyruntime/v1"
	"github.com/byte-v-forge/proxy-runtime/internal/sourceplane"
)

func (d *Driver) SourceNodes(ctx context.Context, sourceID string) ([]*proxyruntimev1.ProxySourceNode, error) {
	d.mu.Lock()
	running := d.running
	apiAddr := strings.TrimSpace(d.cfg.APIAddr)
	file := cloneSourceFile(d.sources)
	d.mu.Unlock()

	fixedNodes := fixedProxySourceNodes(file.FixedProxies, sourceID)
	allowed := subscriptionIDs(file.Subscriptions)
	if len(allowed) == 0 {
		return fixedNodes, nil
	}
	if id := safeID(sourceID); id != "" {
		if _, exists := allowed[id]; !exists && !fixedSourceExists(file.FixedProxies, id) {
			return nil, nil
		}
		if fixedSourceExists(file.FixedProxies, id) {
			return fixedNodes, nil
		}
	}
	if !running {
		return nil, errors.New("mihomo source runtime is not running")
	}
	if apiAddr == "" {
		return nil, errors.New("mihomo api address is required")
	}
	nodes, err := fetchSourceNodes(ctx, d.apiClient, apiAddr, sourceID, allowed)
	if err != nil {
		return nil, err
	}
	return append(fixedNodes, nodes...), nil
}

func (d *Driver) ResolveNodePublicIP(ctx context.Context, sourceID string, nodeID string, nodeDisplayName string) (string, error) {
	d.mu.Lock()
	file := cloneSourceFile(d.sources)
	fixed := fixedProxyByID(file.FixedProxies, sourceID)
	if fixed != nil {
		uri := fixed.URI
		d.mu.Unlock()
		return publicIP(ctx, fixedProxyHost(uri))
	}
	providerPaths := d.providerFileCandidatesLocked(file.Subscriptions, sourceID)
	d.mu.Unlock()
	host := providerNodeHost(providerPaths, nodeID, nodeDisplayName)
	return publicIP(ctx, host)
}

func fixedProxySourceNodes(items []sourceplane.FixedProxy, sourceID string) []*proxyruntimev1.ProxySourceNode {
	filterID := safeID(sourceID)
	out := make([]*proxyruntimev1.ProxySourceNode, 0, len(items))
	for _, item := range items {
		id := safeID(item.ID)
		if id == "" || filterID != "" && id != filterID {
			continue
		}
		out = append(out, &proxyruntimev1.ProxySourceNode{
			SourceId:    id,
			NodeId:      id,
			DisplayName: firstNonEmpty(item.DisplayName, id),
			NodeType:    "vless",
			Status:      proxyruntimev1.ProxySourceNodeStatus_PROXY_SOURCE_NODE_STATUS_UNKNOWN,
		})
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].GetSourceId() < out[j].GetSourceId()
	})
	return out
}

func fixedSourceExists(items []sourceplane.FixedProxy, sourceID string) bool {
	sourceID = safeID(sourceID)
	if sourceID == "" {
		return false
	}
	for _, item := range items {
		if safeID(item.ID) == sourceID {
			return true
		}
	}
	return false
}
