import type {
  EgressProfileSettings,
  ProxySessionPolicy,
} from '~/types/byte/v/forge/contracts/proxyruntime/v1/proxy_runtime'
import {
  EgressProfileExitKind,
  ProxyRotationMode,
  ProxySessionMode,
  ProxyUpstreamKind,
} from '~/types/byte/v/forge/contracts/proxyruntime/v1/proxy_runtime'

export const dynamicIPEndpointLabel = 'dynamic_ip_endpoint_id'
export const dynamicIPSessionLabel = 'session_id'

export interface DynamicProfilePolicyForm {
  exit_kind: EgressProfileExitKind
  exit_dynamic_session_mode: ProxySessionMode
  exit_dynamic_endpoint_id: string
  exit_dynamic_region: string
  exit_dynamic_state: string
  exit_dynamic_city: string
  exit_dynamic_asn: string
  exit_dynamic_sticky_minutes: number
  exit_dynamic_session_id: string
}

export function dynamicIPPolicy(
  form: DynamicProfilePolicyForm,
): ProxySessionPolicy | undefined {
  if (form.exit_kind !== EgressProfileExitKind.EGRESS_PROFILE_EXIT_KIND_DYNAMIC_IP) {
    return undefined
  }
  const mode = dynamicIPSessionMode(form.exit_dynamic_session_mode)
  return {
    mode,
    region: form.exit_dynamic_region.trim().toUpperCase(),
    state: form.exit_dynamic_state.trim(),
    city: form.exit_dynamic_city.trim(),
    asn: form.exit_dynamic_asn.trim(),
    sticky_ttl:
      mode === ProxySessionMode.PROXY_SESSION_MODE_STICKY
        ? durationFromMinutes(form.exit_dynamic_sticky_minutes)
        : undefined,
    labels: dynamicIPPolicyLabels(form),
    upstream_kind: ProxyUpstreamKind.PROXY_UPSTREAM_KIND_DYNAMIC_IP,
    rotation_mode:
      mode === ProxySessionMode.PROXY_SESSION_MODE_ROTATING
        ? ProxyRotationMode.PROXY_ROTATION_MODE_PER_REQUEST
        : ProxyRotationMode.PROXY_ROTATION_MODE_STICKY_SESSION,
  }
}

export function dynamicIPExitText(profile: EgressProfileSettings) {
  const policy = profile.exit?.dynamic_ip_policy
  const mode = dynamicIPSessionMode(policy?.mode)
  const parts = [
    profile.exit?.dynamic_provider_id || '动态 IP',
    mode === ProxySessionMode.PROXY_SESSION_MODE_ROTATING ? '轮转IP' : '粘性IP',
    policy?.labels?.[dynamicIPEndpointLabel],
    policy?.region,
    policy?.state,
    policy?.city,
    policy?.asn ? `ASN ${policy.asn}` : '',
    mode === ProxySessionMode.PROXY_SESSION_MODE_STICKY
      ? durationText(policy?.sticky_ttl)
      : '',
  ].filter(Boolean)
  return parts.join(' / ')
}

export function dynamicIPPolicySessionID(policy: ProxySessionPolicy | undefined) {
  return (
    policy?.labels?.[dynamicIPSessionLabel] ||
    policy?.labels?.sticky_session_id ||
    policy?.labels?.sticky_id ||
    policy?.labels?.sid ||
    policy?.labels?.session ||
    ''
  ).trim()
}

export function dynamicIPSessionMode(value: ProxySessionMode | undefined) {
  return value === ProxySessionMode.PROXY_SESSION_MODE_ROTATING
    ? ProxySessionMode.PROXY_SESSION_MODE_ROTATING
    : ProxySessionMode.PROXY_SESSION_MODE_STICKY
}

export function durationMinutes(value: string | undefined, fallback: number) {
  const raw = (value || '').trim()
  if (!raw) return fallback
  const seconds = Number(raw.replace(/s$/, ''))
  if (!Number.isFinite(seconds) || seconds <= 0) return fallback
  return Math.max(1, Math.ceil(seconds / 60))
}

function durationFromMinutes(value: number) {
  const minutes = Math.max(1, Math.min(120, Number(value) || 10))
  return `${minutes * 60}s`
}

function dynamicIPPolicyLabels(form: DynamicProfilePolicyForm) {
  const labels: Record<string, string> = {}
  const endpointID = form.exit_dynamic_endpoint_id.trim()
  const sessionID = form.exit_dynamic_session_id.trim()
  if (endpointID) labels[dynamicIPEndpointLabel] = endpointID
  if (
    sessionID &&
    form.exit_dynamic_session_mode === ProxySessionMode.PROXY_SESSION_MODE_STICKY
  ) {
    labels[dynamicIPSessionLabel] = sessionID
  }
  return labels
}

function durationText(value: string | undefined) {
  const minutes = durationMinutes(value, 0)
  return minutes > 0 ? `${minutes}m` : ''
}

export function endpointIDFromURL(value: string) {
  const raw = value.trim()
  if (!raw) return ''
  return `endpoint-${shortHash(raw)}`
}

function shortHash(value: string) {
  return hashModulo(value, 0xffffffff).toString(16).padStart(8, '0')
}

function hashModulo(value: string, modulo: number) {
  let hash = 2166136261
  for (const char of new TextEncoder().encode(value)) {
    hash ^= char
    hash = Math.imul(hash, 16777619) >>> 0
  }
  return modulo > 0 ? hash % modulo : hash
}
