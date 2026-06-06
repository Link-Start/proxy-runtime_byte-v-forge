import type {
  EgressProfileSettings,
  ProxyDynamicIPProviderSettings,
  ProxyIngressRuleSettings,
} from '~/types/byte/v/forge/contracts/proxyruntime/v1/proxy_runtime'
import type { MihomoConfigNode } from '~/composables/proxyRuntimeMihomoController'
import {
  type MihomoEgressOwner,
  mihomoOwnersFromController,
  newEgressProfileForm,
  profileFormFromSettings,
  profileRequest,
} from '~/composables/proxyRuntimeEgressProfileHelpers'

export function useProxyRuntimeInUserRules() {
  const api = useProxyRuntimeApi()
  const rules = ref<ProxyIngressRuleSettings[]>([])
  const profiles = ref<EgressProfileSettings[]>([])
  const dynamicProviders = ref<ProxyDynamicIPProviderSettings[]>([])
  const owners = ref<MihomoEgressOwner[]>([])
  const mihomoNodes = reactive<Record<string, MihomoConfigNode[]>>({})
  const mihomoNodeLoading = reactive<Record<string, boolean>>({})
  const form = reactive(newInUserRuleForm())
  const loading = ref(false)
  const saving = ref(false)
  const error = ref('')
  const ruleCount = computed(() => rules.value.length)
  const rows = computed(() =>
    rules.value.map((rule) => ({
      rule,
      profile: profiles.value.find(
        (profile) => profile.profile_id === rule.profile_id,
      ),
    })),
  )

  async function load() {
    loading.value = true
    error.value = ''
    try {
      const [settings, mihomoOwners] = await Promise.all([
        api.getSettings(),
        api.listMihomoEgressOwners(),
      ])
      rules.value = settings.settings?.ingress_rules || []
      profiles.value = settings.settings?.egress_profiles || []
      dynamicProviders.value = settings.settings?.dynamic_ip_providers || []
      owners.value = mihomoOwnersFromController(mihomoOwners)
    } catch (err) {
      error.value = err instanceof Error ? err.message : String(err)
    } finally {
      loading.value = false
    }
  }

  function resetForm() {
    Object.assign(form, newInUserRuleForm())
  }

  function editRule(rule: ProxyIngressRuleSettings) {
    const profile = profiles.value.find(
      (item) => item.profile_id === rule.profile_id,
    )
    const profileForm = profile
      ? profileFormFromSettings(profile)
      : newEgressProfileForm()
    Object.assign(form, newInUserRuleForm(), profileForm, {
      display_name: rule.display_name || profile?.display_name || rule.username,
      enabled: rule.enabled,
      password_value: rule.password_value,
      profile_id: rule.profile_id,
      rule_id: rule.rule_id,
      username: rule.username,
    })
  }

  async function saveRule() {
    await withSave(async () => {
      const next = formRequest(form)
      const nextProfiles = profiles.value.filter(
        (profile) => profile.profile_id !== next.profile.profile_id,
      )
      const nextRules = rules.value.filter(
        (rule) => rule.rule_id !== next.rule.rule_id,
      )
      await api.updateInUserRules(
        [...nextProfiles, next.profile],
        [...nextRules, next.rule],
      )
      resetForm()
      await load()
    })
  }

  async function deleteRule(rule: ProxyIngressRuleSettings) {
    await withSave(async () => {
      const nextRules = rules.value.filter((item) => item.rule_id !== rule.rule_id)
      const profileStillUsed = nextRules.some(
        (item) => item.profile_id === rule.profile_id,
      )
      const nextProfiles = profileStillUsed
        ? profiles.value
        : profiles.value.filter(
            (profile) => profile.profile_id !== rule.profile_id,
          )
      await api.updateInUserRules(nextProfiles, nextRules)
      await load()
    })
  }

  async function loadMihomoNodes(ownerID: string) {
    const key = ownerID.trim()
    if (!key || mihomoNodes[key] || mihomoNodeLoading[key]) return
    mihomoNodeLoading[key] = true
    try {
      const response = await api.listMihomoConfigNodes(key)
      mihomoNodes[key] = response.nodes || []
    } catch (err) {
      error.value = err instanceof Error ? err.message : String(err)
    } finally {
      mihomoNodeLoading[key] = false
    }
  }

  return {
    deleteRule,
    dynamicProviderOptions: computed(() => dynamicProviders.value),
    editRule,
    error,
    form,
    lineSources: computed(() => owners.value),
    load,
    loading,
    loadMihomoNodes,
    mihomoNodesLoading: (ownerID: string) => !!mihomoNodeLoading[ownerID.trim()],
    nodesForMihomoOwner: (ownerID: string) => mihomoNodes[ownerID.trim()] || [],
    resetForm,
    rows,
    ruleCount,
    saveRule,
    saving,
    staticExitSources: computed(() => owners.value),
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

function newInUserRuleForm() {
  return {
    ...newEgressProfileForm(),
    password_value: '',
    rule_id: '',
    username: '',
  }
}

function formRequest(form: ReturnType<typeof newInUserRuleForm>) {
  const ruleID = form.rule_id.trim() || generatedRuleID(form.username)
  const profileID = form.profile_id.trim() || `${ruleID}-egress`
  const displayName = form.display_name.trim() || form.username.trim()
  const profile = profileRequest({
    ...form,
    display_name: displayName,
    profile_id: profileID,
  })
  const rule: ProxyIngressRuleSettings = {
    rule_id: ruleID,
    display_name: displayName,
    enabled: form.enabled,
    username: form.username.trim(),
    password_value: form.password_value,
    profile_id: profile.profile_id,
  }
  return { profile, rule }
}

function generatedRuleID(username: string) {
  const slug = username
    .trim()
    .toLowerCase()
    .replace(/[^a-z0-9_-]+/g, '-')
    .replace(/^-+|-+$/g, '')
  return `in-user-${slug || 'user'}-${Date.now().toString(36)}`
}

export type ProxyRuntimeInUserRulesState = ReturnType<
  typeof useProxyRuntimeInUserRules
>
