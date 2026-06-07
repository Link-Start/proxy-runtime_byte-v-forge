import type { ProxyDynamicIPProviderSettings } from '~/types/byte/v/forge/contracts/proxyruntime/v1/proxy_runtime'
import {
  cloneDynamicIPProvider,
  dynamicProviderFromForm,
  endpointFromForm,
  hasDynamicIPProviderConfig,
  providerEndpoints,
  removeProviderEndpoint,
  syncDynamicIPProviderEndpoints,
  upsertProviderEndpoint,
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
      const providerID = state.endpointForm.provider_id.trim()
      if (!providerID) throw new Error('provider_id is required')
      const next = state.dynamicIpProviders.value.map(cloneDynamicIPProvider)
      if (!next.some((provider) => provider.provider_id === providerID)) {
        throw new Error('请先添加动态代理提供商')
      }
      await api.updateDynamicIPProviders(
        upsertProviderEndpoint(
          next,
          providerID,
          endpoint,
          state.endpointForm.original_endpoint_url.trim(),
        ),
      )
      state.resetEndpointForm()
      await load()
    })
  }

  async function saveProvider() {
    await withSave(async () => {
      const next = syncDynamicIPProviderEndpoints(
        state.dynamicIpProviders.value.map(cloneDynamicIPProvider),
      )
      const current = next.find(
        (provider) =>
          provider.dynamic_provider_id ===
          state.providerForm.dynamic_provider_id.trim(),
      )
      const providerID = state.providerForm.provider_id.trim()
      const sharedEndpoints = providerEndpoints(next, providerID)
      const provider = dynamicProviderFromForm(
        state.providerForm,
        current,
        sharedEndpoints,
      )
      const existing = next.filter(
        (item) => item.dynamic_provider_id !== provider.dynamic_provider_id,
      )
      await api.updateDynamicIPProviders(
        syncDynamicIPProviderEndpoints([...existing, provider]),
      )
      state.resetProviderForm()
      await load()
    })
  }

  async function deleteProvider(provider: ProxyDynamicIPProviderSettings) {
    await withSave(async () => {
      const dynamicProviderID = provider.dynamic_provider_id
      for (const account of state.accounts.value) {
        if (account.dynamic_provider_id === dynamicProviderID) {
          await api.deleteProviderAccount({ account_id: account.account_id })
        }
      }
      await api.updateDynamicIPProviders(
        syncDynamicIPProviderEndpoints(
          state.dynamicIpProviders.value
            .filter((item) => item.dynamic_provider_id !== dynamicProviderID)
            .map(cloneDynamicIPProvider),
        ),
      )
      await load()
    })
  }

  async function deleteEndpoint(providerID: string, endpointURL: string) {
    await withSave(async () => {
      const next = removeProviderEndpoint(
        state.dynamicIpProviders.value.map(cloneDynamicIPProvider),
        providerID,
        endpointURL,
      ).filter(hasDynamicIPProviderConfig)
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
        dynamic_provider_id: state.accountForm.dynamic_provider_id,
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
    deleteProvider,
    error,
    load,
    loading,
    saveAccount,
    saveEndpoint,
    saveProvider,
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
