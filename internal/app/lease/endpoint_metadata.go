package lease

import (
	"strings"

	proxygatewayv1 "github.com/byte-v-forge/proxy-gateway/gen/go/byte/v/forge/contracts/proxygateway/v1"
)

type DynamicEndpointMetadataInput struct {
	Request           *proxygatewayv1.AcquireProxyLeaseRequest
	SelectionPlan     *proxygatewayv1.ProxyDynamicIPSelectionPlan
	ProviderID        string
	ProviderAccountID string
	ConcurrencyHolder string
	SessionID         string
	LineLabels        map[string]string
}

type AcquiredEndpointMetadataInput struct {
	Request           *proxygatewayv1.AcquireProxyLeaseRequest
	SelectionPlan     *proxygatewayv1.ProxyDynamicIPSelectionPlan
	ProviderClient    SessionProvider
	ProviderAccountID string
	ConcurrencyHolder string
	Session           *proxygatewayv1.ProxySession
	LineLabels        map[string]string
}

func ApplyAcquiredEndpointMetadata(endpoint *proxygatewayv1.ProxyEndpoint, input AcquiredEndpointMetadataInput) {
	ApplyDynamicEndpointMetadata(endpoint, DynamicEndpointMetadataInput{
		Request:           input.Request,
		SelectionPlan:     input.SelectionPlan,
		ProviderID:        ProviderName(input.ProviderClient),
		ProviderAccountID: input.ProviderAccountID,
		ConcurrencyHolder: input.ConcurrencyHolder,
		SessionID:         input.Session.GetSessionId(),
		LineLabels:        input.LineLabels,
	})
}

func ApplyDynamicEndpointMetadata(endpoint *proxygatewayv1.ProxyEndpoint, input DynamicEndpointMetadataInput) {
	if endpoint == nil {
		return
	}
	if endpoint.Labels == nil {
		endpoint.Labels = map[string]string{}
	}
	endpoint.ProviderId = input.ProviderID
	endpoint.UpstreamKind = proxygatewayv1.ProxyUpstreamKind_PROXY_UPSTREAM_KIND_DYNAMIC_IP
	endpoint.RotationMode = input.Request.GetPolicy().GetRotationMode()
	endpoint.SessionId = input.SessionID
	endpoint.Labels[LabelAccountID] = input.Request.GetAccountId()
	endpoint.Labels[LabelPurpose] = input.Request.GetPurpose()
	endpoint.Labels[LabelProviderAccountID] = input.ProviderAccountID
	endpoint.Labels[LabelSelectionID] = input.SelectionPlan.GetSelectionId()
	endpoint.Labels[LabelDynamicProviderID] = input.SelectionPlan.GetSelectedEndpoint().GetDynamicProviderId()
	endpoint.Labels[LabelDynamicIPEndpointID] = input.SelectionPlan.GetSelectedEndpoint().GetEndpointId()
	endpoint.Labels[LabelProviderAccountConcurrencyHolder] = input.ConcurrencyHolder
	applyDynamicEndpointLocationLabels(endpoint.Labels, input.Request.GetPolicy(), input.SelectionPlan.GetPolicy())
	for key, value := range input.LineLabels {
		endpoint.Labels[key] = value
	}
}

func applyDynamicEndpointLocationLabels(labels map[string]string, requestPolicy *proxygatewayv1.ProxySessionPolicy, selectedPolicy *proxygatewayv1.ProxyDynamicIPSelectionPolicy) {
	if countryCode := strings.TrimSpace(selectedPolicy.GetCountryCode()); countryCode != "" {
		labels["country_code"] = countryCode
	}
	if region := dynamicEndpointRegion(requestPolicy, selectedPolicy); region != "" {
		labels["region"] = region
	}
	if state := strings.TrimSpace(requestPolicy.GetState()); state != "" {
		labels["state"] = state
	}
	if city := strings.TrimSpace(requestPolicy.GetCity()); city != "" {
		labels["city"] = city
	}
	if asn := strings.TrimSpace(requestPolicy.GetAsn()); asn != "" {
		labels["asn"] = asn
	}
}

func dynamicEndpointRegion(requestPolicy *proxygatewayv1.ProxySessionPolicy, selectedPolicy *proxygatewayv1.ProxyDynamicIPSelectionPolicy) string {
	if region := strings.TrimSpace(requestPolicy.GetRegion()); region != "" {
		return region
	}
	return strings.TrimSpace(selectedPolicy.GetRegion())
}
