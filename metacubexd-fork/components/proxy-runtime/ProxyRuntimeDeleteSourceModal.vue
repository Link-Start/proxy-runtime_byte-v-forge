<script setup lang="ts">
import type { ProxyRuntimeSourcesState } from '~/composables/useProxyRuntimeSources'
import type { ProxySourceDescriptor } from '~/types/byte/v/forge/contracts/proxyruntime/v1/proxy_runtime'
import { IconTrash, IconX } from '@tabler/icons-vue'

const props = defineProps<{
  runtime: ProxyRuntimeSourcesState
  source: ProxySourceDescriptor | null
}>()

const modalRef = ref<{ open: () => void; close: () => void }>()

function open() {
  modalRef.value?.open()
}

function close() {
  modalRef.value?.close()
}

async function remove() {
  if (!props.source) return
  await props.runtime.deleteSource(props.source)
  if (!props.runtime.error.value) close()
}

defineExpose({ open, close })
</script>

<template>
  <Modal ref="modalRef" title="删除来源配置">
    <template #icon>
      <IconTrash :size="22" />
    </template>

    <div class="space-y-2 text-sm">
      <div class="opacity-70">确认删除这个来源配置？</div>
      <div class="truncate rounded-lg bg-base-200 px-3 py-2 font-medium">
        {{ source?.display_name || source?.source_id }}
      </div>
    </div>

    <template #actions>
      <button
        aria-label="取消"
        class="btn btn-ghost btn-sm btn-square"
        title="取消"
        type="button"
        @click="close"
      >
        <IconX :size="16" />
      </button>
      <button
        aria-label="删除"
        class="btn btn-error btn-sm btn-square"
        :disabled="runtime.saving.value || !source"
        title="删除"
        type="button"
        @click="remove"
      >
        <IconTrash :size="16" />
      </button>
    </template>
  </Modal>
</template>
