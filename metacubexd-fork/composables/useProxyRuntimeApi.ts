import type {
  DeleteProxyProviderAccountRequest,
  DeleteProxyProviderAccountResponse,
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
  GetProxyRuntimeMihomoNativeConfigResponse,
  GetProxyRuntimeSettingsResponse,
  ListProxyIPFraudProvidersResponse,
  ListProxyIPGeoProvidersResponse,
  ListProxyProviderAccountsResponse,
  ListProxyProvidersResponse,
  ProxyDynamicIPProviderSettings,
  ProxyIngressRuleSettings,
  ProxyRuntimeMihomoNativeConfig,
  UpdateProxyRuntimeMihomoNativeConfigResponse,
  UpdateProxyRuntimeSettingsRequest,
  UpdateProxyRuntimeSettingsResponse,
  UpsertProxyProviderAccountRequest,
  UpsertProxyProviderAccountResponse,
} from '~/types/byte/v/forge/contracts/proxyruntime/v1/proxy_runtime'
import {
  listMihomoConfigNodes,
  listMihomoEgressOwners,
} from '~/composables/proxyRuntimeMihomoController'
import {
  proxyRuntimeFetchJson,
  proxyRuntimeJsonBody,
} from '~/composables/proxyRuntimeFetch'

const base = '/api'

async function proxyRuntimeRequest<T>(
  path: string,
  init: RequestInit = {},
): Promise<T> {
  return proxyRuntimeFetchJson<T>(base, path, init, { json: true })
}

const emptyMihomoNativeConfig = (): ProxyRuntimeMihomoNativeConfig => ({
  fixed_proxies: [],
  subscriptions: [],
})

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
          body: proxyRuntimeJsonBody(req),
        },
      ),
    deleteProviderAccount: (req: DeleteProxyProviderAccountRequest) =>
      proxyRuntimeRequest<DeleteProxyProviderAccountResponse>(
        '/provider-accounts',
        {
          method: 'DELETE',
          body: proxyRuntimeJsonBody(req),
        },
      ),
    getSettings: () =>
      proxyRuntimeRequest<GetProxyRuntimeSettingsResponse>('/settings'),
    updateRuntimeSettings: (req: UpdateProxyRuntimeSettingsRequest) =>
      proxyRuntimeRequest<UpdateProxyRuntimeSettingsResponse>('/settings', {
        method: 'PUT',
        body: proxyRuntimeJsonBody(req),
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
        body: proxyRuntimeJsonBody(req),
      }),
    getProxyExitIP: (req: GetProxyExitIPRequest) =>
      proxyRuntimeRequest<GetProxyExitIPResponse>('/proxy_exit_ip', {
        method: 'POST',
        body: proxyRuntimeJsonBody(req),
      }),
    checkProxyExitGeo: (req: GetProxyExitGeoRequest) =>
      proxyRuntimeRequest<GetProxyExitGeoResponse>('/proxy_exit_geo', {
        method: 'POST',
        body: proxyRuntimeJsonBody(req),
      }),
    checkProxyEdgeAccess: (req: CheckProxyEdgeAccessRequest) =>
      proxyRuntimeRequest<CheckProxyEdgeAccessResponse>(
        '/check_cf_access_risk',
        {
          method: 'POST',
          body: proxyRuntimeJsonBody(req),
        },
      ),
    getProxyExitCheckSnapshot: (req: GetProxyExitCheckSnapshotRequest) =>
      proxyRuntimeRequest<GetProxyExitCheckSnapshotResponse>(
        '/proxy_exit_check_snapshot',
        {
          method: 'POST',
          body: proxyRuntimeJsonBody(req),
        },
      ),
    getNativeConfig: () =>
      proxyRuntimeRequest<GetProxyRuntimeMihomoNativeConfigResponse>(
        '/settings/mihomo-native',
      ).then((response) => response.config || emptyMihomoNativeConfig()),
    updateNativeConfig: (config: ProxyRuntimeMihomoNativeConfig) =>
      proxyRuntimeRequest<UpdateProxyRuntimeMihomoNativeConfigResponse>(
        '/settings/mihomo-native',
        {
          method: 'PUT',
          body: proxyRuntimeJsonBody({ config }),
        },
      ).then((response) => response.config || emptyMihomoNativeConfig()),
    listMihomoEgressOwners,
    listMihomoConfigNodes,
    updateDynamicIPProviders: (
      dynamicIpProviders: ProxyDynamicIPProviderSettings[],
    ) =>
      proxyRuntimeRequest<UpdateProxyRuntimeSettingsResponse>(
        '/settings/dynamic-ip-providers',
        {
          method: 'PUT',
          body: proxyRuntimeJsonBody({
            dynamic_ip_providers: dynamicIpProviders,
          }),
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
          body: proxyRuntimeJsonBody({
            egress_profiles: egressProfiles,
            ingress_rules: ingressRules,
          }),
        },
      ),
  }
}
