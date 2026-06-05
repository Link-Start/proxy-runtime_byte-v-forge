import type {
  DeleteProxyProviderAccountRequest,
  DeleteProxyProviderAccountResponse,
  EgressProfileSettings,
  GetProxyRuntimeSettingsResponse,
  ListProxySourcesResponse,
  ListProxySourceNodesResponse,
  ListProxyDynamicLeasesResponse,
  ListProxyProviderAccountsResponse,
  ListProxyProvidersResponse,
  ProxyDynamicIPProviderSettings,
  ProxyIngressRuleSettings,
  DeleteProxySourceRequest,
  DeleteProxySourceResponse,
  ReleaseProxyLeaseRequest,
  ReleaseProxyLeaseResponse,
  UpdateProxyIngressRulesResponse,
  UpdateProxyRuntimeSettingsResponse,
  UpdateProxyEgressProfilesResponse,
  UpsertProxyFixedSourceRequest,
  UpsertProxyFixedSourceResponse,
  UpsertProxyProviderAccountRequest,
  UpsertProxyProviderAccountResponse,
  UpsertProxySubscriptionSourceRequest,
  UpsertProxySubscriptionSourceResponse,
} from '~/types/byte/v/forge/contracts/proxyruntime/v1/proxy_runtime'

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
    listSources: () =>
      proxyRuntimeRequest<ListProxySourcesResponse>('/sources'),
    listSourceNodes: (sourceId: string) =>
      proxyRuntimeRequest<ListProxySourceNodesResponse>(
        `/sources/nodes?source_id=${encodeURIComponent(sourceId)}`,
      ),
    upsertSubscriptionSource: (req: UpsertProxySubscriptionSourceRequest) =>
      proxyRuntimeRequest<UpsertProxySubscriptionSourceResponse>('/sources', {
        method: 'PUT',
        body: jsonBody(req),
      }),
    upsertFixedSource: (req: UpsertProxyFixedSourceRequest) =>
      proxyRuntimeRequest<UpsertProxyFixedSourceResponse>('/sources/fixed', {
        method: 'PUT',
        body: jsonBody(req),
      }),
    deleteSource: (req: DeleteProxySourceRequest) =>
      proxyRuntimeRequest<DeleteProxySourceResponse>('/sources', {
        method: 'DELETE',
        body: jsonBody(req),
      }),
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
    updateEgressProfiles: (egressProfiles: EgressProfileSettings[]) =>
      proxyRuntimeRequest<UpdateProxyEgressProfilesResponse>(
        '/settings/egress-profiles',
        {
          method: 'PUT',
          body: jsonBody({ egress_profiles: egressProfiles }),
        },
      ),
    updateIngressRules: (ingressRules: ProxyIngressRuleSettings[]) =>
      proxyRuntimeRequest<UpdateProxyIngressRulesResponse>(
        '/settings/ingress-rules',
        {
          method: 'PUT',
          body: jsonBody({ ingress_rules: ingressRules }),
        },
      ),
    listLeases: () =>
      proxyRuntimeRequest<ListProxyDynamicLeasesResponse>(
        '/leases?include_inactive=true',
      ),
    releaseLease: (req: ReleaseProxyLeaseRequest) =>
      proxyRuntimeRequest<ReleaseProxyLeaseResponse>('/leases/release', {
        method: 'POST',
        body: jsonBody(req),
      }),
  }
}
