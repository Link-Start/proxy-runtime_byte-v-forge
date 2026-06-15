import type {
  GetProxyRuntimeMihomoNativeConfigResponse,
  ProxyRuntimeMihomoNativeConfig,
} from '~/types/byte/v/forge/contracts/proxyruntime/v1/proxy_runtime'
import {
  proxyRuntimeFetchJson,
  type ProxyRuntimeRequestOptions,
} from '~/composables/proxyRuntimeFetch'

const mihomoControllerBase = '/mihomo/controller'
const proxyRuntimeBase = '/api'

export interface MihomoConfigNode {
  owner_id: string
  node_id: string
  display_name: string
  node_type: string
  status: string
  delay_ms: number
  error_message?: string
}

export interface MihomoConfigNodesResponse {
  nodes: MihomoConfigNode[]
}

interface MihomoProxyState {
  name?: string
  type?: string
  alive?: boolean
  hidden?: boolean
  all?: string[]
  history?: { delay?: number; message?: string }[]
  'provider-name'?: string
}

interface MihomoProxiesResponse {
  proxies?: Record<string, MihomoProxyState>
}

interface MihomoProviderState {
  name?: string
  type?: string
  vehicleType?: string
  proxies?: MihomoProxyState[]
}

interface MihomoProvidersResponse {
  providers?: Record<string, MihomoProviderState>
}

export async function listMihomoEgressOwners(
  options: ProxyRuntimeRequestOptions = {},
) {
  const [proxies, providers, nativeConfig] = await Promise.all([
    mihomoControllerRequest<MihomoProxiesResponse>('/proxies', options),
    mihomoControllerRequest<MihomoProvidersResponse>('/providers/proxies', options),
    getMihomoNativeConfig(options),
  ])
  return {
    proxies: proxies.proxies || {},
    providers: providers.providers || {},
    nativeConfig,
  }
}

export async function listMihomoConfigNodes(
  ownerId: string,
  options: ProxyRuntimeRequestOptions = {},
) {
  const [{ proxies }, { providers }, nativeConfig] = await Promise.all([
    mihomoControllerRequest<MihomoProxiesResponse>('/proxies', options),
    mihomoControllerRequest<MihomoProvidersResponse>('/providers/proxies', options),
    getMihomoNativeConfig(options),
  ])
  return nodesFromMihomo(ownerId, proxies || {}, providers || {}, nativeConfig)
}

async function getMihomoNativeConfig(
  options: ProxyRuntimeRequestOptions = {},
): Promise<ProxyRuntimeMihomoNativeConfig> {
  const response = await proxyRuntimeRequest<GetProxyRuntimeMihomoNativeConfigResponse>(
    '/settings/mihomo-native',
    options,
  )
  return response.config || { fixed_proxies: [], subscriptions: [] }
}

async function mihomoControllerRequest<T>(
  path: string,
  options: ProxyRuntimeRequestOptions = {},
): Promise<T> {
  return proxyRuntimeFetchJson<T>(
    mihomoControllerBase,
    path,
    { signal: options.signal },
    { timeoutMs: options.timeoutMs },
  )
}

async function proxyRuntimeRequest<T>(
  path: string,
  options: ProxyRuntimeRequestOptions = {},
): Promise<T> {
  return proxyRuntimeFetchJson<T>(
    proxyRuntimeBase,
    path,
    { signal: options.signal },
    { timeoutMs: options.timeoutMs },
  )
}

function nodesFromMihomo(
  ownerId: string,
  proxies: Record<string, MihomoProxyState>,
  providers: Record<string, MihomoProviderState>,
  nativeConfig: ProxyRuntimeMihomoNativeConfig,
): MihomoConfigNodesResponse {
  const key = ownerId.trim()
  const owner = nativeOwner(key, nativeConfig)
  if (owner?.kind === 'static_proxy') {
    const proxy = proxies[owner.name]
    if (!proxy) return { nodes: [] }
    return { nodes: [mihomoNode(key, owner.id, owner.name, proxy)] }
  }
  const providerKey = owner?.name || key
  const provider = providers[providerKey]
  if (provider) {
    return {
      nodes: (provider.proxies || []).map((node) =>
        mihomoNode(key, `${key}/${node.name || ''}`, node.name || '', node),
      ),
    }
  }
  const proxy = proxies[key]
  if (!proxy) return { nodes: [] }
  if (proxy.all?.length) {
    return {
      nodes: proxy.all
        .map((name) => proxies[name] && mihomoNode(key, `${key}/${name}`, name, proxies[name]))
        .filter((node): node is MihomoConfigNode => !!node),
    }
  }
  return { nodes: [mihomoNode(key, key, key, proxy)] }
}

function mihomoNode(
  ownerId: string,
  nodeId: string,
  name: string,
  proxy: MihomoProxyState,
): MihomoConfigNode {
  const history = proxy.history?.[proxy.history.length - 1]
  return {
    owner_id: ownerId,
    node_id: nodeId,
    display_name: name,
    node_type: proxy.type || '',
    status:
      typeof proxy.alive === 'boolean'
        ? proxy.alive
          ? 'available'
          : 'unavailable'
        : 'unknown',
    delay_ms: history?.delay || 0,
    error_message: history?.message || '',
  }
}

function nativeOwner(
  ownerId: string,
  nativeConfig: ProxyRuntimeMihomoNativeConfig,
) {
  for (const proxy of nativeConfig.fixed_proxies || []) {
    if (proxy.id === ownerId || proxy.name === ownerId) {
      return {
        id: proxy.id || proxy.name,
        name: proxy.name,
        kind: 'static_proxy' as const,
      }
    }
  }
  for (const subscription of nativeConfig.subscriptions || []) {
    if (subscription.id === ownerId || subscription.name === ownerId) {
      return {
        id: subscription.id || subscription.name,
        name: subscription.name,
        kind: 'proxy_provider' as const,
      }
    }
  }
  return undefined
}
