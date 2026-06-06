import type {
  DeleteProxyProviderAccountRequest,
  DeleteProxyProviderAccountResponse,
  EgressProfileSettings,
  GetProxyRuntimeSettingsResponse,
  ListProxyDynamicLeasesResponse,
  ListProxyProviderAccountsResponse,
  ListProxyProvidersResponse,
  ProxyDynamicIPProviderSettings,
  ProxyIngressRuleSettings,
  UpdateProxyRuntimeSettingsResponse,
  UpsertProxyProviderAccountRequest,
  UpsertProxyProviderAccountResponse,
} from '~/types/byte/v/forge/contracts/proxyruntime/v1/proxy_runtime'
import type { ProxyRuntimeNativeConfig } from '~/composables/proxyRuntimeNativeConfigTypes'
import {
  listMihomoConfigNodes,
  listMihomoEgressOwners,
} from '~/composables/proxyRuntimeMihomoController'

const base = '/api/proxy-runtime'

async function proxyRuntimeRequest<T>(
  path: string,
  init: RequestInit = {},
): Promise<T> {
  const response = await fetch(`${base}${path}`, {
    ...init,
    headers: {
      'Content-Type': 'application/json',
      ...(init.headers || {}),
    },
  })
  if (!response.ok) {
    let message = `${response.status} ${response.statusText}`
    try {
      const body = await response.json()
      if (body?.message) message = body.message
    } catch {
      const body = await response.text()
      if (body) message = body
    }
    throw new Error(message)
  }
  if (response.status === 204) return {} as T
  return (await response.json()) as T
}

const jsonBody = (value: unknown) => JSON.stringify(value)

export function useProxyRuntimeApi() {
  return {
    listProviders: () =>
      proxyRuntimeRequest<ListProxyProvidersResponse>('/providers'),
    listProviderAccounts: () =>
      proxyRuntimeRequest<ListProxyProviderAccountsResponse>(
        '/provider-accounts',
      ),
    upsertProviderAccount: (req: UpsertProxyProviderAccountRequest) =>
      proxyRuntimeRequest<UpsertProxyProviderAccountResponse>(
        '/provider-accounts',
        {
          method: 'PUT',
          body: jsonBody(req),
        },
      ),
    deleteProviderAccount: (req: DeleteProxyProviderAccountRequest) =>
      proxyRuntimeRequest<DeleteProxyProviderAccountResponse>(
        '/provider-accounts',
        {
          method: 'DELETE',
          body: jsonBody(req),
        },
      ),
    getSettings: () =>
      proxyRuntimeRequest<GetProxyRuntimeSettingsResponse>('/settings'),
    getNativeConfig: () =>
      proxyRuntimeRequest<ProxyRuntimeNativeConfig>('/settings/mihomo-native'),
    updateNativeConfig: (config: ProxyRuntimeNativeConfig) =>
      proxyRuntimeRequest<ProxyRuntimeNativeConfig>('/settings/mihomo-native', {
        method: 'PUT',
        body: jsonBody(config),
      }),
    listMihomoEgressOwners,
    listMihomoConfigNodes,
    updateDynamicIPProviders: (
      dynamicIpProviders: ProxyDynamicIPProviderSettings[],
    ) =>
      proxyRuntimeRequest<UpdateProxyRuntimeSettingsResponse>(
        '/settings/dynamic-ip-providers',
        {
          method: 'PUT',
          body: jsonBody({ dynamic_ip_providers: dynamicIpProviders }),
        },
      ),
    updateInUserRules: (
      egressProfiles: EgressProfileSettings[],
      ingressRules: ProxyIngressRuleSettings[],
    ) =>
      proxyRuntimeRequest<UpdateProxyRuntimeSettingsResponse>(
        '/settings/in-user-rules',
        {
          method: 'PUT',
          body: jsonBody({
            egress_profiles: egressProfiles,
            ingress_rules: ingressRules,
          }),
        },
      ),
    listLeases: () =>
      proxyRuntimeRequest<ListProxyDynamicLeasesResponse>('/leases'),
  }
}
