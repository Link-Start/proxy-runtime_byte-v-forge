import type { ProxyDynamicIPProviderSettings } from '~/types/byte/v/forge/contracts/proxyruntime/v1/proxy_runtime'
import {
  cloneDynamicIPProvider,
  endpointFromForm,
  hasDynamicIPProviderConfig,
  upsertDynamicIPProvider,
} from '~/composables/proxyRuntimeDynamicIPHelpers'

export function useProxyRuntimeDynamicIPProviders() {
  const api = useProxyRuntimeApi()
  const state = useProxyRuntimeDynamicIPState()
  const loading = ref(false)
  const saving = ref(false)
  const error = ref('')

  async function load() {
    loading.value = true
    error.value = ''
    try {
      const [providerRes, settingsRes, accountRes] = await Promise.all([
        api.listProviders(),
        api.getSettings(),
        api.listProviderAccounts(),
      ])
      state.providers.value = providerRes.providers || []
      state.dynamicIpProviders.value =
        settingsRes.settings?.dynamic_ip_providers?.map(cloneDynamicIPProvider) ||
        []
      state.accounts.value = accountRes.accounts || []
      state.resetEmptyProviders()
    } catch (err) {
      error.value = err instanceof Error ? err.message : String(err)
    } finally {
      loading.value = false
    }
  }

  async function saveEndpoint() {
    await withSave(async () => {
      const endpoint = endpointFromForm(state.endpointForm)
      const next = state.dynamicIpProviders.value.map(cloneDynamicIPProvider)
      const provider = upsertDynamicIPProvider(
        next,
        state.endpointForm.provider_id.trim(),
      )
      const originalURL = state.endpointForm.original_endpoint_url.trim()
      const index = provider.gateways.findIndex(
        (item) =>
          item.endpoint_url === (originalURL || endpoint.endpoint_url),
      )
      if (index >= 0) provider.gateways[index] = endpoint
      else provider.gateways.push(endpoint)
      await api.updateDynamicIPProviders(next)
      state.resetEndpointForm()
      await load()
    })
  }

  async function deleteEndpoint(providerID: string, endpointURL: string) {
    await withSave(async () => {
      const next = state.dynamicIpProviders.value
        .map(cloneDynamicIPProvider)
        .map((provider) => removeEndpoint(provider, providerID, endpointURL))
        .filter(hasDynamicIPProviderConfig)
      await api.updateDynamicIPProviders(next)
      await load()
    })
  }

  async function saveAccount() {
    await withSave(async () => {
      const passwordChanged =
        state.accountForm.password_value !==
        state.accountForm.original_password_value
      await api.upsertProviderAccount({
        account_id: state.accountForm.account_id,
        display_name: state.accountForm.display_name,
        enabled: state.accountForm.enabled,
        provider_id: state.accountForm.provider_id,
        username: state.accountForm.username,
        password_secret_ref: undefined,
        clear_password: false,
        password_value: passwordChanged ? state.accountForm.password_value : '',
      })
      state.resetAccountForm()
      await load()
    })
  }

  async function deleteAccount(accountID: string) {
    await withSave(async () => {
      await api.deleteProviderAccount({ account_id: accountID })
      await load()
    })
  }

  return {
    ...state,
    deleteAccount,
    deleteEndpoint,
    error,
    load,
    loading,
    saveAccount,
    saveEndpoint,
    saving,
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

export type ProxyRuntimeDynamicIPProvidersState = ReturnType<
  typeof useProxyRuntimeDynamicIPProviders
>

function removeEndpoint(
  provider: ProxyDynamicIPProviderSettings,
  providerID: string,
  endpointURL: string,
) {
  if (provider.provider_id !== providerID) return provider
  return {
    ...provider,
    gateways: provider.gateways.filter(
      (gateway) => gateway.endpoint_url !== endpointURL,
    ),
  }
}
