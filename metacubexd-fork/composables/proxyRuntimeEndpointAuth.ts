export const proxyRuntimeMihomoEndpointID = 'proxy-runtime-mihomo'

export function isUnauthorizedStatus(status: number) {
  return status === 401 || status === 403
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
