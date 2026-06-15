import {
  isProxyRuntimeCancellation,
  proxyRuntimeUserMessage,
} from '~/composables/proxyRuntimeFetch'
import type { ProxyRuntimeStatus } from '~/types/byte/v/forge/contracts/proxyruntime/v1/proxy_runtime'

const runtimeStatusRefreshIntervalMs = 5000

export function useProxyRuntimeStatus() {
  const api = useProxyRuntimeStatusApi()
  const status = ref<ProxyRuntimeStatus>()
  const loading = ref(false)
  const error = ref('')
  let refreshTimer: ReturnType<typeof setInterval> | undefined
  let loadController: AbortController | undefined
  let loadSequence = 0

  async function load(options: { preserveError?: boolean } = {}) {
    const sequence = nextLoadSequence()
    const controller = new AbortController()
    loadController = controller
    loading.value = true
    if (!options.preserveError) error.value = ''
    try {
      const response = await api.getStatus({ signal: controller.signal })
      if (!currentLoad(sequence)) return
      status.value = response.status
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
      runtimeStatusRefreshIntervalMs,
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

  return { error, load, loading, status }
}

export type ProxyRuntimeStatusState = ReturnType<typeof useProxyRuntimeStatus>
