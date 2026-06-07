import type { ProxyIngressRuleSettings } from '~/types/byte/v/forge/contracts/proxyruntime/v1/proxy_runtime'

export const proxyRuntimePlaygroundRuleID = 'in-user-playground'
export const proxyRuntimePlaygroundUsername = 'playground'
export const proxyRuntimePlaygroundProfileID = 'playground-egress'

const legacyDynamicRuleID = 'in-user-playground-dynamic'
const legacyDynamicUsername = 'playground-dynamic'

export function isProxyRuntimePlaygroundRule(
  rule?: Pick<ProxyIngressRuleSettings, 'rule_id' | 'username'> | null,
) {
  if (!rule) return false
  return (
    rule.rule_id === proxyRuntimePlaygroundRuleID ||
    rule.rule_id === legacyDynamicRuleID ||
    isPlaygroundUsername(rule.username)
  )
}

function isPlaygroundUsername(username: string | undefined) {
  const value = (username || '').trim()
  return (
    value === proxyRuntimePlaygroundUsername ||
    value.startsWith(`${proxyRuntimePlaygroundUsername}-session-`) ||
    value === legacyDynamicUsername ||
    value.startsWith(`${legacyDynamicUsername}-session-`)
  )
}
