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
  GetProxyGatewayMihomoNativeConfigResponse,
  GetProxyGatewaySettingsResponse,
  ListProxyIPFraudProvidersResponse,
  ListProxyIPGeoProvidersResponse,
  ListProxyProviderAccountsResponse,
  ListProxyProvidersResponse,
  ProxyDynamicIPProviderSettings,
  ProxyIngressRuleSettings,
  ProxyGatewayMihomoNativeConfig,
  UpdateProxyGatewayMihomoNativeConfigResponse,
  UpdateProxyGatewaySettingsRequest,
  UpdateProxyGatewaySettingsResponse,
  UpsertProxyProviderAccountRequest,
  UpsertProxyProviderAccountResponse,
} from '~/types/byte/v/forge/contracts/proxygateway/v1/proxy_gateway'
import {
  listMihomoConfigNodes,
  listMihomoEgressOwners,
} from '~/composables/proxyGatewayMihomoController'
import {
  proxyGatewayFetchJson,
  proxyGatewayJsonBody,
  type ProxyGatewayRequestOptions,
} from '~/composables/proxyGatewayFetch'

const base = '/api'

async function proxyGatewayRequest<T>(
  path: string,
  init: RequestInit = {},
  options: ProxyGatewayRequestOptions = {},
): Promise<T> {
  return proxyGatewayFetchJson<T>(
    base,
    path,
    { ...init, signal: options.signal },
    { json: true, timeoutMs: options.timeoutMs },
  )
}

const emptyMihomoNativeConfig = (): ProxyGatewayMihomoNativeConfig => ({
  fixed_proxies: [],
  subscriptions: [],
})

export function useProxyGatewayApi() {
  return {
    listProviders: (options: ProxyGatewayRequestOptions = {}) =>
      proxyGatewayRequest<ListProxyProvidersResponse>(
        '/providers',
        {},
        options,
      ),
    listProviderAccounts: (options: ProxyGatewayRequestOptions = {}) =>
      proxyGatewayRequest<ListProxyProviderAccountsResponse>(
        '/provider-accounts',
        {},
        options,
      ),
    upsertProviderAccount: (req: UpsertProxyProviderAccountRequest) =>
      proxyGatewayRequest<UpsertProxyProviderAccountResponse>(
        '/provider-accounts',
        {
          method: 'PUT',
          body: proxyGatewayJsonBody(req),
        },
      ),
    deleteProviderAccount: (req: DeleteProxyProviderAccountRequest) =>
      proxyGatewayRequest<DeleteProxyProviderAccountResponse>(
        '/provider-accounts',
        {
          method: 'DELETE',
          body: proxyGatewayJsonBody(req),
        },
      ),
    getSettings: (options: ProxyGatewayRequestOptions = {}) =>
      proxyGatewayRequest<GetProxyGatewaySettingsResponse>(
        '/settings',
        {},
        options,
      ),
    updateRuntimeSettings: (req: UpdateProxyGatewaySettingsRequest) =>
      proxyGatewayRequest<UpdateProxyGatewaySettingsResponse>('/settings', {
        method: 'PUT',
        body: proxyGatewayJsonBody(req),
      }),
    listIPFraudProviders: (options: ProxyGatewayRequestOptions = {}) =>
      proxyGatewayRequest<ListProxyIPFraudProvidersResponse>(
        '/settings/ip-fraud-providers',
        {},
        options,
      ),
    listIPGeoProviders: (options: ProxyGatewayRequestOptions = {}) =>
      proxyGatewayRequest<ListProxyIPGeoProvidersResponse>(
        '/settings/ip-geo-providers',
        {},
        options,
      ),
    checkProxyIPFraud: (
      req: CheckProxyIPFraudRequest,
      options: ProxyGatewayRequestOptions = {},
    ) =>
      proxyGatewayRequest<CheckProxyIPFraudResponse>(
        '/ip_fraud_check',
        {
          method: 'POST',
          body: proxyGatewayJsonBody(req),
        },
        options,
      ),
    getProxyExitIP: (
      req: GetProxyExitIPRequest,
      options: ProxyGatewayRequestOptions = {},
    ) =>
      proxyGatewayRequest<GetProxyExitIPResponse>(
        '/proxy_exit_ip',
        {
          method: 'POST',
          body: proxyGatewayJsonBody(req),
        },
        options,
      ),
    checkProxyExitGeo: (
      req: GetProxyExitGeoRequest,
      options: ProxyGatewayRequestOptions = {},
    ) =>
      proxyGatewayRequest<GetProxyExitGeoResponse>(
        '/proxy_exit_geo',
        {
          method: 'POST',
          body: proxyGatewayJsonBody(req),
        },
        options,
      ),
    checkProxyEdgeAccess: (
      req: CheckProxyEdgeAccessRequest,
      options: ProxyGatewayRequestOptions = {},
    ) =>
      proxyGatewayRequest<CheckProxyEdgeAccessResponse>(
        '/check_cf_access_risk',
        {
          method: 'POST',
          body: proxyGatewayJsonBody(req),
        },
        options,
      ),
    getProxyExitCheckSnapshot: (
      req: GetProxyExitCheckSnapshotRequest,
      options: ProxyGatewayRequestOptions = {},
    ) =>
      proxyGatewayRequest<GetProxyExitCheckSnapshotResponse>(
        '/proxy_exit_check_snapshot',
        {
          method: 'POST',
          body: proxyGatewayJsonBody(req),
        },
        options,
      ),
    getNativeConfig: (options: ProxyGatewayRequestOptions = {}) =>
      proxyGatewayRequest<GetProxyGatewayMihomoNativeConfigResponse>(
        '/settings/mihomo-native',
        {},
        options,
      ).then((response) => response.config || emptyMihomoNativeConfig()),
    updateNativeConfig: (config: ProxyGatewayMihomoNativeConfig) =>
      proxyGatewayRequest<UpdateProxyGatewayMihomoNativeConfigResponse>(
        '/settings/mihomo-native',
        {
          method: 'PUT',
          body: proxyGatewayJsonBody({ config }),
        },
      ).then((response) => response.config || emptyMihomoNativeConfig()),
    listMihomoEgressOwners,
    listMihomoConfigNodes,
    updateDynamicIPProviders: (
      dynamicIpProviders: ProxyDynamicIPProviderSettings[],
    ) =>
      proxyGatewayRequest<UpdateProxyGatewaySettingsResponse>(
        '/settings/dynamic-ip-providers',
        {
          method: 'PUT',
          body: proxyGatewayJsonBody({
            dynamic_ip_providers: dynamicIpProviders,
          }),
        },
      ),
    updateInUserRules: (
      egressProfiles: EgressProfileSettings[],
      ingressRules: ProxyIngressRuleSettings[],
    ) =>
      proxyGatewayRequest<UpdateProxyGatewaySettingsResponse>(
        '/settings/in-user-rules',
        {
          method: 'PUT',
          body: proxyGatewayJsonBody({
            egress_profiles: egressProfiles,
            ingress_rules: ingressRules,
          }),
        },
      ),
  }
}
