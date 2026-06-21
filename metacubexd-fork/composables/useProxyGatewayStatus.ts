import {
  isProxyGatewayCancellation,
  proxyGatewayUserMessage,
} from '~/composables/proxyGatewayFetch'
import type { ProxyGatewayStatus } from '~/types/byte/v/forge/contracts/proxygateway/v1/proxy_gateway'

const runtimeStatusRefreshIntervalMs = 5000

export function useProxyGatewayStatus() {
  const api = useProxyGatewayStatusApi()
  const status = ref<ProxyGatewayStatus>()
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

export type ProxyGatewayStatusState = ReturnType<typeof useProxyGatewayStatus>
