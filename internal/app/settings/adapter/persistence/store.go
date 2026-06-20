package persistence

import (
	"sync"

	"github.com/byte-v-forge/proxy-runtime/internal/ipfraud"
	"github.com/byte-v-forge/proxy-runtime/internal/ipgeo"
	providerregistry "github.com/byte-v-forge/proxy-runtime/internal/provider/registry"
	"github.com/byte-v-forge/proxy-runtime/internal/secretref"

	"github.com/byte-v-forge/proxy-runtime/internal/app/store"
)

// Store is the runtime-settings persistence adapter: it implements the
// settings.Repository outbound port over the control-plane store and secret
// writer, serialising mutations behind mu.
type Store struct {
	store            store.RuntimeSettingsPersistence
	secretWriter     secretref.Writer
	accountProviders *providerregistry.Registry
	ipFraudProviders *ipfraud.Registry
	ipGeoProviders   *ipgeo.Registry
	mu               sync.Mutex
}

func NewStore(stores *store.RuntimeStores, accountProviders *providerregistry.Registry, ipFraudProviders *ipfraud.Registry, ipGeoProviders *ipgeo.Registry) *Store {
	var persistence store.RuntimeSettingsPersistence
	var secretWriter secretref.Writer
	if stores != nil {
		persistence = stores.RuntimeSettingsPersistence
		secretWriter = stores.SecretStore
	}
	return &Store{store: persistence, secretWriter: secretWriter, accountProviders: accountProviders, ipFraudProviders: ipFraudProviders, ipGeoProviders: ipGeoProviders}
}
