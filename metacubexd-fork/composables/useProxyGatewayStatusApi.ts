import type { GetProxyGatewayStatusResponse } from '~/types/byte/v/forge/contracts/proxygateway/v1/proxy_gateway'
import {
  proxyGatewayFetchJson,
  type ProxyGatewayRequestOptions,
} from '~/composables/proxyGatewayFetch'

const base = '/api'
const runtimeStatusTimeoutMs = 10000

export function useProxyGatewayStatusApi() {
  return {
    getStatus: (options: ProxyGatewayRequestOptions = {}) =>
      proxyGatewayFetchJson<GetProxyGatewayStatusResponse>(
        base,
        '/runtime/status',
        { signal: options.signal },
        { json: true, timeoutMs: options.timeoutMs || runtimeStatusTimeoutMs },
      ),
  }
}
