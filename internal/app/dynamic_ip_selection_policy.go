package app

import (
	"errors"
	"strconv"
	"strings"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
	"github.com/byte-v-forge/proxy-runtime/internal/geox"
)

func normalizeDynamicIPSelectionPolicy(req *proxyruntimev1.AcquireProxyLeaseRequest) *proxyruntimev1.ProxyDynamicIPSelectionPolicy {
	in := req.GetSelectionPolicy()
	policy := &proxyruntimev1.ProxyDynamicIPSelectionPolicy{}
	if in != nil {
		policy.CountryCode = strings.TrimSpace(in.GetCountryCode())
		policy.Region = strings.TrimSpace(in.GetRegion())
		policy.Purpose = strings.TrimSpace(in.GetPurpose())
		policy.MaxAttempts = in.GetMaxAttempts()
	}
	if policy.CountryCode == "" {
		policy.CountryCode = firstNonEmpty(req.GetPolicy().GetLabels()["country_code"], req.GetPolicy().GetRegion())
	}
	if policy.Region == "" {
		policy.Region = firstNonEmpty(req.GetPolicy().GetLabels()["region"], req.GetPolicy().GetRegion())
	}
	if policy.Purpose == "" {
		policy.Purpose = strings.TrimSpace(req.GetPurpose())
	}
	policy.CountryCode = geox.NormalizeCountryAlpha2(policy.CountryCode)
	policy.Region = strings.ToUpper(strings.TrimSpace(policy.Region))
	if policy.MaxAttempts == 0 {
		policy.MaxAttempts = 10
	}
	return policy
}

func dynamicIPSelectionAttempt(req *proxyruntimev1.AcquireProxyLeaseRequest) int {
	if req == nil || req.GetPolicy() == nil {
		return 1
	}
	value := strings.TrimSpace(req.GetPolicy().GetLabels()["attempt"])
	if value == "" {
		return 1
	}
	attempt, err := strconv.Atoi(value)
	if err != nil || attempt < 1 {
		return 1
	}
	return attempt
}

func dynamicIPSelectionMaxAttempts(policy *proxyruntimev1.ProxyDynamicIPSelectionPolicy) int {
	attempts := int(policy.GetMaxAttempts())
	if attempts < 1 {
		return 1
	}
	return attempts
}

func retryLeaseAcquireAttempt(err error) bool {
	var appErr *appError
	if !errors.As(err, &appErr) {
		return false
	}
	return appErr.code == errCodeUnavailable || appErr.code == errCodeFailedPrecondition
}

func dynamicIPSelectionKey(req *proxyruntimev1.AcquireProxyLeaseRequest) string {
	if req == nil {
		return ""
	}
	labels := req.GetPolicy().GetLabels()
	return firstNonEmpty(
		labels["selection_seed"],
		labels["proxy_selection_seed"],
		labels["job_id"],
		req.GetAccountId(),
		req.GetPurpose(),
	)
}
