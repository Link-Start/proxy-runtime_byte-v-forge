export const proxyRuntimeMihomoEndpointID = 'proxy-runtime-mihomo'
export const proxyRuntimeAuthRequiredStorageKey =
  'proxyRuntimeControlAuthRequired'

export function isUnauthorizedStatus(status: number) {
  return status === 401 || status === 403
}

export function proxyRuntimeAuthRequired() {
  if (typeof window === 'undefined') return false
  return (
    window.localStorage.getItem(proxyRuntimeAuthRequiredStorageKey) === 'true'
  )
}

export function proxyRuntimeEndpointNeedsSetup(
  endpoint: { id?: string; url?: string; secret?: string } | null | undefined,
) {
  if (!proxyRuntimeAuthRequired() || endpointSecret(endpoint)) return false
  return proxyRuntimeEndpointMatches(endpoint)
}

export function proxyRuntimeEndpointSecret(
  endpointID = proxyRuntimeMihomoEndpointID,
) {
  if (typeof window === 'undefined') return ''
  const selectedEndpoint = window.localStorage.getItem('selectedEndpoint') || ''
  try {
    const endpoints = JSON.parse(
      window.localStorage.getItem('endpointList') || '[]',
    )
    if (!Array.isArray(endpoints)) return ''
    const endpoint = endpoints.find((item) => item?.id === endpointID)
    const selected = endpoints.find((item) => item?.id === selectedEndpoint)
    return endpointSecret(endpoint) || endpointSecret(selected)
  } catch {
    return ''
  }
}

function endpointSecret(endpoint: unknown) {
  if (!endpoint || typeof endpoint !== 'object' || !('secret' in endpoint)) {
    return ''
  }
  const secret = endpoint.secret
  return typeof secret === 'string' ? secret.trim() : ''
}

function proxyRuntimeEndpointMatches(endpoint: unknown) {
  if (!endpoint || typeof endpoint !== 'object') return false
  if ('id' in endpoint && endpoint.id === proxyRuntimeMihomoEndpointID) {
    return true
  }
  if (!('url' in endpoint) || typeof endpoint.url !== 'string') return false
  return proxyRuntimeEndpointURL(endpoint.url) === proxyRuntimeEndpointURL()
}

function proxyRuntimeEndpointURL(value = '/mihomo/controller') {
  if (typeof window === 'undefined') return ''
  try {
    return new URL(value, window.location.origin).href.replace(/\/$/, '')
  } catch {
    return ''
  }
}

export function proxyRuntimeAuthHeaders(init?: HeadersInit) {
  const headers = new Headers(init)
  const secret = proxyRuntimeEndpointSecret()
  if (secret && !headers.has('Authorization')) {
    headers.set('Authorization', `Bearer ${secret}`)
  }
  return headers
}

export function redirectToProxyRuntimeSetup(
  endpointID = proxyRuntimeMihomoEndpointID,
) {
  if (typeof window === 'undefined') return
  if ((window.location.hash || '').startsWith('#/setup')) return
  if (window.localStorage.getItem('selectedEndpoint') === endpointID) {
    window.localStorage.removeItem('selectedEndpoint')
  }
  const target = new URL(window.location.href)
  target.search = ''
  target.hash = `/setup?endpoint=${encodeURIComponent(endpointID)}`
  window.location.replace(target.href)
}
