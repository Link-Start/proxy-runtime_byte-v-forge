package app

import (
	"errors"
	"strings"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
	leaseapp "github.com/byte-v-forge/proxy-runtime/internal/app/lease"
)

func leaseProfilePolicyError(err error) error {
	switch {
	case errors.Is(err, leaseapp.ErrProfileDynamicIPNotConfigured), errors.Is(err, leaseapp.ErrProfileLeaseRequiresSticky):
		return failedPrecondition(err.Error(), err)
	case errors.Is(err, leaseapp.ErrRequestRequiresStickyDynamicIP):
		return invalidArgument(err.Error(), err)
	default:
		return err
	}
}

func playgroundLeaseNeedsReplacement(req *proxyruntimev1.AcquireProxyLeaseRequest, lease *proxyruntimev1.ProxyDynamicLease) bool {
	if req.GetAccountId() != playgroundProfileID {
		return false
	}
	return strings.TrimSpace(lease.GetListener().GetLabels()[leaseapp.LabelProxyUsername]) != playgroundUsername
}
