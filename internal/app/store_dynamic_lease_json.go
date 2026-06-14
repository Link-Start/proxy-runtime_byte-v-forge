package app

import (
	"strings"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
	"github.com/byte-v-forge/proxy-runtime/internal/protojsoncodec"
)

func decodeDynamicLeaseFactJSON(raw string) (*proxyruntimev1.ProxyDynamicLease, error) {
	lease := &proxyruntimev1.ProxyDynamicLease{}
	if strings.TrimSpace(raw) == "" {
		return lease, nil
	}
	if err := protojsoncodec.Unmarshal([]byte(raw), lease); err != nil {
		return nil, err
	}
	return lease, nil
}
