package store

import (
	"strings"

	proxygatewayv1 "github.com/byte-v-forge/proxy-gateway/gen/go/byte/v/forge/contracts/proxygateway/v1"
	"github.com/byte-v-forge/proxy-gateway/internal/protojsoncodec"
)

func DecodeDynamicLeaseFactJSON(raw string) (*proxygatewayv1.ProxyDynamicLease, error) {
	lease := &proxygatewayv1.ProxyDynamicLease{}
	if strings.TrimSpace(raw) == "" {
		return lease, nil
	}
	if err := protojsoncodec.Unmarshal([]byte(raw), lease); err != nil {
		return nil, err
	}
	return lease, nil
}
