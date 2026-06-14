package lease

import proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"

type AcquiredEndpointInput struct {
	Listener          Listener
	Egress            *proxyruntimev1.ProxyEndpoint
	Request           *proxyruntimev1.AcquireProxyLeaseRequest
	SelectionPlan     *proxyruntimev1.ProxyDynamicIPSelectionPlan
	ProviderClient    SessionProvider
	ProviderAccountID string
	ConcurrencyHolder string
	Session           *proxyruntimev1.ProxySession
	LineLabels        map[string]string
	Managed           bool
	FallbackProtocol  string
}

type AcquiredEndpoint struct {
	Listener      Listener
	ListenerProto *proxyruntimev1.EgressListener
	Egress        *proxyruntimev1.ProxyEndpoint
}

func NewAcquiredEndpoint(input AcquiredEndpointInput) AcquiredEndpoint {
	listenerProto := EgressListenerProto(input.Listener, input.Managed, input.FallbackProtocol)
	ApplyAcquiredEndpointMetadata(input.Egress, AcquiredEndpointMetadataInput{
		Request:           input.Request,
		SelectionPlan:     input.SelectionPlan,
		ProviderClient:    input.ProviderClient,
		ProviderAccountID: input.ProviderAccountID,
		ConcurrencyHolder: input.ConcurrencyHolder,
		Session:           input.Session,
		LineLabels:        input.LineLabels,
	})
	return AcquiredEndpoint{Listener: input.Listener, ListenerProto: listenerProto, Egress: input.Egress}
}
