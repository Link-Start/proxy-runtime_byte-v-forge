import type {
  AcquireProxyLeaseRequest,
  AcquireProxyLeaseResponse,
  ListProxyDynamicLeasesResponse,
  ReleaseProxyLeaseRequest,
  ReleaseProxyLeaseResponse,
} from '~/types/byte/v/forge/contracts/proxyruntime/v1/proxy_runtime'
import {
  proxyRuntimeFetchJson,
  proxyRuntimeJsonBody,
  type ProxyRuntimeRequestOptions,
} from '~/composables/proxyRuntimeFetch'

interface ProxyRuntimeLeaseListOptions {
  limit?: number
  status?: 'active' | 'history' | 'recent'
}

const base = '/api'

export function useProxyRuntimeLeaseApi() {
  return {
    acquireLease: (
      req: AcquireProxyLeaseRequest,
      options: ProxyRuntimeRequestOptions = {},
    ) =>
      proxyRuntimeRequest<AcquireProxyLeaseResponse>(
        '/leases/acquire',
        {
          method: 'POST',
          body: proxyRuntimeJsonBody(req),
        },
        options,
      ),
    listLeases: (
      listOptions: ProxyRuntimeLeaseListOptions = {},
      requestOptions: ProxyRuntimeRequestOptions = {},
    ) =>
      proxyRuntimeRequest<ListProxyDynamicLeasesResponse>(
        leaseListPath(listOptions),
        {},
        requestOptions,
      ),
    releaseLease: (
      req: ReleaseProxyLeaseRequest,
      options: ProxyRuntimeRequestOptions = {},
    ) =>
      proxyRuntimeRequest<ReleaseProxyLeaseResponse>(
        '/leases/release',
        {
          method: 'POST',
          body: proxyRuntimeJsonBody(req),
        },
        options,
      ),
  }
}

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

function leaseListPath(options: ProxyRuntimeLeaseListOptions) {
  const query = new URLSearchParams()
  query.set('status', options.status || 'active')
  query.set('limit', String(options.limit || 50))
  return `/leases?${query.toString()}`
}
