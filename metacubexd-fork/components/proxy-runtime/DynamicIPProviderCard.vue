<script setup lang="ts">
import type {
  ProxyDynamicIPGatewaySettings,
  ProxyDynamicIPProviderSettings,
  ProxyProviderAccount,
} from '~/types/byte/v/forge/contracts/proxyruntime/v1/proxy_runtime'
import { IconCloud, IconKey } from '@tabler/icons-vue'

const props = defineProps<{
  row: {
    provider_id: string
    provider_name: string
    settings?: ProxyDynamicIPProviderSettings
    gateways: ProxyDynamicIPGatewaySettings[]
    accounts: ProxyProviderAccount[]
  }
}>()

defineEmits<{
  addAccount: [providerID: string]
  addEndpoint: [providerID: string]
  deleteAccount: [accountID: string]
  deleteEndpoint: [endpointURL: string]
  editAccount: [account: ProxyProviderAccount]
  editEndpoint: [
    provider: ProxyDynamicIPProviderSettings,
    endpoint: ProxyDynamicIPGatewaySettings,
  ]
}>()

const providerSettings = computed<ProxyDynamicIPProviderSettings>(() => {
  return (
    props.row.settings || {
      provider_id: props.row.provider_id,
      gateways: [],
    }
  )
})
</script>

<template>
  <article class="rounded-lg border border-base-content/8 bg-base-100/70">
    <header
      class="flex flex-col gap-3 border-b border-base-content/8 p-4 md:flex-row md:items-center md:justify-between"
    >
      <div class="min-w-0">
        <h3 class="truncate text-base font-semibold">{{ row.provider_name }}</h3>
        <div class="mt-2 flex flex-wrap gap-1">
          <span class="badge badge-sm">
            <IconCloud :size="13" />
            {{ row.gateways.length }}
          </span>
          <span class="badge badge-sm">
            <IconKey :size="13" />
            {{ row.accounts.length }}
          </span>
        </div>
      </div>

      <div class="flex items-center gap-1">
        <button
          aria-label="添加端点"
          class="btn btn-ghost btn-sm btn-square"
          title="添加端点"
          type="button"
          @click="$emit('addEndpoint', row.provider_id)"
        >
          <IconCloud :size="16" />
        </button>
        <button
          aria-label="添加账号"
          class="btn btn-primary btn-sm btn-square"
          title="添加账号"
          type="button"
          @click="$emit('addAccount', row.provider_id)"
        >
          <IconKey :size="16" />
        </button>
      </div>
    </header>

    <div class="grid gap-4 p-4 xl:grid-cols-2">
      <DynamicIPProviderEndpointList
        :gateways="row.gateways"
        :provider="providerSettings"
        @delete-endpoint="$emit('deleteEndpoint', $event)"
        @edit-endpoint="(provider, endpoint) => $emit('editEndpoint', provider, endpoint)"
      />
      <DynamicIPProviderAccountList
        :accounts="row.accounts"
        @delete-account="$emit('deleteAccount', $event)"
        @edit-account="$emit('editAccount', $event)"
      />
    </div>
  </article>
</template>
