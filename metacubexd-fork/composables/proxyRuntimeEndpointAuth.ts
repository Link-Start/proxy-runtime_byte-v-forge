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

export function proxyRuntimeEndpointNeedsSetup() {
  return false
}

export function proxyRuntimeEndpointSecret() {
  return ''
}

export function proxyRuntimeAuthHeaders(init?: HeadersInit) {
  return new Headers(init)
}

export function redirectToProxyRuntimeSetup() {
  redirectToProxyRuntimeLogin()
}

export function redirectToProxyRuntimeLogin() {
  if (typeof window === 'undefined') return
  if (window.location.pathname === '/login') return
  const target = new URL('/login', window.location.origin)
  target.searchParams.set('next', proxyRuntimeCurrentPath())
  window.location.replace(target.href)
}

function proxyRuntimeCurrentPath() {
  return `${window.location.pathname}${window.location.search}${window.location.hash}`
}
