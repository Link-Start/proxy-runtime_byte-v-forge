import type { ProxyRuntimeInUserRulesState } from '~/composables/useProxyRuntimeInUserRules'
import {
  dynamicIPEndpointLabel,
  dynamicIPPolicy,
  endpointIDFromURL,
} from '~/composables/proxyRuntimeDynamicProfilePolicyHelpers'
import {
  EgressProfileExitKind,
  ProxyDynamicLeaseStatus,
  ProxySessionMode,
  type ProxyDynamicLease,
  type ProxySessionPolicy,
} from '~/types/byte/v/forge/contracts/proxyruntime/v1/proxy_runtime'

const playgroundPurpose = 'playground'

export function useProxyRuntimePlaygroundLeases(
  runtime: ProxyRuntimeInUserRulesState,
  persist: () => Promise<void>,
) {
  const api = useProxyRuntimeApi()
  const leases = ref<ProxyDynamicLease[]>([])
  const loading = ref(false)
  const busy = ref(false)
  const error = ref('')
  const profileID = computed(() => runtime.form.profile_id.trim() || 'playground-egress')
  const rows = computed(() =>
    leases.value
      .filter((lease) => lease.account_id === profileID.value)
      .sort((left, right) => timeValue(right.acquired_at) - timeValue(left.acquired_at)),
  )
  const activeRows = computed(() => rows.value.filter(isActiveLease))
  const currentLease = computed(() => activeRows.value[0])
  const canAcquire = computed(() => playgroundAcquireDisabledReason(runtime) === '' && activeRows.value.length === 0)
  const acquireDisabledReason = computed(() => activeRows.value.length > 0 ? 'PlayGround 已有活跃租约' : playgroundAcquireDisabledReason(runtime))

  async function load(options: { preserveError?: boolean } = {}) {
    loading.value = true
    const previousError = error.value
    if (!options.preserveError) error.value = ''
    try {
      leases.value =
        (await api.listLeases({ status: 'active', limit: 50 })).leases || []
      if (options.preserveError) error.value = previousError
    } catch (err) {
      error.value = err instanceof Error ? err.message : String(err)
    } finally {
      loading.value = false
    }
  }

  async function acquire() {
    if (!canAcquire.value) {
      error.value = acquireDisabledReason.value
      return
    }
    await withBusy(async () => {
      await persist()
      await api.acquireLease({
        account_id: profileID.value,
        purpose: playgroundPurpose,
        policy: playgroundLeasePolicy(runtime.form, profileID.value),
        force_new: false,
        selection_policy: undefined,
      })
      await load({ preserveError: true })
    })
  }

  async function release(lease: ProxyDynamicLease) {
    await withBusy(async () => {
      await api.releaseLease({
        account_id: lease.account_id,
        lease_id: lease.lease_id,
        purpose: lease.purpose || playgroundPurpose,
      })
      await load()
    })
  }

  return { acquire, acquireDisabledReason, activeRows, busy, canAcquire, currentLease, error, load, loading, profileID, release, rows }

  async function withBusy(action: () => Promise<void>) {
    busy.value = true
    error.value = ''
    try {
      await action()
    } catch (err) {
      error.value = err instanceof Error ? err.message : String(err)
    } finally {
      busy.value = false
    }
  }
}

function playgroundAcquireDisabledReason(runtime: ProxyRuntimeInUserRulesState) {
  if (runtime.form.exit_kind !== EgressProfileExitKind.EGRESS_PROFILE_EXIT_KIND_DYNAMIC_IP) return 'PlayGround 未配置动态 IP 出口'
  if (runtime.form.exit_dynamic_session_mode !== ProxySessionMode.PROXY_SESSION_MODE_STICKY) return 'PlayGround 租约只支持粘性动态 IP'
  const providers = runtime.dynamicProviderOptions.value.filter((provider) => (provider.endpoints || []).length > 0)
  if (providers.length === 0) return '没有可用的动态 IP Provider'
  const providerID = runtime.form.exit_dynamic_provider_id.trim()
  const selectedProvider = providerID ? providers.find((provider) => provider.dynamic_provider_id === providerID) : undefined
  if (providerID && !selectedProvider) return '当前动态 IP Provider 不可用'
  const endpointID = runtime.form.exit_dynamic_endpoint_id.trim()
  if (!endpointID) return ''
  const endpointProviders = selectedProvider ? [selectedProvider] : providers
  const endpointExists = endpointProviders.some((provider) => (provider.endpoints || []).some((endpoint) => endpointIDFromURL(endpoint.endpoint_url || '') === endpointID))
  return endpointExists ? '' : '当前动态 IP 端点不可用'
}

function playgroundLeasePolicy(form: ProxyRuntimeInUserRulesState['form'], profileID: string): ProxySessionPolicy {
  const policy = dynamicIPPolicy(form) as ProxySessionPolicy
  policy.labels = { ...(policy.labels || {}), selection_seed: profileID }
  setLabel(policy, 'dynamic_provider_id', form.exit_dynamic_provider_id)
  setLabel(policy, dynamicIPEndpointLabel, form.exit_dynamic_endpoint_id)
  return policy
}

function setLabel(policy: ProxySessionPolicy, key: string, value: string) {
  const normalized = value.trim()
  if (normalized) policy.labels[key] = normalized
}

function timeValue(value: string | undefined) {
  return value ? Date.parse(value) || 0 : 0
}

function isActiveLease(lease: ProxyDynamicLease) {
  if (lease.status !== ProxyDynamicLeaseStatus.PROXY_DYNAMIC_LEASE_STATUS_ACTIVE) {
    return false
  }
  const expiresAt = timeValue(lease.expires_at)
  return expiresAt === 0 || expiresAt > Date.now()
}

export type ProxyRuntimePlaygroundLeasesState = ReturnType<typeof useProxyRuntimePlaygroundLeases>
