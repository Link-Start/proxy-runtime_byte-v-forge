import type {
  ProxyDynamicLease,
  ProxyProviderDescriptor,
} from '~/types/byte/v/forge/contracts/proxyruntime/v1/proxy_runtime'
import { ProxyDynamicLeaseStatus } from '~/types/byte/v/forge/contracts/proxyruntime/v1/proxy_runtime'

export function useProxyRuntimeDynamicLeases() {
  const api = useProxyRuntimeApi()
  const providers = ref<ProxyProviderDescriptor[]>([])
  const leases = ref<ProxyDynamicLease[]>([])
  const loading = ref(false)
  const saving = ref(false)
  const error = ref('')
  const activeLeases = computed(() =>
    leases.value.filter(
      (item) =>
        item.status === ProxyDynamicLeaseStatus.PROXY_DYNAMIC_LEASE_STATUS_ACTIVE,
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

  async function releaseLease(lease: ProxyDynamicLease) {
    saving.value = true
    error.value = ''
    try {
      await api.releaseLease({
        lease_id: lease.lease_id,
        account_id: lease.account_id,
        purpose: lease.purpose,
      })
      await load()
    } catch (err) {
      error.value = err instanceof Error ? err.message : String(err)
    } finally {
      saving.value = false
    }
  }

  function providerName(providerID: string) {
    const found = providers.value.find((item) => item.provider_id === providerID)
    return found?.display_name || providerID
  }

  return {
    activeLeases,
    error,
    load,
    loading,
    providerName,
    releaseLease,
    saving,
  }
}

export type ProxyRuntimeDynamicLeasesState = ReturnType<
  typeof useProxyRuntimeDynamicLeases
>
