import {
  EgressProfileExitKind,
  EgressProfileLineKind,
  ProxySessionMode,
} from '~/types/byte/v/forge/contracts/proxyruntime/v1/proxy_runtime'
import { writeProxyRuntimeClipboard } from '~/composables/proxyRuntimeClipboard'
import { isProxyRuntimePlaygroundRule } from '~/composables/proxyRuntimePlaygroundRule'

export function useProxyRuntimePlayground() {
  const runtime = useProxyRuntimeInUserRules()
  const gatewayHost = ref('')
  const gatewayPort = '31081'
  const copied = ref('')
  const leases = useProxyRuntimePlaygroundLeases(runtime, save)

  const lineUsesNode = computed(
    () =>
      runtime.form.line_kind ===
      EgressProfileLineKind.EGRESS_PROFILE_LINE_KIND_MIHOMO_NODE,
  )
  const stickyMode = computed(
    () =>
      runtime.form.exit_dynamic_session_mode ===
      ProxySessionMode.PROXY_SESSION_MODE_STICKY,
  )
  const canSave = computed(() => {
    if (!runtime.form.username.trim() || !runtime.form.password_value.trim()) {
      return false
    }
    if (lineUsesNode.value && !(runtime.form.line_resource_id && runtime.form.line_node_id)) {
      return false
    }
    if (
      runtime.form.exit_kind ===
        EgressProfileExitKind.EGRESS_PROFILE_EXIT_KIND_STATIC_IP &&
      !(runtime.form.exit_resource_id && runtime.form.exit_node_id)
    ) {
      return false
    }
    return true
  })
  const checks = useProxyRuntimePlaygroundChecks(runtime, save, canSave)
  const dynamicExit = computed(
    () =>
      runtime.form.exit_kind ===
      EgressProfileExitKind.EGRESS_PROFILE_EXIT_KIND_DYNAMIC_IP,
  )
  const proxyAuthority = computed(() => {
    const host = gatewayHost.value.trim() || '<gateway-host>'
    return `${host}:${gatewayPort}`
  })
  const credentials = computed(
    () => `${runtime.form.username}:${runtime.form.password_value}`,
  )
  const proxyURL = computed(
    () =>
      `http://${encodeURIComponent(runtime.form.username)}:${encodeURIComponent(
        runtime.form.password_value,
      )}@${proxyAuthority.value}`,
  )
  const curlCommand = computed(
    () => `curl -x '${proxyURL.value}' https://ipv4.icanhazip.com`,
  )

  onMounted(async () => {
    gatewayHost.value = window.location.hostname
    await refresh()
  })

  watch(
    () => runtime.form.exit_dynamic_provider_id,
    () => {
      runtime.form.exit_dynamic_endpoint_id = ''
    },
  )

  watch(dynamicExit, async (enabled) => {
    enforcePlaygroundSticky()
    if (enabled) await leases.load()
  })

  watch(
    () => runtime.form.exit_dynamic_session_mode,
    () => {
      enforcePlaygroundSticky()
    },
  )

  async function refresh() {
    await runtime.load()
    const created = hydratePlayground()
    if (created) {
      await save()
      return
    }
    if (dynamicExit.value) await leases.load()
  }

  function hydratePlayground() {
    const row = runtime.findRuleRow((item) => isProxyRuntimePlaygroundRule(item.rule))
    if (row) {
      runtime.editRule(row.rule)
      return enforcePlaygroundSticky()
    }
    runtime.resetForm()
    Object.assign(runtime.form, {
      display_name: 'PlayGround',
      exit_kind: EgressProfileExitKind.EGRESS_PROFILE_EXIT_KIND_DIRECT,
      exit_dynamic_session_mode: ProxySessionMode.PROXY_SESSION_MODE_STICKY,
      password_value: newPassword(),
      profile_id: 'playground-egress',
      rule_id: 'in-user-playground',
      username: 'playground',
    })
    return true
  }

  function enforcePlaygroundSticky() {
    if (
      runtime.form.exit_kind !==
      EgressProfileExitKind.EGRESS_PROFILE_EXIT_KIND_DYNAMIC_IP
    ) {
      return false
    }
    if (
      runtime.form.exit_dynamic_session_mode ===
      ProxySessionMode.PROXY_SESSION_MODE_STICKY
    ) {
      return false
    }
    runtime.form.exit_dynamic_session_mode = ProxySessionMode.PROXY_SESSION_MODE_STICKY
    return true
  }

  async function save() {
    if (!canSave.value) return
    await runtime.saveRule()
    if (!runtime.error.value) hydratePlayground()
    if (dynamicExit.value) await leases.load()
  }

  function regeneratePassword() {
    runtime.form.password_value = newPassword()
  }

  async function copyText(key: string, value: string) {
    await writeProxyRuntimeClipboard(value)
    copied.value = key
    window.setTimeout(() => {
      if (copied.value === key) copied.value = ''
    }, 1200)
  }

  return {
    canSave,
    checks,
    copied,
    copyText,
    credentials,
    curlCommand,
    dynamicExit,
    gatewayHost,
    gatewayPort,
    leases,
    proxyAuthority,
    regeneratePassword,
    refresh,
    runtime,
    save,
    stickyMode,
  }
}

function newPassword() {
  const bytes = new Uint8Array(12)
  crypto.getRandomValues(bytes)
  return Array.from(bytes, (item) => item.toString(16).padStart(2, '0')).join('')
}
