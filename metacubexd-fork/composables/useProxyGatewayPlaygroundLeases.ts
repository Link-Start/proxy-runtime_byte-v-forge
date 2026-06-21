import {
  isProxyGatewayCancellation,
  proxyGatewayUserMessage,
} from '~/composables/proxyGatewayFetch'
import type { ProxyGatewayInUserRulesState } from '~/composables/useProxyGatewayInUserRules'
import {
  dynamicIPEndpointLabel,
  dynamicIPPolicy,
  endpointIDFromURL,
} from '~/composables/proxyGatewayDynamicProfilePolicyHelpers'
import {
  EgressProfileExitKind,
  ProxyDynamicLeaseStatus,
  ProxySessionMode,
  type ProxyDynamicLease,
  type ProxySessionPolicy,
} from '~/types/byte/v/forge/contracts/proxygateway/v1/proxy_gateway'

const playgroundPurpose = 'playground'
const leaseRefreshTimeoutMs = 8_000

export function useProxyGatewayPlaygroundLeases(
  runtime: ProxyGatewayInUserRulesState,
  persist: () => Promise<void>,
) {
  const leaseApi = useProxyGatewayLeaseApi()
  const leases = ref<ProxyDynamicLease[]>([])
  const loading = ref(false)
  const busy = ref(false)
  const error = ref('')
  let loadController: AbortController | undefined
  let loadSequence = 0
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
    const sequence = nextLoadSequence()
    const controller = new AbortController()
    loadController = controller
    loading.value = true
    const previousError = error.value
    if (!options.preserveError) error.value = ''
    try {
      const response = await leaseApi.listLeases(
        { status: 'active', limit: 50 },
        { signal: controller.signal, timeoutMs: leaseRefreshTimeoutMs },
      )
      if (!currentLoad(sequence)) return
      leases.value = response.leases || []
      if (options.preserveError) error.value = previousError
    } catch (err) {
      if (!currentLoad(sequence) || isProxyGatewayCancellation(err)) return
      const message = proxyGatewayUserMessage(err)
      if (options.preserveError) {
        error.value = previousError || `租约已更新，刷新失败：${message}`
        return
      }
      error.value = message
    } finally {
      if (currentLoad(sequence)) {
        loading.value = false
        loadController = undefined
      }
    }
  }

  async function acquire() {
    if (!canAcquire.value) {
      error.value = acquireDisabledReason.value
      return
    }
    await withBusy(async () => {
      await persist()
      const response = await leaseApi.acquireLease({
        account_id: profileID.value,
        purpose: playgroundPurpose,
        policy: playgroundLeasePolicy(runtime.form, profileID.value),
        force_new: false,
        selection_policy: undefined,
      })
      if (response.lease) upsertLease(response.lease)
      await load({ preserveError: true })
    })
  }

  async function release(lease: ProxyDynamicLease) {
    await withBusy(async () => {
      const response = await leaseApi.releaseLease({
        account_id: lease.account_id,
        lease_id: lease.lease_id,
        purpose: lease.purpose || playgroundPurpose,
      })
      if (response.lease) upsertLease(response.lease)
      await load()
    })
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

  onBeforeUnmount(abortLoad)

  return { acquire, acquireDisabledReason, activeRows, busy, canAcquire, currentLease, error, load, loading, profileID, release, rows }

  async function withBusy(action: () => Promise<void>) {
    busy.value = true
    error.value = ''
    try {
      await action()
    } catch (err) {
      if (!isProxyGatewayCancellation(err)) {
        error.value = proxyGatewayUserMessage(err)
      }
    } finally {
      busy.value = false
    }
  }
}

function playgroundAcquireDisabledReason(runtime: ProxyGatewayInUserRulesState) {
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

function playgroundLeasePolicy(form: ProxyGatewayInUserRulesState['form'], profileID: string): ProxySessionPolicy {
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

export type ProxyGatewayPlaygroundLeasesState = ReturnType<typeof useProxyGatewayPlaygroundLeases>
