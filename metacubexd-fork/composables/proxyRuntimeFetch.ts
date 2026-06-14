import {
  isUnauthorizedStatus,
  proxyRuntimeAuthHeaders,
  redirectToProxyRuntimeSetup,
} from '~/composables/proxyRuntimeEndpointAuth'

interface ProxyRuntimeFetchOptions {
  json?: boolean
}

export async function proxyRuntimeFetchJson<T>(
  base: string,
  path: string,
  init: RequestInit = {},
  options: ProxyRuntimeFetchOptions = {},
): Promise<T> {
  const response = await fetch(`${base}${path}`, {
    ...init,
    headers: proxyRuntimeFetchHeaders(init.headers, options),
  })
  const text = await response.text()
  if (!response.ok) {
    if (isUnauthorizedStatus(response.status)) {
      redirectToProxyRuntimeSetup()
    }
    throw new Error(proxyRuntimeErrorMessage(response, text))
  }
  if (response.status === 204 || !text.trim()) return {} as T
  return JSON.parse(text) as T
}

export const proxyRuntimeJsonBody = (value: unknown) => JSON.stringify(value)

function proxyRuntimeFetchHeaders(
  init: HeadersInit | undefined,
  options: ProxyRuntimeFetchOptions,
) {
  const headers = proxyRuntimeAuthHeaders(init)
  if (options.json) {
    headers.set('Content-Type', 'application/json')
  }
  return headers
}

function proxyRuntimeErrorMessage(response: Response, text: string) {
  if (text.trim()) {
    try {
      const body = JSON.parse(text)
      if (body?.message) return body.message
    } catch {
      return text
    }
  }
  return `${response.status} ${response.statusText}`
}
