<script setup lang="ts">
import type { ProxyRuntimeNativeConfigState } from '~/composables/useProxyRuntimeNativeConfig'
import { IconDeviceFloppy, IconServer, IconX } from '@tabler/icons-vue'

const props = defineProps<{ runtime: ProxyRuntimeNativeConfigState }>()
const modalRef = ref<{ open: () => void; close: () => void }>()

function open() {
  modalRef.value?.open()
}

function close() {
  modalRef.value?.close()
}

async function save() {
  await props.runtime.saveFixedProxy()
  if (!props.runtime.error.value) close()
}

defineExpose({ open, close })
</script>

<template>
  <Modal ref="modalRef" title="固定代理">
    <template #icon>
      <IconServer :size="22" />
    </template>

    <form class="grid grid-cols-1 gap-3" @submit.prevent="save">
      <input
        v-model.trim="runtime.fixedForm.name"
        class="input-bordered input w-full"
        placeholder="名称"
        required
        type="text"
      />
      <textarea
        v-model.trim="runtime.fixedForm.uri"
        class="textarea-bordered textarea min-h-28 w-full resize-y"
        placeholder="vless://..."
        required
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
          runtime.saving.value ||
          !runtime.fixedForm.name.trim() ||
          !runtime.fixedForm.uri.trim()
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
