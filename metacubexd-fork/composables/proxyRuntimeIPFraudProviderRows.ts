import type {
  ProxyIPFraudProviderDescriptor,
  ProxyIPFraudProviderKind,
  ProxyRuntimeSettings,
} from '~/types/byte/v/forge/contracts/proxyruntime/v1/proxy_runtime'

export type ProxyRuntimeIPFraudProviderRow =
  ReturnType<typeof proxyRuntimeIPFraudEmptyRow>

export function proxyRuntimeIPFraudRowFromSettings(
  item: NonNullable<ProxyRuntimeSettings['ip_fraud_providers']>[number],
) {
  return {
    ...proxyRuntimeIPFraudEmptyRow(),
    provider_id: item.provider_id,
    display_name: item.display_name || '',
    kind: item.kind,
    weight: item.weight || 100,
    anonymous: item.anonymous,
    api_key_configured: !!item.api_key_configured,
    api_key_count: item.api_key_count || 0,
  }
}

export function proxyRuntimeIPFraudRowFromDescriptor(
  descriptor: ProxyIPFraudProviderDescriptor,
) {
  return {
    ...proxyRuntimeIPFraudEmptyRow(),
    provider_id: descriptor.provider_id,
    display_name: descriptor.display_name || descriptor.provider_id,
    kind: descriptor.kind,
    weight: descriptor.default_weight || 100,
    anonymous: descriptor.supports_anonymous && !descriptor.supports_api_key,
  }
}

export function proxyRuntimeUniqueIPFraudProviderID(
  base: string,
  rows: ProxyRuntimeIPFraudProviderRow[],
  currentID = '',
) {
  const prefix = base || 'provider'
  const used = new Set(rows.map((row) => row.provider_id).filter((id) => id !== currentID))
  if (!used.has(prefix)) return prefix
  for (let index = 2; ; index += 1) {
    const candidate = `${prefix}-${index}`
    if (!used.has(candidate)) return candidate
  }
}

function proxyRuntimeIPFraudEmptyRow() {
  return {
    provider_id: '',
    display_name: '',
    kind: 'PROXY_IP_FRAUD_PROVIDER_KIND_UNSPECIFIED' as ProxyIPFraudProviderKind,
    weight: 100,
    anonymous: false,
    api_key_value: '',
    api_key_configured: false,
    api_key_count: 0,
    clear_api_keys: false,
  }
}
