<script setup lang="ts">
import type { ProxyRuntimeDynamicIPProvidersState } from '~/composables/useProxyRuntimeDynamicIPProviders'
import { IconCloud, IconDeviceFloppy, IconX } from '@tabler/icons-vue'

const props = defineProps<{ runtime: ProxyRuntimeDynamicIPProvidersState }>()

const modalRef = ref<{ open: () => void; close: () => void }>()

function open() {
  modalRef.value?.open()
}

function close() {
  modalRef.value?.close()
}

async function save() {
  await props.runtime.saveEndpoint()
  if (!props.runtime.error.value) close()
}

defineExpose({ open, close })
</script>

<template>
  <Modal ref="modalRef" title="提供商端点">
    <template #icon>
      <IconCloud :size="22" />
    </template>

    <form class="grid grid-cols-1 gap-3" @submit.prevent="save">
      <input
        v-model.trim="runtime.endpointForm.endpoint_url"
        class="input-bordered input w-full"
        placeholder="端点 URL"
        required
        type="text"
      />
    </form>

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
        aria-label="保存"
        class="btn btn-primary btn-sm btn-square"
        :disabled="
          runtime.saving.value || !runtime.endpointForm.endpoint_url.trim()
        "
        title="保存"
        type="button"
        @click="save"
      >
        <IconDeviceFloppy :size="16" />
      </button>
    </template>
  </Modal>
</template>
