<script setup lang="ts">
import type {
  ProxyDynamicIPGatewaySettings,
  ProxyDynamicIPProviderSettings,
} from '~/types/byte/v/forge/contracts/proxyruntime/v1/proxy_runtime'
import { IconCloud, IconPencil, IconTrash } from '@tabler/icons-vue'

defineProps<{
  provider: ProxyDynamicIPProviderSettings
  gateways: ProxyDynamicIPGatewaySettings[]
}>()

defineEmits<{
  deleteEndpoint: [endpointURL: string]
  editEndpoint: [
    provider: ProxyDynamicIPProviderSettings,
    endpoint: ProxyDynamicIPGatewaySettings,
  ]
}>()
</script>

<template>
  <section class="min-w-0">
    <div class="mb-2 flex items-center gap-2 text-sm font-medium">
      <IconCloud :size="16" />
      端点
    </div>
    <div v-if="gateways.length === 0" class="py-5 text-sm opacity-55">
      暂无端点
    </div>
    <div v-else class="flex flex-wrap gap-2">
      <div
        v-for="endpoint in gateways"
        :key="endpoint.endpoint_url"
        class="inline-flex max-w-full items-center gap-1 rounded-lg border border-base-content/10 bg-base-200/60 px-2 py-1"
      >
        <span
          class="max-w-[18rem] truncate text-sm"
          :title="endpoint.endpoint_url"
        >
          {{ endpoint.endpoint_url }}
        </span>
        <button
          aria-label="编辑端点"
          class="btn btn-ghost btn-xs btn-square"
          title="编辑端点"
          type="button"
          @click="$emit('editEndpoint', provider, endpoint)"
        >
          <IconPencil :size="14" />
        </button>
        <button
          aria-label="删除端点"
          class="btn btn-ghost btn-xs btn-square text-error"
          title="删除端点"
          type="button"
          @click="$emit('deleteEndpoint', endpoint.endpoint_url)"
        >
          <IconTrash :size="14" />
        </button>
      </div>
    </div>
  </section>
</template>
