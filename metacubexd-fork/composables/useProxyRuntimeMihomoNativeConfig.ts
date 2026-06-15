import {
  isProxyRuntimeCancellation,
  proxyRuntimeUserMessage,
} from '~/composables/proxyRuntimeFetch'
import type {
  ProxyRuntimeMihomoNativeFixedProxy,
  ProxyRuntimeMihomoNativeSubscription,
} from '~/types/byte/v/forge/contracts/proxyruntime/v1/proxy_runtime'
import {
  type ProxyRuntimeNativeItemType,
  type ProxyRuntimeNativeRow,
  nativeRows,
} from '~/composables/proxyRuntimeNativeRows'

export function useProxyRuntimeMihomoNativeConfig() {
  const api = useProxyRuntimeApi()
  const fixedProxies = ref<ProxyRuntimeMihomoNativeFixedProxy[]>([])
  const subscriptions = ref<ProxyRuntimeMihomoNativeSubscription[]>([])
  const form = reactive({
    editing: false,
    original_id: '',
    name: '',
    type: 'fixed_proxy' as ProxyRuntimeNativeItemType,
    value: '',
  })
  const loading = ref(false)
  const saving = ref(false)
  const error = ref('')
  let loadController: AbortController | undefined
  let loadSequence = 0
  const rows = computed(() => nativeRows(fixedProxies.value, subscriptions.value))
  const itemCount = computed(() => rows.value.length)

  async function load() {
    const sequence = nextLoadSequence()
    const controller = new AbortController()
    loadController = controller
    loading.value = true
    error.value = ''
    try {
      const config = await api.getNativeConfig({ signal: controller.signal })
      if (!currentLoad(sequence)) return
      fixedProxies.value = config.fixed_proxies || []
      subscriptions.value = config.subscriptions || []
    } catch (err) {
      if (!currentLoad(sequence) || isProxyRuntimeCancellation(err)) return
      error.value = proxyRuntimeUserMessage(err)
    } finally {
      if (currentLoad(sequence)) {
        loading.value = false
        loadController = undefined
      }
    }
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

  onBeforeUnmount(abortLoad)

  function resetForm(type: ProxyRuntimeNativeItemType = 'fixed_proxy') {
    Object.assign(form, {
      editing: false,
      name: '',
      original_id: '',
      type,
      value: '',
    })
  }

  function editRow(row: ProxyRuntimeNativeRow) {
    Object.assign(form, {
      editing: true,
      original_id: row.id.replace(`${row.type}:`, ''),
      name: row.name,
      type: row.type,
      value: row.value,
    })
  }

  async function saveRow() {
    if (form.type === 'fixed_proxy') {
      const fixed = fixedProxies.value.filter(
        (item) => itemKey(item) !== form.original_id,
      )
      fixed.push({
        id: form.original_id,
        name: form.name.trim(),
        uri: form.value.trim(),
      })
      await saveConfig(fixed, subscriptions.value)
    } else {
      const subs = subscriptions.value.filter(
        (item) => itemKey(item) !== form.original_id,
      )
      subs.push({
        id: form.original_id,
        name: form.name.trim(),
        url: form.value.trim(),
      })
      await saveConfig(fixedProxies.value, subs)
    }
    resetForm(form.type)
  }

  async function deleteRow(row: ProxyRuntimeNativeRow) {
    const key = row.id.replace(`${row.type}:`, '')
    await saveConfig(
      fixedProxies.value.filter(
        (item) => row.type !== 'fixed_proxy' || itemKey(item) !== key,
      ),
      subscriptions.value.filter(
        (item) => row.type !== 'subscription' || itemKey(item) !== key,
      ),
    )
  }

  async function saveConfig(
    nextFixed: ProxyRuntimeMihomoNativeFixedProxy[],
    nextSubscriptions: ProxyRuntimeMihomoNativeSubscription[],
  ) {
    saving.value = true
    error.value = ''
    try {
      const config = await api.updateNativeConfig({
        fixed_proxies: nextFixed,
        subscriptions: nextSubscriptions,
      })
      fixedProxies.value = config.fixed_proxies || []
      subscriptions.value = config.subscriptions || []
    } catch (err) {
      error.value = proxyRuntimeUserMessage(err)
    } finally {
      saving.value = false
    }
  }

  return {
    deleteRow,
    editRow,
    error,
    form,
    itemCount,
    load,
    loading,
    resetForm,
    rows,
    saveRow,
    saving,
  }
}

function itemKey(item: { id?: string; name: string }) {
  return item.id || item.name
}

export type ProxyRuntimeMihomoNativeState = ReturnType<typeof useProxyRuntimeMihomoNativeConfig>
