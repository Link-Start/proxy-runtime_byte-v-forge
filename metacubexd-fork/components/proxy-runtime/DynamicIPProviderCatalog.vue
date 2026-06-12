<script setup lang="ts">
import type { ProxyRuntimeDynamicIPProvidersState } from '~/composables/useProxyRuntimeDynamicIPProviders'
import type {
  ProxyDynamicIPEndpointSettings,
  ProxyDynamicIPProviderSettings,
  ProxyProviderAccount,
} from '~/types/byte/v/forge/contracts/proxyruntime/v1/proxy_runtime'

const props = defineProps<{ runtime: ProxyRuntimeDynamicIPProvidersState }>()
const providerModal = ref<{ open: () => void; close: () => void }>()
const endpointModal = ref<{ open: () => void; close: () => void }>()
const accountModal = ref<{ open: () => void; close: () => void }>()
const expanded = reactive<Record<string, boolean>>({})
const groups = computed(() =>
  props.runtime.providerEndpointGroups.value.map((group) => ({
    ...group,
    accounts: props.runtime.accounts.value.filter(
      (account) => account.provider_id === group.provider_id,
    ),
  })),
)
const evenGroups = computed(() => groups.value.filter((_, index) => index % 2 === 0))
const oddGroups = computed(() => groups.value.filter((_, index) => index % 2 === 1))

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
function editEndpoint(providerID: string, endpoint: ProxyDynamicIPEndpointSettings) {
  props.runtime.editEndpoint(providerID, endpoint)
  endpointModal.value?.open()
}
function editProvider(provider: ProxyDynamicIPProviderSettings) {
  props.runtime.editProvider(provider)
  providerModal.value?.open()
}
function editAccount(account: ProxyProviderAccount) {
  props.runtime.editAccount(account)
  accountModal.value?.open()
}

defineExpose({ openCreate: addProvider })
</script>

<template>
  <section class="flex min-h-0 w-full flex-col gap-3">
    <div v-if="groups.length === 0" class="py-8 text-center text-sm opacity-60">
      暂无动态IP提供商
    </div>
    <ProxiesRenderWrapper v-else>
      <template #even>
        <DynamicIPProviderGroup
          v-for="(group, index) in evenGroups"
          :key="group.provider_id"
          :expanded="expanded[group.provider_id] || false"
          :group="group"
          :index="index"
          @add-account="addAccount"
          @add-endpoint="addEndpoint"
          @collapse="expanded[group.provider_id] = $event"
          @delete-account="runtime.deleteAccount"
          @delete-endpoint="runtime.deleteEndpoint"
          @delete-provider="runtime.deleteProvider"
          @edit-account="editAccount"
          @edit-endpoint="editEndpoint"
          @edit-provider="editProvider"
        />
      </template>
      <template #odd>
        <DynamicIPProviderGroup
          v-for="(group, index) in oddGroups"
          :key="group.provider_id"
          :expanded="expanded[group.provider_id] || false"
          :group="group"
          :index="index"
          @add-account="addAccount"
          @add-endpoint="addEndpoint"
          @collapse="expanded[group.provider_id] = $event"
          @delete-account="runtime.deleteAccount"
          @delete-endpoint="runtime.deleteEndpoint"
          @delete-provider="runtime.deleteProvider"
          @edit-account="editAccount"
          @edit-endpoint="editEndpoint"
          @edit-provider="editProvider"
        />
      </template>
      <template #default>
        <DynamicIPProviderGroup
          v-for="(group, index) in groups"
          :key="group.provider_id"
          :expanded="expanded[group.provider_id] || false"
          :group="group"
          :index="index"
          @add-account="addAccount"
          @add-endpoint="addEndpoint"
          @collapse="expanded[group.provider_id] = $event"
          @delete-account="runtime.deleteAccount"
          @delete-endpoint="runtime.deleteEndpoint"
          @delete-provider="runtime.deleteProvider"
          @edit-account="editAccount"
          @edit-endpoint="editEndpoint"
          @edit-provider="editProvider"
        />
      </template>
    </ProxiesRenderWrapper>

    <DynamicIPProviderModal ref="providerModal" :runtime="runtime" />
    <DynamicIPProviderEndpointModal ref="endpointModal" :runtime="runtime" />
    <DynamicIPProviderAccountModal ref="accountModal" :runtime="runtime" />
  </section>
</template>
