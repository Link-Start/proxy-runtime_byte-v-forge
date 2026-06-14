import {
  isUnauthorizedStatus,
  proxyRuntimeAuthHeaders,
  redirectToProxyRuntimeSetup,
} from '~/composables/proxyRuntimeEndpointAuth'

interface ProxyRuntimeFetchOptions {
  json?: boolean
  timeoutMs?: number
}

const defaultProxyRuntimeTimeoutMs = 30000

export async function proxyRuntimeFetchJson<T>(
  base: string,
  path: string,
  init: RequestInit = {},
  options: ProxyRuntimeFetchOptions = {},
): Promise<T> {
  const controller = new AbortController()
  const timeoutMs = options.timeoutMs ?? defaultProxyRuntimeTimeoutMs
  let timedOut = false
  const timeoutID = globalThis.setTimeout(() => {
    timedOut = true
    controller.abort()
  }, timeoutMs)
  const abort = () => controller.abort()
  if (init.signal?.aborted) {
    controller.abort()
  } else {
    init.signal?.addEventListener('abort', abort, { once: true })
  }
  try {
    const response = await fetch(`${base}${path}`, {
      ...init,
      credentials: 'same-origin',
      headers: proxyRuntimeFetchHeaders(init.headers, options),
      signal: controller.signal,
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
  } catch (err) {
    if (timedOut) throw new Error('Proxy Runtime 请求超时')
    if (controller.signal.aborted) throw new Error('Proxy Runtime 请求已取消')
    if (err instanceof TypeError) throw new Error('Proxy Runtime 后端不可达')
    throw err
  } finally {
    globalThis.clearTimeout(timeoutID)
    init.signal?.removeEventListener('abort', abort)
  }
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
