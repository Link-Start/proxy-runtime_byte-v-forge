import type {
  ProxyRuntimeNativeFixedProxy,
  ProxyRuntimeNativeSubscription,
} from '~/composables/proxyRuntimeNativeConfigTypes'

export type ProxyRuntimeNativeItemType = 'fixed_proxy' | 'subscription'

export interface ProxyRuntimeNativeRow {
  id: string
  name: string
  type: ProxyRuntimeNativeItemType
  typeLabel: string
  value: string
  valueLabel: string
}

export function nativeRows(
  fixedProxies: ProxyRuntimeNativeFixedProxy[],
  subscriptions: ProxyRuntimeNativeSubscription[],
) {
  return [
    ...fixedProxies.map((item) => fixedProxyRow(item)),
    ...subscriptions.map((item) => subscriptionRow(item)),
  ].sort((a, b) => a.name.localeCompare(b.name))
}

export function fixedProxyRow(item: ProxyRuntimeNativeFixedProxy) {
  return {
    id: nativeItemID('fixed_proxy', item.id || item.name),
    name: item.name,
    type: 'fixed_proxy' as const,
    typeLabel: '固定代理',
    value: item.uri,
    valueLabel: item.type || 'proxy',
  }
}

export function subscriptionRow(item: ProxyRuntimeNativeSubscription) {
  return {
    id: nativeItemID('subscription', item.id || item.name),
    name: item.name,
    type: 'subscription' as const,
    typeLabel: '订阅',
    value: item.url,
    valueLabel: 'proxy-provider',
  }
}

export function nativeItemID(type: ProxyRuntimeNativeItemType, id: string) {
  return `${type}:${id}`
}
