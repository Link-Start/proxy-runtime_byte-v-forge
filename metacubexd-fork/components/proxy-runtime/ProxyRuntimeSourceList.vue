<script setup lang="ts">
import type { ProxySourceDescriptor } from '~/types/byte/v/forge/contracts/proxyruntime/v1/proxy_runtime'
import {
  IconPencil,
  IconPlug,
  IconTrash,
} from '@tabler/icons-vue'

defineProps<{
  emptyText: string
  sources: ProxySourceDescriptor[]
}>()

defineEmits<{
  deleteSource: [source: ProxySourceDescriptor]
  editSource: [source: ProxySourceDescriptor]
}>()

function subtitle(source: ProxySourceDescriptor) {
  if (source.subscription?.url) {
    return displayURL(source.subscription.url, '订阅 URL 已配置')
  }
  if (source.fixed_proxy?.uri) {
    return displayURL(source.fixed_proxy.uri, '固定代理 URI 已配置')
  }
  return source.source_id
}

function countText(source: ProxySourceDescriptor) {
  if (source.subscription) return '订阅'
  if (source.fixed_proxy) return `${source.fixed_proxy.endpoint_count} 节点`
  return '来源'
}

function displayURL(rawURL: string, fallback: string) {
  try {
    const parsed = new URL(rawURL)
    const path = parsed.pathname === '/' ? '' : parsed.pathname
    return `${parsed.protocol}//${parsed.host}${path}`
  } catch {
    return fallback
  }
}
</script>

<template>
  <div v-if="sources.length === 0" class="py-5 text-sm opacity-55">
    {{ emptyText }}
  </div>
  <div v-else class="divide-y divide-base-content/8">
    <div
      v-for="source in sources"
      :key="source.source_id"
      class="grid gap-2 py-3 md:grid-cols-[1fr_auto]"
    >
      <div class="min-w-0 text-sm">
        <div class="truncate font-medium">
          {{ source.display_name || source.source_id }}
        </div>
        <div class="truncate text-xs opacity-60" :title="subtitle(source)">
          {{ subtitle(source) }}
        </div>
        <div class="mt-2 flex flex-wrap gap-1">
          <span class="badge badge-sm">{{ source.enabled ? '启用' : '停用' }}</span>
          <span class="badge badge-ghost badge-sm">
            <IconPlug :size="12" />
            {{ countText(source) }}
          </span>
        </div>
      </div>
      <div class="flex items-center gap-1">
        <button
          aria-label="编辑"
          class="btn btn-ghost btn-sm btn-square"
          title="编辑"
          type="button"
          @click="$emit('editSource', source)"
        >
          <IconPencil :size="16" />
        </button>
        <button
          aria-label="删除"
          class="btn btn-ghost btn-sm btn-square text-error"
          title="删除"
          type="button"
          @click="$emit('deleteSource', source)"
        >
          <IconTrash :size="16" />
        </button>
      </div>
    </div>
  </div>
</template>
