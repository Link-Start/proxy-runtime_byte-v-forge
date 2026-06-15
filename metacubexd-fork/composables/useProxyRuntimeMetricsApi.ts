import type { GetProxyRuntimeMetricsSummaryResponse } from '~/types/byte/v/forge/contracts/proxyruntime/v1/proxy_runtime'
import {
  proxyRuntimeFetchJson,
  type ProxyRuntimeRequestOptions,
} from '~/composables/proxyRuntimeFetch'

const base = '/api'

export function useProxyRuntimeMetricsApi() {
  return {
    getMetricsSummary: (options: ProxyRuntimeRequestOptions = {}) =>
      proxyRuntimeFetchJson<GetProxyRuntimeMetricsSummaryResponse>(
        base,
        '/runtime/metrics/summary',
        { signal: options.signal },
        { json: true, timeoutMs: options.timeoutMs },
      ),
  }
}
