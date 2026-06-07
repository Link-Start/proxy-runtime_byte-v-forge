import type { ProxyRuntimeInUserRulesState } from '~/composables/useProxyRuntimeInUserRules'
import type { Ref } from 'vue'
import type {
  ProxyExitGeo,
  ProxyExitIP,
  ProxyIPFraudCheck,
} from '~/types/byte/v/forge/contracts/proxyruntime/v1/proxy_runtime'

export function useProxyRuntimePlaygroundChecks(
  runtime: ProxyRuntimeInUserRulesState,
  persist: () => Promise<void>,
  canSave: Ref<boolean>,
) {
  const api = useProxyRuntimeApi()
  const busy = ref(false)
  const error = ref('')
  const exitIP = ref<ProxyExitIP>()
  const geo = ref<ProxyExitGeo>()
  const fraud = ref<ProxyIPFraudCheck>()

  async function load() {
    const username = runtime.form.username.trim()
    if (!username) {
      resetResults()
      return
    }
    try {
      const result = await api.getProxyExitCheckSnapshot({
        listener_id: `in-user:${username}`,
      })
      const snapshot = result.snapshot
      if (!snapshot?.proxy_exit_ip?.ip) {
        resetResults()
        return
      }
      exitIP.value = snapshot.proxy_exit_ip
      geo.value = snapshot.proxy_exit_geo
      fraud.value = snapshot.ip_fraud_check
      error.value = ''
    } catch {
      resetResults()
    }
  }

  async function run() {
    if (!canSave.value) {
      error.value = '先补全并保存 PlayGround 配置'
      return
    }
    busy.value = true
    error.value = ''
    try {
      await persist()
      if (runtime.error.value) {
        error.value = runtime.error.value
        return
      }
      resetResults()
      const ipResult = await api.getProxyExitIP({
        listener_id: `in-user:${runtime.form.username.trim()}`,
      })
      exitIP.value = ipResult.proxy_exit_ip
      const ip = exitIP.value?.ip?.trim()
      if (!ip) {
        error.value = exitIP.value?.error_message || '未检测到出口 IP'
        return
      }
      await enrichIP(ip)
    } catch (err) {
      error.value = err instanceof Error ? err.message : String(err)
    } finally {
      busy.value = false
    }
  }

  async function enrichIP(ip: string) {
    const messages: string[] = []
    const [geoResult, fraudResult] = await Promise.allSettled([
      api.checkProxyExitGeo({ ip }),
      api.checkProxyIPFraud({ ip }),
    ])
    if (geoResult.status === 'fulfilled') {
      geo.value = geoResult.value.proxy_exit_geo
    } else {
      messages.push(`Geo: ${errorText(geoResult.reason)}`)
    }
    if (fraudResult.status === 'fulfilled') {
      fraud.value = fraudResult.value.check
    } else {
      messages.push(`Fraud: ${errorText(fraudResult.reason)}`)
    }
    error.value = messages.join('；')
  }

  function resetResults() {
    exitIP.value = undefined
    geo.value = undefined
    fraud.value = undefined
  }

  return {
    busy,
    error,
    exitIP,
    fraud,
    geo,
    load,
    run,
  }
}

function errorText(value: unknown) {
  return value instanceof Error ? value.message : String(value)
}

export type ProxyRuntimePlaygroundChecksState = ReturnType<
  typeof useProxyRuntimePlaygroundChecks
>
