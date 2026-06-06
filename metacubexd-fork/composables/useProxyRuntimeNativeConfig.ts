import type {
  ProxyRuntimeNativeFixedProxy,
  ProxyRuntimeNativeSubscription,
} from '~/composables/proxyRuntimeNativeConfigTypes'

export function useProxyRuntimeNativeConfig() {
  const api = useProxyRuntimeApi()
  const fixedProxies = ref<ProxyRuntimeNativeFixedProxy[]>([])
  const subscriptions = ref<ProxyRuntimeNativeSubscription[]>([])
  const fixedForm = reactive({ original_name: '', name: '', uri: '' })
  const subscriptionForm = reactive({ original_name: '', name: '', url: '' })
  const loading = ref(false)
  const saving = ref(false)
  const error = ref('')
  const itemCount = computed(
    () => fixedProxies.value.length + subscriptions.value.length,
  )

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

  function resetFixedForm() {
    Object.assign(fixedForm, { original_name: '', name: '', uri: '' })
  }

  function editFixedProxy(proxy: ProxyRuntimeNativeFixedProxy) {
    Object.assign(fixedForm, {
      original_name: proxy.name,
      name: proxy.name,
      uri: proxy.uri,
    })
  }

  function resetSubscriptionForm() {
    Object.assign(subscriptionForm, { original_name: '', name: '', url: '' })
  }

  function editSubscription(subscription: ProxyRuntimeNativeSubscription) {
    Object.assign(subscriptionForm, {
      original_name: subscription.name,
      name: subscription.name,
      url: subscription.url,
    })
  }

  async function saveFixedProxy() {
    const item = { name: fixedForm.name.trim(), uri: fixedForm.uri.trim() }
    await saveConfig(
      fixedProxies.value
        .filter((proxy) => proxy.name !== fixedForm.original_name)
        .concat(item),
      subscriptions.value,
    )
    resetFixedForm()
  }

  async function saveSubscription() {
    const item = {
      name: subscriptionForm.name.trim(),
      url: subscriptionForm.url.trim(),
    }
    await saveConfig(
      fixedProxies.value,
      subscriptions.value
        .filter((subscription) => subscription.name !== subscriptionForm.original_name)
        .concat(item),
    )
    resetSubscriptionForm()
  }

  async function deleteFixedProxy(proxy: ProxyRuntimeNativeFixedProxy) {
    await saveConfig(
      fixedProxies.value.filter((item) => item.name !== proxy.name),
      subscriptions.value,
    )
  }

  async function deleteSubscription(subscription: ProxyRuntimeNativeSubscription) {
    await saveConfig(
      fixedProxies.value,
      subscriptions.value.filter((item) => item.name !== subscription.name),
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
    deleteFixedProxy,
    deleteSubscription,
    editFixedProxy,
    editSubscription,
    error,
    fixedForm,
    fixedProxies,
    itemCount,
    load,
    loading,
    resetFixedForm,
    resetSubscriptionForm,
    saveFixedProxy,
    saveSubscription,
    saving,
    subscriptionForm,
    subscriptions,
  }
}

export type ProxyRuntimeNativeConfigState = ReturnType<typeof useProxyRuntimeNativeConfig>
