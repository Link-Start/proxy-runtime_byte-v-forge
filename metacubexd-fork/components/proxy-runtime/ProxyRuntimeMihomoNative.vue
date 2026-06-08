<script setup lang="ts">
import type { ProxyRuntimeNativeRow } from '~/composables/proxyRuntimeNativeRows'
import type { ProxyRuntimeMihomoNativeState } from '~/composables/useProxyRuntimeMihomoNativeConfig'
import { IconPlus } from '@tabler/icons-vue'

const props = defineProps<{ runtime: ProxyRuntimeMihomoNativeState }>()
const modal = ref<{ open: () => void; close: () => void }>()
const expanded = reactive<Record<string, boolean>>({})

const evenRows = computed(() =>
  props.runtime.rows.value.filter((_, index) => index % 2 === 0),
)
const oddRows = computed(() =>
  props.runtime.rows.value.filter((_, index) => index % 2 === 1),
)

function addItem() {
  props.runtime.resetForm()
  modal.value?.open()
}

function editItem(row: ProxyRuntimeNativeRow) {
  props.runtime.editRow(row)
  modal.value?.open()
}
</script>

<template>
  <section class="flex min-h-0 w-full flex-col gap-3">
    <div class="animate-fade-slide-in flex shrink-0 items-center justify-end gap-2">
      <button
        aria-label="添加原生配置"
        class="btn btn-primary btn-sm btn-square"
        title="添加原生配置"
        type="button"
        @click="addItem"
      >
        <IconPlus :size="16" />
      </button>
    </div>

    <div v-if="runtime.error.value" class="alert alert-error text-sm">
      {{ runtime.error.value }}
    </div>

    <div v-if="runtime.rows.value.length === 0" class="py-8 text-center text-sm opacity-60">
      暂无原生配置
    </div>

    <ProxiesRenderWrapper v-else>
      <template #even>
        <ProxyRuntimeNativeRowCard
          v-for="(row, index) in evenRows"
          :key="row.id"
          :expanded="expanded[row.id] || false"
          :index="index"
          :row="row"
          :saving="runtime.saving.value"
          @collapse="expanded[row.id] = $event"
          @delete="runtime.deleteRow(row)"
          @edit="editItem(row)"
        />
      </template>
      <template #odd>
        <ProxyRuntimeNativeRowCard
          v-for="(row, index) in oddRows"
          :key="row.id"
          :expanded="expanded[row.id] || false"
          :index="index"
          :row="row"
          :saving="runtime.saving.value"
          @collapse="expanded[row.id] = $event"
          @delete="runtime.deleteRow(row)"
          @edit="editItem(row)"
        />
      </template>
      <template #default>
        <ProxyRuntimeNativeRowCard
          v-for="(row, index) in runtime.rows.value"
          :key="row.id"
          :expanded="expanded[row.id] || false"
          :index="index"
          :row="row"
          :saving="runtime.saving.value"
          @collapse="expanded[row.id] = $event"
          @delete="runtime.deleteRow(row)"
          @edit="editItem(row)"
        />
      </template>
    </ProxiesRenderWrapper>

    <ProxyRuntimeNativeItemModal ref="modal" :runtime="runtime" />
  </section>
</template>
