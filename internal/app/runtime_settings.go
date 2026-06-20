package app

import (
	"sync"

	proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"
	"github.com/byte-v-forge/proxy-runtime/internal/ipfraud"
	"github.com/byte-v-forge/proxy-runtime/internal/ipgeo"
	providerregistry "github.com/byte-v-forge/proxy-runtime/internal/provider/registry"
	"github.com/byte-v-forge/proxy-runtime/internal/secretref"
)

type runtimeSettingsFile = proxyruntimev1.ProxyRuntimePersistentSettings

type runtimeSettingsStore struct {
	store            runtimeSettingsPersistence
	secretWriter     secretref.Writer
	accountProviders *providerregistry.Registry
	ipFraudProviders *ipfraud.Registry
	ipGeoProviders   *ipgeo.Registry
	mu               sync.Mutex
}

func newRuntimeSettingsStore(stores *RuntimeStores, accountProviders *providerregistry.Registry, ipFraudProviders *ipfraud.Registry, ipGeoProviders *ipgeo.Registry) *runtimeSettingsStore {
	var store runtimeSettingsPersistence
	var secretWriter secretref.Writer
	if stores != nil {
		store = stores.runtimeSettingsPersistence
		secretWriter = stores.secretStore
	}
	return &runtimeSettingsStore{store: store, secretWriter: secretWriter, accountProviders: accountProviders, ipFraudProviders: ipFraudProviders, ipGeoProviders: ipGeoProviders}
}
