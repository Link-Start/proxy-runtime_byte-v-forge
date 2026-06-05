import type {
  ProxyDynamicIPGatewaySettings,
  ProxyDynamicIPProviderSettings,
} from '~/types/byte/v/forge/contracts/proxyruntime/v1/proxy_runtime'

export function newEndpointForm(providerID = '') {
  return {
    provider_id: providerID,
    original_endpoint_url: '',
    endpoint_url: '',
  }
}

export function newAccountForm(providerID = '') {
  return {
    account_id: '',
    provider_id: providerID,
    display_name: '',
    enabled: true,
    username: '',
    original_password_value: '',
    password_value: '',
  }
}

export function endpointFromForm(
  form: ReturnType<typeof newEndpointForm>,
): ProxyDynamicIPGatewaySettings {
  return {
    endpoint_url: form.endpoint_url.trim(),
  }
}

export function cloneDynamicIPProvider(
  provider: ProxyDynamicIPProviderSettings,
): ProxyDynamicIPProviderSettings {
  return {
    provider_id: provider.provider_id,
    gateways: provider.gateways.map((gateway) => ({ ...gateway })),
  }
}

export function upsertDynamicIPProvider(
  providers: ProxyDynamicIPProviderSettings[],
  providerID: string,
) {
  let provider = providers.find((item) => item.provider_id === providerID)
  if (!provider) {
    provider = { provider_id: providerID, gateways: [] }
    providers.push(provider)
  }
  return provider
}

export function hasDynamicIPProviderConfig(
  provider: ProxyDynamicIPProviderSettings,
) {
  return provider.gateways.length > 0
}
