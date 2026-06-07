import type {
  ProxyDynamicIPEndpointSettings,
  ProxyDynamicIPProviderSettings,
} from '~/types/byte/v/forge/contracts/proxyruntime/v1/proxy_runtime'

export function newDynamicProviderForm(providerID = '') {
  return {
    dynamic_provider_id: '',
    provider_id: providerID,
    display_name: '',
    rotating_concurrency_limit: 10,
    sticky_concurrency_limit: 2,
  }
}

export function newEndpointForm(providerID = '') {
  return {
    provider_id: providerID,
    original_endpoint_url: '',
    endpoint_url: '',
  }
}

export function newAccountForm(dynamicProviderID = '', providerID = '') {
  return {
    account_id: '',
    dynamic_provider_id: dynamicProviderID,
    provider_id: providerID,
    display_name: '',
    enabled: true,
    username: '',
    original_password_value: '',
    password_value: '',
  }
}

export function dynamicProviderFromForm(
  form: ReturnType<typeof newDynamicProviderForm>,
  current?: ProxyDynamicIPProviderSettings,
  sharedEndpoints: ProxyDynamicIPEndpointSettings[] = [],
): ProxyDynamicIPProviderSettings {
  const providerID = form.provider_id.trim()
  const dynamicProviderID =
    form.dynamic_provider_id.trim() || generatedDynamicProviderID(providerID)
  return {
    dynamic_provider_id: dynamicProviderID,
    provider_id: providerID,
    display_name: form.display_name.trim() || dynamicProviderID,
    rotating_concurrency_limit: form.rotating_concurrency_limit || 10,
    sticky_concurrency_limit: form.sticky_concurrency_limit || 2,
    endpoints: (
      sharedEndpoints.length > 0 ? sharedEndpoints : current?.endpoints || []
    ).map((endpoint) => ({ ...endpoint })),
  }
}

export function endpointFromForm(
  form: ReturnType<typeof newEndpointForm>,
): ProxyDynamicIPEndpointSettings {
  return {
    endpoint_url: form.endpoint_url.trim(),
  }
}

export function cloneDynamicIPProvider(
  provider: ProxyDynamicIPProviderSettings,
): ProxyDynamicIPProviderSettings {
  return {
    dynamic_provider_id: provider.dynamic_provider_id,
    provider_id: provider.provider_id,
    display_name: provider.display_name,
    rotating_concurrency_limit: provider.rotating_concurrency_limit || 10,
    sticky_concurrency_limit: provider.sticky_concurrency_limit || 2,
    endpoints: (provider.endpoints || []).map((endpoint) => ({ ...endpoint })),
  }
}

export function providerEndpoints(
  providers: ProxyDynamicIPProviderSettings[],
  providerID: string,
) {
  return uniqueEndpoints(
    providers
      .filter((item) => item.provider_id === providerID)
      .flatMap((item) => item.endpoints || []),
  )
}

export function syncDynamicIPProviderEndpoints(
  providers: ProxyDynamicIPProviderSettings[],
) {
  const endpointsByProviderID = new Map<
    string,
    ProxyDynamicIPEndpointSettings[]
  >()
  for (const provider of providers) {
    if (!provider.provider_id) continue
    endpointsByProviderID.set(
      provider.provider_id,
      uniqueEndpoints([
        ...(endpointsByProviderID.get(provider.provider_id) || []),
        ...(provider.endpoints || []),
      ]),
    )
  }
  return providers.map((provider) => ({
    ...provider,
    endpoints: provider.provider_id
      ? [...(endpointsByProviderID.get(provider.provider_id) || [])]
      : (provider.endpoints || []).map((endpoint) => ({ ...endpoint })),
  }))
}

export function upsertProviderEndpoint(
  providers: ProxyDynamicIPProviderSettings[],
  providerID: string,
  endpoint: ProxyDynamicIPEndpointSettings,
  originalEndpointURL = '',
) {
  return syncDynamicIPProviderEndpoints(
    providers.map((provider) => {
      if (provider.provider_id !== providerID) return provider
      const endpoints = [...(provider.endpoints || [])]
      const endpointURL = originalEndpointURL || endpoint.endpoint_url
      const index = endpoints.findIndex(
        (item) => item.endpoint_url === endpointURL,
      )
      if (index >= 0) endpoints[index] = endpoint
      else endpoints.push(endpoint)
      return { ...provider, endpoints: uniqueEndpoints(endpoints) }
    }),
  )
}

export function removeProviderEndpoint(
  providers: ProxyDynamicIPProviderSettings[],
  providerID: string,
  endpointURL: string,
) {
  return syncDynamicIPProviderEndpoints(
    providers.map((provider) => {
      if (provider.provider_id !== providerID) return provider
      return {
        ...provider,
        endpoints: (provider.endpoints || []).filter(
          (endpoint) => endpoint.endpoint_url !== endpointURL,
        ),
      }
    }),
  )
}

export function hasDynamicIPProviderConfig(
  provider: ProxyDynamicIPProviderSettings,
) {
  return !!provider.dynamic_provider_id && !!provider.provider_id
}

function generatedDynamicProviderID(providerID: string) {
  const provider = providerID.trim() || 'dynamic'
  const uuid = globalThis.crypto?.randomUUID?.()
  const entropy = uuid?.slice(0, 8) || Math.random().toString(16).slice(2, 10)
  return `${provider}-${entropy}`
}

function uniqueEndpoints(endpoints: ProxyDynamicIPEndpointSettings[]) {
  const seen = new Set<string>()
  const out: ProxyDynamicIPEndpointSettings[] = []
  for (const endpoint of endpoints) {
    const endpointURL = endpoint.endpoint_url.trim()
    if (!endpointURL || seen.has(endpointURL)) continue
    seen.add(endpointURL)
    out.push({ endpoint_url: endpointURL })
  }
  return out
}
