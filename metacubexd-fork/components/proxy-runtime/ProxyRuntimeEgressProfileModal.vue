<script setup lang="ts">
import type { ProxyRuntimeEgressProfilesState } from '~/composables/useProxyRuntimeEgressProfiles'
import {
  EgressProfileExitKind,
  EgressProfileLineKind,
} from '~/types/byte/v/forge/contracts/proxyruntime/v1/proxy_runtime'
import { IconDeviceFloppy, IconRoute, IconX } from '@tabler/icons-vue'

const props = defineProps<{ runtime: ProxyRuntimeEgressProfilesState }>()
const modalRef = ref<{ open: () => void; close: () => void }>()

const lineKinds = [
  [EgressProfileLineKind.EGRESS_PROFILE_LINE_KIND_DIRECT, '无代理路线'],
  [EgressProfileLineKind.EGRESS_PROFILE_LINE_KIND_SOURCE, '选择节点'],
] as const
const exitKinds = [
  [EgressProfileExitKind.EGRESS_PROFILE_EXIT_KIND_DIRECT, '使用路线出口'],
  [EgressProfileExitKind.EGRESS_PROFILE_EXIT_KIND_STATIC_IP, '静态 IP'],
  [EgressProfileExitKind.EGRESS_PROFILE_EXIT_KIND_DYNAMIC_IP, '动态 IP'],
] as const

const lineUsesSource = computed(
  () =>
    props.runtime.form.line_kind ===
    EgressProfileLineKind.EGRESS_PROFILE_LINE_KIND_SOURCE,
)
const exitUsesSource = computed(
  () =>
    props.runtime.form.exit_kind ===
    EgressProfileExitKind.EGRESS_PROFILE_EXIT_KIND_STATIC_IP,
)
const exitUsesDynamicProvider = computed(
  () =>
    props.runtime.form.exit_kind ===
    EgressProfileExitKind.EGRESS_PROFILE_EXIT_KIND_DYNAMIC_IP,
)
const canSave = computed(() => {
  const form = props.runtime.form
  if (!form.display_name.trim()) return false
  if (lineUsesSource.value && (!form.line_source_id || !form.line_node_id)) {
    return false
  }
  if (exitUsesSource.value && (!form.exit_source_id || !form.exit_node_id)) {
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
  await props.runtime.saveProfile()
  if (!props.runtime.error.value) close()
}

defineExpose({ open, close })
</script>

<template>
  <Modal ref="modalRef" title="出口 Profile">
    <template #icon>
      <IconRoute :size="22" />
    </template>

    <form class="grid grid-cols-1 gap-4" @submit.prevent="save">
      <div class="grid grid-cols-1 gap-3 sm:grid-cols-2">
        <input
          v-model.trim="runtime.form.display_name"
          class="input-bordered input w-full"
          placeholder="显示名"
          required
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

      <div class="rounded-lg border border-base-content/10 p-3">
        <div class="mb-2 text-sm font-medium">代理路线</div>
        <div class="grid gap-2 sm:grid-cols-2">
          <select v-model="runtime.form.line_kind" class="select-bordered select w-full">
            <option v-for="[value, label] in lineKinds" :key="value" :value="value">
              {{ label }}
            </option>
          </select>
          <ProxyRuntimeSourceSelect
            v-if="lineUsesSource"
            v-model="runtime.form.line_source_id"
            placeholder="路线来源"
            :sources="runtime.lineSources.value"
            @update:model-value="runtime.form.line_node_id = ''"
          />
          <ProxyRuntimeSourceNodeSelect
            v-if="lineUsesSource"
            v-model="runtime.form.line_node_id"
            placeholder="路线节点"
            :runtime="runtime"
            :source-id="runtime.form.line_source_id"
          />
        </div>
      </div>

      <div class="rounded-lg border border-base-content/10 p-3">
        <div class="mb-2 text-sm font-medium">出口</div>
        <div class="grid gap-2 sm:grid-cols-2">
          <select v-model="runtime.form.exit_kind" class="select-bordered select w-full">
            <option v-for="[value, label] in exitKinds" :key="value" :value="value">
              {{ label }}
            </option>
          </select>
          <ProxyRuntimeSourceSelect
            v-if="exitUsesSource"
            v-model="runtime.form.exit_source_id"
            placeholder="静态出口来源"
            :sources="runtime.staticExitSources.value"
            @update:model-value="runtime.form.exit_node_id = ''"
          />
          <ProxyRuntimeSourceNodeSelect
            v-if="exitUsesSource"
            v-model="runtime.form.exit_node_id"
            placeholder="静态出口节点"
            :runtime="runtime"
            :source-id="runtime.form.exit_source_id"
          />
          <ProxyRuntimeDynamicProviderSelect
            v-if="exitUsesDynamicProvider"
            v-model="runtime.form.exit_dynamic_provider_id"
            :providers="runtime.dynamicProviderOptions.value"
          />
        </div>
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
