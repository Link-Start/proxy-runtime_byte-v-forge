import type {
  ProxyRuntimeMihomoNativeFixedProxy,
  ProxyRuntimeMihomoNativeSubscription,
} from '~/types/byte/v/forge/contracts/proxyruntime/v1/proxy_runtime'

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
  fixedProxies: ProxyRuntimeMihomoNativeFixedProxy[],
  subscriptions: ProxyRuntimeMihomoNativeSubscription[],
) {
  return [
    ...fixedProxies.map((item) => fixedProxyRow(item)),
    ...subscriptions.map((item) => subscriptionRow(item)),
  ].sort((a, b) => a.name.localeCompare(b.name))
}

export function fixedProxyRow(item: ProxyRuntimeMihomoNativeFixedProxy) {
  return {
    id: nativeItemID('fixed_proxy', item.id || item.name),
    name: item.name,
    type: 'fixed_proxy' as const,
    typeLabel: '固定代理',
    value: item.uri,
    valueLabel: item.type || 'proxy',
  }
}

export function subscriptionRow(item: ProxyRuntimeMihomoNativeSubscription) {
  return {
    id: nativeItemID('subscription', item.id || item.name),
    name: item.name,
    type: 'subscription' as const,
    typeLabel: '代理集合订阅源',
    value: item.url,
    valueLabel: '代理集合',
  }
}

export function nativeItemID(type: ProxyRuntimeNativeItemType, id: string) {
  return `${type}:${id}`
}
