import type {
  ProxySourceDescriptor,
  UpsertProxyFixedSourceRequest,
  UpsertProxySubscriptionSourceRequest,
} from '~/types/byte/v/forge/contracts/proxyruntime/v1/proxy_runtime'

export function newSubscriptionForm() {
  return {
    source_id: '',
    display_name: '',
    enabled: true,
    url: '',
    filter: '',
    exclude_filter: '',
  }
}

export function newFixedSourceForm() {
  return {
    source_id: '',
    display_name: '',
    enabled: true,
    uri: '',
  }
}

export function subscriptionFormFromSource(source: ProxySourceDescriptor) {
  return {
    source_id: source.source_id,
    display_name: source.display_name,
    enabled: source.enabled,
    url: source.subscription?.url || '',
    filter: source.subscription?.filter || '',
    exclude_filter: source.subscription?.exclude_filter || '',
  }
}

export function fixedFormFromSource(source: ProxySourceDescriptor) {
  return {
    source_id: source.source_id,
    display_name: source.display_name,
    enabled: source.enabled,
    uri: source.fixed_proxy?.uri || '',
  }
}

export function subscriptionRequest(
  form: ReturnType<typeof newSubscriptionForm>,
): UpsertProxySubscriptionSourceRequest {
  return {
    source_id: form.source_id,
    display_name: form.display_name.trim(),
    enabled: form.enabled,
    url: form.url.trim(),
    clear_url: false,
    interval: undefined,
    filter: form.filter.trim(),
    exclude_filter: form.exclude_filter.trim(),
    health_check_url: '',
    health_interval: undefined,
    health_timeout: undefined,
    health_lazy: true,
    expected_status: 0,
    region_codes: [],
  }
}

export function fixedSourceRequest(
  form: ReturnType<typeof newFixedSourceForm>,
): UpsertProxyFixedSourceRequest {
  return {
    source_id: form.source_id,
    display_name: form.display_name.trim(),
    enabled: form.enabled,
    uri: form.uri.trim(),
    clear_uri: false,
    region_codes: [],
  }
}
