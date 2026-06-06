<script setup lang="ts">
import type { ProxyRuntimeInUserRulesState } from '~/composables/useProxyRuntimeInUserRules'
import {
  EgressProfileExitKind,
  EgressProfileLineKind,
} from '~/types/byte/v/forge/contracts/proxyruntime/v1/proxy_runtime'
import { IconDeviceFloppy, IconKey, IconX } from '@tabler/icons-vue'

const props = defineProps<{ runtime: ProxyRuntimeInUserRulesState }>()
const modalRef = ref<{ open: () => void; close: () => void }>()

const canSave = computed(() => {
  const form = props.runtime.form
  if (!form.username.trim()) return false
  if (
    form.line_kind === EgressProfileLineKind.EGRESS_PROFILE_LINE_KIND_MIHOMO_NODE &&
    (!form.line_resource_id || !form.line_node_id)
  ) {
    return false
  }
  if (
    form.exit_kind === EgressProfileExitKind.EGRESS_PROFILE_EXIT_KIND_STATIC_IP &&
    (!form.exit_resource_id || !form.exit_node_id)
  ) {
    return false
  }
  return true
})

function open() {
  modalRef.value?.open()
}

function close() {
  modalRef.value?.close()
}

async function save() {
  if (!canSave.value) return
  await props.runtime.saveRule()
  if (!props.runtime.error.value) close()
}

defineExpose({ open, close })
</script>

<template>
  <Modal ref="modalRef" title="IN-USER规则">
    <template #icon>
      <IconKey :size="22" />
    </template>

    <form class="grid grid-cols-1 gap-4" @submit.prevent="save">
      <div class="grid grid-cols-1 gap-3 sm:grid-cols-2">
        <input
          v-model.trim="runtime.form.display_name"
          class="input-bordered input w-full"
          placeholder="显示名"
          type="text"
        />
        <label class="flex h-12 items-center gap-2 text-sm">
          <input
            v-model="runtime.form.enabled"
            class="toggle toggle-primary"
            type="checkbox"
          />
          启用
        </label>
      </div>

      <div class="grid grid-cols-1 gap-3 sm:grid-cols-2">
        <input
          v-model.trim="runtime.form.username"
          class="input-bordered input w-full"
          placeholder="IN-USER"
          required
          type="text"
        />
        <input
          v-model="runtime.form.password_value"
          class="input-bordered input w-full"
          placeholder="密码"
          type="text"
        />
      </div>

      <ProxyRuntimeInUserRouteFields :runtime="runtime" />
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
