package proxycheck

import (
	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"

	"github.com/byte-v-forge/proxy-runtime/internal/app/appcore"
)

type ExitGeo struct {
	IP          string
	CountryCode string
	Region      string
	City        string
}

func ExitGeoFromProto(ip string, geo *proxyruntimev1.ProxyExitGeo) ExitGeo {
	return ExitGeo{
		IP:          appcore.FirstNonEmpty(geo.GetIp(), ip),
		CountryCode: geo.GetCountryCode(),
		Region:      geo.GetRegion(),
		City:        geo.GetCity(),
	}
}
