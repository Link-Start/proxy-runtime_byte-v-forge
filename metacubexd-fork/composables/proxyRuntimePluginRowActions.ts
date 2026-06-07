import type { ProxyIPFraudProviderDescriptor, ProxyIPGeoProviderDescriptor } from '~/types/byte/v/forge/contracts/proxyruntime/v1/proxy_runtime'
import type { ProxyRuntimeIPFraudProviderRow } from '~/composables/proxyRuntimeIPFraudProviderRows'
import type { ProxyRuntimeIPGeoProviderRow } from '~/composables/proxyRuntimeIPGeoProviderRows'
import { proxyRuntimeUniqueIPFraudProviderID } from '~/composables/proxyRuntimeIPFraudProviderRows'
import { proxyRuntimeUniqueIPGeoProviderID } from '~/composables/proxyRuntimeIPGeoProviderRows'

export function applyIPFraudDescriptor(
  row: ProxyRuntimeIPFraudProviderRow,
  descriptor: ProxyIPFraudProviderDescriptor,
  rows: ProxyRuntimeIPFraudProviderRow[],
) {
  row.provider_id = proxyRuntimeUniqueIPFraudProviderID(descriptor.provider_id, rows, row.provider_id)
  applyProviderDescriptor(row, descriptor)
}

export function applyIPGeoDescriptor(
  row: ProxyRuntimeIPGeoProviderRow,
  descriptor: ProxyIPGeoProviderDescriptor,
  rows: ProxyRuntimeIPGeoProviderRow[],
) {
  row.provider_id = proxyRuntimeUniqueIPGeoProviderID(descriptor.provider_id, rows, row.provider_id)
  applyProviderDescriptor(row, descriptor)
}

function applyProviderDescriptor(
  row: ProxyRuntimeIPFraudProviderRow | ProxyRuntimeIPGeoProviderRow,
  descriptor: ProxyIPFraudProviderDescriptor | ProxyIPGeoProviderDescriptor,
) {
  row.display_name = descriptor.display_name || descriptor.provider_id
  row.weight = row.weight || descriptor.default_weight || 100
  row.anonymous = descriptor.supports_anonymous && !descriptor.supports_api_key
  row.api_key_value = ''
  row.api_key_configured = false
  row.api_key_count = 0
  row.clear_api_keys = false
}
