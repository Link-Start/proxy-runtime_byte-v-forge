import type { ProxyIngressRuleSettings } from '~/types/byte/v/forge/contracts/proxygateway/v1/proxy_gateway'

export const proxyGatewayPlaygroundRuleID = 'in-user-playground'
export const proxyGatewayPlaygroundUsername = 'playground'
export const proxyGatewayPlaygroundProfileID = 'playground-egress'

const legacyDynamicRuleID = 'in-user-playground-dynamic'
const legacyDynamicUsername = 'playground-dynamic'

export function isProxyGatewayPlaygroundRule(
  rule?: Pick<ProxyIngressRuleSettings, 'rule_id' | 'username'> | null,
) {
  if (!rule) return false
  return (
    rule.rule_id === proxyGatewayPlaygroundRuleID ||
    rule.rule_id === legacyDynamicRuleID ||
    isPlaygroundUsername(rule.username)
  )
}

function isPlaygroundUsername(username: string | undefined) {
  const value = (username || '').trim()
  return (
    value === proxyGatewayPlaygroundUsername ||
    value.startsWith(`${proxyGatewayPlaygroundUsername}-session-`) ||
    value === legacyDynamicUsername ||
    value.startsWith(`${legacyDynamicUsername}-session-`)
  )
}
