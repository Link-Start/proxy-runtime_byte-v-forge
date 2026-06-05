import type {
  EgressProfileSettings,
  ProxyDynamicIPProviderSettings,
  ProxyProviderDescriptor,
  ProxySourceDescriptor,
  ProxySourceNode,
} from '~/types/byte/v/forge/contracts/proxyruntime/v1/proxy_runtime'
import { ProxySourceKind } from '~/types/byte/v/forge/contracts/proxyruntime/v1/proxy_runtime'
import {
  newEgressProfileForm,
  profileFormFromSettings,
  profileRequest,
} from '~/composables/proxyRuntimeEgressProfileHelpers'

export function useProxyRuntimeEgressProfiles() {
  const api = useProxyRuntimeApi()
  const profiles = ref<EgressProfileSettings[]>([])
  const providers = ref<ProxyProviderDescriptor[]>([])
  const dynamicProviderIDs = ref<string[]>([])
  const sources = ref<ProxySourceDescriptor[]>([])
  const sourceNodes = reactive<Record<string, ProxySourceNode[]>>({})
  const sourceNodeLoading = reactive<Record<string, boolean>>({})
  const form = reactive(newEgressProfileForm())
  const loading = ref(false)
  const saving = ref(false)
  const error = ref('')
  const profileCount = computed(() => profiles.value.length)
  const lineSources = computed(() => sources.value.filter(isLineSource))
  const staticExitSources = computed(() => sources.value.filter(isLineSource))
  const dynamicProviderOptions = computed(() => {
    const ids = new Set(dynamicProviderIDs.value)
    return [...ids].map((id) => providerByID(id))
  })

  async function load() {
    loading.value = true
    error.value = ''
    try {
      const [settings, sourceList, providerList] = await Promise.all([
        api.getSettings(),
        api.listSources(),
        api.listProviders(),
      ])
      profiles.value = settings.settings?.egress_profiles || []
      sources.value = (sourceList.sources || []).filter((source) => source.enabled)
      providers.value = providerList.providers || []
      dynamicProviderIDs.value = dynamicProviderIDsFrom(
        settings.settings?.dynamic_ip_providers || [],
        sources.value,
      )
    } catch (err) {
      error.value = err instanceof Error ? err.message : String(err)
    } finally {
      loading.value = false
    }
  }

  function resetForm() {
    Object.assign(form, newEgressProfileForm())
  }

  function editProfile(profile: EgressProfileSettings) {
    Object.assign(form, profileFormFromSettings(profile))
  }

  async function saveProfile() {
    await withSave(async () => {
      const next = profileRequest(form)
      const existing = profiles.value.filter(
        (profile) => profile.profile_id !== next.profile_id,
      )
      await api.updateEgressProfiles([...existing, next])
      resetForm()
      await load()
    })
  }

  async function deleteProfile(profile: EgressProfileSettings) {
    await withSave(async () => {
      await api.updateEgressProfiles(
        profiles.value.filter((item) => item.profile_id !== profile.profile_id),
      )
      await load()
    })
  }

  async function loadSourceNodes(sourceID: string) {
    const key = sourceID.trim()
    if (!key || sourceNodes[key] || sourceNodeLoading[key]) return
    sourceNodeLoading[key] = true
    try {
      const response = await api.listSourceNodes(key)
      sourceNodes[key] = response.nodes || []
    } catch (err) {
      error.value = err instanceof Error ? err.message : String(err)
    } finally {
      sourceNodeLoading[key] = false
    }
  }

  function nodesForSource(sourceID: string) {
    return sourceNodes[sourceID.trim()] || []
  }

  function sourceNodesLoading(sourceID: string) {
    return !!sourceNodeLoading[sourceID.trim()]
  }

  function providerByID(id: string): ProxyProviderDescriptor {
    const provider = providers.value.find((item) => item.provider_id === id)
    return provider || emptyProviderDescriptor(id)
  }

  return {
    deleteProfile,
    dynamicProviderOptions,
    editProfile,
    error,
    form,
    load,
    lineSources,
    loadSourceNodes,
    loading,
    nodesForSource,
    profileCount,
    profiles,
    resetForm,
    saveProfile,
    saving,
    sources,
    sourceNodesLoading,
    staticExitSources,
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

function dynamicProviderIDsFrom(
  settings: ProxyDynamicIPProviderSettings[],
  sources: ProxySourceDescriptor[],
) {
  const ids = new Set<string>()
  for (const item of settings) {
    if (item.provider_id) ids.add(item.provider_id)
  }
  for (const source of sources) {
    if (source.kind === ProxySourceKind.PROXY_SOURCE_KIND_DYNAMIC_IP) {
      ids.add(source.provider_id)
    }
  }
  return [...ids].sort()
}

function emptyProviderDescriptor(id: string): ProxyProviderDescriptor {
  return {
    provider_id: id,
    display_name: id,
    capabilities: [],
    protocols: [],
    min_sticky_ttl: undefined,
    max_sticky_ttl: undefined,
    upstream_kinds: [],
    rotation_modes: [],
  }
}

function isLineSource(source: ProxySourceDescriptor) {
  return (
    source.enabled &&
    (source.kind === ProxySourceKind.PROXY_SOURCE_KIND_FIXED_PROXY ||
      source.kind === ProxySourceKind.PROXY_SOURCE_KIND_SUBSCRIPTION)
  )
}

export type ProxyRuntimeEgressProfilesState = ReturnType<
  typeof useProxyRuntimeEgressProfiles
>
