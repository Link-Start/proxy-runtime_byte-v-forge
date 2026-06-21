import type {
  ProxyGatewayMihomoNativeFixedProxy,
  ProxyGatewayMihomoNativeSubscription,
} from '~/types/byte/v/forge/contracts/proxygateway/v1/proxy_gateway'

export type ProxyGatewayNativeItemType = 'fixed_proxy' | 'subscription'

export interface ProxyGatewayNativeRow {
  id: string
  name: string
  type: ProxyGatewayNativeItemType
  typeLabel: string
  value: string
  valueLabel: string
}

export function nativeRows(
  fixedProxies: ProxyGatewayMihomoNativeFixedProxy[],
  subscriptions: ProxyGatewayMihomoNativeSubscription[],
) {
  return [
    ...fixedProxies.map((item) => fixedProxyRow(item)),
    ...subscriptions.map((item) => subscriptionRow(item)),
  ].sort((a, b) => a.name.localeCompare(b.name))
}

export function fixedProxyRow(item: ProxyGatewayMihomoNativeFixedProxy) {
  return {
    id: nativeItemID('fixed_proxy', item.id || item.name),
    name: item.name,
    type: 'fixed_proxy' as const,
    typeLabel: '固定代理',
    value: item.uri,
    valueLabel: item.type || 'proxy',
  }
}

export function subscriptionRow(item: ProxyGatewayMihomoNativeSubscription) {
  return {
    id: nativeItemID('subscription', item.id || item.name),
    name: item.name,
    type: 'subscription' as const,
    typeLabel: '代理集合订阅源',
    value: item.url,
    valueLabel: '代理集合',
  }
}

export function nativeItemID(type: ProxyGatewayNativeItemType, id: string) {
  return `${type}:${id}`
}
