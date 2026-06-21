<script setup lang="ts">
import type { ProxyGatewayPluginsState } from '~/composables/useProxyGatewayPlugins'
import { IconPlus, IconShieldCheck, IconX } from '@tabler/icons-vue'

const props = defineProps<{ runtime: ProxyGatewayPluginsState }>()
const modalRef = ref<{ open: () => void; close: () => void }>()
const providerID = ref('')

function open() {
  providerID.value = props.runtime.availableFraudProviderOptions.value[0]?.provider_id || ''
  modalRef.value?.open()
}

function close() {
  modalRef.value?.close()
}

function add() {
  if (!providerID.value) return
  props.runtime.addFraudProvider(providerID.value)
  close()
}

defineExpose({ open, close })
</script>

<template>
  <Modal ref="modalRef" title="添加 IP Fraud Provider">
    <template #icon>
      <IconShieldCheck :size="22" />
    </template>

    <form class="grid grid-cols-1 gap-3" @submit.prevent="add">
      <label class="grid gap-1">
        <span class="text-xs opacity-60">Fraud 插件类型</span>
        <select
          v-model="providerID"
          class="select-bordered select w-full"
          required
        >
          <option
            v-for="descriptor in runtime.availableFraudProviderOptions.value"
            :key="descriptor.provider_id"
            :value="descriptor.provider_id"
          >
            {{ descriptor.display_name || descriptor.provider_id }}
          </option>
        </select>
      </label>
      <div
        v-if="runtime.availableFraudProviderOptions.value.length === 0"
        class="py-4 text-center text-sm opacity-60"
      >
        暂无 Provider 类型
      </div>
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
        aria-label="添加"
        class="btn btn-primary btn-sm btn-square"
        :disabled="!providerID"
        title="添加"
        type="button"
        @click="add"
      >
        <IconPlus :size="16" />
      </button>
    </template>
  </Modal>
</template>
