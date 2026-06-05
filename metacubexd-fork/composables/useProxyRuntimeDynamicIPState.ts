import type {
  ProxyDynamicIPGatewaySettings,
  ProxyDynamicIPProviderSettings,
  ProxyProviderAccount,
  ProxyProviderDescriptor,
} from '~/types/byte/v/forge/contracts/proxyruntime/v1/proxy_runtime'
import { ProxyProviderAccountStatus } from '~/types/byte/v/forge/contracts/proxyruntime/v1/proxy_runtime'
import {
  newAccountForm,
  newEndpointForm,
} from '~/composables/proxyRuntimeDynamicIPHelpers'

export function useProxyRuntimeDynamicIPState() {
  const providers = ref<ProxyProviderDescriptor[]>([])
  const dynamicIpProviders = ref<ProxyDynamicIPProviderSettings[]>([])
  const accounts = ref<ProxyProviderAccount[]>([])
  const endpointForm = reactive(newEndpointForm())
  const accountForm = reactive(newAccountForm())
  const dynamicProviderCount = computed(() => dynamicIpProviders.value.length)
  const dynamicEndpointCount = computed(() =>
    dynamicIpProviders.value.reduce(
      (total, provider) => total + provider.gateways.length,
      0,
    ),
  )
  const providerOptions = computed(() =>
    providers.value.map((provider) => ({
      id: provider.provider_id,
      name: provider.display_name || provider.provider_id,
    })),
  )

  function resetEmptyProviders() {
    const first = providerOptions.value[0]?.id || ''
    if (!endpointForm.provider_id) endpointForm.provider_id = first
    if (!accountForm.provider_id) accountForm.provider_id = first
  }

  function editEndpoint(
    provider: ProxyDynamicIPProviderSettings,
    endpoint: ProxyDynamicIPGatewaySettings,
  ) {
    Object.assign(endpointForm, {
      provider_id: provider.provider_id,
      original_endpoint_url: endpoint.endpoint_url,
      endpoint_url: endpoint.endpoint_url,
    })
  }

  function editAccount(account: ProxyProviderAccount) {
    Object.assign(accountForm, {
      account_id: account.account_id,
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

  function resetEndpointForm(providerID = providerOptions.value[0]?.id || '') {
    Object.assign(endpointForm, newEndpointForm(providerID))
  }

  function resetAccountForm(providerID = providerOptions.value[0]?.id || '') {
    Object.assign(accountForm, newAccountForm(providerID))
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
    endpointForm,
    providerName,
    providerOptions,
    providers,
    resetAccountForm,
    resetEmptyProviders,
    resetEndpointForm,
  }
}
