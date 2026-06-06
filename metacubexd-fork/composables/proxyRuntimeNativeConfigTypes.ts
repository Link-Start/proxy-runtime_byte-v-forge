export interface ProxyRuntimeNativeFixedProxy {
  name: string
  type?: string
  uri: string
}

export interface ProxyRuntimeNativeSubscription {
  name: string
  url: string
}

export interface ProxyRuntimeNativeConfig {
  fixed_proxies: ProxyRuntimeNativeFixedProxy[]
  subscriptions: ProxyRuntimeNativeSubscription[]
}
