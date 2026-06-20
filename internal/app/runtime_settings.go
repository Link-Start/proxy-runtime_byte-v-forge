package app

import proxyruntimev1 "github.com/byte-v-forge/proxy-runtime/gen/go/byte/v/forge/contracts/proxyruntime/v1"

// runtimeSettingsFile aliases the persisted runtime-settings proto for the many
// composition-root sites that pass it around; the settings persistence adapter
// lives in internal/app/settings/adapter/persistence.
type runtimeSettingsFile = proxyruntimev1.ProxyRuntimePersistentSettings
