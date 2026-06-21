import { EgressProfileExitKind, EgressProfileLineKind } from '~/types/byte/v/forge/contracts/proxygateway/v1/proxy_gateway'
import { writeProxyGatewayClipboard } from '~/composables/proxyGatewayClipboard'
import { isProxyGatewayPlaygroundRule, proxyGatewayPlaygroundProfileID, proxyGatewayPlaygroundRuleID, proxyGatewayPlaygroundUsername } from '~/composables/proxyGatewayPlaygroundRule'

export function useProxyGatewayPlayground() {
  const runtime = useProxyGatewayInUserRules()
  const gatewayHost = ref('')
  const gatewayPort = '30081'
  const copied = ref('')
  const refreshing = ref(false)
  const leases = useProxyGatewayPlaygroundLeases(runtime, () =>
    save({ refreshLeases: false }),
  )
  const dynamicExit = computed(() => runtime.form.exit_kind === EgressProfileExitKind.EGRESS_PROFILE_EXIT_KIND_DYNAMIC_IP)
  const lineUsesNode = computed(() => runtime.form.line_kind === EgressProfileLineKind.EGRESS_PROFILE_LINE_KIND_MIHOMO_NODE)
  const canSave = computed(() => {
    if (!runtime.form.username.trim() || !runtime.form.password_value.trim()) return false
    if (lineUsesNode.value && !(runtime.form.line_resource_id && runtime.form.line_node_id)) return false
    if (runtime.form.exit_kind === EgressProfileExitKind.EGRESS_PROFILE_EXIT_KIND_STATIC_IP && !(runtime.form.exit_resource_id && runtime.form.exit_node_id)) return false
    return true
  })
  const checks = useProxyGatewayPlaygroundChecks(runtime, save, canSave)
  const proxyAuthority = computed(() => `${gatewayHost.value.trim() || '<gateway-host>'}:${gatewayPort}`)
  const credentials = computed(() => `${runtime.form.username}:${runtime.form.password_value}`)
  const proxyURL = computed(() => `http://${encodeURIComponent(runtime.form.username)}:${encodeURIComponent(runtime.form.password_value)}@${proxyAuthority.value}`)
  const curlCommand = computed(() => `curl -x '${proxyURL.value}' https://ipv4.icanhazip.com`)

  onMounted(async () => {
    gatewayHost.value = window.location.hostname
    await refresh()
  })

  async function refresh() {
    if (refreshing.value) return
    refreshing.value = true
    try {
      await runtime.load()
      const changed = hydratePlayground()
      if (changed) {
        await save()
      } else {
        await leases.load()
      }
      await checks.load()
    } finally {
      refreshing.value = false
    }
  }

  function hydratePlayground() {
    const row = runtime.findRuleRow((item) => isProxyGatewayPlaygroundRule(item.rule))
    if (row) {
      runtime.editRule(row.rule)
      return normalizePlayground()
    }
    runtime.resetForm()
    Object.assign(runtime.form, {
      display_name: 'PlayGround',
      exit_kind: EgressProfileExitKind.EGRESS_PROFILE_EXIT_KIND_DIRECT,
      password_value: newPassword(),
      profile_id: proxyGatewayPlaygroundProfileID,
      rule_id: proxyGatewayPlaygroundRuleID,
      username: proxyGatewayPlaygroundUsername,
    })
    normalizePlayground()
    return true
  }

  function normalizePlayground() {
    let changed = false
    changed = assignIfChanged('rule_id', proxyGatewayPlaygroundRuleID) || changed
    changed = assignIfChanged('profile_id', proxyGatewayPlaygroundProfileID) || changed
    changed = assignIfChanged('display_name', 'PlayGround') || changed
    changed = assignIfChanged('username', proxyGatewayPlaygroundUsername) || changed
    changed = assignIfChanged('exit_dynamic_session_id', '') || changed
    return changed
  }

  async function save(options: { refreshLeases?: boolean } = {}) {
    normalizePlayground()
    if (!canSave.value) return
    await runtime.saveRule()
    if (!runtime.error.value) {
      hydratePlayground()
      if (!dynamicExit.value) await releaseActiveLeases()
      if (options.refreshLeases !== false) await leases.load()
    }
  }

  async function releaseActiveLeases() {
    await leases.load({ preserveError: true })
    for (const lease of leases.activeRows.value) await leases.release(lease)
  }

  function assignIfChanged(key: PlaygroundTextField, value: string) {
    if (runtime.form[key] === value) return false
    runtime.form[key] = value
    return true
  }

  function regeneratePassword() {
    runtime.form.password_value = newPassword()
  }

  async function copyText(key: string, value: string) {
    await writeProxyGatewayClipboard(value)
    copied.value = key
    window.setTimeout(() => {
      if (copied.value === key) copied.value = ''
    }, 1200)
  }

  return { canSave, checks, copied, copyText, credentials, curlCommand, dynamicExit, gatewayHost, gatewayPort, leases, proxyAuthority, refresh, refreshing, regeneratePassword, runtime, save }
}

function newPassword() {
  const bytes = new Uint8Array(12)
  crypto.getRandomValues(bytes)
  return Array.from(bytes, (item) => item.toString(16).padStart(2, '0')).join('')
}

type PlaygroundTextField = 'display_name' | 'exit_dynamic_session_id' | 'profile_id' | 'rule_id' | 'username'
