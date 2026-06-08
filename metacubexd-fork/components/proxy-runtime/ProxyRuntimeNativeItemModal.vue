<script setup lang="ts">
import type { ProxyRuntimeMihomoNativeState } from '~/composables/useProxyRuntimeMihomoNativeConfig'
import { IconDeviceFloppy, IconServer, IconX } from '@tabler/icons-vue'

const props = defineProps<{ runtime: ProxyRuntimeMihomoNativeState }>()
const modalRef = ref<{ open: () => void; close: () => void }>()
const typeOptions = [
  ['fixed_proxy', '固定代理'],
  ['subscription', '订阅'],
] as const
const currentTypeLabel = computed(
  () => typeOptions.find(([type]) => type === props.runtime.form.type)?.[1] || '类型',
)
const valuePlaceholder = computed(() =>
  props.runtime.form.type === 'fixed_proxy' ? 'vless://...' : '订阅 URL',
)
const valueInputType = computed(() =>
  props.runtime.form.type === 'fixed_proxy' ? 'text' : 'url',
)
const canSave = computed(
  () => !!props.runtime.form.name.trim() && !!props.runtime.form.value.trim(),
)

function open() {
  modalRef.value?.open()
}

function close() {
  modalRef.value?.close()
}

async function save() {
  if (!canSave.value) return
  await props.runtime.saveRow()
  if (!props.runtime.error.value) close()
}

defineExpose({ open, close })
</script>

<template>
  <Modal ref="modalRef" title="原生配置">
    <template #icon>
      <IconServer :size="22" />
    </template>

    <form class="grid grid-cols-1 gap-3" @submit.prevent="save">
      <select
        v-if="!runtime.form.editing"
        v-model="runtime.form.type"
        class="select-bordered select w-full"
      >
        <option v-for="[value, label] in typeOptions" :key="value" :value="value">
          {{ label }}
        </option>
      </select>
      <div v-else class="flex items-center">
        <span class="badge badge-primary">{{ currentTypeLabel }}</span>
      </div>
      <input
        v-model.trim="runtime.form.name"
        class="input-bordered input w-full"
        placeholder="名称"
        required
        type="text"
      />
      <input
        v-model.trim="runtime.form.value"
        class="input-bordered input w-full"
        :placeholder="valuePlaceholder"
        required
        :type="valueInputType"
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
        :disabled="runtime.saving.value || !canSave"
        title="保存"
        type="button"
        @click="save"
      >
        <IconDeviceFloppy :size="16" />
      </button>
    </template>
  </Modal>
</template>
