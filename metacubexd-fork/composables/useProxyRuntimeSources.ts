import type { ProxySourceDescriptor } from '~/types/byte/v/forge/contracts/proxyruntime/v1/proxy_runtime'
import { ProxySourceKind } from '~/types/byte/v/forge/contracts/proxyruntime/v1/proxy_runtime'
import {
  fixedFormFromSource,
  fixedSourceRequest,
  newFixedSourceForm,
  newSubscriptionForm,
  subscriptionFormFromSource,
  subscriptionRequest,
} from '~/composables/proxyRuntimeSourceHelpers'

export function useProxyRuntimeSources() {
  const api = useProxyRuntimeApi()
  const sources = ref<ProxySourceDescriptor[]>([])
  const subscriptionForm = reactive(newSubscriptionForm())
  const fixedForm = reactive(newFixedSourceForm())
  const loading = ref(false)
  const saving = ref(false)
  const error = ref('')

  const subscriptions = computed(() =>
    sources.value.filter(
      (source) =>
        source.kind ===
        ProxySourceKind.PROXY_SOURCE_KIND_SUBSCRIPTION,
    ),
  )
  const fixedSources = computed(() =>
    sources.value.filter(
      (source) =>
        source.kind ===
        ProxySourceKind.PROXY_SOURCE_KIND_FIXED_PROXY,
    ),
  )
  const sourceCount = computed(
    () => subscriptions.value.length + fixedSources.value.length,
  )

  async function load() {
    loading.value = true
    error.value = ''
    try {
      sources.value = (await api.listSources()).sources || []
    } catch (err) {
      error.value = err instanceof Error ? err.message : String(err)
    } finally {
      loading.value = false
    }
  }

  function resetSubscriptionForm() {
    Object.assign(subscriptionForm, newSubscriptionForm())
  }

  function editSubscription(source: ProxySourceDescriptor) {
    Object.assign(subscriptionForm, subscriptionFormFromSource(source))
  }

  function resetFixedForm() {
    Object.assign(fixedForm, newFixedSourceForm())
  }

  function editFixed(source: ProxySourceDescriptor) {
    Object.assign(fixedForm, fixedFormFromSource(source))
  }

  async function saveSubscription() {
    await withSave(async () => {
      await api.upsertSubscriptionSource(subscriptionRequest(subscriptionForm))
      resetSubscriptionForm()
      await load()
    })
  }

  async function saveFixed() {
    await withSave(async () => {
      await api.upsertFixedSource(fixedSourceRequest(fixedForm))
      resetFixedForm()
      await load()
    })
  }

  async function deleteSource(source: ProxySourceDescriptor) {
    await withSave(async () => {
      await api.deleteSource({ source_id: source.source_id })
      await load()
    })
  }

  return {
    deleteSource,
    editFixed,
    editSubscription,
    error,
    fixedForm,
    fixedSources,
    load,
    loading,
    resetFixedForm,
    resetSubscriptionForm,
    saveFixed,
    saveSubscription,
    saving,
    sourceCount,
    subscriptionForm,
    subscriptions,
  }

  async function withSave(action: () => Promise<void>) {
    saving.value = true
    error.value = ''
    try {
      await action()
    } catch (err) {
      error.value = err instanceof Error ? err.message : String(err)
    } finally {
      saving.value = false
    }
  }
}

export type ProxyRuntimeSourcesState = ReturnType<typeof useProxyRuntimeSources>
