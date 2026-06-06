import type {
  ProxyDynamicLease,
  ProxyProviderDescriptor,
} from '~/types/byte/v/forge/contracts/proxyruntime/v1/proxy_runtime'
import { ProxyDynamicLeaseStatus } from '~/types/byte/v/forge/contracts/proxyruntime/v1/proxy_runtime'

const refreshIntervalMs = 15_000

export function useProxyRuntimeDynamicLeases() {
  const api = useProxyRuntimeApi()
  const providers = ref<ProxyProviderDescriptor[]>([])
  const leases = ref<ProxyDynamicLease[]>([])
  const loading = ref(false)
  const error = ref('')
  const now = ref(Date.now())
  let refreshTimer: ReturnType<typeof setInterval> | undefined
  const activeLeases = computed(() =>
    leases.value.filter(
      (item) =>
        item.status ===
          ProxyDynamicLeaseStatus.PROXY_DYNAMIC_LEASE_STATUS_ACTIVE &&
        leaseNotExpired(item, now.value),
    ),
  )

  async function load() {
    loading.value = true
    error.value = ''
    try {
      const [providerRes, leaseRes] = await Promise.all([
        api.listProviders(),
        api.listLeases(),
      ])
      providers.value = providerRes.providers || []
      leases.value = leaseRes.leases || []
    } catch (err) {
      error.value = err instanceof Error ? err.message : String(err)
    } finally {
      loading.value = false
    }
  }

  function providerName(providerID: string) {
    const found = providers.value.find((item) => item.provider_id === providerID)
    return found?.display_name || providerID
  }

  function refreshNow() {
    now.value = Date.now()
  }

  function startAutoRefresh() {
    if (refreshTimer) return
    refreshTimer = setInterval(() => {
      refreshNow()
      void load()
    }, refreshIntervalMs)
  }

  function stopAutoRefresh() {
    if (!refreshTimer) return
    clearInterval(refreshTimer)
    refreshTimer = undefined
  }

  onMounted(() => {
    refreshNow()
    startAutoRefresh()
  })

  onBeforeUnmount(stopAutoRefresh)

  return {
    activeLeases,
    error,
    leases,
    load,
    loading,
    providerName,
  }
}

export type ProxyRuntimeDynamicLeasesState = ReturnType<
  typeof useProxyRuntimeDynamicLeases
>

function leaseNotExpired(lease: ProxyDynamicLease, now: number) {
  if (!lease.expires_at) return true
  const expiresAt = Date.parse(lease.expires_at)
  if (Number.isNaN(expiresAt)) return true
  return expiresAt > now
}
