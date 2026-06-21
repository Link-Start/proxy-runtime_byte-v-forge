import {
  isProxyGatewayCancellation,
  proxyGatewayUserMessage,
  type ProxyGatewayRequestOptions,
} from '~/composables/proxyGatewayFetch'
import type { ProxyGatewayInUserRulesState } from '~/composables/useProxyGatewayInUserRules'
import type { Ref } from 'vue'
import type {
  ProxyExitGeo,
  ProxyExitIP,
  ProxyIPFraudCheck,
} from '~/types/byte/v/forge/contracts/proxygateway/v1/proxy_gateway'

export function useProxyGatewayPlaygroundChecks(
  runtime: ProxyGatewayInUserRulesState,
  persist: () => Promise<void>,
  canSave: Ref<boolean>,
) {
  const api = useProxyGatewayApi()
  const busy = ref(false)
  const error = ref('')
  const exitIP = ref<ProxyExitIP>()
  const geo = ref<ProxyExitGeo>()
  const fraud = ref<ProxyIPFraudCheck>()
  let requestController: AbortController | undefined
  let requestSequence = 0

  async function load() {
    const username = runtime.form.username.trim()
    if (!username) {
      resetResults()
      return
    }
    if (busy.value) return
    const sequence = nextRequestSequence()
    const controller = new AbortController()
    requestController = controller
    try {
      const result = await api.getProxyExitCheckSnapshot(
        { listener_id: `in-user:${username}` },
        { signal: controller.signal },
      )
      if (!currentRequest(sequence)) return
      const snapshot = result.snapshot
      if (!snapshot?.proxy_exit_ip?.ip) {
        resetResults()
        return
      }
      exitIP.value = snapshot.proxy_exit_ip
      geo.value = snapshot.proxy_exit_geo
      fraud.value = snapshot.ip_fraud_check
      error.value = ''
    } catch (err) {
      if (!currentRequest(sequence) || isProxyGatewayCancellation(err)) return
      resetResults()
    } finally {
      if (currentRequest(sequence)) requestController = undefined
    }
  }

  async function run() {
    if (!canSave.value) {
      error.value = '先补全并保存 PlayGround 配置'
      return
    }
    const sequence = nextRequestSequence()
    const controller = new AbortController()
    requestController = controller
    busy.value = true
    error.value = ''
    try {
      await persist()
      if (!currentRequest(sequence)) return
      if (runtime.error.value) {
        error.value = runtime.error.value
        return
      }
      resetResults()
      const ipResult = await api.getProxyExitIP(
        { listener_id: `in-user:${runtime.form.username.trim()}` },
        { signal: controller.signal },
      )
      if (!currentRequest(sequence)) return
      exitIP.value = ipResult.proxy_exit_ip
      const ip = exitIP.value?.ip?.trim()
      if (!ip) {
        error.value = exitIP.value?.error_message || '未检测到出口 IP'
        return
      }
      await enrichIP(ip, sequence, { signal: controller.signal })
    } catch (err) {
      if (!currentRequest(sequence) || isProxyGatewayCancellation(err)) return
      error.value = proxyGatewayUserMessage(err)
    } finally {
      if (currentRequest(sequence)) {
        busy.value = false
        requestController = undefined
      }
    }
  }

  async function enrichIP(ip: string, sequence: number, options: ProxyGatewayRequestOptions) {
    const messages: string[] = []
    const [geoResult, fraudResult] = await Promise.allSettled([
      api.checkProxyExitGeo({ ip }, options),
      api.checkProxyIPFraud({ ip }, options),
    ])
    if (!currentRequest(sequence)) return
    if (geoResult.status === 'fulfilled') {
      geo.value = geoResult.value.proxy_exit_geo
    } else if (!isProxyGatewayCancellation(geoResult.reason)) {
      messages.push(`Geo: ${errorText(geoResult.reason)}`)
    }
    if (fraudResult.status === 'fulfilled') {
      fraud.value = fraudResult.value.check
    } else if (!isProxyGatewayCancellation(fraudResult.reason)) {
      messages.push(`Fraud: ${errorText(fraudResult.reason)}`)
    }
    error.value = messages.join('；')
  }

  function abortRequest() {
    requestController?.abort()
    requestController = undefined
  }

  function nextRequestSequence() {
    abortRequest()
    requestSequence += 1
    return requestSequence
  }

  function currentRequest(sequence: number) {
    return sequence === requestSequence
  }

  function resetResults() {
    exitIP.value = undefined
    geo.value = undefined
    fraud.value = undefined
  }

  onBeforeUnmount(abortRequest)

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
  return proxyGatewayUserMessage(value)
}

export type ProxyGatewayPlaygroundChecksState = ReturnType<
  typeof useProxyGatewayPlaygroundChecks
>
