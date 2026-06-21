<script setup lang="ts">
import type {
  ProxyDynamicIPEndpointSettings,
  ProxyDynamicIPProviderSettings,
  ProxyProviderAccount,
} from '~/types/byte/v/forge/contracts/proxygateway/v1/proxy_gateway'
import {
  IconCloud,
  IconKey,
  IconPencil,
  IconServer,
  IconTrash,
} from '@tabler/icons-vue'

defineProps<{
  expanded: boolean
  index: number
  group: {
    provider_id: string
    provider_name: string
    providers: ProxyDynamicIPProviderSettings[]
    endpoints: ProxyDynamicIPEndpointSettings[]
    accounts: ProxyProviderAccount[]
  }
}>()

defineEmits<{
  collapse: [value: boolean]
  addAccount: [providerID: string]
  addEndpoint: [providerID: string]
  deleteAccount: [accountID: string]
  deleteEndpoint: [providerID: string, endpointURL: string]
  deleteProvider: [provider: ProxyDynamicIPProviderSettings]
  editAccount: [account: ProxyProviderAccount]
  editEndpoint: [
    providerID: string,
    endpoint: ProxyDynamicIPEndpointSettings,
  ]
  editProvider: [provider: ProxyDynamicIPProviderSettings]
}>()

</script>

<template>
  <Collapse
    class="animate-fade-slide-in w-full"
    :style="{ animationDelay: `${index * 40}ms` }"
    :is-open="expanded"
    @collapse="$emit('collapse', $event)"
  >
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
            <span class="badge badge-sm">
              <IconServer :size="13" />
              {{ group.providers.length }}
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

    <div class="col-span-full grid min-w-0 gap-4 text-sm">
      <section class="min-w-0">
        <div class="divider my-0 text-xs uppercase opacity-40">提供商配置</div>
        <div v-if="group.providers.length === 0" class="py-3 text-sm opacity-55">
          暂无配置
        </div>
        <div v-else class="grid gap-2">
          <div
            v-for="provider in group.providers"
            :key="provider.dynamic_provider_id"
            class="grid gap-2 py-2 md:grid-cols-[1fr_auto]"
          >
            <div class="min-w-0 text-sm">
              <div class="truncate font-medium">
                {{ provider.display_name || provider.dynamic_provider_id }}
              </div>
              <div class="truncate text-xs opacity-60">
                {{ provider.dynamic_provider_id }}
              </div>
              <div class="mt-2 flex flex-wrap gap-1">
                <span class="badge badge-ghost badge-sm">
                  轮转 {{ provider.rotating_concurrency_limit || 10 }}
                </span>
                <span class="badge badge-ghost badge-sm">
                  粘性 {{ provider.sticky_concurrency_limit || 2 }}
                </span>
              </div>
            </div>
            <div class="flex items-center gap-1">
              <button
                aria-label="编辑提供商配置"
                class="btn btn-ghost btn-sm btn-square"
                title="编辑提供商配置"
                type="button"
                @click="$emit('editProvider', provider)"
              >
                <IconPencil :size="16" />
              </button>
              <button
                aria-label="删除提供商配置"
                class="btn btn-ghost btn-sm btn-square text-error"
                title="删除提供商配置"
                type="button"
                @click="$emit('deleteProvider', provider)"
              >
                <IconTrash :size="16" />
              </button>
            </div>
          </div>
        </div>
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
      </section>
      <DynamicIPProviderAccountList
        :accounts="group.accounts"
        @delete-account="$emit('deleteAccount', $event)"
        @edit-account="$emit('editAccount', $event)"
      />
    </div>
  </Collapse>
</template>
