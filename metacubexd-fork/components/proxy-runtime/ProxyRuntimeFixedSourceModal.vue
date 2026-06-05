<script setup lang="ts">
import type { ProxyRuntimeSourcesState } from '~/composables/useProxyRuntimeSources'
import { IconDeviceFloppy, IconPlug, IconX } from '@tabler/icons-vue'

const props = defineProps<{ runtime: ProxyRuntimeSourcesState }>()

const modalRef = ref<{ open: () => void; close: () => void }>()

function open() {
  modalRef.value?.open()
}

function close() {
  modalRef.value?.close()
}

async function save() {
  await props.runtime.saveFixed()
  if (!props.runtime.error.value) close()
}

defineExpose({ open, close })
</script>

<template>
  <Modal ref="modalRef" title="固定代理">
    <template #icon>
      <IconPlug :size="22" />
    </template>

    <form class="grid grid-cols-1 gap-3 sm:grid-cols-2" @submit.prevent="save">
      <input
        v-model.trim="runtime.fixedForm.display_name"
        class="input-bordered input w-full"
        placeholder="显示名"
        type="text"
      />
      <label class="flex h-12 items-center gap-2 text-sm">
        <input
          v-model="runtime.fixedForm.enabled"
          class="toggle toggle-primary"
          type="checkbox"
        />
        启用
      </label>
      <input
        v-model.trim="runtime.fixedForm.uri"
        class="input-bordered input w-full sm:col-span-2"
        placeholder="固定代理 URI"
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
        :disabled="runtime.saving.value || !runtime.fixedForm.uri.trim()"
        title="保存"
        type="button"
        @click="save"
      >
        <IconDeviceFloppy :size="16" />
      </button>
    </template>
  </Modal>
</template>
