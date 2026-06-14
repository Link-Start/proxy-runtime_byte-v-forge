import type { GetProxyRuntimeStatusResponse } from '~/types/byte/v/forge/contracts/proxyruntime/v1/proxy_runtime'
import { proxyRuntimeFetchJson } from '~/composables/proxyRuntimeFetch'

const base = '/api'

export function useProxyRuntimeStatusApi() {
  return {
    getStatus: () =>
      proxyRuntimeFetchJson<GetProxyRuntimeStatusResponse>(
        base,
        '/runtime/status',
        {},
        { json: true, timeoutMs: 10000 },
      ),
  }
}
