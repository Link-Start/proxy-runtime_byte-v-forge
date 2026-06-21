import {
  ProxyEdgeAccessRiskLevel,
  ProxyIPFraudProviderKind,
  ProxyIPGeoProviderKind,
  ProxyIPFraudRiskLevel,
} from '~/types/byte/v/forge/contracts/proxygateway/v1/proxy_gateway'

export function fraudPluginTypeLabel(kind: ProxyIPFraudProviderKind) {
  return kind.replace('PROXY_IP_FRAUD_PROVIDER_KIND_', '').replaceAll('_', '-')
}

export function geoPluginTypeLabel(kind: ProxyIPGeoProviderKind) {
  return kind.replace('PROXY_IP_GEO_PROVIDER_KIND_', '').replaceAll('_', '-')
}

export function fraudRiskLabel(level?: ProxyIPFraudRiskLevel) {
  switch (level) {
    case ProxyIPFraudRiskLevel.PROXY_IP_FRAUD_RISK_LEVEL_LOW:
      return '低风险'
    case ProxyIPFraudRiskLevel.PROXY_IP_FRAUD_RISK_LEVEL_MEDIUM:
      return '中风险'
    case ProxyIPFraudRiskLevel.PROXY_IP_FRAUD_RISK_LEVEL_HIGH:
      return '高风险'
    case ProxyIPFraudRiskLevel.PROXY_IP_FRAUD_RISK_LEVEL_CRITICAL:
      return '严重'
    case ProxyIPFraudRiskLevel.PROXY_IP_FRAUD_RISK_LEVEL_UNSUPPORTED:
      return '未配置'
    default:
      return '未知'
  }
}

export function edgeRiskLabel(level?: ProxyEdgeAccessRiskLevel) {
  switch (level) {
    case ProxyEdgeAccessRiskLevel.PROXY_EDGE_ACCESS_RISK_LEVEL_LOW:
      return '通过'
    case ProxyEdgeAccessRiskLevel.PROXY_EDGE_ACCESS_RISK_LEVEL_MEDIUM:
      return '中风险'
    case ProxyEdgeAccessRiskLevel.PROXY_EDGE_ACCESS_RISK_LEVEL_HIGH:
      return '高风险'
    case ProxyEdgeAccessRiskLevel.PROXY_EDGE_ACCESS_RISK_LEVEL_CHALLENGE_LIKELY:
      return '可能挑战'
    case ProxyEdgeAccessRiskLevel.PROXY_EDGE_ACCESS_RISK_LEVEL_BLOCK_LIKELY:
      return '可能阻断'
    case ProxyEdgeAccessRiskLevel.PROXY_EDGE_ACCESS_RISK_LEVEL_UNSUPPORTED:
      return '未配置'
    default:
      return '未知'
  }
}

export function riskBadgeClass(value: string) {
  if (value.includes('通过') || value.includes('低')) return 'badge-success'
  if (value.includes('中') || value.includes('挑战')) return 'badge-warning'
  if (value.includes('高') || value.includes('严重') || value.includes('阻断')) return 'badge-error'
  return 'badge-ghost'
}
