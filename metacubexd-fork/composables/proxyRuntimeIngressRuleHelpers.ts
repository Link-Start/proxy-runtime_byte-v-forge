import type {
  EgressProfileSettings,
  ProxyIngressRuleSettings,
} from '~/types/byte/v/forge/contracts/proxyruntime/v1/proxy_runtime'

export function newIngressRuleForm() {
  return {
    rule_id: '',
    display_name: '',
    enabled: true,
    username: '',
    password_value: '',
    profile_id: '',
  }
}

export function ingressRuleFormFromSettings(rule: ProxyIngressRuleSettings) {
  return {
    ...newIngressRuleForm(),
    rule_id: rule.rule_id,
    display_name: rule.display_name,
    enabled: rule.enabled,
    username: rule.username,
    password_value: rule.password_value,
    profile_id: rule.profile_id,
  }
}

export function ingressRuleRequest(
  form: ReturnType<typeof newIngressRuleForm>,
): ProxyIngressRuleSettings {
  return {
    rule_id: form.rule_id || generatedRuleID(form.username),
    display_name: form.display_name.trim(),
    enabled: form.enabled,
    username: form.username.trim(),
    password_value: form.password_value,
    profile_id: form.profile_id.trim(),
  }
}

export function profileLabel(profile: EgressProfileSettings) {
  return profile.display_name || profile.profile_id
}

export function profileLabelByID(
  profiles: EgressProfileSettings[],
  profileID: string,
) {
  const profile = profiles.find((item) => item.profile_id === profileID)
  return profile ? profileLabel(profile) : profileID || '出口 Profile'
}

function generatedRuleID(username: string) {
  const slug = username
    .trim()
    .toLowerCase()
    .replace(/[^a-z0-9_-]+/g, '-')
    .replace(/^-+|-+$/g, '')
  return `ingress-${slug || 'user'}-${Date.now().toString(36)}`
}
