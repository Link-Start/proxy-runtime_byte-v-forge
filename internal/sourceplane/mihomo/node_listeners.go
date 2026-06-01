package mihomo

import (
	"context"
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"fmt"
	"net"
	"sort"
	"strings"
	"time"

	proxyruntimev1 "github.com/byte-v-forge/common-lib/gen/go/byte/v/forge/contracts/proxyruntime/v1"
	"github.com/byte-v-forge/proxy-runtime/internal/sourceplane"
)

func (d *Driver) nodeListenersLocked(ctx context.Context, endpoint sourceplane.Endpoint, providers []sourceplane.SubscriptionProvider, fixedProxies []sourceplane.FixedProxy) ([]nodeListener, error) {
	host, basePort, err := splitEndpoint(endpoint.Addr)
	if err != nil {
		return nil, err
	}
	bindings := make([]nodeListener, 0, len(fixedProxies))
	for _, item := range fixedProxies {
		sourceID := safeID(item.ID)
		if sourceID == "" {
			continue
		}
		bindings = append(bindings, nodeListener{
			SourceID:       sourceID,
			NodeID:         sourceID,
			DisplayName:    firstNonEmpty(item.DisplayName, sourceID),
			ProxyName:      sourceID,
			ProviderBacked: false,
		})
	}
	if len(providers) > 0 {
		apiAddr := strings.TrimSpace(d.cfg.APIAddr)
		if apiAddr == "" {
			return nil, errors.New("mihomo api address is required")
		}
		nodes, err := fetchSourceNodesWhenReady(ctx, apiAddr, subscriptionIDs(providers))
		if err != nil {
			return nil, err
		}
		for _, node := range nodes {
			if node.GetStatus() != proxyruntimev1.ProxySourceNodeStatus_PROXY_SOURCE_NODE_STATUS_AVAILABLE {
				continue
			}
			proxyName := strings.TrimSpace(node.GetDisplayName())
			if proxyName == "" {
				continue
			}
			bindings = append(bindings, nodeListener{
				SourceID:       node.GetSourceId(),
				NodeID:         node.GetNodeId(),
				DisplayName:    proxyName,
				ProxyName:      proxyName,
				ProviderBacked: true,
			})
		}
	}
	sort.Slice(bindings, func(i, j int) bool {
		if bindings[i].SourceID == bindings[j].SourceID {
			return bindings[i].NodeID < bindings[j].NodeID
		}
		return bindings[i].SourceID < bindings[j].SourceID
	})
	if err := assignNodeListenerPorts(bindings, host, basePort); err != nil {
		return nil, err
	}
	return bindings, nil
}

func fetchSourceNodesWhenReady(ctx context.Context, apiAddr string, allowed map[string]struct{}) ([]*proxyruntimev1.ProxySourceNode, error) {
	deadline := time.Now().Add(12 * time.Second)
	var out []*proxyruntimev1.ProxySourceNode
	var lastErr error
	for {
		nodes, err := fetchSourceNodes(ctx, apiAddr, "", allowed)
		if err == nil && len(nodes) > 0 && sourceNodesHealthObserved(nodes) {
			return nodes, nil
		}
		if err != nil {
			lastErr = err
		} else {
			out = nodes
		}
		if time.Now().After(deadline) {
			if lastErr != nil {
				return nil, lastErr
			}
			return out, nil
		}
		timer := time.NewTimer(250 * time.Millisecond)
		select {
		case <-ctx.Done():
			timer.Stop()
			return nil, ctx.Err()
		case <-timer.C:
		}
	}
}

func sourceNodesHealthObserved(nodes []*proxyruntimev1.ProxySourceNode) bool {
	for _, node := range nodes {
		if node.GetCheckedAt() == nil {
			return false
		}
		switch node.GetStatus() {
		case proxyruntimev1.ProxySourceNodeStatus_PROXY_SOURCE_NODE_STATUS_AVAILABLE,
			proxyruntimev1.ProxySourceNodeStatus_PROXY_SOURCE_NODE_STATUS_UNAVAILABLE:
			continue
		default:
			return false
		}
	}
	return true
}

func assignNodeListenerPorts(bindings []nodeListener, host string, basePort int) error {
	if len(bindings) > nodeListenerPortSpan {
		return fmt.Errorf("mihomo node listeners exceed reserved port span: nodes=%d span=%d", len(bindings), nodeListenerPortSpan)
	}
	used := make(map[int]struct{}, len(bindings))
	for index := range bindings {
		start := int(hashPortSlot(bindings[index].SourceID+"/"+bindings[index].NodeID, uint32(nodeListenerPortSpan)))
		port := 0
		for offset := 0; offset < nodeListenerPortSpan; offset++ {
			candidate := basePort + nodeListenerPortStart + ((start + offset) % nodeListenerPortSpan)
			if candidate <= 0 || candidate > 65535 {
				return fmt.Errorf("mihomo node listener port overflow: base=%d port=%d", basePort, candidate)
			}
			if _, exists := used[candidate]; exists {
				continue
			}
			used[candidate] = struct{}{}
			port = candidate
			break
		}
		if port == 0 {
			return fmt.Errorf("mihomo node listener port allocation failed: nodes=%d span=%d", len(bindings), nodeListenerPortSpan)
		}
		bindings[index].Endpoint = sourceplane.Endpoint{Addr: net.JoinHostPort(host, fmt.Sprintf("%d", port)), Protocol: "socks5"}
	}
	return nil
}

func hashPortSlot(value string, modulo uint32) uint32 {
	if modulo == 0 {
		return 0
	}
	sum := sha256.Sum256([]byte(strings.TrimSpace(value)))
	return binary.BigEndian.Uint32(sum[:4]) % modulo
}
