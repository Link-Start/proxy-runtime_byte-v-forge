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
  const busyLeaseID = ref('')
  const error = ref('')
  let refreshTimer: ReturnType<typeof setInterval> | undefined
  const activeLeases = computed(() =>
    leases.value.filter(isActiveLease),
  )

  async function load() {
    loading.value = true
    error.value = ''
    try {
      const [providerRes, leaseRes] = await Promise.all([
        api.listProviders(),
        api.listLeases({ status: 'active', limit: 50 }),
      ])
      providers.value = providerRes.providers || []
      leases.value = leaseRes.leases || []
    } catch (err) {
      error.value = err instanceof Error ? err.message : String(err)
    } finally {
      loading.value = false
    }
  }

  async function release(lease: ProxyDynamicLease) {
    busyLeaseID.value = lease.lease_id
    error.value = ''
    try {
      await api.releaseLease({
        account_id: lease.account_id,
        lease_id: lease.lease_id,
        purpose: lease.purpose,
      })
      await load()
    } catch (err) {
      error.value = err instanceof Error ? err.message : String(err)
    } finally {
      busyLeaseID.value = ''
    }
  }

  function providerName(providerID: string) {
    const found = providers.value.find((item) => item.provider_id === providerID)
    return found?.display_name || providerID
  }

  function startAutoRefresh() {
    if (refreshTimer) return
    refreshTimer = setInterval(() => void load(), refreshIntervalMs)
  }

  function stopAutoRefresh() {
    if (!refreshTimer) return
    clearInterval(refreshTimer)
    refreshTimer = undefined
  }

  onMounted(startAutoRefresh)
  onBeforeUnmount(stopAutoRefresh)

  return {
    activeLeases,
    busyLeaseID,
    error,
    leases,
    load,
    loading,
    providerName,
    release,
  }
}

function isActiveLease(lease: ProxyDynamicLease) {
  if (lease.status !== ProxyDynamicLeaseStatus.PROXY_DYNAMIC_LEASE_STATUS_ACTIVE) {
    return false
  }
  const expiresAt = lease.expires_at ? Date.parse(lease.expires_at) || 0 : 0
  return expiresAt === 0 || expiresAt > Date.now()
}

export type ProxyRuntimeDynamicLeasesState = ReturnType<
  typeof useProxyRuntimeDynamicLeases
>
