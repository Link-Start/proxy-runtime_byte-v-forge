export interface ProxyRuntimeNativeFixedProxy {
  id?: string
  name: string
  type?: string
  uri: string
}

export interface ProxyRuntimeNativeSubscription {
  id?: string
  name: string
  url: string
}

export interface ProxyRuntimeNativeConfig {
  fixed_proxies: ProxyRuntimeNativeFixedProxy[]
  subscriptions: ProxyRuntimeNativeSubscription[]
}
