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
  type ProxyRuntimeRequestOptions,
} from '~/composables/proxyRuntimeFetch'

const base = '/api'

async function proxyRuntimeRequest<T>(
  path: string,
  init: RequestInit = {},
  options: ProxyRuntimeRequestOptions = {},
): Promise<T> {
  return proxyRuntimeFetchJson<T>(
    base,
    path,
    { ...init, signal: options.signal },
    { json: true, timeoutMs: options.timeoutMs },
  )
}

const emptyMihomoNativeConfig = (): ProxyRuntimeMihomoNativeConfig => ({
  fixed_proxies: [],
  subscriptions: [],
})

export function useProxyRuntimeApi() {
  return {
    listProviders: (options: ProxyRuntimeRequestOptions = {}) =>
      proxyRuntimeRequest<ListProxyProvidersResponse>(
        '/providers',
        {},
        options,
      ),
    listProviderAccounts: (options: ProxyRuntimeRequestOptions = {}) =>
      proxyRuntimeRequest<ListProxyProviderAccountsResponse>(
        '/provider-accounts',
        {},
        options,
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
    getSettings: (options: ProxyRuntimeRequestOptions = {}) =>
      proxyRuntimeRequest<GetProxyRuntimeSettingsResponse>(
        '/settings',
        {},
        options,
      ),
    updateRuntimeSettings: (req: UpdateProxyRuntimeSettingsRequest) =>
      proxyRuntimeRequest<UpdateProxyRuntimeSettingsResponse>('/settings', {
        method: 'PUT',
        body: proxyRuntimeJsonBody(req),
      }),
    listIPFraudProviders: (options: ProxyRuntimeRequestOptions = {}) =>
      proxyRuntimeRequest<ListProxyIPFraudProvidersResponse>(
        '/settings/ip-fraud-providers',
        {},
        options,
      ),
    listIPGeoProviders: (options: ProxyRuntimeRequestOptions = {}) =>
      proxyRuntimeRequest<ListProxyIPGeoProvidersResponse>(
        '/settings/ip-geo-providers',
        {},
        options,
      ),
    checkProxyIPFraud: (
      req: CheckProxyIPFraudRequest,
      options: ProxyRuntimeRequestOptions = {},
    ) =>
      proxyRuntimeRequest<CheckProxyIPFraudResponse>(
        '/ip_fraud_check',
        {
          method: 'POST',
          body: proxyRuntimeJsonBody(req),
        },
        options,
      ),
    getProxyExitIP: (
      req: GetProxyExitIPRequest,
      options: ProxyRuntimeRequestOptions = {},
    ) =>
      proxyRuntimeRequest<GetProxyExitIPResponse>(
        '/proxy_exit_ip',
        {
          method: 'POST',
          body: proxyRuntimeJsonBody(req),
        },
        options,
      ),
    checkProxyExitGeo: (
      req: GetProxyExitGeoRequest,
      options: ProxyRuntimeRequestOptions = {},
    ) =>
      proxyRuntimeRequest<GetProxyExitGeoResponse>(
        '/proxy_exit_geo',
        {
          method: 'POST',
          body: proxyRuntimeJsonBody(req),
        },
        options,
      ),
    checkProxyEdgeAccess: (
      req: CheckProxyEdgeAccessRequest,
      options: ProxyRuntimeRequestOptions = {},
    ) =>
      proxyRuntimeRequest<CheckProxyEdgeAccessResponse>(
        '/check_cf_access_risk',
        {
          method: 'POST',
          body: proxyRuntimeJsonBody(req),
        },
        options,
      ),
    getProxyExitCheckSnapshot: (
      req: GetProxyExitCheckSnapshotRequest,
      options: ProxyRuntimeRequestOptions = {},
    ) =>
      proxyRuntimeRequest<GetProxyExitCheckSnapshotResponse>(
        '/proxy_exit_check_snapshot',
        {
          method: 'POST',
          body: proxyRuntimeJsonBody(req),
        },
        options,
      ),
    getNativeConfig: (options: ProxyRuntimeRequestOptions = {}) =>
      proxyRuntimeRequest<GetProxyRuntimeMihomoNativeConfigResponse>(
        '/settings/mihomo-native',
        {},
        options,
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
