import type { ProxyRuntimeStatus } from '~/types/byte/v/forge/contracts/proxyruntime/v1/proxy_runtime'

const runtimeStatusRefreshIntervalMs = 5000

export function useProxyRuntimeStatus() {
  const api = useProxyRuntimeStatusApi()
  const status = ref<ProxyRuntimeStatus>()
  const loading = ref(false)
  const error = ref('')
  let refreshTimer: ReturnType<typeof setInterval> | undefined

  async function load(options: { preserveError?: boolean } = {}) {
    loading.value = true
    if (!options.preserveError) error.value = ''
    try {
      status.value = (await api.getStatus()).status
      error.value = ''
    } catch (err) {
      error.value = err instanceof Error ? err.message : String(err)
    } finally {
      loading.value = false
    }
  }

  function startAutoRefresh() {
    if (refreshTimer) return
    void load()
    refreshTimer = setInterval(
      () => void load({ preserveError: true }),
      runtimeStatusRefreshIntervalMs,
    )
  }

  function stopAutoRefresh() {
    if (!refreshTimer) return
    clearInterval(refreshTimer)
    refreshTimer = undefined
  }

  onMounted(startAutoRefresh)
  onBeforeUnmount(stopAutoRefresh)

  return { error, load, loading, status }
}

export type ProxyRuntimeStatusState = ReturnType<typeof useProxyRuntimeStatus>
