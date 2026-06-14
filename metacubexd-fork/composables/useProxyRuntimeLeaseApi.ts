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
} from '~/composables/proxyRuntimeFetch'

interface ProxyRuntimeLeaseListOptions {
  limit?: number
  status?: 'active' | 'history' | 'recent'
}

const base = '/api'

export function useProxyRuntimeLeaseApi() {
  return {
    acquireLease: (req: AcquireProxyLeaseRequest) =>
      proxyRuntimeRequest<AcquireProxyLeaseResponse>('/leases/acquire', {
        method: 'POST',
        body: proxyRuntimeJsonBody(req),
      }),
    listLeases: (options: ProxyRuntimeLeaseListOptions = {}) =>
      proxyRuntimeRequest<ListProxyDynamicLeasesResponse>(
        leaseListPath(options),
      ),
    releaseLease: (req: ReleaseProxyLeaseRequest) =>
      proxyRuntimeRequest<ReleaseProxyLeaseResponse>('/leases/release', {
        method: 'POST',
        body: proxyRuntimeJsonBody(req),
      }),
  }
}

async function proxyRuntimeRequest<T>(
  path: string,
  init: RequestInit = {},
): Promise<T> {
  return proxyRuntimeFetchJson<T>(base, path, init, { json: true })
}

function leaseListPath(options: ProxyRuntimeLeaseListOptions) {
  const query = new URLSearchParams()
  query.set('status', options.status || 'active')
  query.set('limit', String(options.limit || 50))
  return `/leases?${query.toString()}`
}
