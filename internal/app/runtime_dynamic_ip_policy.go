package app

import (
	"strings"
	"time"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
	"google.golang.org/protobuf/types/known/durationpb"
)

const defaultDynamicIPStickyTTL = 10 * time.Minute

func normalizeDynamicIPSessionPolicy(input *proxyruntimev1.ProxySessionPolicy) *proxyruntimev1.ProxySessionPolicy {
	policy := &proxyruntimev1.ProxySessionPolicy{
		Mode:         dynamicIPSessionMode(input.GetMode()),
		UpstreamKind: proxyruntimev1.ProxyUpstreamKind_PROXY_UPSTREAM_KIND_DYNAMIC_IP,
		Labels:       cloneStringMap(input.GetLabels()),
	}
	policy.RotationMode = dynamicIPRotationMode(policy.GetMode(), input.GetRotationMode())
	policy.Region = strings.TrimSpace(input.GetRegion())
	policy.State = strings.TrimSpace(input.GetState())
	policy.City = strings.TrimSpace(input.GetCity())
	policy.Asn = strings.TrimSpace(input.GetAsn())
	if input.GetStickyTtl() != nil && input.GetStickyTtl().AsDuration() > 0 {
		policy.StickyTtl = cloneDuration(input.GetStickyTtl())
	} else if policy.GetMode() == proxyruntimev1.ProxySessionMode_PROXY_SESSION_MODE_STICKY {
		policy.StickyTtl = durationpb.New(defaultDynamicIPStickyTTL)
	}
	if policy.Labels == nil {
		policy.Labels = map[string]string{}
	}
	return policy
}

func dynamicIPSessionMode(mode proxyruntimev1.ProxySessionMode) proxyruntimev1.ProxySessionMode {
	switch mode {
	case proxyruntimev1.ProxySessionMode_PROXY_SESSION_MODE_ROTATING:
		return proxyruntimev1.ProxySessionMode_PROXY_SESSION_MODE_ROTATING
	default:
		return proxyruntimev1.ProxySessionMode_PROXY_SESSION_MODE_STICKY
	}
}

func dynamicIPRotationMode(mode proxyruntimev1.ProxySessionMode, rotationMode proxyruntimev1.ProxyRotationMode) proxyruntimev1.ProxyRotationMode {
	if mode == proxyruntimev1.ProxySessionMode_PROXY_SESSION_MODE_ROTATING {
		return proxyruntimev1.ProxyRotationMode_PROXY_ROTATION_MODE_PER_REQUEST
	}
	return proxyruntimev1.ProxyRotationMode_PROXY_ROTATION_MODE_STICKY_SESSION
}

func dynamicIPPolicyDurationText(policy *proxyruntimev1.ProxySessionPolicy) string {
	if policy == nil || policy.GetStickyTtl() == nil || policy.GetStickyTtl().AsDuration() <= 0 {
		return ""
	}
	return policy.GetStickyTtl().AsDuration().String()
}
