import { proxyRuntimeUserMessage } from '~/composables/proxyRuntimeFetch'
import type {
  EgressProfileSettings,
  ProxyDynamicIPProviderSettings,
  ProxyIngressRuleSettings,
} from '~/types/byte/v/forge/contracts/proxyruntime/v1/proxy_runtime'
import type { MihomoConfigNode } from '~/composables/proxyRuntimeMihomoController'
import { type MihomoEgressOwner, mihomoOwnersFromController } from '~/composables/proxyRuntimeMihomoOwnerHelpers'
import {
  exitText,
  lineText,
  newEgressProfileForm,
  profileFormFromSettings,
} from '~/composables/proxyRuntimeEgressProfileHelpers'
import {
  inUserRuleFormRequest,
  newInUserRuleForm,
} from '~/composables/proxyRuntimeInUserRuleForm'
import { isProxyRuntimePlaygroundRule } from '~/composables/proxyRuntimePlaygroundRule'

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
  const allRows = computed(() =>
    rules.value.map((rule) => ({
      rule,
      profile: profiles.value.find(
        (profile) => profile.profile_id === rule.profile_id,
      ),
    })),
  )
  const rows = computed(() =>
    allRows.value.filter(({ rule }) => !isProxyRuntimePlaygroundRule(rule)),
  )
  const ruleCount = computed(() => rows.value.length)

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
      error.value = proxyRuntimeUserMessage(err)
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
    const { enabled: _enabled, ...profileForm } = profile
      ? profileFormFromSettings(profile)
      : newEgressProfileForm()
    Object.assign(form, newInUserRuleForm(), profileForm, {
      display_name: rule.display_name || profile?.display_name || rule.username,
      password_value: rule.password_value,
      profile_id: rule.profile_id,
      rule_id: rule.rule_id,
      username: rule.username,
    })
  }

  async function saveRule() {
    await withSave(async () => {
      const next = inUserRuleFormRequest(form)
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
    if (isProxyRuntimePlaygroundRule(rule)) return
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

  function findRuleRow(predicate: (item: (typeof allRows.value)[number]) => boolean) {
    return allRows.value.find(predicate)
  }

  async function loadMihomoNodes(ownerID: string) {
    const key = ownerID.trim()
    if (!key || mihomoNodes[key] || mihomoNodeLoading[key]) return
    mihomoNodeLoading[key] = true
    try {
      const response = await api.listMihomoConfigNodes(key)
      mihomoNodes[key] = response.nodes || []
    } catch (err) {
      error.value = proxyRuntimeUserMessage(err)
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
    findRuleRow,
    lineSources: computed(() => owners.value),
    lineText: (profile: EgressProfileSettings) => lineText(profile, owners.value),
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
    exitText: (profile: EgressProfileSettings) => exitText(profile, owners.value),
    staticExitSources: computed(() => owners.value),
  }

  async function withSave(action: () => Promise<void>) {
    saving.value = true
    error.value = ''
    try {
      await action()
    } catch (err) {
      error.value = proxyRuntimeUserMessage(err)
    } finally {
      saving.value = false
    }
  }
}

export type ProxyRuntimeInUserRulesState = ReturnType<typeof useProxyRuntimeInUserRules>
