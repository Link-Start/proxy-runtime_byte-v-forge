import type { ProxyIngressRuleSettings } from '~/types/byte/v/forge/contracts/proxygateway/v1/proxy_gateway'
import {
  newEgressProfileForm,
  profileRequest,
} from '~/composables/proxyGatewayEgressProfileHelpers'

export type InUserRuleForm = Omit<
  ReturnType<typeof newEgressProfileForm>,
  'enabled'
> & {
  password_value: string
  rule_id: string
  username: string
}

export function newInUserRuleForm() {
  const { enabled: _enabled, ...profileForm } = newEgressProfileForm()
  return {
    ...profileForm,
    password_value: '',
    rule_id: '',
    username: '',
  }
}

export function inUserRuleFormRequest(form: InUserRuleForm) {
  const ruleID = form.rule_id.trim() || generatedRuleID(form.username)
  const profileID = form.profile_id.trim() || `${ruleID}-egress`
  const displayName = form.display_name.trim() || form.username.trim()
  const profile = profileRequest({
    ...form,
    display_name: displayName,
    enabled: true,
    profile_id: profileID,
  })
  const rule: ProxyIngressRuleSettings = {
    rule_id: ruleID,
    display_name: displayName,
    enabled: true,
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
