import type {
  ProxyDynamicIPEndpointSettings,
  ProxyDynamicIPProviderSettings,
  ProxyProviderAccount,
  ProxyProviderDescriptor,
} from '~/types/byte/v/forge/contracts/proxyruntime/v1/proxy_runtime'
import { ProxyProviderAccountStatus } from '~/types/byte/v/forge/contracts/proxyruntime/v1/proxy_runtime'
import {
  newDynamicProviderForm,
  newAccountForm,
  newEndpointForm,
  providerEndpoints,
} from '~/composables/proxyRuntimeDynamicIPHelpers'

export function useProxyRuntimeDynamicIPState() {
  const providers = ref<ProxyProviderDescriptor[]>([])
  const dynamicIpProviders = ref<ProxyDynamicIPProviderSettings[]>([])
  const accounts = ref<ProxyProviderAccount[]>([])
  const providerForm = reactive(newDynamicProviderForm())
  const endpointForm = reactive(newEndpointForm())
  const accountForm = reactive(newAccountForm())
  const dynamicProviderCount = computed(() => dynamicIpProviders.value.length)
  const dynamicEndpointCount = computed(() =>
    providerEndpointGroups.value.reduce(
      (total, group) => total + group.endpoints.length,
      0,
    ),
  )
  const providerOptions = computed(() =>
    providers.value.map((provider) => ({
      id: provider.provider_id,
      name: provider.display_name || provider.provider_id,
    })),
  )
  const providerEndpointGroups = computed(() => {
    const providerIDs = [
      ...new Set(
        dynamicIpProviders.value
          .map((provider) => provider.provider_id)
          .filter((providerID): providerID is string => Boolean(providerID)),
      ),
    ]
    return providerIDs.map((providerID) => ({
      provider_id: providerID,
      provider_name: providerName(providerID),
      providers: dynamicIpProviders.value.filter(
        (provider) => provider.provider_id === providerID,
      ),
      endpoints: providerEndpoints(dynamicIpProviders.value, providerID),
      dynamic_provider_count: dynamicIpProviders.value.filter(
        (provider) => provider.provider_id === providerID,
      ).length,
    }))
  })

  function resetEmptyProviders() {
    const first = providerOptions.value[0]?.id || ''
    if (!providerForm.provider_id) providerForm.provider_id = first
    if (!accountForm.provider_id) accountForm.provider_id = first
  }

  function editProvider(provider: ProxyDynamicIPProviderSettings) {
    Object.assign(providerForm, {
      dynamic_provider_id: provider.dynamic_provider_id,
      provider_id: provider.provider_id,
      display_name: provider.display_name,
      rotating_concurrency_limit: provider.rotating_concurrency_limit || 10,
      sticky_concurrency_limit: provider.sticky_concurrency_limit || 2,
    })
  }

  function editEndpoint(
    providerID: string,
    endpoint: ProxyDynamicIPEndpointSettings,
  ) {
    Object.assign(endpointForm, {
      provider_id: providerID,
      original_endpoint_url: endpoint.endpoint_url,
      endpoint_url: endpoint.endpoint_url,
    })
  }

  function editAccount(account: ProxyProviderAccount) {
    Object.assign(accountForm, {
      account_id: account.account_id,
      dynamic_provider_id: account.dynamic_provider_id,
      provider_id: account.provider_id,
      display_name: account.display_name,
      enabled:
        account.status ===
        ProxyProviderAccountStatus.PROXY_PROVIDER_ACCOUNT_STATUS_ENABLED,
      username: account.username,
      original_password_value: account.password_value,
      password_value: account.password_value,
    })
  }

  function resetProviderForm(providerID = providerOptions.value[0]?.id || '') {
    Object.assign(providerForm, newDynamicProviderForm(providerID))
  }

  function resetEndpointForm(providerID = '') {
    Object.assign(endpointForm, newEndpointForm(providerID))
  }

  function resetAccountForm(
    dynamicProviderID = '',
    providerID = providerOptions.value[0]?.id || '',
  ) {
    Object.assign(accountForm, newAccountForm(dynamicProviderID, providerID))
  }

  function providerName(providerID: string) {
    return (
      providerOptions.value.find((item) => item.id === providerID)?.name ||
      providerID
    )
  }

  return {
    accountForm,
    accounts,
    dynamicEndpointCount,
    dynamicIpProviders,
    dynamicProviderCount,
    editAccount,
    editEndpoint,
    editProvider,
    endpointForm,
    providerForm,
    providerEndpointGroups,
    providerName,
    providerOptions,
    providers,
    resetAccountForm,
    resetEmptyProviders,
    resetEndpointForm,
    resetProviderForm,
  }
}
