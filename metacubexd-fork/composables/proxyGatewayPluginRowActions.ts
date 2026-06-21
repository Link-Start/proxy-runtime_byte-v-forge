import type { ProxyIPFraudProviderDescriptor, ProxyIPGeoProviderDescriptor } from '~/types/byte/v/forge/contracts/proxygateway/v1/proxy_gateway'
import type { ProxyGatewayIPFraudProviderRow } from '~/composables/proxyGatewayIPFraudProviderRows'
import type { ProxyGatewayIPGeoProviderRow } from '~/composables/proxyGatewayIPGeoProviderRows'
import { proxyGatewayUniqueIPFraudProviderID } from '~/composables/proxyGatewayIPFraudProviderRows'
import { proxyGatewayUniqueIPGeoProviderID } from '~/composables/proxyGatewayIPGeoProviderRows'

export function applyIPFraudDescriptor(
  row: ProxyGatewayIPFraudProviderRow,
  descriptor: ProxyIPFraudProviderDescriptor,
  rows: ProxyGatewayIPFraudProviderRow[],
) {
  row.provider_id = proxyGatewayUniqueIPFraudProviderID(descriptor.provider_id, rows, row.provider_id)
  applyProviderDescriptor(row, descriptor)
}

export function applyIPGeoDescriptor(
  row: ProxyGatewayIPGeoProviderRow,
  descriptor: ProxyIPGeoProviderDescriptor,
  rows: ProxyGatewayIPGeoProviderRow[],
) {
  row.provider_id = proxyGatewayUniqueIPGeoProviderID(descriptor.provider_id, rows, row.provider_id)
  applyProviderDescriptor(row, descriptor)
}

function applyProviderDescriptor(
  row: ProxyGatewayIPFraudProviderRow | ProxyGatewayIPGeoProviderRow,
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
