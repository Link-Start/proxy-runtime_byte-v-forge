<script setup lang="ts">
import type { ProxyRuntimeDynamicIPProvidersState } from '~/composables/useProxyRuntimeDynamicIPProviders'
import type {
  ProxyDynamicIPEndpointSettings,
  ProxyProviderAccount,
} from '~/types/byte/v/forge/contracts/proxyruntime/v1/proxy_runtime'
import { IconPlus, IconServer } from '@tabler/icons-vue'

const props = defineProps<{ runtime: ProxyRuntimeDynamicIPProvidersState }>()

const providerModal = ref<{ open: () => void; close: () => void }>()
const endpointModal = ref<{ open: () => void; close: () => void }>()
const accountModal = ref<{ open: () => void; close: () => void }>()

const groups = computed(() => {
  return props.runtime.providerEndpointGroups.value.map((group) => ({
    ...group,
    accounts: props.runtime.accounts.value.filter(
      (account) => account.provider_id === group.provider_id,
    ),
  }))
})

function addProvider() {
  props.runtime.resetProviderForm()
  providerModal.value?.open()
}

function addEndpoint(providerID: string) {
  props.runtime.resetEndpointForm(providerID)
  endpointModal.value?.open()
}

function addAccount(providerID: string) {
  props.runtime.resetAccountForm('', providerID)
  accountModal.value?.open()
}

function editEndpoint(
  providerID: string,
  endpoint: ProxyDynamicIPEndpointSettings,
) {
  props.runtime.editEndpoint(providerID, endpoint)
  endpointModal.value?.open()
}

function editAccount(account: ProxyProviderAccount) {
  props.runtime.editAccount(account)
  accountModal.value?.open()
}
</script>

<template>
  <section
    class="flex min-h-0 flex-col gap-3 p-2"
  >
    <div class="flex items-center justify-between gap-3">
      <h2 class="flex items-center gap-2 text-base font-semibold">
        <IconServer :size="18" />
        动态IP提供商
      </h2>
      <div class="flex items-center gap-2">
        <span class="badge badge-primary badge-sm">{{ groups.length }}</span>
        <button
          aria-label="添加动态代理提供商"
          class="btn btn-primary btn-sm btn-square"
          title="添加动态代理提供商"
          type="button"
          @click="addProvider"
        >
          <IconPlus :size="16" />
        </button>
      </div>
    </div>

    <div
      v-if="groups.length === 0"
      class="rounded-lg border border-dashed border-base-content/15 p-6 text-center text-sm opacity-60"
    >
      暂无动态IP提供商
    </div>
    <div v-else class="grid min-h-0 gap-3 overflow-y-auto xl:grid-cols-2">
      <DynamicIPProviderGroup
        v-for="group in groups"
        :key="group.provider_id"
        :group="group"
        @add-account="addAccount"
        @add-endpoint="addEndpoint"
        @delete-account="runtime.deleteAccount"
        @delete-endpoint="runtime.deleteEndpoint"
        @edit-account="editAccount"
        @edit-endpoint="editEndpoint"
      />
    </div>

    <DynamicIPProviderModal ref="providerModal" :runtime="runtime" />
    <DynamicIPProviderEndpointModal ref="endpointModal" :runtime="runtime" />
    <DynamicIPProviderAccountModal ref="accountModal" :runtime="runtime" />
  </section>
</template>
