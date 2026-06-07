import type {
  EgressProfileMihomoNodeRef,
  EgressProfileSettings,
} from '~/types/byte/v/forge/contracts/proxyruntime/v1/proxy_runtime'
import type { MihomoEgressOwner } from '~/composables/proxyRuntimeMihomoOwnerHelpers'
import {
  EgressProfileExitKind,
  EgressProfileLineKind,
  ProxySessionMode,
} from '~/types/byte/v/forge/contracts/proxyruntime/v1/proxy_runtime'
import {
  durationMinutes,
  dynamicIPEndpointLabel,
  dynamicIPExitText,
  dynamicIPPolicySessionID,
  dynamicIPPolicy,
  dynamicIPSessionMode,
} from '~/composables/proxyRuntimeDynamicProfilePolicyHelpers'

export function newEgressProfileForm() {
  return {
    profile_id: '',
    display_name: '',
    enabled: true,
    line_kind: EgressProfileLineKind.EGRESS_PROFILE_LINE_KIND_DIRECT,
    line_resource_id: '',
    line_node_id: '',
    exit_kind: EgressProfileExitKind.EGRESS_PROFILE_EXIT_KIND_DIRECT,
    exit_dynamic_session_mode: ProxySessionMode.PROXY_SESSION_MODE_STICKY,
    exit_dynamic_provider_id: '',
    exit_dynamic_endpoint_id: '',
    exit_dynamic_region: '',
    exit_dynamic_state: '',
    exit_dynamic_city: '',
    exit_dynamic_asn: '',
    exit_dynamic_sticky_minutes: 10,
    exit_dynamic_session_id: '',
    exit_resource_id: '',
    exit_node_id: '',
  }
}

export function profileFormFromSettings(profile: EgressProfileSettings) {
  const fallback = newEgressProfileForm()
  return {
    ...fallback,
    profile_id: profile.profile_id,
    display_name: profile.display_name,
    enabled: profile.enabled,
    line_kind: profile.line?.kind || fallback.line_kind,
    line_resource_id: profile.line?.mihomo_node?.resource_id || '',
    line_node_id: profile.line?.mihomo_node?.node_id || '',
    exit_kind: profile.exit?.kind || fallback.exit_kind,
    exit_dynamic_session_mode: dynamicIPSessionMode(
      profile.exit?.dynamic_ip_policy?.mode,
    ),
    exit_dynamic_provider_id: profile.exit?.dynamic_provider_id || '',
    exit_dynamic_endpoint_id:
      profile.exit?.dynamic_ip_policy?.labels?.[dynamicIPEndpointLabel] || '',
    exit_dynamic_region: profile.exit?.dynamic_ip_policy?.region || '',
    exit_dynamic_state: profile.exit?.dynamic_ip_policy?.state || '',
    exit_dynamic_city: profile.exit?.dynamic_ip_policy?.city || '',
    exit_dynamic_asn: profile.exit?.dynamic_ip_policy?.asn || '',
    exit_dynamic_sticky_minutes: durationMinutes(
      profile.exit?.dynamic_ip_policy?.sticky_ttl,
      fallback.exit_dynamic_sticky_minutes,
    ),
    exit_dynamic_session_id: dynamicIPPolicySessionID(
      profile.exit?.dynamic_ip_policy,
    ),
    exit_resource_id: profile.exit?.mihomo_node?.resource_id || '',
    exit_node_id: profile.exit?.mihomo_node?.node_id || '',
  }
}

export function profileRequest(
  form: ReturnType<typeof newEgressProfileForm>,
): EgressProfileSettings {
  return {
    profile_id: form.profile_id.trim() || generatedProfileID(form.display_name),
    display_name: form.display_name.trim(),
    enabled: form.enabled,
    line: {
      kind: form.line_kind,
      mihomo_node: mihomoNodeRef(form.line_resource_id, form.line_node_id),
      health_check_url: '',
      health_interval: undefined,
      health_timeout: undefined,
      expected_status: 0,
    },
    exit: {
      kind: form.exit_kind,
      dynamic_provider_id: dynamicProviderID(form),
      mihomo_node: mihomoNodeRef(form.exit_resource_id, form.exit_node_id),
      health_check_url: '',
      health_interval: undefined,
      health_timeout: undefined,
      expected_status: 0,
      dynamic_ip_policy: dynamicIPPolicy(form),
    },
  }
}

export function lineText(profile: EgressProfileSettings, owners: MihomoEgressOwner[] = []) {
  if (
    profile.line?.kind === EgressProfileLineKind.EGRESS_PROFILE_LINE_KIND_MIHOMO_NODE
  ) {
    return mihomoNodeRefText(profile.line.mihomo_node, 'Mihomo 节点', owners)
  }
  return '无代理路线'
}

export function exitText(profile: EgressProfileSettings, owners: MihomoEgressOwner[] = []) {
  switch (profile.exit?.kind) {
    case EgressProfileExitKind.EGRESS_PROFILE_EXIT_KIND_DIRECT:
      if (
        profile.line?.kind ===
        EgressProfileLineKind.EGRESS_PROFILE_LINE_KIND_MIHOMO_NODE
      ) {
        return '使用路线出口'
      }
      return '直连出口'
    case EgressProfileExitKind.EGRESS_PROFILE_EXIT_KIND_DYNAMIC_IP:
      return dynamicIPExitText(profile)
    case EgressProfileExitKind.EGRESS_PROFILE_EXIT_KIND_STATIC_IP:
      return mihomoNodeRefText(profile.exit.mihomo_node, '静态 IP', owners)
    default:
      return '出口'
  }
}

function dynamicProviderID(form: ReturnType<typeof newEgressProfileForm>) {
  if (form.exit_kind !== EgressProfileExitKind.EGRESS_PROFILE_EXIT_KIND_DYNAMIC_IP) {
    return ''
  }
  return form.exit_dynamic_provider_id.trim()
}

function mihomoNodeRef(resourceID: string, nodeID: string) {
  return {
    resource_id: resourceID.trim(),
    node_id: nodeID.trim(),
  }
}

function mihomoNodeRefText(
  ref: EgressProfileMihomoNodeRef | undefined,
  fallback: string,
  owners: MihomoEgressOwner[],
) {
  const resourceID = ref?.resource_id || ''
  const nodeID = ref?.node_id || ''
  const owner = owners.find((item) => item.owner_id === resourceID)
  const ownerName = owner?.display_name || resourceID
  if (resourceID && nodeID) {
    if (nodeID === resourceID || nodeID === ownerName) return ownerName
    const stablePrefix = `${resourceID}/`
    const displayPrefix = `${ownerName}/`
    if (nodeID.startsWith(stablePrefix)) return `${ownerName}/${nodeID.slice(stablePrefix.length)}`
    if (nodeID.startsWith(displayPrefix)) return nodeID
    return `${ownerName}/${nodeID}`
  }
  return ownerName || fallback
}

function generatedProfileID(displayName: string) {
  const slug = displayName
    .trim()
    .toLowerCase()
    .replace(/[^a-z0-9_-]+/g, '-')
    .replace(/^-+|-+$/g, '')
  return `profile-${slug || 'egress'}-${Date.now().toString(36)}`
}
