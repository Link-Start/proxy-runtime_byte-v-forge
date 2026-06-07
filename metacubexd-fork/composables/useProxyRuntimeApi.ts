import type {
  DeleteProxyProviderAccountRequest,
  DeleteProxyProviderAccountResponse,
  AcquireProxyLeaseRequest,
  AcquireProxyLeaseResponse,
  EgressProfileSettings,
  CheckProxyEdgeAccessRequest,
  CheckProxyEdgeAccessResponse,
  CheckProxyIPFraudRequest,
  CheckProxyIPFraudResponse,
  GetProxyExitCheckSnapshotRequest,
  GetProxyExitCheckSnapshotResponse,
  GetProxyExitIPRequest,
  GetProxyExitIPResponse,
  GetProxyExitGeoRequest,
  GetProxyExitGeoResponse,
  GetProxyRuntimeSettingsResponse,
  ListProxyIPFraudProvidersResponse,
  ListProxyIPGeoProvidersResponse,
  ListProxyDynamicLeasesResponse,
  ListProxyProviderAccountsResponse,
  ListProxyProvidersResponse,
  ProxyDynamicIPProviderSettings,
  ProxyIngressRuleSettings,
  ReleaseProxyLeaseRequest,
  ReleaseProxyLeaseResponse,
  UpdateProxyRuntimeSettingsRequest,
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
    updateRuntimeSettings: (req: UpdateProxyRuntimeSettingsRequest) =>
      proxyRuntimeRequest<UpdateProxyRuntimeSettingsResponse>('/settings', {
        method: 'PUT',
        body: jsonBody(req),
      }),
    listIPFraudProviders: () =>
      proxyRuntimeRequest<ListProxyIPFraudProvidersResponse>(
        '/settings/ip-fraud-providers',
      ),
    listIPGeoProviders: () =>
      proxyRuntimeRequest<ListProxyIPGeoProvidersResponse>(
        '/settings/ip-geo-providers',
      ),
    checkProxyIPFraud: (req: CheckProxyIPFraudRequest) =>
      proxyRuntimeRequest<CheckProxyIPFraudResponse>('/ip_fraud_check', {
        method: 'POST',
        body: jsonBody(req),
      }),
    getProxyExitIP: (req: GetProxyExitIPRequest) =>
      proxyRuntimeRequest<GetProxyExitIPResponse>('/proxy_exit_ip', {
        method: 'POST',
        body: jsonBody(req),
      }),
    checkProxyExitGeo: (req: GetProxyExitGeoRequest) =>
      proxyRuntimeRequest<GetProxyExitGeoResponse>('/proxy_exit_geo', {
        method: 'POST',
        body: jsonBody(req),
      }),
    checkProxyEdgeAccess: (req: CheckProxyEdgeAccessRequest) =>
      proxyRuntimeRequest<CheckProxyEdgeAccessResponse>(
        '/check_cf_access_risk',
        {
          method: 'POST',
          body: jsonBody(req),
        },
      ),
    getProxyExitCheckSnapshot: (req: GetProxyExitCheckSnapshotRequest) =>
      proxyRuntimeRequest<GetProxyExitCheckSnapshotResponse>(
        '/proxy_exit_check_snapshot',
        {
          method: 'POST',
          body: jsonBody(req),
        },
      ),
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
      proxyRuntimeRequest<ListProxyDynamicLeasesResponse>(
        '/leases?include_inactive=true',
      ),
    acquireLease: (req: AcquireProxyLeaseRequest) =>
      proxyRuntimeRequest<AcquireProxyLeaseResponse>('/leases/acquire', {
        method: 'POST',
        body: jsonBody(req),
      }),
    releaseLease: (req: ReleaseProxyLeaseRequest) =>
      proxyRuntimeRequest<ReleaseProxyLeaseResponse>('/leases/release', {
        method: 'POST',
        body: jsonBody(req),
      }),
  }
}
