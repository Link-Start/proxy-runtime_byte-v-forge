import type {
  ProxyRuntimeMetricsSummary,
  ProxyRuntimeOperationMetric,
} from '~/types/byte/v/forge/contracts/proxyruntime/v1/proxy_runtime'

export interface ProxyRuntimeMetricRow {
  operation: string
  status: string
  count: number
  slowCount: number
  durationSeconds: number
  averageSeconds: number
}

export interface ProxyRuntimeMetricOverview {
  totalCount: number
  errorCount: number
  slowCount: number
  durationSeconds: number
  averageSeconds: number
}

const operationLabels: Record<string, string> = {
  dataplane_apply_desired_config: '数据面应用配置',
  dataplane_delete_session_route: '删除会话路由',
  dataplane_upsert_session_route: '更新会话路由',
  lease_acquire: '申请租约',
  lease_list: '读取租约',
  lease_release: '释放租约',
  lease_worker_cleanup_pending: '清理待处理租约',
  lease_worker_expire_due: '租约过期回收',
  lease_worker_restore_active: '恢复活跃租约',
  provider_fetch_base: '拉取 Provider 基础信息',
  provider_session_create: '创建 Provider 会话',
  provider_session_factory: '创建 Provider 工厂',
  provider_session_fetch: '拉取 Provider 会话',
  provider_session_release: '释放 Provider 会话',
  settings_apply: '应用设置',
}

export function proxyRuntimeMetricRows(
  summary: ProxyRuntimeMetricsSummary | undefined,
): ProxyRuntimeMetricRow[] {
  return (summary?.operations || []).map(metricRow).sort(compareMetricRows)
}

export function proxyRuntimeMetricsOverview(
  rows: ProxyRuntimeMetricRow[],
): ProxyRuntimeMetricOverview {
  const overview = rows.reduce(
    (acc, row) => {
      acc.totalCount += row.count
      acc.durationSeconds += row.durationSeconds
      acc.slowCount += row.slowCount
      if (row.status !== 'success') acc.errorCount += row.count
      return acc
    },
    { totalCount: 0, errorCount: 0, slowCount: 0, durationSeconds: 0, averageSeconds: 0 },
  )
  overview.averageSeconds =
    overview.totalCount > 0 ? overview.durationSeconds / overview.totalCount : 0
  return overview
}

export function proxyRuntimeMetricOperationLabel(operation: string) {
  return operationLabels[operation] || operation.replaceAll('_', ' ')
}

export function proxyRuntimeMetricStatusLabel(status: string) {
  switch (status) {
    case 'success':
      return '成功'
    case 'canceled':
      return '取消'
    case 'timeout':
      return '超时'
    case 'error':
      return '错误'
    default:
      return status || '未知'
  }
}

export function proxyRuntimeMetricStatusClass(status: string) {
  switch (status) {
    case 'success':
      return 'badge-success'
    case 'canceled':
      return 'badge-warning'
    case 'timeout':
      return 'badge-error'
    case 'error':
      return 'badge-error badge-outline'
    default:
      return 'badge-ghost'
  }
}

export function formatProxyRuntimeMetricCount(value: number) {
  return new Intl.NumberFormat().format(Math.round(value))
}

export function formatProxyRuntimeMetricPercent(value: number) {
  if (!Number.isFinite(value)) return '0%'
  return `${(value * 100).toFixed(value > 0 && value < 0.01 ? 2 : 1)}%`
}

export function formatProxyRuntimeMetricSeconds(value: number) {
  if (!Number.isFinite(value) || value <= 0) return '0ms'
  if (value < 1) return `${Math.round(value * 1000)}ms`
  return `${value.toFixed(value < 10 ? 2 : 1)}s`
}

export function formatProxyRuntimeMetricTime(value: string | undefined) {
  if (!value) return '-'
  const timestamp = Date.parse(value)
  if (!Number.isFinite(timestamp) || timestamp <= 0) return '-'
  return new Date(timestamp).toLocaleString()
}

function metricRow(item: ProxyRuntimeOperationMetric): ProxyRuntimeMetricRow {
  const count = metricNumber(item.count)
  const durationSeconds = metricNumber(item.duration_seconds)
  return {
    operation: item.operation || 'unknown',
    status: item.status || 'unknown',
    count,
    slowCount: metricNumber(item.slow_count),
    durationSeconds,
    averageSeconds: count > 0 ? durationSeconds / count : 0,
  }
}

function compareMetricRows(a: ProxyRuntimeMetricRow, b: ProxyRuntimeMetricRow) {
  if (b.count !== a.count) return b.count - a.count
  if (a.operation === b.operation) return a.status.localeCompare(b.status)
  return a.operation.localeCompare(b.operation)
}

function metricNumber(value: unknown) {
  if (typeof value === 'bigint') return Number(value)
  const parsed = Number(value || 0)
  return Number.isFinite(parsed) ? parsed : 0
}
