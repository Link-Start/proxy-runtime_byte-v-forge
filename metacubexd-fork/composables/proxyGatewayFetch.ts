import {
  isUnauthorizedStatus,
  proxyGatewayAuthHeaders,
  redirectToProxyGatewaySetup,
} from '~/composables/proxyGatewayEndpointAuth'

export type ProxyGatewayErrorKind =
  | 'timeout'
  | 'cancelled'
  | 'unauthorized'
  | 'backend_unreachable'
  | 'validation_error'
  | 'provider_error'
  | 'internal_error'

interface ProxyGatewayFetchOptions {
  json?: boolean
  timeoutMs?: number
}

export interface ProxyGatewayRequestOptions {
  signal?: AbortSignal
  timeoutMs?: number
}

export class ProxyGatewayRequestError extends Error {
  readonly kind: ProxyGatewayErrorKind
  readonly status?: number

  constructor(kind: ProxyGatewayErrorKind, message: string, status?: number) {
    super(message)
    this.name = 'ProxyGatewayRequestError'
    this.kind = kind
    this.status = status
  }
}

const defaultProxyGatewayTimeoutMs = 30000

export async function proxyGatewayFetchJson<T>(
  base: string,
  path: string,
  init: RequestInit = {},
  options: ProxyGatewayFetchOptions = {},
): Promise<T> {
  const controller = new AbortController()
  const timeoutMs = options.timeoutMs ?? defaultProxyGatewayTimeoutMs
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
      headers: proxyGatewayFetchHeaders(init.headers, options),
      signal: controller.signal,
    })
    const text = await response.text()
    if (!response.ok) {
      const message = proxyGatewayErrorMessage(response, text)
      const kind = proxyGatewayHTTPErrorKind(response.status)
      if (kind === 'unauthorized') {
        redirectToProxyGatewaySetup()
      }
      throw new ProxyGatewayRequestError(kind, message, response.status)
    }
    if (response.status === 204 || !text.trim()) return {} as T
    return JSON.parse(text) as T
  } catch (err) {
    if (timedOut) {
      throw new ProxyGatewayRequestError('timeout', 'Proxy Runtime 请求超时')
    }
    if (controller.signal.aborted) {
      throw new ProxyGatewayRequestError('cancelled', 'Proxy Runtime 请求已取消')
    }
    if (err instanceof TypeError) {
      throw new ProxyGatewayRequestError('backend_unreachable', 'Proxy Runtime 后端不可达')
    }
    throw err
  } finally {
    globalThis.clearTimeout(timeoutID)
    init.signal?.removeEventListener('abort', abort)
  }
}

export const proxyGatewayJsonBody = (value: unknown) => JSON.stringify(value)

function proxyGatewayFetchHeaders(
  init: HeadersInit | undefined,
  options: ProxyGatewayFetchOptions,
) {
  const headers = proxyGatewayAuthHeaders(init)
  if (options.json) {
    headers.set('Content-Type', 'application/json')
  }
  return headers
}

function proxyGatewayErrorMessage(response: Response, text: string) {
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


export function proxyGatewayUserMessage(err: unknown) {
  if (err instanceof ProxyGatewayRequestError) return err.message
  if (err instanceof Error) return err.message
  return String(err)
}

export function isProxyGatewayCancellation(err: unknown) {
  return (
    err instanceof ProxyGatewayRequestError && err.kind === 'cancelled'
  )
}

function proxyGatewayHTTPErrorKind(status: number): ProxyGatewayErrorKind {
  if (isUnauthorizedStatus(status)) return 'unauthorized'
  if (status === 400 || status === 422) return 'validation_error'
  if (status === 502 || status === 503 || status === 504) return 'provider_error'
  return 'internal_error'
}
