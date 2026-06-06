<script setup lang="ts">
import type { ProxyRuntimeDynamicIPProvidersState } from '~/composables/useProxyRuntimeDynamicIPProviders'
import { IconDeviceFloppy, IconServer, IconX } from '@tabler/icons-vue'

const props = defineProps<{ runtime: ProxyRuntimeDynamicIPProvidersState }>()

const modalRef = ref<{ open: () => void; close: () => void }>()
const editing = computed(() => !!props.runtime.providerForm.dynamic_provider_id)

function open() {
  modalRef.value?.open()
}

function close() {
  modalRef.value?.close()
}

async function save() {
  await props.runtime.saveProvider()
  if (!props.runtime.error.value) close()
}

defineExpose({ open, close })
</script>

<template>
  <Modal ref="modalRef" title="动态代理提供商">
    <template #icon>
      <IconServer :size="22" />
    </template>

    <form class="grid grid-cols-1 gap-3" @submit.prevent="save">
      <select
        v-model="runtime.providerForm.provider_id"
        class="select-bordered select w-full"
        :disabled="editing"
        required
      >
        <option
          v-for="provider in runtime.providerOptions.value"
          :key="provider.id"
          :value="provider.id"
        >
          {{ provider.name }}
        </option>
      </select>
      <input
        v-model.trim="runtime.providerForm.display_name"
        class="input-bordered input w-full"
        placeholder="显示名"
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
        :disabled="runtime.saving.value || !runtime.providerForm.provider_id"
        title="保存"
        type="button"
        @click="save"
      >
        <IconDeviceFloppy :size="16" />
      </button>
    </template>
  </Modal>
</template>
