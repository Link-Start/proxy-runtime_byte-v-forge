<script setup lang="ts">
import type { ProxyGatewayPluginsState } from '~/composables/useProxyGatewayPlugins'
import { IconDeviceFloppy, IconKey, IconMapPin, IconTrash } from '@tabler/icons-vue'

const props = defineProps<{
  index: number
  row: ProxyGatewayPluginsState['geoRows']['value'][number]
  runtime: ProxyGatewayPluginsState
}>()

const expanded = ref(false)
const title = computed(() => props.row.display_name.trim() || descriptorName(props.row.kind))

function descriptorName(kind: string) {
  const descriptor = props.runtime.geoDescriptors.value.find((item) => item.kind === kind)
  return descriptor?.display_name || kind
}

function apiKeyStatusText() {
  if (props.row.anonymous) return '无需 API Key'
  if (props.row.clear_api_keys) return '保存后清除'
  if (props.row.api_key_value.trim()) return '待保存新 Key'
  if (props.row.api_key_configured) {
    return props.row.api_key_count ? `已配置 ${props.row.api_key_count} 个` : '已配置'
  }
  return '未配置 API Key'
}

function apiKeyStatusClass() {
  if (props.row.anonymous) return 'badge-info'
  if (props.row.clear_api_keys) return 'badge-warning'
  if (props.row.api_key_value.trim()) return 'badge-info'
  if (props.row.api_key_configured) return 'badge-success'
  return 'badge-error badge-outline'
}
</script>

<template>
  <Collapse :is-open="expanded" @collapse="expanded = $event">
    <template #title>
      <div class="flex min-w-0 flex-1 items-center justify-between gap-3">
        <div class="min-w-0">
          <h3 class="truncate text-base font-semibold">{{ title }}</h3>
          <div class="mt-2 flex flex-wrap gap-1">
            <span class="badge badge-sm" :class="apiKeyStatusClass()">
              <IconKey :size="12" />
              {{ apiKeyStatusText() }}
            </span>
            <span class="badge badge-ghost badge-sm">weight {{ row.weight || 100 }}</span>
            <span class="badge badge-ghost badge-sm">{{ geoPluginTypeLabel(row.kind) }}</span>
          </div>
        </div>
        <div class="flex shrink-0 items-center gap-1" @click.stop>
          <button
            aria-label="保存 Geo Provider 配置"
            class="btn btn-primary btn-sm btn-square"
            :disabled="runtime.saving.value"
            title="保存 Geo Provider 配置"
            type="button"
            @click="runtime.saveGeoProvider(row)"
          >
            <IconDeviceFloppy :size="16" />
          </button>
          <button
            aria-label="删除 Geo Provider 配置"
            class="btn btn-ghost btn-sm btn-square text-error"
            :disabled="runtime.saving.value"
            title="删除 Geo Provider 配置"
            type="button"
            @click="runtime.deleteGeoProvider(index)"
          >
            <IconTrash :size="16" />
          </button>
        </div>
      </div>
    </template>

    <div class="col-span-full grid gap-3">
      <div class="grid gap-3 md:grid-cols-[1fr_14rem]">
        <label class="grid gap-1">
          <span class="text-xs opacity-60">名称</span>
          <input
            v-model.trim="row.display_name"
            class="input input-bordered input-sm bg-base-100/80"
            :placeholder="descriptorName(row.kind)"
          />
        </label>
        <label class="grid gap-1">
          <span class="text-xs opacity-60">Geo Provider 类型</span>
          <select
            v-model="row.kind"
            class="select select-bordered select-sm bg-base-100/80"
            @change="runtime.applyGeoDescriptor(row)"
          >
            <option
              v-for="descriptor in runtime.geoDescriptors.value"
              :key="descriptor.provider_id"
              :value="descriptor.kind"
            >
              {{ descriptor.display_name || descriptor.provider_id }}
            </option>
          </select>
        </label>
      </div>

      <div class="flex flex-wrap gap-1">
        <span class="badge badge-sm">
          <IconMapPin :size="12" />
          {{ descriptorName(row.kind) }}
        </span>
        <span class="badge badge-ghost badge-sm">{{ geoPluginTypeLabel(row.kind) }}</span>
        <span class="badge badge-ghost badge-sm">weight {{ row.weight || 100 }}</span>
        <span class="badge badge-sm" :class="apiKeyStatusClass()">
          <IconKey :size="12" />
          {{ apiKeyStatusText() }}
        </span>
      </div>

      <div class="grid gap-3 md:grid-cols-[8rem_1fr_auto]">
        <label class="grid gap-1">
          <span class="text-xs opacity-60">Weight</span>
          <input v-model.number="row.weight" class="input input-bordered input-sm" min="1" type="number" />
        </label>
        <label class="grid gap-1">
          <span class="flex items-center justify-between gap-2 text-xs">
            <span class="opacity-60">API Key</span>
            <span class="badge badge-sm" :class="apiKeyStatusClass()">{{ apiKeyStatusText() }}</span>
          </span>
          <input
            v-model="row.api_key_value"
            class="input input-bordered input-sm"
            :disabled="row.anonymous"
            :placeholder="row.api_key_configured ? '留空保留已配置 API Key' : '输入 API Key'"
            type="password"
          />
        </label>
        <label class="flex items-end gap-2 pb-2 text-sm">
          <input v-model="row.anonymous" class="checkbox checkbox-sm" type="checkbox" />
          anonymous
        </label>
      </div>

      <label v-if="row.api_key_configured" class="flex items-center gap-2 text-xs opacity-75">
        <input v-model="row.clear_api_keys" class="checkbox checkbox-xs" type="checkbox" />
        保存时清除已配置 API Key
      </label>
    </div>
  </Collapse>
</template>
