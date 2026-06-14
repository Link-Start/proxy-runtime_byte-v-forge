package lease

import proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"

func ApplyAcquireRequestPolicies(req *proxyruntimev1.AcquireProxyLeaseRequest, profiles []*proxyruntimev1.EgressProfileSettings) (*proxyruntimev1.ProxyDynamicIPSelectionPolicy, error) {
	req.Policy = NormalizeDynamicIPSessionPolicy(req.GetPolicy())
	ApplyRequestLabels(req)
	if err := ApplyProfileDynamicIPPolicy(profiles, req); err != nil {
		return nil, err
	}
	return NormalizeDynamicIPSelectionPolicy(req), nil
}
