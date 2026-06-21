package kernel

import (
	"strings"
	"time"

	proxygatewayv1 "github.com/byte-v-forge/proxy-gateway/gen/go/byte/v/forge/contracts/proxygateway/v1"
	"google.golang.org/protobuf/types/known/durationpb"

	"github.com/byte-v-forge/proxy-gateway/internal/app/appcore"
)

const DefaultDynamicIPStickyTTL = 10 * time.Minute

func NormalizeDynamicIPSessionPolicy(input *proxygatewayv1.ProxySessionPolicy) *proxygatewayv1.ProxySessionPolicy {
	policy := &proxygatewayv1.ProxySessionPolicy{
		Mode:         dynamicIPSessionMode(input.GetMode()),
		UpstreamKind: proxygatewayv1.ProxyUpstreamKind_PROXY_UPSTREAM_KIND_DYNAMIC_IP,
		Labels:       appcore.CloneStringMap(input.GetLabels()),
	}
	policy.RotationMode = dynamicIPRotationMode(policy.GetMode())
	policy.Region = strings.TrimSpace(input.GetRegion())
	policy.State = strings.TrimSpace(input.GetState())
	policy.City = strings.TrimSpace(input.GetCity())
	policy.Asn = strings.TrimSpace(input.GetAsn())
	if input.GetStickyTtl() != nil && input.GetStickyTtl().AsDuration() > 0 {
		policy.StickyTtl = appcore.CloneDuration(input.GetStickyTtl())
	} else if policy.GetMode() == proxygatewayv1.ProxySessionMode_PROXY_SESSION_MODE_STICKY {
		policy.StickyTtl = durationpb.New(DefaultDynamicIPStickyTTL)
	}
	if policy.Labels == nil {
		policy.Labels = map[string]string{}
	}
	return policy
}

func dynamicIPSessionMode(mode proxygatewayv1.ProxySessionMode) proxygatewayv1.ProxySessionMode {
	switch mode {
	case proxygatewayv1.ProxySessionMode_PROXY_SESSION_MODE_ROTATING:
		return proxygatewayv1.ProxySessionMode_PROXY_SESSION_MODE_ROTATING
	default:
		return proxygatewayv1.ProxySessionMode_PROXY_SESSION_MODE_STICKY
	}
}

func dynamicIPRotationMode(mode proxygatewayv1.ProxySessionMode) proxygatewayv1.ProxyRotationMode {
	if mode == proxygatewayv1.ProxySessionMode_PROXY_SESSION_MODE_ROTATING {
		return proxygatewayv1.ProxyRotationMode_PROXY_ROTATION_MODE_PER_REQUEST
	}
	return proxygatewayv1.ProxyRotationMode_PROXY_ROTATION_MODE_STICKY_SESSION
}
