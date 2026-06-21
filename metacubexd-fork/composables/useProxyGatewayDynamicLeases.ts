import {
  isProxyGatewayCancellation,
  proxyGatewayUserMessage,
} from '~/composables/proxyGatewayFetch'
import type {
  ProxyDynamicLease,
  ProxyProviderDescriptor,
} from '~/types/byte/v/forge/contracts/proxygateway/v1/proxy_gateway'
import { ProxyDynamicLeaseStatus } from '~/types/byte/v/forge/contracts/proxygateway/v1/proxy_gateway'

const refreshIntervalMs = 15_000
const leaseRefreshTimeoutMs = 8_000

export function useProxyGatewayDynamicLeases() {
  const api = useProxyGatewayApi()
  const leaseApi = useProxyGatewayLeaseApi()
  const providers = ref<ProxyProviderDescriptor[]>([])
  const leases = ref<ProxyDynamicLease[]>([])
  const loading = ref(false)
  const busyLeaseID = ref('')
  const error = ref('')
  let refreshTimer: ReturnType<typeof setInterval> | undefined
  let loadController: AbortController | undefined
  let loadSequence = 0
  const activeLeases = computed(() =>
    leases.value.filter(isActiveLease),
  )

  async function load() {
    const sequence = nextLoadSequence()
    const controller = new AbortController()
    loadController = controller
    loading.value = true
    error.value = ''
    try {
      const [providerRes, leaseRes] = await Promise.all([
        api.listProviders({ signal: controller.signal }),
        leaseApi.listLeases(
          { status: 'active', limit: 50 },
          { signal: controller.signal, timeoutMs: leaseRefreshTimeoutMs },
        ),
      ])
      if (!currentLoad(sequence)) return
      providers.value = providerRes.providers || []
      leases.value = leaseRes.leases || []
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

  async function release(lease: ProxyDynamicLease) {
    busyLeaseID.value = lease.lease_id
    error.value = ''
    try {
      const response = await leaseApi.releaseLease({
        account_id: lease.account_id,
        lease_id: lease.lease_id,
        purpose: lease.purpose,
      })
      if (response.lease) upsertLease(response.lease)
      await load()
    } catch (err) {
      if (!isProxyGatewayCancellation(err)) {
        error.value = proxyGatewayUserMessage(err)
      }
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

  function upsertLease(lease: ProxyDynamicLease) {
    const index = leases.value.findIndex((item) => item.lease_id === lease.lease_id)
    if (index >= 0) {
      leases.value.splice(index, 1, lease)
      return
    }
    leases.value = [lease, ...leases.value]
  }

  onMounted(startAutoRefresh)
  onBeforeUnmount(() => {
    stopAutoRefresh()
    abortLoad()
  })

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

export type ProxyGatewayDynamicLeasesState = ReturnType<
  typeof useProxyGatewayDynamicLeases
>
