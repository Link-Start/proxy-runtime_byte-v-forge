<script setup lang="ts">
import type { ProxyGatewayNativeRow } from '~/composables/proxyGatewayNativeRows'
import type { ProxyGatewayMihomoNativeState } from '~/composables/useProxyGatewayMihomoNativeConfig'

const props = defineProps<{ runtime: ProxyGatewayMihomoNativeState }>()
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

function editItem(row: ProxyGatewayNativeRow) {
  props.runtime.editRow(row)
  modal.value?.open()
}

defineExpose({ openCreate: addItem })
</script>

<template>
  <section class="flex min-h-0 w-full flex-col gap-3">
    <div v-if="runtime.error.value" class="alert alert-error text-sm">
      {{ runtime.error.value }}
    </div>

    <div v-if="runtime.rows.value.length === 0" class="py-8 text-center text-sm opacity-60">
      暂无 Mihomo 资源
    </div>

    <ProxiesRenderWrapper v-else>
      <template #even>
        <ProxyGatewayNativeRowCard
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
        <ProxyGatewayNativeRowCard
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
        <ProxyGatewayNativeRowCard
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

    <ProxyGatewayNativeItemModal ref="modal" :runtime="runtime" />
  </section>
</template>
