import {
  isProxyRuntimeCancellation,
  proxyRuntimeUserMessage,
} from '~/composables/proxyRuntimeFetch'
import type { ProxyRuntimeMetricsSummary } from '~/types/byte/v/forge/contracts/proxyruntime/v1/proxy_runtime'
import {
  formatProxyRuntimeMetricSeconds,
  formatProxyRuntimeMetricTime,
  proxyRuntimeMetricRows,
  proxyRuntimeMetricsOverview,
} from '~/composables/proxyRuntimeMetricsRows'

const proxyRuntimeMetricsRefreshIntervalMs = 10000

export function useProxyRuntimeMetrics() {
  const api = useProxyRuntimeMetricsApi()
  const summary = ref<ProxyRuntimeMetricsSummary>()
  const loading = ref(false)
  const error = ref('')
  let refreshTimer: ReturnType<typeof setInterval> | undefined
  let loadController: AbortController | undefined
  let loadSequence = 0

  const rows = computed(() => proxyRuntimeMetricRows(summary.value))
  const overview = computed(() => proxyRuntimeMetricsOverview(rows.value))
  const slowThresholdLabel = computed(() =>
    formatProxyRuntimeMetricSeconds(summary.value?.slow_threshold_seconds || 0),
  )
  const updatedAtLabel = computed(() =>
    formatProxyRuntimeMetricTime(summary.value?.updated_at),
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
      if (!currentLoad(sequence) || isProxyRuntimeCancellation(err)) return
      error.value = proxyRuntimeUserMessage(err)
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
      proxyRuntimeMetricsRefreshIntervalMs,
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

export type ProxyRuntimeMetricsState = ReturnType<typeof useProxyRuntimeMetrics>
