import type { MihomoConfigNode } from '~/composables/proxyGatewayMihomoController'
import type { ProxyGatewayMihomoNativeConfig } from '~/types/byte/v/forge/contracts/proxygateway/v1/proxy_gateway'

export interface MihomoEgressOwner {
  owner_id: string
  display_name: string
  kind: 'static_proxy' | 'proxy_provider'
}

interface MihomoOwnerSource {
  proxies: Record<string, { type?: string }>
  providers: Record<string, { name?: string; proxies?: unknown[] }>
  nativeConfig?: ProxyGatewayMihomoNativeConfig
}

const unsupportedProxyTypes = new Set([
  '',
  'compatible',
  'direct',
  'pass',
  'reject',
  'rejectdrop',
])

export function mihomoOwnerLabel(owner: MihomoEgressOwner) {
  const kind = owner.kind === 'proxy_provider' ? 'Provider' : 'Proxy'
  return `${owner.display_name || owner.owner_id} / ${kind}`
}

export function nodeLabel(node: MihomoConfigNode) {
  const name = node.display_name || node.node_id
  if (node.delay_ms > 0) return `${name} / ${node.delay_ms}ms`
  return name
}

export function mihomoOwnersFromController(input: MihomoOwnerSource) {
  return [
    ...staticProxyOwners(input),
    ...subscriptionOwners(input),
  ].sort((left, right) => left.display_name.localeCompare(right.display_name))
}

function staticProxyOwners(input: MihomoOwnerSource) {
  const out: MihomoEgressOwner[] = []
  for (const proxy of input.nativeConfig?.fixed_proxies || []) {
    const id = proxy.id || proxy.name
    if (!id || !proxy.name) continue
    if (!isLineProxy(input.proxies[proxy.name])) continue
    out.push({ owner_id: id, display_name: proxy.name, kind: 'static_proxy' })
  }
  return out
}

function subscriptionOwners(input: MihomoOwnerSource) {
  const out: MihomoEgressOwner[] = []
  for (const subscription of input.nativeConfig?.subscriptions || []) {
    const id = subscription.id || subscription.name
    if (!id || !subscription.name) continue
    const provider = input.providers[subscription.name]
    if (!provider || (provider.proxies || []).length === 0) continue
    out.push({
      owner_id: id,
      display_name: subscription.name,
      kind: 'proxy_provider',
    })
  }
  return out
}

function isLineProxy(proxy: { type?: string } | undefined) {
  if (!proxy) return false
  return !unsupportedProxyTypes.has((proxy.type || '').toLowerCase())
}
