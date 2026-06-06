const mihomoControllerBase = '/api/proxy-runtime/mihomo/controller'

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

export async function listMihomoEgressOwners() {
  const [proxies, providers] = await Promise.all([
    mihomoControllerRequest<MihomoProxiesResponse>('/proxies'),
    mihomoControllerRequest<MihomoProvidersResponse>('/providers/proxies'),
  ])
  return {
    proxies: proxies.proxies || {},
    providers: providers.providers || {},
  }
}

export async function listMihomoConfigNodes(ownerId: string) {
  const [{ proxies }, { providers }] = await Promise.all([
    mihomoControllerRequest<MihomoProxiesResponse>('/proxies'),
    mihomoControllerRequest<MihomoProvidersResponse>('/providers/proxies'),
  ])
  return nodesFromMihomo(ownerId, proxies || {}, providers || {})
}

async function mihomoControllerRequest<T>(path: string): Promise<T> {
  const response = await fetch(`${mihomoControllerBase}${path}`)
  if (!response.ok) {
    let message = `${response.status} ${response.statusText}`
    try {
      const body = await response.json()
      if (body?.message) message = body.message
    } catch {
      const body = await response.text()
      if (body) message = body
    }
    throw new Error(message)
  }
  return (await response.json()) as T
}

function nodesFromMihomo(
  ownerId: string,
  proxies: Record<string, MihomoProxyState>,
  providers: Record<string, MihomoProviderState>,
): MihomoConfigNodesResponse {
  const key = ownerId.trim()
  const provider = providers[key]
  if (provider) {
    return {
      nodes: (provider.proxies || []).map((node) =>
        mihomoNode(key, node.name || '', node),
      ),
    }
  }
  const proxy = proxies[key]
  if (!proxy) return { nodes: [] }
  if (proxy.all?.length) {
    return {
      nodes: proxy.all
        .map((name) => proxies[name] && mihomoNode(key, name, proxies[name]))
        .filter((node): node is MihomoConfigNode => !!node),
    }
  }
  return { nodes: [mihomoNode(key, key, proxy)] }
}

function mihomoNode(
  ownerId: string,
  name: string,
  proxy: MihomoProxyState,
): MihomoConfigNode {
  const history = proxy.history?.[proxy.history.length - 1]
  return {
    owner_id: ownerId,
    node_id: `${ownerId}/${name}`,
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
