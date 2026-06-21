import {
  isProxyGatewayCancellation,
  proxyGatewayUserMessage,
} from '~/composables/proxyGatewayFetch'
import type { ProxyGatewayMetricsSummary } from '~/types/byte/v/forge/contracts/proxygateway/v1/proxy_gateway'
import {
  formatProxyGatewayMetricSeconds,
  formatProxyGatewayMetricTime,
  proxyGatewayMetricRows,
  proxyGatewayMetricsOverview,
} from '~/composables/proxyGatewayMetricsRows'

const proxyGatewayMetricsRefreshIntervalMs = 10000

export function useProxyGatewayMetrics() {
  const api = useProxyGatewayMetricsApi()
  const summary = ref<ProxyGatewayMetricsSummary>()
  const loading = ref(false)
  const error = ref('')
  let refreshTimer: ReturnType<typeof setInterval> | undefined
  let loadController: AbortController | undefined
  let loadSequence = 0

  const rows = computed(() => proxyGatewayMetricRows(summary.value))
  const overview = computed(() => proxyGatewayMetricsOverview(rows.value))
  const slowThresholdLabel = computed(() =>
    formatProxyGatewayMetricSeconds(summary.value?.slow_threshold_seconds || 0),
  )
  const updatedAtLabel = computed(() =>
    formatProxyGatewayMetricTime(summary.value?.updated_at),
  )

  async function load(options: { preserveError?: boolean } = {}) {
    const sequence = nextLoadSequence()
    const controller = new AbortController()
    loadController = controller
    loading.value = true
    if (!options.preserveError) error.value = ''
    try {
      const response = await api.getMetricsSummary({ signal: controller.signal })
      if (!currentLoad(sequence)) return
      summary.value = response.summary
      error.value = ''
    } catch (err) {
      if (!currentLoad(sequence) || isProxyGatewayCancellation(err)) return
      error.value = proxyGatewayUserMessage(err)
    } finally {
      if (currentLoad(sequence)) {
        loading.value = false
        loadController = undefined
      }
    }
  }

  function startAutoRefresh() {
    if (refreshTimer) return
    void load()
    refreshTimer = setInterval(
      () => void load({ preserveError: true }),
      proxyGatewayMetricsRefreshIntervalMs,
    )
  }

  function stopAutoRefresh() {
    if (!refreshTimer) return
    clearInterval(refreshTimer)
    refreshTimer = undefined
  }

  function abortLoad() {
    loadController?.abort()
    loadController = undefined
  }

  function nextLoadSequence() {
    abortLoad()
    loadSequence += 1
    return loadSequence
  }

  function currentLoad(sequence: number) {
    return sequence === loadSequence
  }

  onMounted(startAutoRefresh)
  onBeforeUnmount(() => {
    stopAutoRefresh()
    abortLoad()
  })

  return {
    error,
    load,
    loading,
    overview,
    rows,
    slowThresholdLabel,
    summary,
    updatedAtLabel,
  }
}

export type ProxyGatewayMetricsState = ReturnType<typeof useProxyGatewayMetrics>
