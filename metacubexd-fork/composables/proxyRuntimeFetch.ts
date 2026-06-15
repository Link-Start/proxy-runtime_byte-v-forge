import {
  isUnauthorizedStatus,
  proxyRuntimeAuthHeaders,
  redirectToProxyRuntimeSetup,
} from '~/composables/proxyRuntimeEndpointAuth'

export type ProxyRuntimeErrorKind =
  | 'timeout'
  | 'cancelled'
  | 'unauthorized'
  | 'backend_unreachable'
  | 'validation_error'
  | 'provider_error'
  | 'internal_error'

interface ProxyRuntimeFetchOptions {
  json?: boolean
  timeoutMs?: number
}

export interface ProxyRuntimeRequestOptions {
  signal?: AbortSignal
  timeoutMs?: number
}

export class ProxyRuntimeRequestError extends Error {
  readonly kind: ProxyRuntimeErrorKind
  readonly status?: number

  constructor(kind: ProxyRuntimeErrorKind, message: string, status?: number) {
    super(message)
    this.name = 'ProxyRuntimeRequestError'
    this.kind = kind
    this.status = status
  }
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
      const message = proxyRuntimeErrorMessage(response, text)
      const kind = proxyRuntimeHTTPErrorKind(response.status)
      if (kind === 'unauthorized') {
        redirectToProxyRuntimeSetup()
      }
      throw new ProxyRuntimeRequestError(kind, message, response.status)
    }
    if (response.status === 204 || !text.trim()) return {} as T
    return JSON.parse(text) as T
  } catch (err) {
    if (timedOut) {
      throw new ProxyRuntimeRequestError('timeout', 'Proxy Runtime 请求超时')
    }
    if (controller.signal.aborted) {
      throw new ProxyRuntimeRequestError('cancelled', 'Proxy Runtime 请求已取消')
    }
    if (err instanceof TypeError) {
      throw new ProxyRuntimeRequestError('backend_unreachable', 'Proxy Runtime 后端不可达')
    }
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


export function proxyRuntimeUserMessage(err: unknown) {
  if (err instanceof ProxyRuntimeRequestError) return err.message
  if (err instanceof Error) return err.message
  return String(err)
}

export function isProxyRuntimeCancellation(err: unknown) {
  return (
    err instanceof ProxyRuntimeRequestError && err.kind === 'cancelled'
  )
}

function proxyRuntimeHTTPErrorKind(status: number): ProxyRuntimeErrorKind {
  if (isUnauthorizedStatus(status)) return 'unauthorized'
  if (status === 400 || status === 422) return 'validation_error'
  if (status === 502 || status === 503 || status === 504) return 'provider_error'
  return 'internal_error'
}
