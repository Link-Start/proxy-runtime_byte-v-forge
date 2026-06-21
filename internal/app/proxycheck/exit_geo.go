package proxycheck

import (
	proxygatewayv1 "github.com/byte-v-forge/proxy-gateway/gen/go/byte/v/forge/contracts/proxygateway/v1"

	"github.com/byte-v-forge/proxy-gateway/internal/app/appcore"
)

type ExitGeo struct {
	IP          string
	CountryCode string
	Region      string
	City        string
}

func ExitGeoFromProto(ip string, geo *proxygatewayv1.ProxyExitGeo) ExitGeo {
	return ExitGeo{
		IP:          appcore.FirstNonEmpty(geo.GetIp(), ip),
		CountryCode: geo.GetCountryCode(),
		Region:      geo.GetRegion(),
		City:        geo.GetCity(),
	}
}
