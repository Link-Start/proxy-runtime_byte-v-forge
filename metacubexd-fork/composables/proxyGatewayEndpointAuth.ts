export const proxyGatewayMihomoEndpointID = 'proxy-gateway-mihomo'
export const proxyGatewayAuthRequiredStorageKey =
  'proxyGatewayControlAuthRequired'

export function isUnauthorizedStatus(status: number) {
  return status === 401 || status === 403
}

export function proxyGatewayAuthRequired() {
  if (typeof window === 'undefined') return false
  return (
    window.localStorage.getItem(proxyGatewayAuthRequiredStorageKey) === 'true'
  )
}

export function proxyGatewayEndpointNeedsSetup() {
  return false
}

export function proxyGatewayEndpointSecret() {
  return ''
}

export function proxyGatewayAuthHeaders(init?: HeadersInit) {
  return new Headers(init)
}

export function redirectToProxyGatewaySetup() {
  redirectToProxyGatewayLogin()
}

export function redirectToProxyGatewayLogin() {
  if (typeof window === 'undefined') return
  if (window.location.pathname === '/login') return
  const target = new URL('/login', window.location.origin)
  target.searchParams.set('next', proxyGatewayCurrentPath())
  window.location.replace(target.href)
}

function proxyGatewayCurrentPath() {
  return `${window.location.pathname}${window.location.search}${window.location.hash}`
}

interface ProxyGatewayWebSocketTokenResponse {
  token?: string
}

export async function proxyGatewayWebSocketSessionParam() {
  if (!proxyGatewayAuthRequired()) return ''
  const response = await fetch('/api/auth/ws-token', {
    credentials: 'same-origin',
    headers: { Accept: 'application/json' },
  })
  if (!response.ok) {
    if (isUnauthorizedStatus(response.status)) redirectToProxyGatewayLogin()
    return ''
  }
  const body = (await response.json()) as ProxyGatewayWebSocketTokenResponse
  return body.token ? `session=${encodeURIComponent(body.token)}` : ''
}
