import type {
  EgressProfileSettings,
  ProxyIngressRuleSettings,
} from '~/types/byte/v/forge/contracts/proxyruntime/v1/proxy_runtime'
import {
  ingressRuleFormFromSettings,
  ingressRuleRequest,
  newIngressRuleForm,
} from '~/composables/proxyRuntimeIngressRuleHelpers'

export function useProxyRuntimeIngressRules() {
  const api = useProxyRuntimeApi()
  const rules = ref<ProxyIngressRuleSettings[]>([])
  const profiles = ref<EgressProfileSettings[]>([])
  const form = reactive(newIngressRuleForm())
  const loading = ref(false)
  const saving = ref(false)
  const error = ref('')
  const ruleCount = computed(() => rules.value.length)
  const enabledProfiles = computed(() =>
    profiles.value.filter((profile) => profile.enabled),
  )

  async function load() {
    loading.value = true
    error.value = ''
    try {
      const response = await api.getSettings()
      rules.value = response.settings?.ingress_rules || []
      profiles.value = response.settings?.egress_profiles || []
      ensureProfileSelection()
    } catch (err) {
      error.value = err instanceof Error ? err.message : String(err)
    } finally {
      loading.value = false
    }
  }

  function resetForm() {
    Object.assign(form, newIngressRuleForm())
    ensureProfileSelection()
  }

  function editRule(rule: ProxyIngressRuleSettings) {
    Object.assign(form, ingressRuleFormFromSettings(rule))
  }

  async function saveRule() {
    await withSave(async () => {
      const next = ingressRuleRequest(form)
      const existing = rules.value.filter((rule) => rule.rule_id !== next.rule_id)
      await api.updateIngressRules([...existing, next])
      resetForm()
      await load()
    })
  }

  async function deleteRule(rule: ProxyIngressRuleSettings) {
    await withSave(async () => {
      await api.updateIngressRules(
        rules.value.filter((item) => item.rule_id !== rule.rule_id),
      )
      await load()
    })
  }

  return {
    deleteRule,
    editRule,
    enabledProfiles,
    error,
    form,
    load,
    loading,
    profiles,
    resetForm,
    ruleCount,
    rules,
    saveRule,
    saving,
  }

  function ensureProfileSelection() {
    if (form.profile_id) return
    form.profile_id = enabledProfiles.value[0]?.profile_id || ''
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

export type ProxyRuntimeIngressRulesState = ReturnType<
  typeof useProxyRuntimeIngressRules
>
