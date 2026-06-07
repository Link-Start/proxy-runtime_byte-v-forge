import type { ProxyIngressRuleSettings } from '~/types/byte/v/forge/contracts/proxyruntime/v1/proxy_runtime'

export const proxyRuntimePlaygroundRuleID = 'in-user-playground'
export const proxyRuntimePlaygroundUsername = 'playground'

export function isProxyRuntimePlaygroundRule(
  rule?: Pick<ProxyIngressRuleSettings, 'rule_id' | 'username'> | null,
) {
  return (
    rule?.rule_id === proxyRuntimePlaygroundRuleID ||
    rule?.username === proxyRuntimePlaygroundUsername
  )
}
