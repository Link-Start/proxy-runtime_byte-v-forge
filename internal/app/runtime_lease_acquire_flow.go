package app

import (
	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
	leaseapp "github.com/byte-v-forge/proxy-runtime/internal/app/lease"
	"github.com/byte-v-forge/proxy-runtime/internal/provider"
)

type acquiredLeaseFlow struct {
	advertisedHost    string
	request           *proxyruntimev1.AcquireProxyLeaseRequest
	settings          *runtimeSettingsFile
	selection         dynamicIPSelection
	providerAccountID string
	leaseID           string
	concurrencyHolder string
	providerClient    leaseapp.SessionProvider
	session           *proxyruntimev1.ProxySession
	nodes             []provider.Node
	dialerProxy       string
	lineLabels        map[string]string
	failure           *leaseapp.FailedAcquireRecorder
}
