import type { GetProxyGatewayMetricsSummaryResponse } from '~/types/byte/v/forge/contracts/proxygateway/v1/proxy_gateway'
import {
  proxyGatewayFetchJson,
  type ProxyGatewayRequestOptions,
} from '~/composables/proxyGatewayFetch'

const base = '/api'

export function useProxyGatewayMetricsApi() {
  return {
    getMetricsSummary: (options: ProxyGatewayRequestOptions = {}) =>
      proxyGatewayFetchJson<GetProxyGatewayMetricsSummaryResponse>(
        base,
        '/runtime/metrics/summary',
        { signal: options.signal },
        { json: true, timeoutMs: options.timeoutMs },
      ),
  }
}
