import type {
  AcquireProxyLeaseRequest,
  AcquireProxyLeaseResponse,
  ListProxyDynamicLeasesResponse,
  ReleaseProxyLeaseRequest,
  ReleaseProxyLeaseResponse,
} from '~/types/byte/v/forge/contracts/proxygateway/v1/proxy_gateway'
import {
  proxyGatewayFetchJson,
  proxyGatewayJsonBody,
  type ProxyGatewayRequestOptions,
} from '~/composables/proxyGatewayFetch'

interface ProxyGatewayLeaseListOptions {
  limit?: number
  status?: 'active' | 'history' | 'recent'
}

const base = '/api'

export function useProxyGatewayLeaseApi() {
  return {
    acquireLease: (
      req: AcquireProxyLeaseRequest,
      options: ProxyGatewayRequestOptions = {},
    ) =>
      proxyGatewayRequest<AcquireProxyLeaseResponse>(
        '/leases/acquire',
        {
          method: 'POST',
          body: proxyGatewayJsonBody(req),
        },
        options,
      ),
    listLeases: (
      listOptions: ProxyGatewayLeaseListOptions = {},
      requestOptions: ProxyGatewayRequestOptions = {},
    ) =>
      proxyGatewayRequest<ListProxyDynamicLeasesResponse>(
        leaseListPath(listOptions),
        {},
        requestOptions,
      ),
    releaseLease: (
      req: ReleaseProxyLeaseRequest,
      options: ProxyGatewayRequestOptions = {},
    ) =>
      proxyGatewayRequest<ReleaseProxyLeaseResponse>(
        '/leases/release',
        {
          method: 'POST',
          body: proxyGatewayJsonBody(req),
        },
        options,
      ),
  }
}

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

function leaseListPath(options: ProxyGatewayLeaseListOptions) {
  const query = new URLSearchParams()
  query.set('status', options.status || 'active')
  query.set('limit', String(options.limit || 50))
  return `/leases?${query.toString()}`
}
