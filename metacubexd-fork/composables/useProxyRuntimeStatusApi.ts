import type { GetProxyRuntimeStatusResponse } from '~/types/byte/v/forge/contracts/proxyruntime/v1/proxy_runtime'
import {
  proxyRuntimeFetchJson,
  type ProxyRuntimeRequestOptions,
} from '~/composables/proxyRuntimeFetch'

const base = '/api'
const runtimeStatusTimeoutMs = 10000

export function useProxyRuntimeStatusApi() {
  return {
    getStatus: (options: ProxyRuntimeRequestOptions = {}) =>
      proxyRuntimeFetchJson<GetProxyRuntimeStatusResponse>(
        base,
        '/runtime/status',
        { signal: options.signal },
        { json: true, timeoutMs: options.timeoutMs || runtimeStatusTimeoutMs },
      ),
  }
}
