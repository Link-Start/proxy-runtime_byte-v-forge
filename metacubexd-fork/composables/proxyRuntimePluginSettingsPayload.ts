import type { ProxyRuntimeIPFraudProviderRow } from '~/composables/proxyRuntimeIPFraudProviderRows'
import type { ProxyRuntimeIPGeoProviderRow } from '~/composables/proxyRuntimeIPGeoProviderRows'

export function proxyRuntimeIPFraudSettingsFromRows(rows: ProxyRuntimeIPFraudProviderRow[]) {
  return rows.map((row) => ({
    provider_id: row.provider_id,
    display_name: row.display_name.trim(),
    kind: row.kind,
    weight: row.weight || 100,
    anonymous: row.anonymous,
    api_key_values: row.api_key_value.trim() ? [row.api_key_value.trim()] : [],
    api_key_secret_refs: [],
    clear_api_keys: row.clear_api_keys,
  }))
}

export function proxyRuntimeIPGeoSettingsFromRows(rows: ProxyRuntimeIPGeoProviderRow[]) {
  return rows.map((row) => ({
    provider_id: row.provider_id,
    display_name: row.display_name.trim(),
    kind: row.kind,
    weight: row.weight || 100,
    anonymous: row.anonymous,
    api_key_values: row.api_key_value.trim() ? [row.api_key_value.trim()] : [],
    api_key_secret_refs: [],
    clear_api_keys: row.clear_api_keys,
  }))
}
