<script setup lang="ts">
import type {
  ProxyDynamicIPEndpointSettings,
  ProxyProviderAccount,
} from '~/types/byte/v/forge/contracts/proxyruntime/v1/proxy_runtime'
import { IconCloud, IconKey } from '@tabler/icons-vue'

defineProps<{
  group: {
    provider_id: string
    provider_name: string
    endpoints: ProxyDynamicIPEndpointSettings[]
    accounts: ProxyProviderAccount[]
  }
}>()

defineEmits<{
  addAccount: [providerID: string]
  addEndpoint: [providerID: string]
  deleteAccount: [accountID: string]
  deleteEndpoint: [providerID: string, endpointURL: string]
  editAccount: [account: ProxyProviderAccount]
  editEndpoint: [
    providerID: string,
    endpoint: ProxyDynamicIPEndpointSettings,
  ]
}>()

const expanded = ref(false)
</script>

<template>
  <Collapse :is-open="expanded" @collapse="expanded = $event">
    <template #title>
      <div class="flex min-w-0 flex-1 items-center justify-between gap-3">
        <div class="min-w-0">
          <h3 class="truncate text-base font-semibold">
            {{ group.provider_name }}
          </h3>
          <div class="mt-2 flex flex-wrap gap-1">
            <span class="badge badge-sm">
              <IconCloud :size="13" />
              {{ group.endpoints.length }}
            </span>
            <span class="badge badge-sm">
              <IconKey :size="13" />
              {{ group.accounts.length }}
            </span>
          </div>
        </div>

        <div class="flex shrink-0 items-center gap-1" @click.stop>
          <button
            aria-label="配置端点"
            class="btn btn-ghost btn-sm btn-square"
            title="配置端点"
            type="button"
            @click="$emit('addEndpoint', group.provider_id)"
          >
            <IconCloud :size="16" />
          </button>
          <button
            aria-label="添加账号"
            class="btn btn-primary btn-sm btn-square"
            title="添加账号"
            type="button"
            @click="$emit('addAccount', group.provider_id)"
          >
            <IconKey :size="16" />
          </button>
        </div>
      </div>
    </template>

    <div class="col-span-full grid gap-4 xl:grid-cols-2">
      <DynamicIPProviderEndpointList
        :endpoints="group.endpoints"
        :provider-id="group.provider_id"
        @delete-endpoint="
          (endpointURL) =>
            $emit('deleteEndpoint', group.provider_id, endpointURL)
        "
        @edit-endpoint="
          (providerID, endpoint) =>
            $emit('editEndpoint', providerID, endpoint)
        "
      />
      <DynamicIPProviderAccountList
        :accounts="group.accounts"
        @delete-account="$emit('deleteAccount', $event)"
        @edit-account="$emit('editAccount', $event)"
      />
    </div>
  </Collapse>
</template>
