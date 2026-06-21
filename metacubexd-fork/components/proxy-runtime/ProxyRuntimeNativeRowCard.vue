<script setup lang="ts">
import type { ProxyGatewayNativeRow } from '~/composables/proxyGatewayNativeRows'
import { writeProxyGatewayClipboard } from '~/composables/proxyGatewayClipboard'
import { IconClipboard, IconPencil, IconTrash } from '@tabler/icons-vue'

defineProps<{
  expanded: boolean
  index: number
  row: ProxyGatewayNativeRow
  saving: boolean
}>()
defineEmits<{
  collapse: [value: boolean]
  delete: []
  edit: []
}>()

const copied = ref(false)

async function copyValue(value: string) {
  await writeProxyGatewayClipboard(value)
  copied.value = true
  window.setTimeout(() => {
    copied.value = false
  }, 1200)
}
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
          <div class="flex items-center gap-2">
            <span class="badge badge-primary badge-sm">{{ row.typeLabel }}</span>
            <h3 class="truncate text-base font-semibold">{{ row.name }}</h3>
          </div>
          <div class="mt-2 flex flex-wrap gap-1">
            <span class="badge badge-ghost badge-sm">{{ row.valueLabel }}</span>
          </div>
        </div>
        <div class="flex shrink-0 items-center gap-1" @click.stop>
          <button
            aria-label="编辑"
            class="btn btn-ghost btn-sm btn-square"
            title="编辑"
            type="button"
            @click="$emit('edit')"
          >
            <IconPencil :size="16" />
          </button>
          <button
            aria-label="删除"
            class="btn btn-ghost btn-sm btn-square text-error"
            :disabled="saving"
            title="删除"
            type="button"
            @click="$emit('delete')"
          >
            <IconTrash :size="16" />
          </button>
        </div>
      </div>
    </template>

    <div class="col-span-full grid min-w-0 gap-2 text-sm">
      <div class="text-xs opacity-60">地址</div>
      <div class="flex min-w-0 items-center gap-2" @click.stop>
        <div class="min-w-0 flex-1 break-all font-mono text-xs" :title="row.value">
          {{ row.value }}
        </div>
        <button
          aria-label="复制地址"
          class="btn btn-ghost btn-xs btn-square shrink-0"
          :title="copied ? '已复制' : '复制地址'"
          type="button"
          @click="copyValue(row.value)"
        >
          <IconClipboard :size="14" />
        </button>
      </div>
    </div>
  </Collapse>
</template>
