<script setup lang="ts">
import type { ProxyRuntimeDynamicIPProvidersState } from '~/composables/useProxyRuntimeDynamicIPProviders'
import type {
  ProxyDynamicIPProviderSettings,
  ProxyProviderAccount,
} from '~/types/byte/v/forge/contracts/proxyruntime/v1/proxy_runtime'
import { IconServer } from '@tabler/icons-vue'

const props = defineProps<{ runtime: ProxyRuntimeDynamicIPProvidersState }>()

const endpointModal = ref<{ open: () => void; close: () => void }>()
const accountModal = ref<{ open: () => void; close: () => void }>()

const rows = computed(() => {
  const providerIDs = new Set<string>()
  for (const provider of props.runtime.providerOptions.value) {
    providerIDs.add(provider.id)
  }
  for (const provider of props.runtime.dynamicIpProviders.value) {
    providerIDs.add(provider.provider_id)
  }
  for (const account of props.runtime.accounts.value) {
    providerIDs.add(account.provider_id)
  }
  return [...providerIDs].map((providerID) => {
    const settings = props.runtime.dynamicIpProviders.value.find(
      (provider) => provider.provider_id === providerID,
    )
    const accounts = props.runtime.accounts.value.filter(
      (account) => account.provider_id === providerID,
    )
    return {
      provider_id: providerID,
      provider_name: props.runtime.providerName(providerID),
      settings,
      gateways: settings?.gateways || [],
      accounts,
    }
  })
})

function providerSettings(providerID: string): ProxyDynamicIPProviderSettings {
  return (
    props.runtime.dynamicIpProviders.value.find(
      (provider) => provider.provider_id === providerID,
    ) || { provider_id: providerID, gateways: [] }
  )
}

function addEndpoint(providerID: string) {
  props.runtime.resetEndpointForm(providerID)
  endpointModal.value?.open()
}

function addAccount(providerID: string) {
  props.runtime.resetAccountForm(providerID)
  accountModal.value?.open()
}

function editEndpoint(
  provider: ProxyDynamicIPProviderSettings,
  endpoint: ProxyDynamicIPProviderSettings['gateways'][number],
) {
  props.runtime.editEndpoint(provider, endpoint)
  endpointModal.value?.open()
}

function editAccount(account: ProxyProviderAccount) {
  props.runtime.editAccount(account)
  accountModal.value?.open()
}
</script>

<template>
  <section
    class="flex min-h-0 flex-col gap-4 rounded-xl border border-base-content/8 bg-base-200/60 p-4"
  >
    <div class="flex items-center justify-between gap-3">
      <h2 class="flex items-center gap-2 text-base font-semibold">
        <IconServer :size="18" />
        动态代理提供者
      </h2>
      <span class="badge badge-primary badge-sm">{{ rows.length }}</span>
    </div>

    <div
      v-if="rows.length === 0"
      class="rounded-lg border border-dashed border-base-content/15 p-6 text-center text-sm opacity-60"
    >
      暂无动态代理提供者
    </div>
    <div v-else class="flex min-h-0 flex-col gap-3 overflow-y-auto">
      <DynamicIPProviderCard
        v-for="row in rows"
        :key="row.provider_id"
        :row="row"
        @add-account="addAccount"
        @add-endpoint="addEndpoint"
        @delete-account="runtime.deleteAccount"
        @delete-endpoint="(endpointURL) => runtime.deleteEndpoint(row.provider_id, endpointURL)"
        @edit-account="editAccount"
        @edit-endpoint="editEndpoint"
      />
    </div>

    <DynamicIPProviderEndpointModal ref="endpointModal" :runtime="runtime" />
    <DynamicIPProviderAccountModal ref="accountModal" :runtime="runtime" />
  </section>
</template>
