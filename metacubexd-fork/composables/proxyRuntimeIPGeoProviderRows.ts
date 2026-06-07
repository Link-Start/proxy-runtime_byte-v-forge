import type {
  ProxyIPGeoProviderDescriptor,
  ProxyIPGeoProviderKind,
  ProxyRuntimeSettings,
} from '~/types/byte/v/forge/contracts/proxyruntime/v1/proxy_runtime'

export type ProxyRuntimeIPGeoProviderRow = ReturnType<typeof proxyRuntimeIPGeoEmptyRow>

export function proxyRuntimeIPGeoRowFromSettings(
  item: NonNullable<ProxyRuntimeSettings['ip_geo_providers']>[number],
) {
  return {
    ...proxyRuntimeIPGeoEmptyRow(),
    provider_id: item.provider_id,
    display_name: item.display_name || '',
    kind: item.kind,
    weight: item.weight || 100,
    anonymous: item.anonymous,
    api_key_configured: !!item.api_key_configured,
    api_key_count: item.api_key_count || 0,
  }
}

export function proxyRuntimeIPGeoRowFromDescriptor(
  descriptor: ProxyIPGeoProviderDescriptor,
) {
  return {
    ...proxyRuntimeIPGeoEmptyRow(),
    provider_id: descriptor.provider_id,
    display_name: descriptor.display_name || descriptor.provider_id,
    kind: descriptor.kind,
    weight: descriptor.default_weight || 100,
    anonymous: descriptor.supports_anonymous && !descriptor.supports_api_key,
  }
}

export function proxyRuntimeUniqueIPGeoProviderID(
  base: string,
  rows: ProxyRuntimeIPGeoProviderRow[],
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

function proxyRuntimeIPGeoEmptyRow() {
  return {
    provider_id: '',
    display_name: '',
    kind: 'PROXY_IP_GEO_PROVIDER_KIND_UNSPECIFIED' as ProxyIPGeoProviderKind,
    weight: 100,
    anonymous: false,
    api_key_value: '',
    api_key_configured: false,
    api_key_count: 0,
    clear_api_keys: false,
  }
}
