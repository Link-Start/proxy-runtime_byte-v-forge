<script setup lang="ts">
import type { ProxyRuntimeNativeConfigState } from '~/composables/useProxyRuntimeNativeConfig'
import { IconDeviceFloppy, IconLink, IconX } from '@tabler/icons-vue'

const props = defineProps<{ runtime: ProxyRuntimeNativeConfigState }>()
const modalRef = ref<{ open: () => void; close: () => void }>()

function open() {
  modalRef.value?.open()
}

function close() {
  modalRef.value?.close()
}

async function save() {
  await props.runtime.saveSubscription()
  if (!props.runtime.error.value) close()
}

defineExpose({ open, close })
</script>

<template>
  <Modal ref="modalRef" title="订阅">
    <template #icon>
      <IconLink :size="22" />
    </template>

    <form class="grid grid-cols-1 gap-3" @submit.prevent="save">
      <input
        v-model.trim="runtime.subscriptionForm.name"
        class="input-bordered input w-full"
        placeholder="名称"
        required
        type="text"
      />
      <input
        v-model.trim="runtime.subscriptionForm.url"
        class="input-bordered input w-full"
        placeholder="订阅 URL"
        required
        type="url"
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
          !runtime.subscriptionForm.name.trim() ||
          !runtime.subscriptionForm.url.trim()
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
