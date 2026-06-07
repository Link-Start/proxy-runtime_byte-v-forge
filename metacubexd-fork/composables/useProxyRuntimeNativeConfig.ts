import type {
  ProxyRuntimeNativeFixedProxy,
  ProxyRuntimeNativeSubscription,
} from '~/composables/proxyRuntimeNativeConfigTypes'
import {
  type ProxyRuntimeNativeItemType,
  type ProxyRuntimeNativeRow,
  nativeRows,
} from '~/composables/proxyRuntimeNativeRows'

export function useProxyRuntimeNativeConfig() {
  const api = useProxyRuntimeApi()
  const fixedProxies = ref<ProxyRuntimeNativeFixedProxy[]>([])
  const subscriptions = ref<ProxyRuntimeNativeSubscription[]>([])
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
  const rows = computed(() => nativeRows(fixedProxies.value, subscriptions.value))
  const itemCount = computed(() => rows.value.length)

  async function load() {
    loading.value = true
    error.value = ''
    try {
      const config = await api.getNativeConfig()
      fixedProxies.value = config.fixed_proxies || []
      subscriptions.value = config.subscriptions || []
    } catch (err) {
      error.value = err instanceof Error ? err.message : String(err)
    } finally {
      loading.value = false
    }
  }

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
    nextFixed: ProxyRuntimeNativeFixedProxy[],
    nextSubscriptions: ProxyRuntimeNativeSubscription[],
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
      error.value = err instanceof Error ? err.message : String(err)
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

export type ProxyRuntimeNativeConfigState = ReturnType<typeof useProxyRuntimeNativeConfig>
