package app

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"strings"

	proxyruntimev1 "github.com/byte-v-forge/common-lib/gen/go/byte/v/forge/contracts/proxyruntime/v1"
	"github.com/byte-v-forge/proxy-runtime/internal/provider"
	"github.com/byte-v-forge/proxy-runtime/internal/runtimehttp"
)

func (r *Runtime) refreshDynamicProfileExitIPs(ctx context.Context) {
	settings, err := r.settings.load(ctx)
	if err != nil {
		return
	}
	nodes, _ := r.dynamicProfilePoolSnapshot()
	if len(nodes) == 0 {
		return
	}
	for _, profile := range settings.GetEgressProfiles() {
		r.applyDynamicProfileExitIP(ctx, nodes, profile, settings)
	}
	r.setDynamicProfilePoolSnapshot(nodes)
}

func (r *Runtime) applyDynamicProfileExitIP(ctx context.Context, nodes []provider.Node, profile *proxyruntimev1.EgressProfileSettings, settings *runtimeSettingsFile) {
	if !profile.GetEnabled() || profile.GetExit().GetKind() != proxyruntimev1.EgressProfileExitKind_EGRESS_PROFILE_EXIT_KIND_DYNAMIC_IP {
		return
	}
	clearDynamicProfileExitIP(nodes, profile.GetProfileId())
	groupName := strings.TrimSpace(profile.GetDisplayName())
	selected, err := r.mihomoProxyGroupSelected(ctx, groupName)
	if err != nil || selected == "" {
		return
	}
	ip := r.probeProfileExitIP(ctx, profile.GetProfileId(), settings)
	markDynamicProfileSelection(nodes, profile.GetProfileId(), selected, ip)
}

func clearDynamicProfileExitIP(nodes []provider.Node, profileID string) {
	profileID = runtimeSafeID(profileID)
	for index := range nodes {
		if strings.TrimSpace(nodes[index].Labels["egress_profile_id"]) != profileID {
			continue
		}
		nodes[index].Labels = cloneLabels(nodes[index].Labels)
		delete(nodes[index].Labels, "exit_ip")
		delete(nodes[index].Labels, "selected")
		delete(nodes[index].Labels, "mihomo_proxy_name")
	}
}

func markDynamicProfileSelection(nodes []provider.Node, profileID string, selectedProxy string, exitIP string) {
	index := dynamicProfileNodeIndex(nodes, profileID, selectedProxy)
	if index < 0 {
		return
	}
	nodes[index].Labels = cloneLabels(nodes[index].Labels)
	nodes[index].Labels["selected"] = "true"
	nodes[index].Labels["mihomo_proxy_name"] = strings.TrimSpace(selectedProxy)
	if strings.TrimSpace(exitIP) != "" {
		nodes[index].Labels["exit_ip"] = strings.TrimSpace(exitIP)
	}
}

func dynamicProfileNodeIndex(nodes []provider.Node, profileID string, selectedProxy string) int {
	profileID = runtimeSafeID(profileID)
	for index, node := range nodes {
		if strings.TrimSpace(node.Labels["egress_profile_id"]) != profileID {
			continue
		}
		nodeID := strings.TrimSpace(node.ID)
		if nodeID != "" && strings.Contains(selectedProxy, nodeID) {
			return index
		}
	}
	return -1
}

func (r *Runtime) probeProfileExitIP(ctx context.Context, profileID string, settings *runtimeSettingsFile) string {
	rule := profileIngressRule(settings, profileID)
	if rule == nil {
		return ""
	}
	proxyURL, err := localProfileProxyURL(r.cfg.LocalAddr, r.cfg.LocalProtocol, rule)
	if err != nil {
		return ""
	}
	timeout := proxyExitIPTimeout(settings)
	client, err := runtimehttp.NewWithProxy(timeout, proxyURL, runtimehttp.CommonProxySchemes...)
	if err != nil {
		return ""
	}
	defer client.CloseIdleConnections()
	probeCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	ip, err := r.probeExitIP(probeCtx, client)
	if err != nil {
		return ""
	}
	return ip
}

func profileIngressRule(settings *runtimeSettingsFile, profileID string) *proxyruntimev1.ProxyIngressRuleSettings {
	profileID = runtimeSafeID(profileID)
	for _, rule := range normalizeRuntimeSettings(settings).GetIngressRules() {
		if !rule.GetEnabled() || runtimeSafeID(rule.GetProfileId()) != profileID {
			continue
		}
		if strings.TrimSpace(rule.GetUsername()) == "" {
			continue
		}
		return rule
	}
	return nil
}

func localProfileProxyURL(addr string, protocol string, rule *proxyruntimev1.ProxyIngressRuleSettings) (string, error) {
	hostPort, err := localListenHostPort(addr)
	if err != nil {
		return "", err
	}
	proxyURL := &url.URL{
		Scheme: strings.TrimSpace(protocol),
		Host:   hostPort,
		User:   url.UserPassword(rule.GetUsername(), rule.GetPasswordValue()),
	}
	if proxyURL.Scheme == "" {
		proxyURL.Scheme = "socks5"
	}
	return proxyURL.String(), nil
}

func (r *Runtime) mihomoProxyGroupNow(ctx context.Context, groupName string) (string, error) {
	group, err := r.mihomoProxyGroup(ctx, groupName)
	return group.Now, err
}

func (r *Runtime) mihomoProxyGroupSelected(ctx context.Context, groupName string) (string, error) {
	group, err := r.mihomoProxyGroup(ctx, groupName)
	if err != nil {
		return "", err
	}
	if group.Now != "" {
		return group.Now, nil
	}
	if len(group.All) == 1 {
		return group.All[0], nil
	}
	return "", nil
}

type mihomoProxyGroupState struct {
	Now string   `json:"now"`
	All []string `json:"all"`
}

func (r *Runtime) mihomoProxyGroup(ctx context.Context, groupName string) (mihomoProxyGroupState, error) {
	target, err := mihomoAPIURL(r.cfg.Mihomo.APIAddr)
	if err != nil {
		return mihomoProxyGroupState{}, err
	}
	target.Path = "/proxies/" + url.PathEscape(groupName)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, target.String(), nil)
	if err != nil {
		return mihomoProxyGroupState{}, err
	}
	resp, err := runtimehttp.New(r.cfg.RequestTimeout).Do(req)
	if err != nil {
		return mihomoProxyGroupState{}, err
	}
	defer resp.Body.Close()
	var payload mihomoProxyGroupState
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return mihomoProxyGroupState{}, err
	}
	payload.Now = strings.TrimSpace(payload.Now)
	payload.All = compactStrings(payload.All)
	return payload, nil
}

func compactStrings(values []string) []string {
	out := make([]string, 0, len(values))
	for _, value := range values {
		if value = strings.TrimSpace(value); value != "" {
			out = append(out, value)
		}
	}
	return out
}
