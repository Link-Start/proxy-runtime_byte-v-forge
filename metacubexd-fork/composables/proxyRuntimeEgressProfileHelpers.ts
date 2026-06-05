import type {
  EgressProfileSettings,
  EgressProfileSourceRef,
  ProxySourceDescriptor,
  ProxySourceNode,
} from '~/types/byte/v/forge/contracts/proxyruntime/v1/proxy_runtime'
import {
  EgressProfileExitKind,
  EgressProfileLineKind,
} from '~/types/byte/v/forge/contracts/proxyruntime/v1/proxy_runtime'

export function newEgressProfileForm() {
  return {
    profile_id: '',
    display_name: '',
    enabled: true,
    line_kind: EgressProfileLineKind.EGRESS_PROFILE_LINE_KIND_DIRECT,
    line_source_id: '',
    line_node_id: '',
    exit_kind: EgressProfileExitKind.EGRESS_PROFILE_EXIT_KIND_DIRECT,
    exit_dynamic_provider_id: '',
    exit_source_id: '',
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
    line_source_id: profile.line?.source?.source_id || '',
    line_node_id: profile.line?.source?.node_id || '',
    exit_kind: profile.exit?.kind || fallback.exit_kind,
    exit_dynamic_provider_id: profile.exit?.dynamic_provider_id || '',
    exit_source_id: profile.exit?.source?.source_id || '',
    exit_node_id: profile.exit?.source?.node_id || '',
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
      source: sourceRef(form.line_source_id, form.line_node_id),
      health_check_url: '',
      health_interval: undefined,
      health_timeout: undefined,
      expected_status: 0,
    },
    exit: {
      kind: form.exit_kind,
      dynamic_provider_id: dynamicProviderID(form),
      source: sourceRef(form.exit_source_id, form.exit_node_id),
      health_check_url: '',
      health_interval: undefined,
      health_timeout: undefined,
      expected_status: 0,
    },
  }
}

export function sourceLabel(source: ProxySourceDescriptor) {
  return source.display_name || source.source_id
}

export function nodeLabel(node: ProxySourceNode) {
  const name = node.display_name || node.node_id
  if (node.delay_ms > 0) return `${name} / ${node.delay_ms}ms`
  return name
}

export function lineText(profile: EgressProfileSettings) {
  if (profile.line?.kind === EgressProfileLineKind.EGRESS_PROFILE_LINE_KIND_SOURCE) {
    return sourceRefText(profile.line.source, '来源')
  }
  return '无代理路线'
}

export function exitText(profile: EgressProfileSettings) {
  switch (profile.exit?.kind) {
    case EgressProfileExitKind.EGRESS_PROFILE_EXIT_KIND_DIRECT:
      if (
        profile.line?.kind === EgressProfileLineKind.EGRESS_PROFILE_LINE_KIND_SOURCE
      ) {
        return '使用路线出口'
      }
      return '直连出口'
    case EgressProfileExitKind.EGRESS_PROFILE_EXIT_KIND_DYNAMIC_IP:
      return profile.exit?.dynamic_provider_id || '动态 IP'
    case EgressProfileExitKind.EGRESS_PROFILE_EXIT_KIND_STATIC_IP:
      return sourceRefText(profile.exit.source, '静态 IP')
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

function sourceRef(sourceID: string, nodeID: string) {
  return {
    source_id: sourceID.trim(),
    node_id: nodeID.trim(),
  }
}

function sourceRefText(ref: EgressProfileSourceRef | undefined, fallback: string) {
  const sourceID = ref?.source_id || ''
  const nodeID = ref?.node_id || ''
  if (sourceID && nodeID) return `${sourceID}/${nodeID}`
  return sourceID || fallback
}

function generatedProfileID(displayName: string) {
  const slug = displayName
    .trim()
    .toLowerCase()
    .replace(/[^a-z0-9_-]+/g, '-')
    .replace(/^-+|-+$/g, '')
  return `profile-${slug || 'egress'}-${Date.now().toString(36)}`
}
